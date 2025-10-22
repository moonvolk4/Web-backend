package main

import (
	"log"

	"lab2/internal/api"
)

func main() {
	log.Println("Application started")
	api.StartServer()
	log.Println("Application stopped")
}
