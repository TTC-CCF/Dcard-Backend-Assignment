package models

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var dsn string
	if os.Getenv("APP_ENV") == "test" {
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Taipei",
			os.Getenv("TEST_DB_HOST"),
			os.Getenv("TEST_DB_USER"),
			os.Getenv("TEST_DB_PASSWORD"),
			os.Getenv("TEST_DB_DATABASE"),
			os.Getenv("TEST_DB_PORT"),
		)
	} else {
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Taipei",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_DATABASE"),
			os.Getenv("DB_PORT"),
		)
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

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

	DB = conn
	DB.AutoMigrate(&Banner{}, &Gender{}, &Country{}, &Platform{})
}
