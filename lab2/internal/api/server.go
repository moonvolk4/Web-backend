package api

import (
	"html/template"
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
)

func StartServer() {
	log.Println("Starting server")
	_ = godotenv.Load()

	repo, err := repository.NewRepository(dsn.FromEnv())
	if err != nil {
		logrus.WithError(err).Fatal("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

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

	r.GET("/stages", handler.GetOrders)
	r.GET("/stage/:id", handler.GetOrder)
	r.GET("/record", handler.GetApplication)
	r.GET("/record/:id", handler.GetApplication)
	r.GET("/error", handler.ErrorPage)
	r.POST("/draft/add", handler.AddToDraft)
	r.POST("/draft/delete", handler.DeleteDraft)

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
