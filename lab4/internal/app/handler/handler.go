package handler

import (
	"fmt"
	"lab2/internal/app/auth"
	"lab2/internal/app/repository"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	JWT        *auth.JWTService
}

func NewHandler(r *repository.Repository, jwt *auth.JWTService) *Handler {
	return &Handler{Repository: r, JWT: jwt}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
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
		orders, err = h.Repository.GetOrdersByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	cnt, _ := h.Repository.GetDraftCount(1)
	appCount := int(cnt)
	var draftID uint
	_, id, _ := h.Repository.GetDraftOrders(1)
	draftID = id

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":     time.Now().Format("15:04:05"),
		"orders":   orders,
		"query":    searchQuery,
		"appCount": appCount,
		"hasDraft": draftID > 0,
		"draftID":  draftID,
	})
}

func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
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
	if idStr := ctx.Param("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			ctx.Redirect(http.StatusFound, "/stages")
			return
		}
		app, err := h.Repository.GetRequest(id)
		if err != nil || app.Status == "удалён" {
			ctx.Redirect(http.StatusFound, "/stages")
			return
		}
		orders, err := h.Repository.GetRequestOrders(id)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		h.renderApplication(ctx, app, orders, int(app.ID))
		return
	}
	creatorID := currentUserIDOrDefault(ctx)
	orders, draftID, err := h.Repository.GetDraftOrders(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	var app *repository.Application
	if draftID != 0 {
		if a, err := h.Repository.GetRequest(int(draftID)); err == nil {
			app = a
		}
	}
	h.renderApplication(ctx, app, orders, int(draftID))
}

// currentUserIDOrDefault tries to read user_id from context (set by JWT middleware),
// and falls back to 1 for compatibility with legacy HTML flows.
func currentUserIDOrDefault(ctx *gin.Context) int {
	if v, ok := ctx.Get("user_id"); ok {
		if id, ok2 := v.(int); ok2 && id > 0 {
			return id
		}
	}
	return 1
}

func (h *Handler) renderApplication(ctx *gin.Context, app *repository.Application, orders []repository.Order, requestID int) {
	entries := []map[string]string{}
	for _, o := range orders {
		entries = append(entries, map[string]string{
			"id":        strconv.Itoa(o.ID),
			"title":     o.Title,
			"imageKey":  o.ImageKey,
			"icon":      o.Icon,
			"pressure":  o.Pressure,
			"stage":     o.Title,
			"riskName":  o.RiskName,
			"riskClass": o.RiskClass,
			"quantity":  strconv.Itoa(o.Quantity),
		})
	}

	var maxInfo map[string]string
	if ctx.Query("max") != "" && len(entries) > 0 {
		maxIdx := 0
		maxSys, maxDia := parsePressure(entries[0]["pressure"])
		for i := 1; i < len(entries); i++ {
			s, d := parsePressure(entries[i]["pressure"])
			if s > maxSys || (s == maxSys && d > maxDia) {
				maxSys, maxDia = s, d
				maxIdx = i
			}
		}
		st, rn, rc := classifyPressure(entries[maxIdx]["pressure"])
		maxInfo = map[string]string{
			"pressure":  entries[maxIdx]["pressure"],
			"stage":     st,
			"riskName":  rn,
			"riskClass": rc,
		}
	}

	autoSet := map[string]bool{}
	reqMu.RLock()
	for _, e := range requestStore {
		for _, c := range e.AutoComplications {
			autoSet[c] = true
		}
	}
	reqMu.RUnlock()

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

	patientName := ctx.Query("patient")
	if patientName == "" {
		patientName = "Иванов Иван Иванович"
	}

	bpRaw := ctx.Query("bp")

	appCount := len(entries)
	// flags from application (if loaded)
	flags := map[string]bool{
		"lv_h":      false,
		"renal":     false,
		"stiffness": false,
	}
	if app != nil {
		flags["lv_h"] = app.FlagLvHypertrophy
		flags["renal"] = app.FlagRenalDamage
		flags["stiffness"] = app.FlagArterialStiff
	}
	ctx.HTML(http.StatusOK, "application.html", gin.H{
		"time":         time.Now().Format("15:04:05"),
		"entries":      entries,
		"appCount":     appCount,
		"appCountText": formatServiceCount(appCount),
		"maxInfo":      maxInfo,
		"patientName":  patientName,
		"bp":           bpRaw,
		"draftID":      requestID,
		"flags":        flags,
		"appComment": func() string {
			if app != nil {
				return app.Comment
			}
			return ""
		}(),
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

// (2*ДАД + САД) / 3
func calcMAP(sys, dia int) float64 {
	return float64(2*dia+sys) / 3.0
}

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

type AppEntry struct {
	ID                int
	Patient           string
	Pressure          string
	Stage             string
	RiskName          string
	RiskClass         string
	AutoComplications []string
}

var (
	requestStore = map[int]*AppEntry{}
	reqMu        sync.RWMutex
)

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

func (h *Handler) AddToDraft(ctx *gin.Context) {
	orderIDStr := ctx.PostForm("order_id")
	id, err := strconv.Atoi(orderIDStr)
	if err != nil || id <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid order_id"))
		return
	}
	creatorID := currentUserIDOrDefault(ctx)
	if reqIDStr := ctx.PostForm("request_id"); strings.TrimSpace(reqIDStr) != "" {
		reqID, _ := strconv.Atoi(reqIDStr)
		if reqID > 0 {
			if err := h.Repository.AddOrderToRequest(reqID, id); err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
				return
			}
			ctx.Redirect(http.StatusFound, "/stages")
			return
		}
	}
	if err := h.Repository.AddOrderToDraft(creatorID, id); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Redirect(http.StatusFound, "/stages")
}

func (h *Handler) DeleteDraft(ctx *gin.Context) {
	if reqIDStr := ctx.PostForm("request_id"); strings.TrimSpace(reqIDStr) != "" {
		reqID, _ := strconv.Atoi(reqIDStr)
		if reqID > 0 {
			if err := h.Repository.DeleteRequest(reqID); err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
				return
			}
			ctx.Redirect(http.StatusFound, "/stages")
			return
		}
	}
	creatorID := currentUserIDOrDefault(ctx)
	if err := h.Repository.DeleteDraftSQL(creatorID); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	ctx.Redirect(http.StatusFound, "/stages")
}

func (h *Handler) ErrorPage(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		code = "not_found"
	}
	ctx.HTML(http.StatusNotFound, "error.html", gin.H{
		"code": code,
	})
}
