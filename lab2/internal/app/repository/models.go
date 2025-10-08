package repository

type Application struct {
	ID        uint   `gorm:"primaryKey"`
	Status    string `gorm:"type:varchar(20);not null;index"`
	CreatorID uint   `gorm:"not null;index"`
}

func (Application) TableName() string { return "records" }

type ApplicationOrder struct {
	ID            uint `gorm:"primaryKey"`
	ApplicationID uint `gorm:"not null;uniqueIndex:idx_app_order"`
	OrderID       uint `gorm:"not null;uniqueIndex:idx_app_order"`
}

func (ApplicationOrder) TableName() string { return "records_stages" }
