package handler

import (
	"lab1/internal/app/repository"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func init() {
	reqMu.Lock()
	defer reqMu.Unlock()
	if len(requestStore) == 0 {
		st1, rn1, rc1 := classifyPressure("120/80")
		requestStore[1] = &AppEntry{ID: 1, Pressure: "120/80", Stage: st1, RiskName: rn1, RiskClass: rc1}

		st2, rn2, rc2 := classifyPressure("135/88")
		requestStore[3] = &AppEntry{ID: 3, Pressure: "135/88", Stage: st2, RiskName: rn2, RiskClass: rc2}

		st3, rn3, rc3 := classifyPressure("145/95")
		requestStore[4] = &AppEntry{ID: 4, Pressure: "145/95", Stage: st3, RiskName: rn3, RiskClass: rc3}
	}
}

func (h *Handler) GetOrders(ctx *gin.Context) {
	var orders []repository.Order
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		orders, err = h.Repository.GetOrders()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		orders, err = h.Repository.GetOrdersByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	// количество заявок берём из in-memory словаря заявок
	reqMu.RLock()
	appCount := len(requestStore)
	reqMu.RUnlock()

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":     time.Now().Format("15:04:05"),
		"orders":   orders,
		"query":    searchQuery, // передаем введенный запрос обратно на страницу
		"appCount": appCount,
		// в ином случае оно будет очищаться при нажатии на кнопку
	})
}

func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	order, err := h.Repository.GetOrder(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"order": order,
	})
}

func (h *Handler) GetApplication(ctx *gin.Context) {

	type localEntry struct {
		id        int
		pressure  string
		stage     string
		riskName  string
		riskClass string
	}
	le := []localEntry{}
	reqMu.RLock()
	for id, e := range requestStore {
		le = append(le, localEntry{
			id:        id,
			pressure:  e.Pressure,
			stage:     e.Stage,
			riskName:  e.RiskName,
			riskClass: e.RiskClass,
		})
	}
	appCount := len(requestStore)
	reqMu.RUnlock()

	orders, err := h.Repository.GetOrders()
	if err != nil {
		logrus.Error(err)
	}
	orderMap := map[int]repository.Order{}
	for _, o := range orders {
		orderMap[o.ID] = o
	}

	entries := []map[string]string{}
	for _, e := range le {
		img := ""
		title := ""
		icon := ""
		if ord, ok := orderMap[e.id]; ok {
			img = ord.ImageKey
			title = ord.Title
			icon = ord.Icon
		}
		entries = append(entries, map[string]string{
			"id":        strconv.Itoa(e.id),
			"pressure":  e.pressure,
			"stage":     e.stage,
			"riskName":  e.riskName,
			"riskClass": e.riskClass,
			"imageKey":  img,
			"title":     title,
			"icon":      icon,
		})
	}
	ctx.HTML(http.StatusOK, "application.html", gin.H{
		"time":         time.Now().Format("15:04:05"),
		"entries":      entries,
		"appCount":     appCount,
		"appCountText": formatAppCount(appCount),
	})
}

func classifyPressure(raw string) (stageTitle, riskName, riskClass string) {

	re := regexp.MustCompile(`(?s)(\d+)\D+(\d+)`)
	m := re.FindStringSubmatch(raw)
	if len(m) < 3 {
		return "", "", ""
	}
	sys, _ := strconv.Atoi(m[1])
	dia, _ := strconv.Atoi(m[2])

	if sys >= 180 && dia < 90 {
		return "ИСАГ", "Очень высокий", "risk-vhigh"
	}
	// АГ 3-й стадии
	if sys >= 180 || dia >= 110 {
		return "АГ 3-й стадии", "Очень высокий", "risk-vhigh"
	}
	// АГ 2-й стадии
	if (sys >= 160 && sys <= 179) || (dia >= 100 && dia <= 109) {
		return "АГ 2-й стадии", "Высокий", "risk-high"
	}
	// АГ 1-й стадии
	if (sys >= 140 && sys <= 159) || (dia >= 90 && dia <= 99) {
		return "АГ 1-й стадии", "Умеренный", "risk-mod"
	}
	// Высокое нормальное
	if (sys >= 130 && sys <= 139) || (dia >= 85 && dia <= 89) {
		return "Высокое", "Незначительный", "risk-mid"
	}
	// Нормальное
	if (sys >= 120 && sys <= 129) || (dia >= 80 && dia <= 84) {
		return "Нормальное", "Минимальный", "risk-low"
	}
	// Оптимальное
	if sys < 120 && dia < 80 {
		return "Оптимальное", "Нулевой", "risk-zero"
	}
	return "", "", ""
}

// formatAppCount возвращает строку вида "0 заявок", "1 заявка", "2 заявки", "5 заявок"
func formatAppCount(n int) string {
	rem10 := n % 10
	rem100 := n % 100
	word := "заявок"
	if rem10 == 1 && rem100 != 11 {
		word = "заявка"
	} else if rem10 >= 2 && rem10 <= 4 && (rem100 < 12 || rem100 > 14) {
		word = "заявки"
	}
	return strconv.Itoa(n) + " " + word
}

// ===== In-memory заявка =====
type AppEntry struct {
	ID        int
	Patient   string
	Pressure  string
	Stage     string
	RiskName  string
	RiskClass string
}

var (
	requestStore = map[int]*AppEntry{}
	reqMu        sync.RWMutex
)
