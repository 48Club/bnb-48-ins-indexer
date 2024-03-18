package main

import (
	"fmt"
	"os"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	token := os.Getenv("TOKEN")
	if token == "" {
		panic("TOKEN environment variable is empty")
	}

	b, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}
	bot = b

	mysqlPWD := os.Getenv("MYSQL_PASSWORD")
	if mysqlPWD == "" {
		panic("MYSQL_PASSWORD environment variable is empty")
	}

	dsn := fmt.Sprintf("root:%s@tcp(mysql:3306)/bnb48_inscription?charset=utf8mb4&parseTime=True&loc=Local", mysqlPWD)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	mySql = db.Debug()
}
