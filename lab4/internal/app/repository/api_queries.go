package repository

import (
	"errors"
	"strings"
	"time"
)

const (
	statusFormed   = "сформирована"
	statusFinished = "завершена"
	statusRejected = "отклонена"
)

// Stages (services)

// CreateStage inserts a new stage row writing only stage columns (no m:n payload fields).
func (r *Repository) CreateStage(o *Order) error {
	// Use a narrow struct mapped to stages to avoid writing Quantity/DoctorComment
	type stageRow struct {
		ID          int    `gorm:"column:id;primaryKey"`
		Title       string `gorm:"column:title"`
		Description string `gorm:"column:description"`
		RiskClass   string `gorm:"column:risk_class;not null;default:''"`
		ImageKey    string `gorm:"column:image_key;not null;default:''"`
		IsDeleted   bool   `gorm:"column:is_deleted;not null;default:false"`
		SysFrom     int    `gorm:"column:sys_from"`
		SysTo       int    `gorm:"column:sys_to"`
		DiaFrom     int    `gorm:"column:dia_from"`
		DiaTo       int    `gorm:"column:dia_to"`
	}
	rc := strings.TrimSpace(o.RiskClass)
	img := strings.TrimSpace(o.ImageKey)
	row := stageRow{
		Title:       o.Title,
		Description: o.Description,
		RiskClass:   rc,
		ImageKey:    img,
		SysFrom:     o.SysFrom,
		SysTo:       o.SysTo,
		DiaFrom:     o.DiaFrom,
		DiaTo:       o.DiaTo,
		IsDeleted:   false,
	}
	if err := r.db.Table("stages").Create(&row).Error; err != nil {
		return err
	}
	// hydrate minimal fields back
	o.ID = row.ID
	o.Title = row.Title
	o.Description = row.Description
	o.RiskClass = row.RiskClass
	o.ImageKey = row.ImageKey
	o.SysFrom = row.SysFrom
	o.SysTo = row.SysTo
	o.DiaFrom = row.DiaFrom
	o.DiaTo = row.DiaTo
	return nil
}

func (r *Repository) UpdateStage(id int, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	tx := r.db.Model(&Order{}).Where("id = ?", id).Updates(fields)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("stage not found")
	}
	return nil
}

func (r *Repository) DeleteStage(id int) error {
	// Soft delete: mark as is_deleted and clear image_key
	tx := r.db.Model(&Order{}).Where("id = ?", id).Updates(map[string]any{"is_deleted": true, "image_key": ""})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("stage not found")
	}
	return nil
}

// Applications

type ApplicationFilter struct {
	Status     string
	FormedFrom string // optional, ISO date string; ignored if column absent
	FormedTo   string // optional
	CreatorID  *int   // if set and not moderator, limit to this creator
}

func (r *Repository) ListApplications(f ApplicationFilter) ([]Application, error) {
	var apps []Application
	q := r.db.Model(&Application{})
	// exclude deleted and drafts by default
	q = q.Where("status <> ? AND status <> ?", statusDeleted, statusDraft)
	if s := strings.TrimSpace(f.Status); s != "" {
		q = q.Where("status = ?", s)
	}
	if f.CreatorID != nil {
		q = q.Where("creator_id = ?", *f.CreatorID)
	}
	// formed_at date range filter
	if ff := strings.TrimSpace(f.FormedFrom); ff != "" {
		q = q.Where("formed_at >= ?", ff)
	}
	if ft := strings.TrimSpace(f.FormedTo); ft != "" {
		q = q.Where("formed_at <= ?", ft)
	}
	q = q.Order("id DESC")
	if err := q.Find(&apps).Error; err != nil {
		return nil, err
	}
	// aggregate items_count for returned applications
	if len(apps) > 0 {
		ids := make([]uint, 0, len(apps))
		for _, a := range apps {
			ids = append(ids, a.ID)
		}
		type row struct {
			ApplicationID uint
			Cnt           int
		}
		var rows []row
		if err := r.db.Table("records_stages").
			Select("application_id, COUNT(*) as cnt").
			Where("application_id IN ?", ids).
			Group("application_id").
			Find(&rows).Error; err == nil {
			m := make(map[uint]int, len(rows))
			for _, r1 := range rows {
				m[r1.ApplicationID] = r1.Cnt
			}
			for i := range apps {
				if v, ok := m[apps[i].ID]; ok {
					apps[i].ItemsCount = v
				}
			}
		}
	}
	return apps, nil
}

