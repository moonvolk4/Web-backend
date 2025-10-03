package handler

import (
	"lab1/internal/app/repository"
	"net/http"
	"regexp"
	"sort"
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
	return &Handler{Repository: r}
}

func init() {
	reqMu.Lock()
	defer reqMu.Unlock()
	if len(requestStore) == 0 {
		// Предварительные три записи (выбранные услуги в заявке)
		seedPressures := map[int]string{1: "120/80", 3: "135/88", 4: "145/95"}
		for id, p := range seedPressures {
			st, rn, rc := classifyPressure(p)
			ac := autoComplicationsForStage(st)
			requestStore[id] = &AppEntry{
				ID:                id,
				Pressure:          p,
				Stage:             st,
				RiskName:          rn,
				RiskClass:         rc,
				AutoComplications: ac,
			}
		}
	}
}

func (h *Handler) GetOrders(ctx *gin.Context) {
	var orders []repository.Order
	var err error

	// поддерживаем новый параметр stage; для совместимости оставляем query как fallback
	searchQuery := ctx.Query("stage")
	if searchQuery == "" {
		searchQuery = ctx.Query("query")
	}
	logrus.Infof("searchQuery raw: %q", searchQuery)
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

	// фиксируем число услуг в заявке для ЛР1 (отображение: "3 услуги")
	appCount := 3

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
	// Одна заявка (id игнорируем, поддерживаем /request/:id для совместимости)
	_ = ctx.Param("id")

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
		le = append(le, localEntry{id: id, pressure: e.Pressure, stage: e.Stage, riskName: e.RiskName, riskClass: e.RiskClass})
	}
	reqMu.RUnlock()

	// Стабильный порядок вывода
	sort.Slice(le, func(i, j int) bool { return le[i].id < le[j].id })

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
		title, imageKey, icon := "", "", ""
		if ord, ok := orderMap[e.id]; ok {
			title, imageKey, icon = ord.Title, ord.ImageKey, ord.Icon
		}
		entries = append(entries, map[string]string{
			"id":        strconv.Itoa(e.id),
			"title":     title,
			"imageKey":  imageKey,
			"icon":      icon,
			"pressure":  e.pressure,
			"stage":     e.stage,
			"riskName":  e.riskName,
			"riskClass": e.riskClass,
		})
	}

	// GET-расчёт максимального давления (?max=1)
	var maxInfo map[string]string
	if ctx.Query("max") != "" && len(le) > 0 {
		maxIdx := 0
		maxSys, maxDia := parsePressure(le[0].pressure)
		for i := 1; i < len(le); i++ {
			s, d := parsePressure(le[i].pressure)
			if s > maxSys || (s == maxSys && d > maxDia) {
				maxSys, maxDia = s, d
				maxIdx = i
			}
		}
		st, rn, rc := classifyPressure(le[maxIdx].pressure)
		maxInfo = map[string]string{
			"pressure":  le[maxIdx].pressure,
			"stage":     st,
			"riskName":  rn,
			"riskClass": rc,
		}
	}

	// Авто осложнения (из стадий) + возможность отметить дополнительно (manual) через параметр extra
	autoSet := map[string]bool{}
	reqMu.RLock()
	for _, e := range requestStore {
		for _, c := range e.AutoComplications {
			autoSet[c] = true
		}
	}
	reqMu.RUnlock()

	// Полный перечень для отображения: сначала авто, затем заранее заданные потенциальные дополнительные
	// Полный перечень «реальных» осложнений (единый справочник в фиксированном порядке)
	possible := []string{
		"Пограничный сосудистый риск",
		"Начальные сосудистые изменения",
		"Гипертрофия ЛЖ",
		"Микроальбуминурия",
		"ХПН",
		"Сердечная недостаточность",
		"Органные поражения",
		"Повышенная жёсткость артерий",
	}

	// Дополнительные пользовательские (manual) отметки через ?extra=Название
	manualSel := map[string]bool{}
	for _, v := range ctx.QueryArray("extra") {
		manualSel[v] = true
	}
	compList := []map[string]any{}
	for _, name := range possible {
		if name == "" {
			continue
		}
		auto := autoSet[name]
		checked := auto || manualSel[name]
		compList = append(compList, map[string]any{"title": name, "checked": checked, "auto": auto})
	}

	// ФИО пациента через GET (не сохраняем в store по условиям ЛР1)
	patientName := ctx.Query("patient")
	if patientName == "" {
		patientName = "Иванов Иван Иванович"
	}

	// Параметры для потенциального будущего расчёта MAP (пока только передаём bp обратно)
	bpRaw := ctx.Query("bp")

	appCount := len(le)
	ctx.HTML(http.StatusOK, "application.html", gin.H{
		"time":          time.Now().Format("15:04:05"),
		"entries":       entries,
		"appCount":      appCount,
		"appCountText":  formatServiceCount(appCount),
		"maxInfo":       maxInfo,
		"complications": compList,
		"patientName":   patientName,
		"bp":            bpRaw,
	})
}

// factorState используется только для шаблона.
// (убраны m:n факторы и связанные структуры)

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

// parsePressure извлекает числовые значения САД/ДАД из строки "120/80".
func parsePressure(raw string) (sys int, dia int) {
	re := regexp.MustCompile(`(?s)(\d+)\D+(\d+)`)
	m := re.FindStringSubmatch(raw)
	if len(m) < 3 {
		return 0, 0
	}
	sys, _ = strconv.Atoi(m[1])
	dia, _ = strconv.Atoi(m[2])
	return
}

// formatServiceCount возвращает строку вида "0 услуг", "1 услуга", "2 услуги", "5 услуг"
func formatServiceCount(n int) string {
	rem10 := n % 10
	rem100 := n % 100
	word := "услуг"
	if rem10 == 1 && rem100 != 11 {
		word = "услуга"
	} else if rem10 >= 2 && rem10 <= 4 && (rem100 < 12 || rem100 > 14) {
		word = "услуги"
	}
	return strconv.Itoa(n) + " " + word
}

// ===== In-memory заявка (старая модель) =====
type AppEntry struct {
	ID        int
	Patient   string
	Pressure  string
	Stage     string
	RiskName  string
	RiskClass string
	// Автоматически определённые осложнения (галочки) для данной стадии
	AutoComplications []string
}

var (
	requestStore = map[int]*AppEntry{}
	reqMu        sync.RWMutex
)

// autoComplicationsForStage возвращает список типовых осложнений (заглушка) по названию стадии.
func autoComplicationsForStage(stage string) []string {
	switch stage {
	case "Оптимальное":
		return []string{}
	case "Нормальное":
		return []string{}
	case "Высокое":
		return []string{"Пограничный сосудистый риск"}
	case "АГ 1-й стадии":
		return []string{"Начальные сосудистые изменения"}
	case "АГ 2-й стадии":
		return []string{"Гипертрофия ЛЖ", "Микроальбуминурия"}
	case "АГ 3-й стадии":
		return []string{"ХПН", "Сердечная недостаточность", "Органные поражения"}
	case "ИСАГ":
		return []string{"Повышенная жёсткость артерий"}
	default:
		return []string{}
	}
}
