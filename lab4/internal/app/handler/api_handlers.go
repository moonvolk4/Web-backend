package handler

import (
	"net/http"
	"strconv"
	"strings"

	"lab2/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// helper
func jsonFail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"status": "fail", "message": msg})
}

// current user helpers (populated by JWT middleware)
func CurrentUserID(c *gin.Context) int {
	if v, ok := c.Get("user_id"); ok {
		if id, ok2 := v.(int); ok2 {
			return id
		}
	}
	return 0
}

func IsModerator(c *gin.Context) bool {
	if v, ok := c.Get("is_moderator"); ok {
		if b, ok2 := v.(bool); ok2 {
			return b
		}
	}
	return false
}

// GET /api/records/draft/icon
func (h *Handler) ApiGetDraftIcon(c *gin.Context) {
	uid := CurrentUserID(c)
	cnt, err := h.Repository.GetDraftCount(uid)
	if err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	_, draftID, _ := h.Repository.GetDraftOrders(uid)
	var id *uint
	if draftID != 0 {
		id = &draftID
	}
	c.JSON(http.StatusOK, gin.H{"record_id": id, "items_count": cnt})
}

// GET /api/stages
func (h *Handler) ApiListStages(c *gin.Context) {
	q := strings.TrimSpace(c.Query("query"))
	var res interface{}
	var err error
	if q == "" {
		res, err = h.Repository.GetOrders()
	} else {
		res, err = h.Repository.GetOrdersByTitle(q)
	}
	if err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /api/stages/:id
func (h *Handler) ApiGetStage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	o, err := h.Repository.GetOrder(id)
	if err != nil {
		jsonFail(c, http.StatusNotFound, "stage not found")
		return
	}
	c.JSON(http.StatusOK, o)
}

type stageDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Pressure    string `json:"pressure"`
	RiskName    string `json:"risk_name"`
	RiskClass   string `json:"risk_class"`
	Icon        string `json:"icon"`
	ImageKey    string `json:"image_key"`
	SysFrom     int    `json:"sys_from"`
	SysTo       int    `json:"sys_to"`
	DiaFrom     int    `json:"dia_from"`
	DiaTo       int    `json:"dia_to"`
}

