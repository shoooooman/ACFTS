package model

// Input represents transaction inputs
type Input struct {
	// gorm.Model
	UTXO Output `json:"utxo"`
	Key  string `json:"key"`
}
