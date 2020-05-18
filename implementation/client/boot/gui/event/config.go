package event

import (
	"acfts-client/config"
	"bytes"
	"encoding/json"
	"log"
)

// HandleConfig is called when receiving "config" message from frontend
func HandleConfig(c chan bool) func([]byte) {
	return func(bin []byte) {
		buf := bytes.NewBuffer(bin)
		log.Println(buf)

		resp := struct {
			Cluster int  `json:"cluster"`
			Client  int  `json:"address"`
			All     int  `json:"all"`
			Genesis bool `json:"genesis"`
			GAmount int  `json:"gamount"`
		}{}
		if err := json.Unmarshal(bin, &resp); err != nil {
			log.Fatal(err)
		}
		config.Num = resp.Cluster
		config.NumClients = resp.Client
		config.NumClusters = resp.All
		config.HasGenesis = resp.Genesis
		if config.HasGenesis {
			config.GAmount = resp.GAmount
		}
		c <- true
	}
}
