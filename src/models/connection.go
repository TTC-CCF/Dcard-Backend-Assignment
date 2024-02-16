package models

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Taipei",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_PORT"),
	)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		for i := 0; i < 10; i++ {
			fmt.Println("Failed to connect to database. Retrying...")
			conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err == nil {
				break
			}
			time.Sleep(2 * time.Second)
		}

		if err != nil {
			panic(err)
		}
	}

	sqlDb, _ := conn.DB()

	maxConn, _ := strconv.Atoi(os.Getenv("DB_MAX_CONN"))
	maxIdle, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE"))
	sqlDb.SetMaxIdleConns(maxConn)
	sqlDb.SetMaxOpenConns(maxIdle)

	db = conn
	db.AutoMigrate(&Banner{}, &Gender{}, &Country{}, &Platform{})
}
