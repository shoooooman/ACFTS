package api

import (
	"acfts/db/model"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
)

var conns = make(map[int]*websocket.Conn)
var clusterID = 0

// WsHandler is
func WsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		w := c.Writer
		r := c.Request
		upgrader := websocket.Upgrader{
			ReadBufferSize:  8192,
			WriteBufferSize: 8192,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error: failed to set websocket upgrade: %+v", err)
			return
		}
		// defer conn.Close()
		conns[clusterID] = conn
		log.Printf("conns[%d] is updated: %v\n", clusterID, conn.RemoteAddr().String())

		// Register addresses of clients
		clients := []model.Client{}
		err = conn.ReadJSON(&clients)
		if err != nil {
			log.Println("read:", err)
		}
		for _, v := range clients {
			client := model.Client{}
			client.Address1 = v.Address1
			client.Address2 = v.Address2
			client.ClusterID = uint(clusterID)
			db.Create(&client)
		}
		clusterID++

		obj := gin.H{
			"message": "addresses are registered.",
		}
		err = conn.WriteJSON(obj)
		if err != nil {
			log.Println("write:", err)
		}
	}
}
