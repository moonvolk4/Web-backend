package repository

import "time"

type Application struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Status    string    `gorm:"type:varchar(20);not null;index" json:"status"`
	CreatorID uint      `gorm:"not null;index" json:"creator_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	// System fields
	FormedAt    *time.Time `gorm:"column:formed_at" json:"formed_at,omitempty"`
	FinishedAt  *time.Time `gorm:"column:finished_at" json:"finished_at,omitempty"`
	ModeratorID *uint      `gorm:"column:moderator_id" json:"moderator_id,omitempty"`
	// Result fields
	ResultStage string  `gorm:"column:result_stage" json:"result_stage"`
	ResultMap   float64 `gorm:"column:result_map" json:"result_map"`
	// Flags
	FlagLvHypertrophy bool `gorm:"column:flag_lv_hypertrophy;not null;default:false" json:"flag_lv_hypertrophy"`
	FlagRenalDamage   bool `gorm:"column:flag_renal_damage;not null;default:false" json:"flag_renal_damage"`
	FlagArterialStiff bool `gorm:"column:flag_arterial_stiffness;not null;default:false" json:"flag_arterial_stiffness"`
	// Application-level comment
	Comment string `gorm:"column:comment" json:"comment"`
	// Aggregates (computed)
	ItemsCount int `gorm:"-" json:"items_count"`
	// Logins (computed)
	CreatorLogin   string `gorm:"-" json:"creator_login,omitempty"`
	ModeratorLogin string `gorm:"-" json:"moderator_login,omitempty"`
}

func (Application) TableName() string { return "records" }

type ApplicationOrder struct {
	ID            uint   `gorm:"primaryKey"`
	ApplicationID uint   `gorm:"not null;uniqueIndex:idx_app_order"`
	OrderID       uint   `gorm:"not null;uniqueIndex:idx_app_order"`
	Quantity      int    `gorm:"column:quantity;not null;default:1"`
	DoctorComment string `gorm:"column:doctor_comment"`
}

func (ApplicationOrder) TableName() string { return "records_stages" }
