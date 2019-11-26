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
	db.AutoMigrate(&model.Signature{})

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

func getClientSig(utxo model.Output) signature {
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

func getServerSigs(utxo model.Output) string {
	var signatures []model.Signature
	db.Table("signatures").Where("output_id = ?", utxo.ID).Find(&signatures)

	str := ""
	for _, signature := range signatures {
		s := `
					{
						"address1": "` + signature.Address1 + `",
						"address2": "` + signature.Address2 + `",
						"signature1": "` + signature.Signature1 + `",
						"signature2": "` + signature.Signature2 + `"
					},`
		str += s
	}
	// Remove the last ','
	str = "[" + str[:len(str)-1] + "]"

	return str
}

func getSiblings(utxo model.Output) string {
	var siblings []model.Output
	db.Where("id <> ? AND previous_hash = ?", utxo.ID, utxo.PreviousHash).Find(&siblings)

	str := ""
	for _, sibling := range siblings {
		// used and signatures are initial values because it is when they are signed.
		s := `
				{
					"amount": ` + strconv.Itoa(sibling.Amount) + `,
					"address1": "` + sibling.Address1 + `",
					"address2": "` + sibling.Address2 + `",
					"previous_hash": "` + sibling.PreviousHash + `",
					"used": false,
					"signatures": []
				},`
		str += s
	}
	// Remove the last ','
	if len(str) > 0 {
		str = "[" + str[:len(str)-1] + "]"
	} else {
		str = "[]"
	}

	return str
}

func createInputStr(utxos []model.Output) string {
	inputs := ""
	for _, utxo := range utxos {
		sig := getClientSig(utxo)
		inputStr := `
		{
			"utxo": {
				"address1": "` + utxo.Address1 + `",
				"address2": "` + utxo.Address2 + `",
				"previous_hash": "` + utxo.PreviousHash + `",
				"server_signatures": ` + getServerSigs(utxo) + `
			},
			"siblings": ` + getSiblings(utxo) + `,
			"signature1": "` + sig.r.String() + `",
			"signature2": "` + sig.s.String() + `"
		},`

		inputs += inputStr
	}
	// Remove the last ','
	inputs = "[" + inputs[:len(inputs)-1] + "]"

	return inputs
}

func createOutputStr(ops []model.Output, hash string) string {
	outputs := ""
	for i, op := range ops {
		outputStr := `
		{
			"amount": ` + strconv.Itoa(op.Amount) + `,
			"address1": "` + op.Address1 + `",
			"address2": "` + op.Address2 + `",
			"previous_hash": "` + hash + `",
			"index": ` + strconv.Itoa(i) + `
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

// From, To: Number of clients
type simpleTx struct {
	From    int
	To      []int
	Amounts []int
}

func findUTXOs(publicKey *ecdsa.PublicKey, amount int) ([]model.Output, bool) {
	var candidates []model.Output
	address1 := publicKey.X.String()
	address2 := publicKey.Y.String()
	// TODO: signatures > 2/3のUTXOをSQLで見つける
	db.Where("address1 = ? AND address2 = ? AND used = false", address1, address2).
		Find(&candidates)

	// Select outputs which have enough signatures
	utxos := make([]model.Output, 0)
	for _, candidate := range candidates {
		count := 0
		db.Table("signatures").Where("output_id = ?", candidate.ID).Count(&count)
		if float64(count) >= 2.0*float64(n)/3.0 {
			utxos = append(utxos, candidate)
		}
	}
	if len(utxos) == 0 {
		return nil, false
	}

	// Collect transactions so that the amount of them becomes equal to the arguement
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

func createGenesis(urls []string, owner int, amount int) {
	priKey := keys[owner]
	pubKey := &priKey.PublicKey

	// Dummy signatures for genesis
	sig := model.Signature{
		Address1:   "gene",
		Address2:   "sis",
		Signature1: "dum",
		Signature2: "my",
		OutputID:   1,
	}
	sigs := make([]model.Signature, n)
	for i := 0; i < n; i++ {
		sigs[i] = sig
	}

	genesis := model.Output{
		Amount:       amount,
		Address1:     pubKey.X.String(),
		Address2:     pubKey.Y.String(),
		PreviousHash: "genesis",
		Index:        0,
		Used:         false,
		Signatures:   sigs,
	}
	db.Create(&genesis)

	jsonStr := `
	{
		"amount": ` + strconv.Itoa(amount) + `,
		"address1": "` + pubKey.X.String() + `",
		"address2": "` + pubKey.Y.String() + `"
	}`

	// Create genesis records in servers
	for _, baseURL := range urls {
		url := baseURL + "/genesis"
		body := post(url, jsonStr)
		var g genesisJSON
		err := json.Unmarshal(body, &g)
		if err != nil {
			panic(err)
		}
	}
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

func executeTxs(baseURLs []string, txs []simpleTx, async bool, finished chan bool) {
	for i := 0; i < len(txs); i++ {
		jsonStr := createJSONStr(txs[i])
		fmt.Println(jsonStr)

		// FIXME: 全てのサーバーに送るようにし，2/3以上からの応答を待つ
		// FIXME: リクエストを送ってレスポンスを受け取る部分は並行処理の方がいい？
		for j, baseURL := range baseURLs {
			url := baseURL + "/transaction"
			body := post(url, jsonStr)

			response := model.Response{}
			err := json.Unmarshal(body, &response)
			if err != nil {
				panic(err)
			}
			fmt.Println(response)

			outputs := response.Transaction.Outputs
			sigs := model.Signature{
				Address1:   response.Address1,
				Address2:   response.Address2,
				Signature1: response.Signature1.String(),
				Signature2: response.Signature2.String(),
			}
			for _, output := range outputs {
				// Make a record if the output is not created yet
				if j == 0 {
					db.Create(&output)
				} else {
					db.Where("address1 = ? AND address2 = ? AND previous_hash = ?",
						output.Address1, output.Address2, output.PreviousHash).
						First(&output)
				}
				// Add a signature of server i
				sigs.OutputID = output.ID
				db.Model(&output).Association("Signatures").Append(sigs)
			}
		}
	}
	if async {
		finished <- true
	}
}

var db *gorm.DB

var keys []*ecdsa.PrivateKey

// Get a private key from a public key (PublicKey.X + PublicKey.Y)
var pub2Pri map[string]*ecdsa.PrivateKey

// n is the number of servers
const n = 4

func main() {
	baseURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}
	db = initDB()
	defer db.Close()

	// Generate private keys
	const numClients = 4
	generateClients(numClients)

	// Make a genesis transaction
	createGenesis(baseURLs, 0, 200)

	// Make sample transactions
	// case 1
	txs1 := []simpleTx{
		{From: 0, To: []int{1}, Amounts: []int{200}},
	}
	executeTxs(baseURLs, txs1, false, nil)

	// case 2
	// tx1 := simpleTx{From: 0, To: []int{1}, Amounts: []int{200}}
	// tx2 := simpleTx{From: 1, To: []int{0}, Amounts: []int{200}}
	// txs1 := []simpleTx{}
	// for i := 0; i < 10; i++ {
	// 	txs1 = append(txs1, tx1)
	// 	txs1 = append(txs1, tx2)
	// }
	// executeTxs(baseURLs, txs1, false, nil)

	// case 3
	// finished := make(chan bool)
	//
	// tx0 := simpleTx{From: 0, To: []int{0, 2}, Amounts: []int{50, 150}}
	// txs0 := []simpleTx{}
	// txs0 = append(txs0, tx0)
	// executeTxs(baseURLs, txs0, false, nil)
	//
	// tx1 := simpleTx{From: 0, To: []int{1}, Amounts: []int{50}}
	// tx2 := simpleTx{From: 1, To: []int{0}, Amounts: []int{50}}
	// tx3 := simpleTx{From: 2, To: []int{3}, Amounts: []int{150}}
	// tx4 := simpleTx{From: 3, To: []int{2}, Amounts: []int{150}}
	// txs1 := []simpleTx{}
	// txs2 := []simpleTx{}
	// for i := 0; i < 10; i++ {
	// 	txs1 = append(txs1, tx1)
	// 	txs1 = append(txs1, tx2)
	// 	txs2 = append(txs2, tx3)
	// 	txs2 = append(txs2, tx4)
	// }
	//
	// go executeTxs(baseURLs, txs1, true, finished)
	// go executeTxs(baseURLs, txs2, true, finished)
	// <-finished
	// <-finished
}
