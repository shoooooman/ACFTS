package model

// Input represents transaction inputs
type Input struct {
	UTXO Output `json:"utxo"`
	Key  string `json:"key"`
}
