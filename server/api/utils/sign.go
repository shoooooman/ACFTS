package utils

import (
	"acfts/config"
	"acfts/model"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
)

// CreateSignature creates a signature on the transaction
func CreateSignature(transaction model.Transaction) (*big.Int, *big.Int) {
	// Remove gorm.Model, Used, Signatures from outputs
	simpleOutputs := make([]SimpleOutput, len(transaction.Outputs))
	for i := range simpleOutputs {
		simpleOutputs[i] = ConvertOutput(transaction.Outputs[i])
	}

	// Convert []simpleOutput to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v", simpleOutputs))

	// Get hash using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	fmt.Println("hash when creating")
	fmt.Println(hashed)

	// Get signature using ellipse curve cryptography
	r, s, err := ecdsa.Sign(rand.Reader, config.GetKey(), hashed)
	if err != nil {
		panic(err)
	}

	return r, s
}
