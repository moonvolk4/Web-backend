package api

import (
	"html/template"
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
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

	r.Static("/resources", "/Users/moonvolk4/bmstu/3course/4sem/web/lab1/resources")
	r.StaticFile("/header_icon.png", "/Users/moonvolk4/bmstu/3course/4sem/web/lab1/header_icon.png")
	r.LoadHTMLGlob("/Users/moonvolk4/bmstu/3course/4sem/web/lab1/templates/*")

	r.GET("/stages", handler.GetOrders)
	r.GET("/stage/:id", handler.GetOrder)
	// Корзина-заявка: поддерживаем и /request, и /request/:id
	r.GET("/request", handler.GetApplication)
	r.GET("/request/:id", handler.GetApplication)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
