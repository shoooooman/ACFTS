package config

import (
	"acfts-client/model"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"

	// For gorm mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// SetDB starts DB
func SetDB(num int) *gorm.DB {
	setting := fmt.Sprintf("root:@tcp(127.0.0.1:3306)/acfts_client_%d?charset=utf8&parseTime=True&loc=Local", num)
	var err error
	db, err = gorm.Open("mysql", setting)
	if err != nil {
		log.Println("failed to connect database")
		log.Panicln(err)
	}

	db.AutoMigrate(&model.Output{})
	db.AutoMigrate(&model.Signature{})
	db.AutoMigrate(&model.Cluster{})
	db.AutoMigrate(&model.Client{})

	// Add a index to improve throughput of queries
	db.Model(&model.Output{}).
		AddIndex("idx_address_hash", "address1", "address2", "previous_hash")
	db.Model(&model.Signature{}).AddIndex("idx_signature", "output_id")

	deleteAll()

	return db
}

// GetDB returns a pointer to gorm.DB
func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("GetDB: DB is not set")
	}
	return db
}

// deleteAll deletes all data in DB
func deleteAll() {
	output := model.Output{}
	signature := model.Signature{}
	client := model.Client{}
	cluster := model.Cluster{}

	db.DropTable(&output)
	db.DropTable(&signature)
	db.DropTable(&client)
	db.DropTable(&cluster)

	db.AutoMigrate(&output)
	db.AutoMigrate(&signature)
	db.AutoMigrate(&client)
	db.AutoMigrate(&cluster)

	db.Model(&model.Output{}).
		AddIndex("idx_address_hash", "address1", "address2", "previous_hash")
	db.Model(&model.Signature{}).AddIndex("idx_signature", "output_id")
}
