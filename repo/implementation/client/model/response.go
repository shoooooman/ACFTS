package model

import "math/big"

// Response is
type Response struct {
	Message     string      `json:"message"`
	Address1    string      `json:"address1"`
	Address2    string      `json:"address2"`
	Signature1  *big.Int    `json:"signature1"`
	Signature2  *big.Int    `json:"signature2"`
	Transaction Transaction `json:"transaction"`
}
