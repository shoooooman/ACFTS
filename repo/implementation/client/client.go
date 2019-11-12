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

func getSig(id string) signature {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// Convert transaction struct to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v", id))

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

func createInputStr(ids []string) string {
	sigs := make([]signature, len(ids))
	inputs := ""
	for i, id := range ids {
		sigs[i] = getSig(id)
		inputStr := `
		{
			"utxo": {
				"id": "` + id + `"
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

func main() {
	url := "http://localhost:8080/transaction"

	// Generate input strings
	utxoIDs := []string{"genesis"}
	inputs := createInputStr(utxoIDs)

	jsonStr := `
{
	"inputs": ` + inputs + `,
	"outputs": [
		{
			"amount": 150,
			"address": "hogehoge"
		},
		{
			"amount": 50,
			"address": "foofoo"
		}
	]
}`

	post(url, jsonStr)
}
