package api

import (
	"acfts/db/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// CreateGenesis makes a genesis transaction
func CreateGenesis(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var output model.Output
		c.BindJSON(&output)

		genesis := model.Output{
			Amount:       output.Amount,
			Address1:     output.Address1,
			Address2:     output.Address2,
			PreviousHash: "genesis",
			Used:         false,
		}
		db.Create(&genesis)

		json := convertOutput(genesis)
		c.JSON(http.StatusOK, gin.H{
			"message": "Genesis is created.",
			"genesis": json,
		})
	}
}
