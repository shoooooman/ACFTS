package handler

import (
	"acfts/api/utils"
	"acfts/config"
	"acfts/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// VerifyTransaction verifies transactions from clients
func VerifyTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transaction model.Transaction
		c.BindJSON(&transaction)

		inputs := transaction.Inputs
		if len(inputs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "There are no inputs.",
			})
			return
		}

		outputs := transaction.Outputs
		if len(outputs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "There are no outputs.",
			})
			return
		}

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Transacton error. Rollback.",
				})
				return
			}
		}()

		inputAmount := 0
		for _, input := range inputs {
			utxo := input.UTXO
			count := 0
			tx.Where("address1 = ? AND address2 = ? AND previous_hash = ? AND output_index = ?",
				utxo.Address1, utxo.Address2, utxo.PreviousHash, utxo.Index).
				First(&utxo).Count(&count)
			if count == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Input is not valid.",
				})
				tx.Rollback()
				return
			}
			if utxo.Used {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "UTXO is used.",
				})
				tx.Rollback()
				return
			}
			/* When choosing to delete records of outputs
			which have already been used. */
			// count := 0
			// db.First(&utxo).Count(&count)
			// if count == 0 {
			// 	c.JSON(http.StatusBadRequest, gin.H{
			// 		"message": "There are no valid UTXOs.",
			// 	})
			// 	return
			// }

			if !utils.UnlockUTXO(utxo, input.Signature1, input.Signature2) {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Could not unlock UTXO.",
				})
				tx.Rollback()
				return
			}

			// Update gorm.Model of Siblings
			tx.Where("id <> ? AND previous_hash = ?", utxo.ID, utxo.PreviousHash).
				Find(&input.Siblings)
			if !utils.VerifyUTXO(utxo, input.Siblings) {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "One of signatures of servers is not valid.",
				})
				tx.Rollback()
				return
			}

			inputAmount += utxo.Amount
		}

		outputAmount := 0
		for _, output := range outputs {
			outputAmount += output.Amount
		}

		if inputAmount != outputAmount {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Amount of inputs is different from amount of outputs.",
			})
			tx.Rollback()
			return
		}

		for i, input := range inputs {
			utxo := input.UTXO
			// db.Model(&utxo).Where("address1 = ? AND address2 = ? AND previous_hash = ?",
			// 	utxo.Address1, utxo.Address2, utxo.PreviousHash).Update("used", true)
			// db.Unscoped().Delete(&utxo)

			tx.Where("address1 = ? AND address2 = ? AND previous_hash = ? AND output_index = ?",
				utxo.Address1, utxo.Address2, utxo.PreviousHash, utxo.Index).
				First(&utxo).Update("used", true)
			transaction.Inputs[i].UTXO = utxo
		}
		tx.Commit()

		// Add records of outputs
		for i, output := range outputs {
			db.Create(&output)
			transaction.Outputs[i] = output
		}

		r, s := utils.CreateSignature(transaction)
		simpleTx := utils.ConvertTransaction(transaction)

		key := config.GetKey()
		c.JSON(http.StatusCreated, gin.H{
			"message":     "Verified this transaction.",
			"transaction": simpleTx,
			"address1":    (&key.PublicKey).X.String(),
			"address2":    (&key.PublicKey).Y.String(),
			"signature1":  r,
			"signature2":  s,
		})
	}
}
