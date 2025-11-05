package main

import (
	"log"

	"lab2/internal/api"
)

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform
// @contact.name API Support
// @contact.url https://vk.com/bmstu_schedule
// @contact.email bitop@spatecon.ru
// @license.name AS IS (NO WARRANTY)
// @host 127.0.0.1
// @schemes http https
// @BasePath /
func main() {
	log.Println("Application started")
	api.StartServer()
	log.Println("Application stopped")
}
