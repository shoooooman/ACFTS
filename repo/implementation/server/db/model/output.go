package model

import "github.com/jinzhu/gorm"

// Output represents transaction output
type Output struct {
	gorm.Model
	SearchID string `json:"id"`
	Amount   int    `json:"amount"`
	Used     bool   `json:"used"`
}
