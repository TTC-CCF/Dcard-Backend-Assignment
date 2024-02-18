package main

import (
	"fmt"
	"main/cache"
	"main/models"
	"main/routers"
	"main/tests/load_test"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "load_test" {
			err := godotenv.Load("../.env")
			if err != nil {
				panic(err)
			}

			os.Setenv("APP_ENV", "test")
			models.Init()
			cache.Init()
			router := routers.Init()

			load_test.DeleteAllData()
			load_test.InsertLoadTestData()
			fmt.Println("Load test data inserted")

			port := os.Getenv("TEST_PORT")
			router.Run(":" + port)
		}
	} else {
		err := godotenv.Load()
		if err != nil {
			panic(err)
		}

		router := routers.Init()
		models.Init()
		cache.Init()

		port := os.Getenv("APP_PORT")
		router.Run(":" + port)
	}

}
