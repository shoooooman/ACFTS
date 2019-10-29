package api

import (
	"acfts/db/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SendTransaction is
func SendTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		transaction := model.Transaction{From: "foo", To: "bar", Amount: 100}
		db.Create(&transaction)
		c.JSON(200, gin.H{
			"message": "send transaction",
		})
	}
}
