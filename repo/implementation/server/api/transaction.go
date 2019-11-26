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
	// TODO: Outputsのgorm.Modelは時刻が厄介なので抜いたほうがいいかもしれない
	// Convert transaction.Outputs to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v", transaction.Outputs))
	fmt.Println("outputs when creating")
	fmt.Println(transaction.Outputs)

	// Get hash using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	fmt.Println("hash when creating")
	fmt.Println(hashed)

	// Get signature using ellipse curve cryptography
	r, s, err := ecdsa.Sign(rand.Reader, key, hashed)
	if err != nil {
		panic(err)
	}

	return r, s
}

func unlockUTXO(utxo model.Output, signature1, signature2 string) bool {
	// Convert transaction struct to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v%v%v", utxo.Address1, utxo.Address2, utxo.PreviousHash))

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

	// model.Output.Address1 and Address2 represent a public key
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

func verifyUTXO(utxo model.Output, siblings []model.Output) bool {
	signatures := utxo.Signatures

	// Create the same array as when creating a signature
	outputs := make([]model.Output, 1+len(siblings))
	for i := range outputs {
		if i < int(utxo.Index) {
			outputs[i] = siblings[i]
			outputs[i].Used = false // FIXME
		} else if i > int(utxo.Index) {
			outputs[i] = siblings[i-1]
			outputs[i].Used = false // FIXME
		} else {
			// utxoWithoutSigs := model.Output{
			// 	gorm.Model:   utxo.gorm.Model,
			// 	Amount:       utxo.Amount,
			// 	Address1:     utxo.Address1,
			// 	Address2:     utxo.Address2,
			// 	PreviousHash: utxo.PreviousHash,
			// 	Index:        utxo.Index,
			// }
			// outputs[i] = utxoWithoutSigs
			utxo.Used = false                            // FIXME
			utxo.Signatures = make([]model.Signature, 0) // FIXME
			outputs[i] = utxo
		}
	}

	// Convert transaction struct to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v", outputs))
	fmt.Println("outputs when verification")
	fmt.Println(outputs)

	// Get a hash value using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	fmt.Println("hash when verification")
	fmt.Println(hashed)

	// Genesis is approved without verification
	if utxo.PreviousHash == "genesis" {
		return true
	}

	valid := 0
	for _, signature := range signatures {
		// FIXME: Should get public keys of other servers independently of clients
		address1, _ := new(big.Int).SetString(signature.Address1, 10)
		address2, _ := new(big.Int).SetString(signature.Address2, 10)
		serverPubKey := ecdsa.PublicKey{elliptic.P521(), address1, address2}
		signature1, _ := new(big.Int).SetString(signature.Signature1, 10)
		signature2, _ := new(big.Int).SetString(signature.Signature2, 10)
		if ecdsa.Verify(&serverPubKey, hashed, signature1, signature2) {
			valid++
			fmt.Println("one valid signature")
			if float64(valid) >= 2.0*N/3.0 {
				fmt.Println("Server Verifyed!")
				return true
			}
		} else {
			fmt.Println("not valid")
		}
	}
	return false
}

// Just for responses
type inputJSON struct {
	UTXO       outputJSON `json:"utxo"`
	Signature1 string     `json:"signature1"`
	Signature2 string     `json:"signature2"`
}

// Just for responses
type outputJSON struct {
	Amount       int    `json:"amount"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2"`
	PreviousHash string `json:"previous_hash"`
	Index        uint   `json:"index"`
	Used         bool   `json:"used"`
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
	json.PreviousHash = output.PreviousHash
	json.Index = output.Index
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
		if inputs == nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "There are no inputs.",
			})
			return
		}

		outputs := transaction.Outputs
		if outputs == nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "There are no outputs.",
			})
			return
		}

		inputAmount := 0
		for _, input := range inputs {
			utxo := input.UTXO
			count := 0
			db.Where("address1 = ? AND address2 = ? AND previous_hash = ?", utxo.Address1, utxo.Address2, utxo.PreviousHash).First(&utxo).Count(&count)
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

			// Update gorm.Model of Siblings
			db.Where("id <> ? AND previous_hash = ?", utxo.ID, utxo.PreviousHash).Find(&input.Siblings)
			if !verifyUTXO(utxo, input.Siblings) {
				c.JSON(http.StatusOK, gin.H{
					"message": "UTXO does not have enough signatures.",
				})
				return
			}

			inputAmount += utxo.Amount
		}

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
			// db.Model(&utxo).Where("address1 = ? AND address2 = ? AND previous_hash = ?", utxo.Address1, utxo.Address2, utxo.PreviousHash).Update("used", true)
			// db.Unscoped().Delete(&utxo)

			db.Where("address1 = ? AND address2 = ? AND previous_hash = ?", utxo.Address1, utxo.Address2, utxo.PreviousHash).First(&utxo).Update("used", true)
			transaction.Inputs[i].UTXO = utxo
		}

		for i, output := range outputs {
			db.Create(&output)

			db.Where("address1 = ? AND address2 = ? AND previous_hash = ?", output.Address1, output.Address2, output.PreviousHash).First(&output)
			transaction.Outputs[i] = output
		}

		r, s := createSignature(transaction)

		json := convertTransaction(transaction)
		c.JSON(http.StatusOK, gin.H{
			"message":     "Verified this transaction.",
			"transaction": json,
			"address1":    (&key.PublicKey).X.String(),
			"address2":    (&key.PublicKey).Y.String(),
			"signature1":  r,
			"signature2":  s,
		})
	}
}
