package model

// Transaction is
type Transaction struct {
	// gorm.Model
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}
