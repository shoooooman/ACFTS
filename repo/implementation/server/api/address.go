package api

import (
	"acfts/db/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetAddresses is
func GetAddresses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clients := []model.Client{}
		db.Find(&clients)

		addrs := make([]model.Address, len(clients))
		for i, client := range clients {
			addrs[i].Address1 = client.Address1
			addrs[i].Address2 = client.Address2
		}

		c.JSON(200, gin.H{
			"message":   "success",
			"addresses": addrs,
		})
	}
}
