package repository

import "strings"

func (r *Repository) GetOrders() ([]Order, error) {
	var orders []Order
	tx := r.db.Order("id").Find(&orders)
	return orders, tx.Error
}

func (r *Repository) GetOrder(id int) (Order, error) {
	var o Order
	tx := r.db.First(&o, id)
	return o, tx.Error
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	if strings.TrimSpace(title) == "" {
		return r.GetOrders()
	}
	var orders []Order
	tx := r.db.Where("title ILIKE ?", "%"+title+"%").Order("id").Find(&orders)
	return orders, tx.Error
}
