package api

import (
	"acfts/api/handler"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SetRouter sets APIs
func SetRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// APIs
	r.GET("/", handler.Ping())
	r.POST("/genesis", handler.CreateGenesis(db))
	r.POST("/transaction", handler.VerifyTransaction(db))
	r.DELETE("/all", handler.DeleteAll(db))

	return r
}
