package boot

import (
	"acfts-client/api"

	// mrand "math/rand"

	"github.com/Equanox/gotron"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
