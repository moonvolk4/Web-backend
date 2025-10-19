package repository

type Application struct {
	ID        uint   `gorm:"primaryKey"`
	Status    string `gorm:"type:varchar(20);not null;index"`
	CreatorID uint   `gorm:"not null;index"`
	// Result fields (minimal)
	ResultStageCode string  `gorm:"column:result_stage_code"`
	ResultMap       float64 `gorm:"column:result_map"`
	// Minimal flags
	FlagLvHypertrophy bool `gorm:"column:flag_lv_hypertrophy;not null;default:false"`
	FlagRenalDamage   bool `gorm:"column:flag_renal_damage;not null;default:false"`
	FlagArterialStiff bool `gorm:"column:flag_arterial_stiffness;not null;default:false"`
	// Application-level comment
	Comment string `gorm:"column:comment"`
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