func (r *Repository) UpdateApplicationFields(id int, fields map[string]any) error {
	// Protect system fields at handler layer; here just update map
	tx := r.db.Model(&Application{}).Where("id = ?", id).Updates(fields)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("application not found")
	}
	return nil
}

func (r *Repository) SubmitApplication(id int, creatorID int) error {
	// Only draft owned by creator can be submitted
	var app Application
	if err := r.db.First(&app, id).Error; err != nil {
		return err
	}
	if app.Status != statusDraft || int(app.CreatorID) != creatorID {
		return errors.New("cannot submit: wrong status or owner")
	}
	// ensure there is at least one item
	var cnt int64
	if err := r.db.Model(&ApplicationOrder{}).Where("application_id = ?", id).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return errors.New("cannot submit: empty draft")
	}
	return r.db.Model(&Application{}).Where("id = ?", id).Updates(map[string]any{
		"status":    statusFormed,
		"formed_at": time.Now(),
	}).Error
}

func (r *Repository) ResolveApplication(id int, moderatorID int, approve bool) error {
	var app Application
	if err := r.db.First(&app, id).Error; err != nil {
		return err
	}
	if app.Status != statusFormed {
		return errors.New("cannot resolve: not in 'сформирована'")
	}
	newStatus := statusRejected
	if approve {
		newStatus = statusFinished
	}
	updates := map[string]any{
		"status":       newStatus,
		"moderator_id": moderatorID,
		"finished_at":  time.Now(),
	}
	// If approved, compute result_map from linked stages (simple MAP from Pressure, take max)
	if approve {
		if mp, stageTitle, err := r.computeResultForApplication(id); err == nil {
			updates["result_map"] = mp
			if stageTitle != "" {
				updates["result_stage"] = stageTitle
			}
		}
	}
	return r.db.Model(&Application{}).Where("id = ?", id).Updates(updates).Error
}

// computeResultForApplication calculates a simple result (MAP) from linked stages sys/dia (or falls back).
func (r *Repository) computeResultForApplication(appID int) (mapValue float64, stageTitle string, err error) {
	type row struct {
		SysTo int
		DiaTo int
		Title string
	}
	var rows []row
	q := r.db.Table("records_stages ao").
		Select("o.sys_to, o.dia_to, o.title").
		Joins("join stages o on o.id = ao.order_id").
		Where("ao.application_id = ?", appID)
	if err = q.Find(&rows).Error; err != nil {
		return 0, "", err
	}
	best := 0.0
	bestTitle := ""
	for _, r1 := range rows {
		sys := r1.SysTo
		dia := r1.DiaTo
		if sys == 0 || dia == 0 {
			continue
		}
		mp := (float64(2*dia) + float64(sys)) / 3.0
		if mp > best {
			best = mp
			bestTitle = r1.Title
		}
	}
	return best, bestTitle, nil
}

// m-m helpers on records_stages

func (r *Repository) UpdateApplicationItem(appID, serviceID int, fields map[string]any) error {
	tx := r.db.Model(&ApplicationOrder{}).
		Where("application_id = ? AND order_id = ?", appID, serviceID).
		Updates(fields)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("application item not found")
	}
	return nil
}

func (r *Repository) DeleteApplicationItem(appID, serviceID int) error {
	tx := r.db.Where("application_id = ? AND order_id = ?", appID, serviceID).Delete(&ApplicationOrder{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("application item not found")
	}
	return nil
}
