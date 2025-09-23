package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	// Ищем JSON независимо от каталога запуска
	wd, _ := os.Getwd()
	// Путь от исходника репозитория (этот файл находится в internal/app/repository)
	_, srcFile, _, _ := runtime.Caller(0)
	srcBase := filepath.Dir(filepath.Dir(filepath.Dir(srcFile))) // подняться до корня проекта

	candidates := []string{
		filepath.Join(wd, "resources", "data", "orders.json"),
		filepath.Join(wd, "..", "resources", "data", "orders.json"),
		filepath.Join(wd, "..", "..", "resources", "data", "orders.json"),
		filepath.Join(srcBase, "resources", "data", "orders.json"),
	}

	var dataPath string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			dataPath = p
			break
		}
	}
	if dataPath == "" {
		return nil, fmt.Errorf("не найден файл данных orders.json по путям: %v", candidates)
	}

	f, err := os.Open(dataPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть JSON с заказами: %w", err)
	}
	defer f.Close()

	var orders []Order
	if err := json.NewDecoder(f).Decode(&orders); err != nil {
		return nil, fmt.Errorf("не удалось распарсить JSON: %w", err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
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
