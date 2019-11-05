package api

import (
	"acfts/db/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func makeSignature(transaction model.Transaction) string {
	// privateKey := getPrivateKey()
	return "hogehoge"
}

// [tmp] for search_id of outputs
var seq = 1

// VerifyTransaction is
func VerifyTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transaction model.Transaction
		c.BindJSON(&transaction)

		inputs := transaction.Inputs
		inputAmount := 0
		for _, input := range inputs {
			utxo := input.UTXO
			db.Where("search_id = ?", utxo.SearchID).First(&utxo)
			if utxo.Used {
				c.JSON(http.StatusOK, gin.H{
					"message": "UTXO is used.",
				})
				return
			}
			/* When choosing to delete records of outputs
			which have already been used. */
			// count := 0
			// db.First(&utxo).Count(&count)
			// if count == 0 {
			// 	c.JSON(http.StatusOK, gin.H{
			// 		"message": "There are no valid UTXOs.",
			// 	})
			// 	return
			// }
			inputAmount += utxo.Amount
		}

		outputs := transaction.Outputs
		outputAmount := 0
		for _, output := range outputs {
			outputAmount += output.Amount
		}

		if inputAmount != outputAmount {
			c.JSON(http.StatusOK, gin.H{
				"message": "Amount of inputs is different from amount of outputs.",
			})
			return
		}

		for _, input := range inputs {
			utxo := input.UTXO
			db.Model(&utxo).Where("search_id = ?", utxo.SearchID).Update("used", true)
			// db.Unscoped().Delete(&utxo)
		}
		for _, output := range outputs {
			output.SearchID = "search" + strconv.Itoa(seq)
			seq++
			db.Create(&output)
		}

		signature := makeSignature(transaction)

		c.JSON(http.StatusOK, gin.H{
			"message":   "Verified this transaction.",
			"signature": signature,
		})
	}
}
