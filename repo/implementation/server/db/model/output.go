package model

import (
	"github.com/jinzhu/gorm"
)

// Output represents transaction output
type Output struct {
	gorm.Model
	Amount   int    `json:"amount"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	Used     bool   `json:"used"`
}