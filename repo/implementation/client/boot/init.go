package boot

import (
	"acfts-client/api"
	"acfts-client/model"
	"fmt"
	"log"

	// mrand "math/rand"

	"github.com/Equanox/gotron"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	// For gorm mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// SetRouter sets apis
func SetRouter(db *gorm.DB, window *gotron.BrowserWindow) *gin.Engine {
	r := gin.Default()

	// APIs
	r.GET("/address", api.GetAddrs(db))
	r.POST("/output", api.ReceiveUTXO(db, window))
	r.DELETE("/output", api.ClearOutputs(db))

	return r
}

// SetDB starts DB
func SetDB(num int) *gorm.DB {
	setting := fmt.Sprintf("root:@tcp(127.0.0.1:3306)/acfts_client_%d?charset=utf8&parseTime=True&loc=Local", num)
	db, err := gorm.Open("mysql", setting)
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

	return db
}

// DeleteAll deletes all data in DB
func DeleteAll(db *gorm.DB) {
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
