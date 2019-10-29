package model

import "github.com/jinzhu/gorm"

// Transaction is
type Transaction struct {
	gorm.Model
	From   string
	To     string
	Amount int
}
