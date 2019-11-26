package model

import "github.com/jinzhu/gorm"

// Signature represents signatures of servers
type Signature struct {
	gorm.Model
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Signature1 string `json:"signature1"`
	Signature2 string `json:"signature2"`
	OutputID   uint
}
