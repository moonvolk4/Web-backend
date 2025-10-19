package repository

import (
	"errors"
)

const (
	statusDraft   = "черновик"
	statusDeleted = "удалён"
)

func (r *Repository) GetOrCreateDraft(creatorID int) (*Application, error) {
	var app Application
	tx := r.db.Where("creator_id = ? AND status = ?", creatorID, statusDraft).Limit(1).Find(&app)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected > 0 {
		return &app, nil
	}

	app = Application{
		Status:    statusDraft,
		CreatorID: uint(creatorID),
	}
	if err := r.db.Create(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) CreateNewDraft(creatorID int) (*Application, error) {
	app := Application{Status: statusDraft, CreatorID: uint(creatorID)}
	if err := r.db.Create(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) AddOrderToDraft(creatorID, orderID int) error {
	app, err := r.GetOrCreateDraft(creatorID)
	if err != nil {
		return err
	}

	ao := ApplicationOrder{ApplicationID: app.ID, OrderID: uint(orderID)}
	if err := r.db.Where("application_id = ? AND order_id = ?", app.ID, orderID).FirstOrCreate(&ao).Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetDraftOrders(creatorID int) ([]Order, uint, error) {
	var app Application
	if tx := r.db.Where("creator_id = ? AND status = ?", creatorID, statusDraft).Limit(1).Find(&app); tx.Error != nil {
		return nil, 0, tx.Error
	} else if tx.RowsAffected == 0 {
		return []Order{}, 0, nil
	}

	var orders []Order
	q := r.db.Table("records_stages ao").
		Select("o.id, o.title, o.pressure, o.risk_name, o.risk_class, o.code, o.icon, o.image_key, o.description, ao.quantity, ao.doctor_comment").
		Joins("JOIN stages o ON o.id = ao.order_id").
		Where("ao.application_id = ?", app.ID).
		Order("o.id")
	if err := q.Find(&orders).Error; err != nil {
		return nil, app.ID, err
	}
	return orders, app.ID, nil
}

func (r *Repository) GetDraftCount(creatorID int) (int64, error) {
	var app Application
	if tx := r.db.Where("creator_id = ? AND status = ?", creatorID, statusDraft).Limit(1).Find(&app); tx.Error != nil {
		return 0, tx.Error
	} else if tx.RowsAffected == 0 {
		return 0, nil
	}
	var cnt int64
	if err := r.db.Model(&ApplicationOrder{}).Where("application_id = ?", app.ID).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

// DeleteDraftSQL marks user's draft as deleted using raw SQL (no ORM update).
func (r *Repository) DeleteDraftSQL(creatorID int) error {
	// only affect current draft
	// Use only status column to be compatible with minimal schema
	tx := r.db.Exec("UPDATE records SET status = ? WHERE creator_id = ? AND status = ?", statusDeleted, creatorID, statusDraft)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("draft not found")
	}
	return nil
}

// ---- ID-based requests API ----

// GetRequest returns request by id.
func (r *Repository) GetRequest(id int) (*Application, error) {
	var app Application
	if err := r.db.First(&app, id).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// GetRequestOrders returns orders for a specific request id.
func (r *Repository) GetRequestOrders(requestID int) ([]Order, error) {
	var orders []Order
	q := r.db.Table("records_stages ao").
		Select("o.id, o.title, o.pressure, o.risk_name, o.risk_class, o.code, o.icon, o.image_key, o.description, ao.quantity, ao.doctor_comment").
		Joins("JOIN stages o ON o.id = ao.order_id").
		Where("ao.application_id = ?", requestID).
		Order("o.id")
	if err := q.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// AddOrderToRequest links order to a specific request id.
func (r *Repository) AddOrderToRequest(requestID, orderID int) error {
	ao := ApplicationOrder{ApplicationID: uint(requestID), OrderID: uint(orderID)}
	return r.db.Where("application_id = ? AND order_id = ?", requestID, orderID).FirstOrCreate(&ao).Error
}

// DeleteRequest marks the request as deleted by id using raw SQL.
func (r *Repository) DeleteRequest(id int) error {
	tx := r.db.Exec("UPDATE records SET status = ? WHERE id = ? AND status <> ?", statusDeleted, id, statusDeleted)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("request not found or already deleted")
	}
	return nil
}

// GetCurrentDraft returns current draft for creator, if any.
func (r *Repository) GetCurrentDraft(creatorID int) (*Application, error) {
	var app Application
	if tx := r.db.Where("creator_id = ? AND status = ?", creatorID, statusDraft).Limit(1).Find(&app); tx.Error != nil {
		return nil, tx.Error
	} else if tx.RowsAffected == 0 {
		return nil, errors.New("draft not found")
	}
	return &app, nil
}
