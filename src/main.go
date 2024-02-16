package main

import (
	"main/models"
	"main/routers"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	router := routers.Init()
	models.Init()
	router.Run(":8080")
}
