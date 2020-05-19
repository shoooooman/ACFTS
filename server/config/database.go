package config

import (
	"acfts/model"
	"log"
	"strconv"

	"github.com/jinzhu/gorm"

	// For mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// InitDB connects to the database
func InitDB(port int) *gorm.DB {
	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/acfts_"+strconv.Itoa(port)+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("initDB: failed to connect database")
	}

	db.AutoMigrate(&model.Output{})

	return db
}
