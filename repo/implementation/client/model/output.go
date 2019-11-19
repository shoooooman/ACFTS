package model

// Output represents transaction output
type Output struct {
	Amount       int    `json:"amount"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2"`
	PreviousHash string `json:"previous_hash"`
	Used         bool   `json:"used"`
}
