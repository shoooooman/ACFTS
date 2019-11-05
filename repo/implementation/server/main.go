package main

import (
	"acfts/api"
	"acfts/db/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func initDB() *gorm.DB {
	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/acfts?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Output{})

	return db
}

func initRoute(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// APIs
	r.GET("/", api.Ping())
	r.POST("/transaction", api.VerifyTransaction(db))

	return r
}

func main() {
	db := initDB()
	defer db.Close()

	r := initRoute(db)
	r.Run()
}
