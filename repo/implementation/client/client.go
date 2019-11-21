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
	"strconv"

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

func post(url, jsonStr string) []byte {
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

	return body
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
	inputs := ""
	for _, utxo := range utxos {
		sig := getSig(utxo)
		//  FIXME: 使おうとしているutxoの兄弟のtxもサーバーに送る必要がある
		// utxoと並列でsiblingsのような要素を入れる
		inputStr := `
		{
			"utxo": {
				"address1": "` + utxo.Address1 + `",
				"address2": "` + utxo.Address2 + `",
				"previous_hash": "` + utxo.PreviousHash + `"
			},
			"sig1": "` + sig.r.String() + `",
			"sig2": "` + sig.s.String() + `"
		},`

		inputs += inputStr
	}
	// Remove the last ','
	inputs = "[" + inputs[:len(inputs)-1] + "]"

	return inputs
}

func createOutputStr(ops []model.Output, hash string) string {
	outputs := ""
	for _, op := range ops {
		outputStr := `
		{
			"amount": ` + strconv.Itoa(op.Amount) + `,
			"address1": "` + op.Address1 + `",
			"address2": "` + op.Address2 + `",
			"previous_hash": "` + hash + `"
		},`

		outputs += outputStr
	}
	// Remove the last ','
	outputs = "[" + outputs[:len(outputs)-1] + "]"

	return outputs
}

func getPreviousHash(previous string) string {
	bytes := sha256.Sum256([]byte(previous))
	num := fmt.Sprintf("%x", bytes)
	return string(num)
}

type simpleTx struct {
	// clientの番号(シミュレーション用, 0から)
	From    int
	To      []int
	Amounts []int
}

func findUTXOs(publicKey *ecdsa.PublicKey, amount int) ([]model.Output, bool) {
	var utxos []model.Output
	count := 0
	address1 := publicKey.X.String()
	address2 := publicKey.Y.String()
	// TODO: public keyをキーとしてDBからused = false && signatures > 2/3のUTXOを見つける
	db.Where("address1 = ? AND address2 = ? AND used = false", address1, address2).Find(&utxos).Count(&count)
	if count == 0 {
		return nil, false
	}

	// amount以上になるようにtxを集める
	sum := 0
	for i, utxo := range utxos {
		sum += utxo.Amount
		db.Model(&utxo).Where("address1 = ? AND address2 = ? AND used = false", address1, address2).Update("used", true)
		// FIXME: >= ではなく==にする
		if sum >= amount {
			utxos = utxos[:i+1]
			return utxos, true
		}
	}

	fmt.Println("There are no enough utxos.")
	return nil, false
}

func createJSONStr(tx simpleTx) string {
	priKey := keys[tx.From]
	pubKey := &priKey.PublicKey
	sum := 0
	for _, val := range tx.Amounts {
		sum += val
	}
	utxos, exists := findUTXOs(pubKey, sum)
	if !exists {
		// TODO: このclient番号に紐づいている有効なUTXOがなかった場合
		// トランザクションの作成を保留にすべき？
		fmt.Printf("There are no valid utxos of client %v\n", tx.From)
		return ""
	}
	inputs := createInputStr(utxos)
	hash := getPreviousHash(inputs)

	ops := make([]model.Output, len(tx.To))
	for i, index := range tx.To {
		priKey := keys[index]
		pubKey := &priKey.PublicKey
		ops[i] = model.Output{Amount: tx.Amounts[i], Address1: pubKey.X.String(), Address2: pubKey.Y.String(), PreviousHash: hash}
	}
	outputs := createOutputStr(ops, hash)

	jsonStr := `
{
	"inputs": ` + inputs + `,
	"outputs": ` + outputs + `
}`

	return jsonStr
}

type genesisJSON struct {
	Message string       `json:"message"`
	Genesis model.Output `json:"genesis"`
}

func createGenesis(owner int, amount int) {
	priKey := keys[owner]
	pubKey := &priKey.PublicKey
	genesis := model.Output{
		Amount:       amount,
		Address1:     pubKey.X.String(),
		Address2:     pubKey.Y.String(),
		PreviousHash: "genesis",
		Used:         false,
	}
	db.Create(&genesis)

	jsonStr := `
	{
		"amount": ` + strconv.Itoa(amount) + `,
		"address1": "` + pubKey.X.String() + `",
		"address2": "` + pubKey.Y.String() + `"
	}`

	// Create a genesis record in a server
	// FIXME: 複数のサーバーにリクエストを送る
	url := "http://localhost:8080/genesis"
	body := post(url, jsonStr)
	var g genesisJSON
	err := json.Unmarshal(body, &g)
	if err != nil {
		panic(err)
	}
	fmt.Println(g)
}

func generateClients(num int) {
	keys = make([]*ecdsa.PrivateKey, num)
	pub2Pri = make(map[string]*ecdsa.PrivateKey)
	for i := 0; i < num; i++ {
		var err error
		keys[i], err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			panic(err)
		}
		pub := &keys[i].PublicKey
		pub2Pri[pub.X.String()+pub.Y.String()] = keys[i]
	}
}

var db *gorm.DB

var keys []*ecdsa.PrivateKey

// Get a private key from a public key (PublicKey.X + PublicKey.Y)
var pub2Pri map[string]*ecdsa.PrivateKey

func main() {
	url := "http://localhost:8080/transaction"
	db = initDB()
	defer db.Close()

	// Generate private keys
	const numClients = 3
	generateClients(numClients)

	// Make a genesis transaction
	createGenesis(0, 200)

	// Make sample transactions
	txs := []simpleTx{
		// {From: 0, To: []int{1}, Amounts: []int{200}},
		{From: 0, To: []int{1, 2}, Amounts: []int{150, 50}},
		{From: 1, To: []int{2}, Amounts: []int{150}},
		{From: 2, To: []int{0}, Amounts: []int{200}},
	}

	for i := 0; i < len(txs); i++ {
		jsonStr := createJSONStr(txs[i])
		fmt.Println(jsonStr)

		// FIXME: 全てのサーバーに送るようにし，2/3以上からの応答を待つ
		body := post(url, jsonStr)
		response := model.Response{}
		err := json.Unmarshal(body, &response)
		if err != nil {
			panic(err)
		}
		fmt.Println(response)

		// Make records of new outputs
		outputs := response.Transaction.Outputs
		for _, output := range outputs {
			db.Create(&output)
		}
	}
}
