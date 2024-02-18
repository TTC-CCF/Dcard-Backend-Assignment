package main

import (
	"main/cache"
	"main/models"
	"main/routers"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	router := routers.Init()
	models.Init()
	cache.Init()

	port := os.Getenv("APP_PORT")
	router.Run(":" + port)
}
