package api

import (
	"acfts/db/model"
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func createSignature(transaction model.Transaction) (*big.Int, *big.Int) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// Convert transaction struct to binary
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &transaction)
	fmt.Printf("buf=% x\n", buf.Bytes())

	h := crypto.Hash.New(crypto.SHA256)
	h.Write(([]byte)(buf.Bytes()))
	hashed := h.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed)
	if err != nil {
		panic(err)
	}

	// for verification (client side)
	// if ecdsa.Verify(&privateKey.PublicKey, hashed, r, s) {
	// 	fmt.Println("Verifyed!")
	// }

	return r, s
}

func unlockUTXO(utxo Output, key string) bool {
	return true
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

			if !unlockUTXO(utxo, input.key) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Could not unlock UTXO.",
				})
				return
			}
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

		for i, input := range inputs {
			utxo := input.UTXO
			db.Model(&utxo).Where("search_id = ?", utxo.SearchID).Update("used", true)
			// db.Unscoped().Delete(&utxo)

			transaction.Inputs[i].UTXO = utxo
		}
		for i, output := range outputs {
			output.SearchID = "search" + strconv.Itoa(seq)
			seq++
			db.Create(&output)

			db.Where("search_id = ?", output.SearchID).First(&output)
			transaction.Outputs[i] = output
		}

		r, s := createSignature(transaction)

		c.JSON(http.StatusOK, gin.H{
			"message":     "Verified this transaction.",
			"transaction": transaction,
			"signature1":  r,
			"signature2":  s,
		})
	}
}
