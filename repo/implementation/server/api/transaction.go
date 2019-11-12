package api

import (
	"acfts/db/model"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
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

	// Convert transaction struct to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v", transaction))

	// Get hash using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	// Get signature using ellipse curve cryptography
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed)
	if err != nil {
		panic(err)
	}

	return r, s
}

func unlockUTXO(utxo model.Output, signature1, signature2 string) bool {
	// Convert transaction struct to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v", utxo))

	// Get a hash value using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	// Convert string to big.Int
	sigInt1, err := strconv.Atoi(signature1)
	if err != nil {
		fmt.Println("Signature1 is not valid.")
		return false
	}
	sigBigInt1 := big.NewInt(int64(sigInt1))

	sigInt2, err := strconv.Atoi(signature2)
	if err != nil {
		fmt.Println("Signature2 is not valid.")
		return false
	}
	sigBigInt2 := big.NewInt(int64(sigInt2))

	// model.Output.Address represents public key
	// Convert string to ecdsa.PublicKey
	x509EncodedPub := []byte(utxo.Address)
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		fmt.Println("Address is not valid.")
		return false
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	if ecdsa.Verify(publicKey, hashed, sigBigInt1, sigBigInt2) {
		fmt.Println("Verifyed!")
		return true
	}
	return false
}

// [tmp] for search_id of outputs
var seq = 1

type inputJSON struct {
	UTXO       outputJSON `json:"utxo"`
	Signature1 string     `json:"sig1"`
	Signature2 string     `json:"sig2"`
}

// Just for responses
type outputJSON struct {
	SearchID string `json:"id"`
	Amount   int    `json:"amount"`
	Address  string `json:"address"`
	Used     bool   `json:"used"`
}

// Just for responses
type transactionJSON struct {
	Inputs  []inputJSON  `json:"inputs"`
	Outputs []outputJSON `json:"outputs"`
}

// Convert model.Output to outputJSON
func convertOutput(output model.Output) outputJSON {
	json := outputJSON{}
	json.SearchID = output.SearchID
	json.Amount = output.Amount
	json.Address = output.Address
	json.Used = output.Used
	return json
}

// Convert model.Transaction to transactionJSON
func convertTransaction(transaction model.Transaction) transactionJSON {
	json := transactionJSON{}

	json.Inputs = make([]inputJSON, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		utxo := input.UTXO
		json.Inputs[i].UTXO = convertOutput(utxo)
		json.Inputs[i].Signature1 = input.Signature1
		json.Inputs[i].Signature2 = input.Signature2
	}

	json.Outputs = make([]outputJSON, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		json.Outputs[i] = convertOutput(output)
	}

	return json
}

// VerifyTransaction is
func VerifyTransaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transaction model.Transaction
		c.BindJSON(&transaction)

		inputs := transaction.Inputs
		inputAmount := 0
		for _, input := range inputs {
			utxo := input.UTXO
			count := 0
			db.Where("search_id = ?", utxo.SearchID).First(&utxo).Count(&count)
			if count == 0 {
				c.JSON(http.StatusOK, gin.H{
					"message": "Input is not valid.",
				})
				return
			}
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

			if !unlockUTXO(utxo, input.Signature1, input.Signature2) {
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

			db.Where("search_id = ?", utxo.SearchID).First(&utxo)
			transaction.Inputs[i].UTXO = utxo
		}
		for i, output := range outputs {
			output.SearchID = "output" + strconv.Itoa(seq)
			seq++
			db.Create(&output)

			db.Where("search_id = ?", output.SearchID).First(&output)
			transaction.Outputs[i] = output
		}

		r, s := createSignature(transaction)

		json := convertTransaction(transaction)
		c.JSON(http.StatusOK, gin.H{
			"message":     "Verified this transaction.",
			"transaction": json,
			"signature1":  r,
			"signature2":  s,
		})
	}
}
