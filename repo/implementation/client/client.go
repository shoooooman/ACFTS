package main

import (
	"acfts-client/model"
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func initDB() *gorm.DB {
	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/acfts_client?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Output{})

	return db
}

func post(url, jsonStr string) model.Response {
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	response := model.Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	return response
}

type signature struct {
	r *big.Int
	s *big.Int
}

func getSig(utxo model.Output) signature {
	// Convert addresses to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v%v%v", utxo.Address1, utxo.Address2, utxo.PreviousHash))

	// Get hash using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	// Get signature using ellipse curve cryptography
	pub := utxo.Address1 + utxo.Address2
	r, s, err := ecdsa.Sign(rand.Reader, pub2Pri[pub], hashed)
	if err != nil {
		panic(err)
	}
	sig := signature{r, s}

	return sig
}

func createInputStr(utxos []model.Output) string {
	sigs := make([]signature, len(utxos))
	inputs := ""
	for i, utxo := range utxos {
		sigs[i] = getSig(utxo)
		inputStr := `
		{
			"utxo": {
				"address1": "` + utxo.Address1 + `",
				"address2": "` + utxo.Address2 + `",
				"previous_hash": "` + utxo.PreviousHash + `"
			},
			"sig1": "` + sigs[i].r.String() + `",
			"sig2": "` + sigs[i].s.String() + `"
		},`

		inputs += inputStr
	}
	// Remove the last ','
	inputs = "[" + inputs[:len(inputs)-1] + "]"

	return inputs
}

func getPreviousHash(previous string) string {
	bytes := sha256.Sum256([]byte(previous))
	num := fmt.Sprintf("%x", bytes)
	return string(num)
}

// Get a private key from a public key (PublicKey.X + PublicKey.Y)
var pub2Pri map[string]*ecdsa.PrivateKey

func main() {
	url := "http://localhost:8080/transaction"
	db := initDB()
	defer db.Close()

	// Generate private keys
	const numClients = 3
	keys := make([]*ecdsa.PrivateKey, numClients)
	pub2Pri = make(map[string]*ecdsa.PrivateKey)
	for i := 0; i < numClients; i++ {
		var err error
		keys[i], err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			panic(err)
		}
		pub := &keys[i].PublicKey
		pub2Pri[pub.X.String()+pub.Y.String()] = keys[i]
	}

	// Make transactions
	publicKey := &keys[0].PublicKey
	utxo1 := model.Output{Address1: publicKey.X.String(), Address2: publicKey.Y.String(), PreviousHash: "genesis"}
	utxos := []model.Output{utxo1}
	inputs := createInputStr(utxos)
	hash := getPreviousHash(inputs)

	jsonStr := `
{
	"inputs": ` + inputs + `,
	"outputs": [
		{
			"amount": 150,
			"address1": "foofoo1",
			"address2": "foofoo2",
			"previous_hash": "` + hash + `"
		},
		{
			"amount": 50,
			"address1": "barbar1",
			"address2": "barbar2",
			"previous_hash": "` + hash + `"
		}
	]
}`

	fmt.Println(jsonStr)

	// FIXME: To get addresses to insert them into db as a genesis transaction
	var dummy string
	fmt.Scan(&dummy)

	resp := post(url, jsonStr)
	fmt.Println(resp)
}
