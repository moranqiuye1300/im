package main

import (
	"IM/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("dsn"), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败，详细错误：%v", err)
	}
	db.AutoMigrate(&models.UserBasic{})
	user := &models.UserBasic{
		Name: "张三",
	}
	db.Create(user)
	fmt.Println(db.First(user, 1))
	db.Model(user).Update("name", "张三1")
}
