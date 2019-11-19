package main

import (
	"acfts-client/model"
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
)

func post(url, jsonStr string) {
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		panic(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
		return
	}

	response := model.Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("%#v\n", response)
}

type signature struct {
	r *big.Int
	s *big.Int
}

func getSig(address *ecdsa.PublicKey) signature {
	// Convert addresses to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v%v", address.X, address.Y))

	// Get hash using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	// Get signature using ellipse curve cryptography
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed)
	if err != nil {
		panic(err)
	}
	sig := signature{r, s}

	return sig
}

func createInputStr(addresses []*ecdsa.PublicKey) string {
	sigs := make([]signature, len(addresses))
	inputs := ""
	for i, address := range addresses {
		sigs[i] = getSig(address)
		inputStr := `
		{
			"utxo": {
				"address1": "` + address.X.String() + `",
				"address2": "` + address.Y.String() + `"
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

var privateKey *ecdsa.PrivateKey

func main() {
	url := "http://localhost:8080/transaction"

	// Generate a private key
	// Need to generate every time you make a transaction
	var err error
	privateKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// Get a public key (address)
	publicKey := &privateKey.PublicKey
	utxos := []*ecdsa.PublicKey{publicKey}
	inputs := createInputStr(utxos)

	jsonStr := `
{
	"inputs": ` + inputs + `,
	"outputs": [
		{
			"amount": 150,
			"address1": "foofoo1",
			"address2": "foofoo2"
		},
		{
			"amount": 50,
			"address1": "barbar1",
			"address2": "barbar2"
		}
	]
}`

	fmt.Println(jsonStr)

	// FIXME: To get addresses to insert them into db as a genesis transaction
	var dummy string
	fmt.Scan(&dummy)

	post(url, jsonStr)
}
