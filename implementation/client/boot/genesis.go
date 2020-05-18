package boot

import (
	"acfts-client/config"
	"acfts-client/model"
	"acfts-client/utils"
	"encoding/json"
	"log"
	"strconv"
)

// CreateGenesis creates genesis tx in DB
func CreateGenesis(owner model.Address, amount int) {
	db := config.GetDB()

	// Dummy signatures for genesis
	sig := model.Signature{
		Address: model.Address{
			Address1: "gene",
			Address2: "sis",
		},
		Signature1: "dum",
		Signature2: "my",
		OutputID:   1,
	}
	sigs := make([]model.Signature, config.NumServers)
	for i := 0; i < config.NumServers; i++ {
		sigs[i] = sig
	}

	genesis := model.Output{
		Amount: amount,
		Address: model.Address{
			Address1: owner.Address1,
			Address2: owner.Address2,
		},
		PreviousHash: "genesis",
		Index:        0,
		Used:         false,
		Signatures:   sigs,
	}
	db.Create(&genesis)

	jsonStr := `
	{
		"amount": ` + strconv.Itoa(amount) + `,
		"address1": "` + owner.Address1 + `",
		"address2": "` + owner.Address2 + `"
	}`

	// Create genesis records in servers
	b := make(chan []byte, config.NumServers)
	for _, surl := range config.ServerURLs {
		url := surl + "/genesis"
		go utils.Post(url, jsonStr, b)
	}

	// Wait for responses from all servers
	for i := 0; i < config.NumServers; i++ {
		body := <-b
		gen := struct {
			Message string       `json:"message"`
			Genesis model.Output `json:"genesis"`
		}{}
		err := json.Unmarshal(body, &gen)
		if err != nil {
			log.Panicln(err)
		}
		log.Println(gen.Message)
	}
}