// POST /api/stages
func (h *Handler) ApiCreateStage(c *gin.Context) {
	var dto stageDTO
	if err := c.BindJSON(&dto); err != nil {
		jsonFail(c, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(dto.Title) == "" {
		jsonFail(c, http.StatusBadRequest, "title required")
		return
	}
	o := repository.Order{
		Title:       dto.Title,
		Description: dto.Description,
		RiskClass:   strings.TrimSpace(dto.RiskClass),
		ImageKey:    strings.TrimSpace(dto.ImageKey),
		SysFrom:     dto.SysFrom,
		SysTo:       dto.SysTo,
		DiaFrom:     dto.DiaFrom,
		DiaTo:       dto.DiaTo,
	}
	if err := h.Repository.CreateStage(&o); err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

// PUT /api/stages/:id
func (h *Handler) ApiUpdateStage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var dto stageDTO
	if err := c.BindJSON(&dto); err != nil {
		jsonFail(c, http.StatusBadRequest, "invalid json")
		return
	}
	fields := map[string]any{}
	if dto.Title != "" {
		fields["title"] = dto.Title
	}
	if dto.Description != "" {
		fields["description"] = dto.Description
	}
	// Only allow current schema fields
	if err := h.Repository.UpdateStage(id, fields); err != nil {
		if strings.Contains(err.Error(), "not found") {
			jsonFail(c, http.StatusNotFound, err.Error())
			return
		}
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// DELETE /api/stages/:id
func (h *Handler) ApiDeleteStage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.Repository.DeleteStage(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			jsonFail(c, http.StatusNotFound, err.Error())
			return
		}
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// POST /api/applications/draft/items
type addItemDTO struct {
	StageID       int    `json:"stage_id"`
	Quantity      *int   `json:"quantity"`
	DoctorComment string `json:"doctor_comment"`
}

func (h *Handler) ApiAddDraftItem(c *gin.Context) {
	var dto addItemDTO
	if err := c.BindJSON(&dto); err != nil || dto.StageID <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid json or stage_id")
		return
	}
	if err := h.Repository.AddOrderToDraft(CurrentUserID(c), dto.StageID); err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Optionally update quantity/comment
	if dto.Quantity != nil || dto.DoctorComment != "" {
		draft, err := h.Repository.GetCurrentDraft(CurrentUserID(c))
		if err == nil {
			fields := map[string]any{}
			if dto.Quantity != nil {
				fields["quantity"] = *dto.Quantity
			}
			if dto.DoctorComment != "" {
				fields["doctor_comment"] = dto.DoctorComment
			}
			_ = h.Repository.UpdateApplicationItem(int(draft.ID), dto.StageID, fields)
		}
	}
	c.JSON(http.StatusCreated, gin.H{})
}

// PUT /api/records/:id/items
type updateItemDTO struct {
	StageID       int    `json:"stage_id"`
	Quantity      *int   `json:"quantity"`
	DoctorComment string `json:"doctor_comment"`
}

func (h *Handler) ApiUpdateRecordItem(c *gin.Context) {
	appID, _ := strconv.Atoi(c.Param("id"))
	var dto updateItemDTO
	if err := c.BindJSON(&dto); err != nil || appID <= 0 || dto.StageID <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid json or ids")
		return
	}
	fields := map[string]any{}
	if dto.Quantity != nil {
		fields["quantity"] = *dto.Quantity
	}
	if dto.DoctorComment != "" {
		fields["doctor_comment"] = dto.DoctorComment
	}
	if err := h.Repository.UpdateApplicationItem(appID, dto.StageID, fields); err != nil {
		jsonFail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// DELETE /api/records/:id/items
func (h *Handler) ApiDeleteRecordItem(c *gin.Context) {
	appID, _ := strconv.Atoi(c.Param("id"))
	stageID, _ := strconv.Atoi(c.Query("stage_id"))
	if appID <= 0 || stageID <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid ids")
		return
	}
	if err := h.Repository.DeleteApplicationItem(appID, stageID); err != nil {
		jsonFail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// GET /api/records
func (h *Handler) ApiListRecords(c *gin.Context) {
	f := repository.ApplicationFilter{
		Status:     c.Query("status"),
		FormedFrom: strings.TrimSpace(c.Query("formed_from")),
		FormedTo:   strings.TrimSpace(c.Query("formed_to")),
	}
	// If authenticated and not moderator, restrict to own records
	if !IsModerator(c) {
		if uid := CurrentUserID(c); uid > 0 {
			f.CreatorID = &uid
		}
	}
	apps, err := h.Repository.ListApplications(f)
	if err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, apps)
}

// GET /api/records/:id
func (h *Handler) ApiGetRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	app, err := h.Repository.GetRequest(id)
	if err != nil || app.Status == "удалён" {
		jsonFail(c, http.StatusNotFound, "not found")
		return
	}
	if !IsModerator(c) {
		uid := CurrentUserID(c)
		if uid <= 0 || int(app.CreatorID) != uid {
			jsonFail(c, http.StatusForbidden, "forbidden")
			return
		}
	}
	orders, err := h.Repository.GetRequestOrders(id)
	if err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"application": app, "items": orders})
}

type appFieldsDTO struct {
	Comment           *string `json:"comment"`
	FlagLvHypertrophy *bool   `json:"flag_lv_hypertrophy"`
	FlagRenalDamage   *bool   `json:"flag_renal_damage"`
	FlagArterialStiff *bool   `json:"flag_arterial_stiffness"`
}

// PUT /api/records/:id
func (h *Handler) ApiUpdateRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	// permission: only owner or moderator
	if !IsModerator(c) {
		uid := CurrentUserID(c)
		app, err := h.Repository.GetRequest(id)
		if err != nil || int(app.CreatorID) != uid {
			jsonFail(c, http.StatusForbidden, "forbidden")
			return
		}
	}
	var dto appFieldsDTO
	if err := c.BindJSON(&dto); err != nil {
		jsonFail(c, http.StatusBadRequest, "invalid json")
		return
	}
	fields := map[string]any{}
	if dto.Comment != nil {
		fields["comment"] = *dto.Comment
	}
	if dto.FlagLvHypertrophy != nil {
		fields["flag_lv_hypertrophy"] = *dto.FlagLvHypertrophy
	}
	if dto.FlagRenalDamage != nil {
		fields["flag_renal_damage"] = *dto.FlagRenalDamage
	}
	if dto.FlagArterialStiff != nil {
		fields["flag_arterial_stiffness"] = *dto.FlagArterialStiff
	}
	if err := h.Repository.UpdateApplicationFields(id, fields); err != nil {
		jsonFail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// PUT /api/records/:id/submit
func (h *Handler) ApiSubmitRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.Repository.SubmitApplication(id, CurrentUserID(c)); err != nil {
		jsonFail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

type resolveDTO struct {
	Action string `json:"action"`
}

// PUT /api/records/:id/resolve
func (h *Handler) ApiResolveRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var dto resolveDTO
	if err := c.BindJSON(&dto); err != nil || (dto.Action != "approve" && dto.Action != "reject") {
		jsonFail(c, http.StatusBadRequest, "invalid action")
		return
	}
	approve := dto.Action == "approve"
	// Use current user as moderator to satisfy FK constraint (if present)
	if err := h.Repository.ResolveApplication(id, CurrentUserID(c), approve); err != nil {
		jsonFail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// DELETE /api/records/:id
func (h *Handler) ApiDeleteRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if !IsModerator(c) {
		uid := CurrentUserID(c)
		app, err := h.Repository.GetRequest(id)
		if err != nil || int(app.CreatorID) != uid {
			jsonFail(c, http.StatusForbidden, "forbidden")
			return
		}
	}
	if err := h.Repository.DeleteRequest(id); err != nil {
		jsonFail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// POST /api/stages/:id/image
type imageDTO struct {
	ImageKey string `json:"image_key"`
}

func (h *Handler) ApiSetStageImage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var dto imageDTO
	if err := c.BindJSON(&dto); err != nil {
		jsonFail(c, http.StatusBadRequest, "invalid image_key payload")
		return
	}
	if err := h.Repository.UpdateStage(id, map[string]any{"image_key": dto.ImageKey}); err != nil {
		jsonFail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// POST /api/stages/:id/draft-add
func (h *Handler) ApiAddStageToDraft(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		jsonFail(c, http.StatusBadRequest, "invalid id")
		return
	}
	// optional quantity/comment
	var body struct {
		Quantity      *int   `json:"quantity"`
		DoctorComment string `json:"doctor_comment"`
	}
	_ = c.BindJSON(&body)
	if err := h.Repository.AddOrderToDraft(CurrentUserID(c), id); err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if body.Quantity != nil || strings.TrimSpace(body.DoctorComment) != "" {
		if draft, err := h.Repository.GetCurrentDraft(CurrentUserID(c)); err == nil {
			fields := map[string]any{}
			if body.Quantity != nil {
				fields["quantity"] = *body.Quantity
			}
			if strings.TrimSpace(body.DoctorComment) != "" {
				fields["doctor_comment"] = strings.TrimSpace(body.DoctorComment)
			}
			_ = h.Repository.UpdateApplicationItem(int(draft.ID), id, fields)
		}
	}
	c.JSON(http.StatusCreated, gin.H{})
}

// ---- Ping (demo) ----
type pingResp struct {
	Status string `json:"status"`
}

// Ping godoc
// @Summary      Show hello text
// @Description  very very friendly response
// @Tags         Tests
// @Produce      json
// @Success      200  {object}  pingResp
// @Router       /ping/{name} [get]
func (h *Handler) Ping(c *gin.Context) {
	// name := c.Param("name") // not used in JSON demo
	c.JSON(http.StatusOK, pingResp{Status: "ok"})
}

// ---- User API ----

// POST /api/users/register
func (h *Handler) ApiUserRegister(c *gin.Context) {
	var body struct {
		Username    string `json:"username"`
		Login       string `json:"login"`
		Password    string `json:"password"`
		IsModerator bool   `json:"is_moderator"`
	}
	_ = c.BindJSON(&body)
	login := strings.TrimSpace(body.Username)
	if login == "" {
		login = strings.TrimSpace(body.Login)
	}
	pass := strings.TrimSpace(body.Password)
	if login == "" || pass == "" {
		jsonFail(c, http.StatusBadRequest, "username and password required")
		return
	}
	u, err := h.Repository.CreateUser(login, pass, body.IsModerator)
	if err != nil {
		jsonFail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "username": u.Login, "is_moderator": u.IsModerator})
}

// POST /api/users/login
func (h *Handler) ApiUserLogin(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	_ = c.BindJSON(&body)
	login := strings.TrimSpace(body.Username)
	if login == "" {
		login = strings.TrimSpace(body.Login)
	}
	pass := strings.TrimSpace(body.Password)
	if login == "" || pass == "" {
		jsonFail(c, http.StatusBadRequest, "username and password required")
		return
	}
	u, err := h.Repository.VerifyUser(login, pass)
	if err != nil {
		jsonFail(c, http.StatusUnauthorized, err.Error())
		return
	}
	if h.JWT == nil {
		jsonFail(c, http.StatusInternalServerError, "auth not configured")
		return
	}
	token, claims, err := h.JWT.Generate(int(u.ID), u.Login, u.IsModerator)
	if err != nil {
		jsonFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Set cookie
	maxAge := int(h.JWT.TTL().Seconds())
	c.SetCookie(h.JWT.CookieName(), token, maxAge, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  gin.H{"id": claims.UserID, "username": claims.Login, "is_moderator": claims.IsModerator},
	})
}

// POST /api/users/logout
func (h *Handler) ApiUserLogout(c *gin.Context) {
	// best-effort blacklist: require middleware to have populated jti if logged in
	if h.JWT != nil {
		if v, ok := c.Get("jti"); ok {
			if jti, ok2 := v.(string); ok2 && jti != "" {
				_ = h.JWT.Blacklist(jti)
			}
		}
		// clear cookie
		c.SetCookie(h.JWT.CookieName(), "", -1, "/", "", false, true)
	}
	c.JSON(http.StatusOK, gin.H{})
}

// GET /api/users/me
func (h *Handler) ApiUserMe(c *gin.Context) {
	uid := CurrentUserID(c)
	if uid <= 0 {
		jsonFail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	u, err := h.Repository.GetUserByID(uint(uid))
	if err != nil {
		jsonFail(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "username": u.Login, "is_moderator": u.IsModerator})
}

// PUT /api/users/me
func (h *Handler) ApiUserUpdateMe(c *gin.Context) {
	var body struct {
		Username *string `json:"username"`
	}
	_ = c.BindJSON(&body)
	uid := CurrentUserID(c)
	if uid <= 0 {
		jsonFail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	fields := map[string]any{}
	if body.Username != nil && strings.TrimSpace(*body.Username) != "" {
		fields["login"] = strings.TrimSpace(*body.Username)
	}
	if len(fields) == 0 {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if login, ok := fields["login"].(string); ok {
		if err := h.Repository.UpdateUserLogin(uint(uid), login); err != nil {
			jsonFail(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{})
}
