package utils

import "acfts/model"

// SimpleTransaction is model.Transaction withtout unnecessary properties
type SimpleTransaction struct {
	Inputs  []SimpleInput  `json:"inputs"`
	Outputs []SimpleOutput `json:"outputs"`
}

// SimpleInput is model.Input without unnecessary properties
type SimpleInput struct {
	UTXO       SimpleOutput `json:"utxo"`
	Signature1 string       `json:"signature1"`
	Signature2 string       `json:"signature2"`
}

// SimpleOutput is model.Output without unnecessary properties
type SimpleOutput struct {
	Amount       int    `json:"amount"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2"`
	PreviousHash string `json:"previous_hash"`
	Index        uint   `json:"index"`
}

// ConvertTransaction converts model.Transaction to SimpleTransaction
func ConvertTransaction(transaction model.Transaction) SimpleTransaction {
	simpleTx := SimpleTransaction{}

	simpleTx.Inputs = make([]SimpleInput, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		simpleTx.Inputs[i] = SimpleInput{
			UTXO:       ConvertOutput(input.UTXO),
			Signature1: input.Signature1,
			Signature2: input.Signature2,
		}
	}

	simpleTx.Outputs = make([]SimpleOutput, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		simpleTx.Outputs[i] = ConvertOutput(output)
	}

	return simpleTx
}

// ConvertOutput converts model.Output to simpleOutput
func ConvertOutput(output model.Output) SimpleOutput {
	simpleOutput := SimpleOutput{
		Amount:       output.Amount,
		Address1:     output.Address1,
		Address2:     output.Address2,
		PreviousHash: output.PreviousHash,
		Index:        output.Index,
	}
	return simpleOutput
}
