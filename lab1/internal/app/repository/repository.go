package repository

import (
	"fmt"
	"strings"
)

type Repository struct{}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Order struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Pressure    string `json:"pressure"`
	RiskName    string `json:"riskName"`
	RiskClass   string `json:"riskClass"`
	Code        string `json:"code"`
	Icon        string `json:"icon"`
	ImageKey    string `json:"imageKey"`
	Description string `json:"description"`
}

func (r *Repository) GetOrders() ([]Order, error) {

	// ЛР1: данные берём из in-memory коллекции, без JSON/БД
	orders := []Order{
		{ID: 1, Title: "Оптимальное", Pressure: "120 / 80", RiskName: "Нулевой", RiskClass: "risk-zero", Code: "I10-1", Icon: "🧍", ImageKey: "stages/1.jpg", Description: "Оптимальные значения артериального давления соответствуют хорошему состоянию сердечно‑сосудистой системы. Рекомендуется поддерживать активный образ жизни, сбалансированное питание и контроль факторов риска."},
		{ID: 2, Title: "Нормальное", Pressure: "120 - 129 / 80 - 84", RiskName: "Минимальный", RiskClass: "risk-low", Code: "I10-1", Icon: "➕", ImageKey: "stages/2.jpg", Description: "Нормальные значения АД. Важно сохранять здоровые привычки: достаточная физическая активность, ограничение соли, контроль массы тела и стресса."},
		{ID: 3, Title: "Высокое", Pressure: "130 - 139 / 85 - 89", RiskName: "Незначительный", RiskClass: "risk-mid", Code: "I10-1", Icon: "❤️", ImageKey: "stages/3.jpg", Description: "Погранично высокие значения АД. Рекомендуется более строгий контроль образа жизни и наблюдение. Возможен переход к гипертензии без коррекции факторов риска."},
		{ID: 4, Title: "АГ 1-й стадии", Pressure: "140 - 159 / 90 - 99", RiskName: "Умеренный", RiskClass: "risk-mod", Code: "I10-1", Icon: "🩺", ImageKey: "stages/4.jpg", Description: "Артериальная гипертензия 1‑й стадии. Как правило, поражения органов‑мишеней отсутствуют. Требуются меры по изменению образа жизни, возможен медикаментозный контроль согласно рекомендациям врача."},
		{ID: 5, Title: "АГ 2-й стадии", Pressure: "160 - 179 / 100 - 109", RiskName: "Высокий", RiskClass: "risk-high", Code: "I10-1", Icon: "🫀", ImageKey: "stages/5.jpg", Description: "Артериальная гипертензия 2‑й стадии сопровождается более выраженным повышением АД и ростом сердечно‑сосудистых рисков. Необходима медикаментозная терапия и регулярный мониторинг."},
		{ID: 6, Title: "АГ 3-й стадии", Pressure: "> 180 / > 110", RiskName: "Очень высокий", RiskClass: "risk-vhigh", Code: "I10-1", Icon: "⚠️", ImageKey: "stages/6.jpg", Description: "Тяжёлая гипертензия с очень высоким риском осложнений. Требуется интенсивная терапия и наблюдение специалиста. Высока вероятность поражения органов‑мишеней."},
		{ID: 7, Title: "ИСАГ", Pressure: "> 180 / < 90", RiskName: "Очень высокий", RiskClass: "risk-vhigh", Code: "I10-1", Icon: "🫀", ImageKey: "stages/7.jpg", Description: "Изолированная систолическая артериальная гипертензия — значительно повышено систолическое давление при нормальном или низком диастолическом. Часто встречается у пожилых, требует подбора терапии."},
	}
	return orders, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Order{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return []Order{}, err
	}

	var result []Order
	for _, order := range orders {
		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
			result = append(result, order)
		}
	}

	return result, nil
}
