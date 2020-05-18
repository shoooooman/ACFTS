package transaction

import (
	"acfts-client/config"
	"acfts-client/model"
	"acfts-client/utils"
	"encoding/json"
	"log"
)

// Execute executes transactions
func Execute(txs []GeneralTx) {
	db := config.DB
	for i := 0; i < len(txs); i++ {
		jsonStr, err := createJSONStr(txs[i])
		if err != nil {
			log.Println(err)
			return
		}
		// fmt.Println(jsonStr)

		// Send requests to all servers
		b := make(chan []byte, config.NumServers)
		for _, serverURL := range config.ServerURLs {
			url := serverURL + "/transaction"
			go utils.Post(url, jsonStr, b)
		}

		var outputs []model.Output
		sigs := make([]model.Signature, config.NumServers)
		// Wait for responses from all servers
		for j := 0; j < config.NumServers; j++ {
			body := <-b
			response := model.Response{}
			err := json.Unmarshal(body, &response)
			if err != nil {
				log.Panicln(err)
			}
			log.Println("Message: " + response.Message)

			outputs = response.Transaction.Outputs
			sig := model.Signature{
				Address: model.Address{
					Address1: response.Address1,
					Address2: response.Address2,
				},
				Signature1: response.Signature1.String(),
				Signature2: response.Signature2.String(),
			}
			sigs[j] = sig
		}

		if len(outputs) == 0 {
			log.Panicln("Error: there are no outputs.")
		}

		related := make(map[uint]bool)
		for _, output := range outputs {
			// Update signatures
			db.Where("address1 = ? AND address2 = ? AND previous_hash = ?",
				output.Address1, output.Address2, output.PreviousHash).First(&output)

			// Get a related client
			receiver := model.Client{}
			raddr1 := output.Address1
			raddr2 := output.Address2
			db.Where("address1 = ? AND address2 = ?", raddr1, raddr2).First(&receiver)
			related[receiver.ClusterID] = true
		}

		// log.Printf("related %v\n", related)

		r := make(chan []byte, len(related))
		cnt := 0
		for cid := range related {
			// If myself is related, add outputs to db
			// If else, send outputs to different clusters
			if cid == 1 {
				updateOutputs(outputs, sigs)
			} else {
				cluster := model.Cluster{}
				db.Where("id = ?", cid).First(&cluster)
				url := cluster.URL + "/output"

				hash := outputs[0].PreviousHash
				outputs := createOutputStr(outputs, hash, sigs)
				jsonStr := `
{
	"outputs": ` + outputs + `
}`

				go utils.Post(url, jsonStr, r)
				cnt++
			}
		}

		// Wait for responses from all servers
		for j := 0; j < cnt; j++ {
			body := <-r
			response := struct {
				Message string         `json:"message"`
				UTXOs   []model.Output `json:"utxos"`
			}{}
			err := json.Unmarshal(body, &response)
			if err != nil {
				log.Panicln(err)
			}
			log.Printf("Message: %s\n", response.Message)
			// log.Println("received related utxos")
			// jsonBytes, _ := json.MarshalIndent(response.UTXOs, "", "    ")
			// log.Println(string(jsonBytes))
		}
	}
}
