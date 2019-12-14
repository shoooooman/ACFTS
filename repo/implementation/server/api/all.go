package api

import (
	"acfts/db/model"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
)

// DeleteAll is
func DeleteAll(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		output := model.Output{}
		client := model.Client{}
		db.Unscoped().Delete(&output)
		db.Unscoped().Delete(&client)

		conns = make(map[int]*websocket.Conn)
		clusterID = 0

		// For benchmarking of transactions between clusters
		// db.Unscoped().Where("cluster_id > 0").Delete(&client)
		// conn0 := conns[0]
		// clusterID = 1
		// conns = make(map[int]*websocket.Conn)
		// conns[0] = conn0

		c.JSON(200, gin.H{
			"message": "all data is deleted.",
		})
	}
}
