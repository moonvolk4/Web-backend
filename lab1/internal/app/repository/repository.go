package repository

import (
	"fmt"
	"strings"
)

// Repository инкапсулирует доступ к данным. Для ЛР1 данные — статичный срез в памяти.
type Repository struct{}

func NewRepository() (*Repository, error) { return &Repository{}, nil }

// Order описывает «услугу/стадию» гипертонии, отображаемую на всех трёх страницах.
// Поля экспортируются (с заглавной буквы), чтобы шаблоны могли их читать.
type Order struct {
	ID          int
	Title       string
	Pressure    string // репрезентативный диапазон / пример АД для стадии
	RiskName    string // человеко‑читаемое название риска
	RiskClass   string // CSS‑класс для раскраски (risk-low, risk-mid, ...)
	Code        string // условный код (ICD / внутренний)
	Icon        string // запасной emoji, если нет картинки
	ImageKey    string // ключ объекта в MinIO (путь внутри bucket)
	Description string
}

// seedOrders — единый источник данных. Картинки предполагаются по пути stages/N.jpg в бакете.
// Набор покрывает все градации, которые использует classifyPressure.
var seedOrders = []Order{
	{
		ID:          1,
		Title:       "Оптимальное АД",
		Pressure:    "115/75",
		RiskName:    "Нулевой",
		RiskClass:   "risk-zero",
		Code:        "I10.0",
		Icon:        "✅",
		ImageKey:    "stages/1.jpg",
		Description: "Оптимальные значения артериального давления (<120/<80 мм рт. ст.) без признаков сердечно‑сосудистого риска.",
	},
	{
		ID:          2,
		Title:       "Нормальное АД",
		Pressure:    "125/82",
		RiskName:    "Минимальный",
		RiskClass:   "risk-low",
		Code:        "I10.1",
		Icon:        "🟢",
		ImageKey:    "stages/2.jpg",
		Description: "Нормальные значения (120–129 / 80–84). Рекомендуется поддержание образа жизни и наблюдение.",
	},
	{
		ID:          3,
		Title:       "Высокое нормальное",
		Pressure:    "135/88",
		RiskName:    "Незначительный",
		RiskClass:   "risk-mid",
		Code:        "I10.2",
		Icon:        "🟡",
		ImageKey:    "stages/3.jpg",
		Description: "Пограничные значения (130–139 / 85–89). Требует модификации образа жизни и динамического контроля.",
	},
	{
		ID:          4,
		Title:       "АГ 1-й стадии",
		Pressure:    "145/95",
		RiskName:    "Умеренный",
		RiskClass:   "risk-mod",
		Code:        "I10.3",
		Icon:        "⚠️",
		ImageKey:    "stages/4.jpg",
		Description: "Артериальная гипертензия первой стадии (140–159 / 90–99). Начальный уровень повышения давления.",
	},
	{
		ID:          5,
		Title:       "АГ 2-й стадии",
		Pressure:    "165/105",
		RiskName:    "Высокий",
		RiskClass:   "risk-high",
		Code:        "I10.4",
		Icon:        "⚠️",
		ImageKey:    "stages/5.jpg",
		Description: "Артериальная гипертензия второй стадии (160–179 / 100–109). Существенно повышенный риск осложнений.",
	},
	{
		ID:          6,
		Title:       "АГ 3-й стадии",
		Pressure:    "185/115",
		RiskName:    "Очень высокий",
		RiskClass:   "risk-vhigh",
		Code:        "I10.5",
		Icon:        "❗",
		ImageKey:    "stages/6.jpg",
		Description: "Тяжёлая гипертензия (≥180 / ≥110). Требует незамедлительной коррекции терапии и наблюдения.",
	},
	{
		ID:          7,
		Title:       "ИСАГ",
		Pressure:    "180/85",
		RiskName:    "Очень высокий",
		RiskClass:   "risk-vhigh",
		Code:        "I10.6",
		Icon:        "❗",
		ImageKey:    "stages/7.jpg",
		Description: "Изолированная систолическая артериальная гипертензия (≥180 при диастолическом <90). Часто у пожилых пациентов.",
	},
}

func (r *Repository) GetOrders() ([]Order, error) {
	if len(seedOrders) == 0 {
		return nil, fmt.Errorf("dataset empty")
	}
	return seedOrders, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	for _, o := range seedOrders {
		if o.ID == id {
			return o, nil
		}
	}
	return Order{}, fmt.Errorf("order %d not found", id)
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	if title == "" {
		return r.GetOrders()
	}
	low := strings.ToLower(title)
	res := []Order{}
	for _, o := range seedOrders {
		if strings.Contains(strings.ToLower(o.Title), low) {
			res = append(res, o)
		}
	}
	return res, nil
}
