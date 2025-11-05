package api

import (
	"html/template"
	"lab2/internal/app/auth"
	"lab2/internal/app/config"
	"lab2/internal/app/dsn"
	"lab2/internal/app/handler"
	"lab2/internal/app/repository"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func StartServer() {
	log.Println("Starting server")
	_ = godotenv.Load()

	repo, err := repository.NewRepository(dsn.FromEnv())
	if err != nil {
		logrus.WithError(err).Fatal("ошибка инициализации репозитория")
	}

	// load config
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		logrus.WithError(cfgErr).Warn("cannot load config; using defaults")
		cfg = &config.Config{ServiceHost: "0.0.0.0", ServicePort: 8080, JWTSecret: "dev-secret-change-me", JWTTTLMinutes: 60, CookieName: "access_token", RedisAddr: "127.0.0.1:6379"}
	}

	// auth services
	bl := auth.NewRedisBlacklist(cfg.RedisAddr, cfg.RedisPassword)
	jwt := auth.NewJWTService(cfg.JWTSecret, cfg.JWTTTLMinutes, cfg.CookieName, bl)

	handler := handler.NewHandler(repo, jwt)

	r := gin.Default()
	// Make JWT optionally available for all routes (including HTML), so we can get user_id from cookie
	r.Use(auth.OptionalJWT(jwt))

	base := strings.TrimRight(getEnv("MINIO_PUBLIC_BASE", "http://localhost:9000"), "/")
	bucket := strings.Trim(getEnv("MINIO_BUCKET", "images"), "/")
	r.SetFuncMap(template.FuncMap{
		"minioURL": func(key string) string {
			if key == "" {
				return ""
			}
			k := strings.TrimLeft(key, "/")
			return base + "/" + bucket + "/" + k
		},
	})

	resDir, ok := firstExisting(
		"resources",
		"../resources",
		"../../resources",
		"../../../resources",
	)
	if !ok {
		logrus.Fatal("resources directory not found")
	}
	headerPath, ok := firstExisting(
		"header_icon.png",
		"../header_icon.png",
		"../../header_icon.png",
		"../../../header_icon.png",
	)
	if !ok {
		logrus.Fatal("header_icon.png not found")
	}
	tmplGlob, ok := firstGlobExisting(
		"templates/*",
		"../templates/*",
		"../../templates/*",
		"../../../templates/*",
	)
	if !ok {
		logrus.Fatal("templates directory not found")
	}

	r.Static("/resources", resDir)
	r.StaticFile("/header_icon.png", headerPath)
	r.LoadHTMLGlob(tmplGlob)

	// Swagger UI at /swagger/*any. Point UI to static /api/swagger.json so it works without generated docs.
	// Also keep serving static /api/swagger.json when present under resources.
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api/swagger.json")))

	// Simple ping endpoint from methodichka
	r.GET("/ping/:name", handler.Ping)

	r.GET("/stages", handler.GetOrders)
	r.GET("/stage/:id", handler.GetOrder)
	r.GET("/record", handler.GetApplication)
	r.GET("/record/:id", handler.GetApplication)
	r.GET("/error", handler.ErrorPage)
	r.POST("/draft/add", handler.AddToDraft)
	r.POST("/draft/delete", handler.DeleteDraft)

	// JSON API
	api := r.Group("/api")
	api.Use(auth.OptionalJWT(jwt))
	{
		// Пользователи
		api.POST("/users/register", handler.ApiUserRegister)
		api.POST("/users/login", handler.ApiUserLogin)
		api.POST("/users/logout", auth.RequireJWT(jwt), handler.ApiUserLogout)
		api.GET("/users/me", auth.RequireJWT(jwt), handler.ApiUserMe)
		api.PUT("/users/me", auth.RequireJWT(jwt), handler.ApiUserUpdateMe)

		// Корзина/черновик
		api.GET("/records/draft/icon", auth.RequireJWT(jwt), handler.ApiGetDraftIcon)

		// Справочник стадий (услуг)
		api.GET("/stages", handler.ApiListStages)
		api.GET("/stages/:id", handler.ApiGetStage)
		api.POST("/stages", auth.RequireJWT(jwt), auth.RequireModerator(), handler.ApiCreateStage)
		api.PUT("/stages/:id", auth.RequireJWT(jwt), auth.RequireModerator(), handler.ApiUpdateStage)
		api.DELETE("/stages/:id", auth.RequireJWT(jwt), auth.RequireModerator(), handler.ApiDeleteStage)
		api.POST("/stages/:id/image", auth.RequireJWT(jwt), auth.RequireModerator(), handler.ApiSetStageImage)
		api.POST("/stages/:id/draft-add", auth.RequireJWT(jwt), handler.ApiAddStageToDraft)

		// Заявка (record)
		api.POST("/records/draft/items", auth.RequireJWT(jwt), handler.ApiAddDraftItem)
		api.PUT("/records/:id/items", auth.RequireJWT(jwt), handler.ApiUpdateRecordItem)
		api.DELETE("/records/:id/items", auth.RequireJWT(jwt), handler.ApiDeleteRecordItem)

		api.GET("/records", auth.RequireJWT(jwt), handler.ApiListRecords)
		api.GET("/records/:id", auth.RequireJWT(jwt), handler.ApiGetRecord)
		api.PUT("/records/:id", auth.RequireJWT(jwt), handler.ApiUpdateRecord)
		api.PUT("/records/:id/submit", auth.RequireJWT(jwt), handler.ApiSubmitRecord)
		api.PUT("/records/:id/resolve", auth.RequireJWT(jwt), auth.RequireModerator(), handler.ApiResolveRecord)
		api.DELETE("/records/:id", auth.RequireJWT(jwt), handler.ApiDeleteRecord)
	}

	// Serve Swagger OpenAPI JSON if present
	swaggerFile := filepath.Join(resDir, "swagger", "openapi.json")
	if _, err := os.Stat(swaggerFile); err == nil {
		r.StaticFile("/api/swagger.json", swaggerFile)
	}

	r.Run()
	log.Println("Server down")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func firstExisting(paths ...string) (string, bool) {
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil {
			_ = fi
			return p, true
		}
	}
	return "", false
}

func firstGlobExisting(patterns ...string) (string, bool) {
	for _, pat := range patterns {
		if matches, _ := filepath.Glob(pat); len(matches) > 0 {
			return pat, true
		}
	}
	return "", false
}
