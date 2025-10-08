package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Repository инкапсулирует доступ к данным. Теперь он работает с PostgreSQL через GORM.
type Repository struct{ db *gorm.DB }

// NewRepository инициализирует подключение к БД Postgres по dsn и возвращает репозиторий.
func NewRepository(dsn string) (*Repository, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("empty DSN for repository")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// Верифицируем соединение
	if sqlDB, err := db.DB(); err == nil {
		if pingErr := sqlDB.Ping(); pingErr != nil {
			return nil, fmt.Errorf("db ping: %w", pingErr)
		}
	}
	return &Repository{db: db}, nil
}

// Order описывает «услугу/стадию». Теги соответствуют колонкам в БД.
type Order struct {
	ID          int    `gorm:"column:id;primaryKey"`
	Title       string `gorm:"column:title"`
	Pressure    string `gorm:"column:pressure"`
	RiskName    string `gorm:"column:risk_name"`
	RiskClass   string `gorm:"column:risk_class"`
	Code        string `gorm:"column:code"`
	Icon        string `gorm:"column:icon"`
	ImageKey    string `gorm:"column:image_key"`
	Description string `gorm:"column:description"`
}

func (Order) TableName() string { return "stages" }
