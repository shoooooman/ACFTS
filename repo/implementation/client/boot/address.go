package boot

import (
	"acfts-client/config"
	"acfts-client/model"
	"acfts-client/utils"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"log"
	"strconv"
)

// GenerateClients creates key pairs and set it to DB
func GenerateClients(url string) {
	db := config.GetDB()

	config.Keys = make([]*ecdsa.PrivateKey, config.NumClients)
	config.Pub2Pri = make(map[string]*ecdsa.PrivateKey)

	cluster := model.Cluster{URL: url}
	if db == nil {
		log.Fatal("GenerateClients: DB is not set")
	}
	db.Create(&cluster)

	for i := 0; i < config.NumClients; i++ {
		var err error
		config.Keys[i], err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			log.Panicln(err)
		}
		pub := &config.Keys[i].PublicKey
		config.Pub2Pri[pub.X.String()+pub.Y.String()] = config.Keys[i]

		client := model.Client{
			Address: model.Address{
				Address1: pub.X.String(),
				Address2: pub.Y.String(),
			},
		}
		db.Model(&cluster).Association("Clients").Append(client)
	}
}

// GetOtherCURLs returns other client URLs
func GetOtherCURLs(num int) []string {
	otherClients := make([]string, 0)
	for i := 0; i < config.NumClusters; i++ {
		if i != num {
			cport := 3000 + i
			otherClients = append(otherClients, config.CBase+":"+strconv.Itoa(cport))
		}
	}
	return otherClients
}

// CollectOtherAddrs requests other clients for their addresses
// and set them to DB
func CollectOtherAddrs(others []string) {
	db := config.GetDB()

	for _, ourl := range others {
		url := ourl + "/address"
		body := utils.Get(url)

		response := struct {
			Message string          `json:"message"`
			Addrs   []model.Address `json:"addresses"`
		}{}
		err := json.Unmarshal(body, &response)
		if err != nil {
			log.Panicln(err)
		}
		log.Printf("Message: %s\n", response.Message)
		log.Printf("Addrs: %s\n", response.Addrs)

		cluster := model.Cluster{URL: ourl}
		db.Create(&cluster)

		addrs := response.Addrs
		for _, addr := range addrs {
			client := model.Client{Address: addr}
			db.Model(&cluster).Association("Clients").Append(client)
		}
	}
}

// GetAllAddrs returns all addresses including mine from DB
func GetAllAddrs() []model.Address {
	db := config.GetDB()

	clients := []model.Client{}
	db.Find(&clients)
	addrs := make([]model.Address, len(clients))
	for i, client := range clients {
		addrs[i] = client.Address
	}
	return addrs
}
