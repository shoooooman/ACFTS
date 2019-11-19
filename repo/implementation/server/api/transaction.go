package api

import (
	"acfts/db/model"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func createSignature(transaction model.Transaction) (*big.Int, *big.Int) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// Convert transaction struct to bytes to get its hash
	// FIXME: should make a hash of outputs
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
	buf := []byte(fmt.Sprintf("%v%v", utxo.Address1, utxo.Address2))

	// Get a hash value using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	// Convert string to big.Int
	var ok bool
	n1 := new(big.Int)
	n1, ok = n1.SetString(signature1, 10)
	if !ok {
		fmt.Println("Signature1 is not valid.")
		return false
	}

	n2 := new(big.Int)
	n2, ok = n2.SetString(signature2, 10)
	if !ok {
		fmt.Println("Signature2 is not valid.")
		return false
	}

	// model.Output.Address1 and Address2 represent public key
	// Convert Address1 and Address2 to ecdsa.PublicKey
	address1, _ := new(big.Int).SetString(utxo.Address1, 10)
	address2, _ := new(big.Int).SetString(utxo.Address2, 10)
	publicKey := ecdsa.PublicKey{elliptic.P521(), address1, address2}

	if ecdsa.Verify(&publicKey, hashed, n1, n2) {
		fmt.Println("Verifyed!")
		return true
	}
	return false
}

type inputJSON struct {
	UTXO       outputJSON `json:"utxo"`
	Signature1 string     `json:"sig1"`
	Signature2 string     `json:"sig2"`
}

// Just for responses
type outputJSON struct {
	Amount   int    `json:"amount"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
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
	json.Amount = output.Amount
	json.Address1 = output.Address1
	json.Address2 = output.Address2
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
			db.Where("address1 = ? AND address2 = ?", utxo.Address1, utxo.Address2).First(&utxo).Count(&count)
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
			db.Model(&utxo).Where("address1 = ? AND address2 = ?", utxo.Address1, utxo.Address2).Update("used", true)
			// db.Unscoped().Delete(&utxo)

			db.Where("address1 = ? AND address2 = ?", utxo.Address1, utxo.Address2).First(&utxo)
			transaction.Inputs[i].UTXO = utxo
		}

		// FIXME: dummy outputs
		// Needs to add outputs to requests
		for i, output := range outputs {
			db.Create(&output)

			db.Where("address1 = ? AND address2 = ?", output.Address1, output.Address2).First(&output)
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
