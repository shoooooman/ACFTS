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
	"log"
	mrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func initDB(num int) *gorm.DB {
	setting := fmt.Sprintf("root:@tcp(127.0.0.1:3306)/acfts_client_%d?charset=utf8&parseTime=True&loc=Local", num)
	db, err := gorm.Open("mysql", setting)
	if err != nil {
		log.Println("failed to connect database")
		log.Panicln(err)
	}

	db.AutoMigrate(&model.Output{})
	db.AutoMigrate(&model.Signature{})

	return db
}

func get(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panicln(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	return body
}

func post(url, jsonStr string) []byte {
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		log.Panicln(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	return body
}

func getClientSig(utxo model.Output) (string, string) {
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
		log.Panicln(err)
	}

	return r.String(), s.String()
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
					"index": ` + strconv.Itoa(int(sibling.Index)) + `,
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
		sig1, sig2 := getClientSig(utxo)
		inputStr := `
		{
			"utxo": {
				"address1": "` + utxo.Address1 + `",
				"address2": "` + utxo.Address2 + `",
				"previous_hash": "` + utxo.PreviousHash + `",
				"index": ` + strconv.Itoa(int(utxo.Index)) + `,
				"server_signatures": ` + getServerSigs(utxo) + `
			},
			"siblings": ` + getSiblings(utxo) + `,
			"signature1": "` + sig1 + `",
			"signature2": "` + sig2 + `"
		},`

		inputs += inputStr
	}
	// Remove the last ','
	if len(inputs) > 0 {
		inputs = "[" + inputs[:len(inputs)-1] + "]"
	} else {
		inputs = "[]"
	}

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
			"previous_hash": "` + hash + `",
			"index": ` + strconv.Itoa(int(op.Index)) + `
		},`

		outputs += outputStr
	}
	// Remove the last ','
	if len(outputs) > 0 {
		outputs = "[" + outputs[:len(outputs)-1] + "]"
	} else {
		outputs = "[]"
	}

	return outputs
}

func getPreviousHash(previous string) string {
	bytes := sha256.Sum256([]byte(previous))
	num := fmt.Sprintf("%x", bytes)
	return string(num)
}

// From, To: Number of clients
type insideTx struct {
	From    int
	To      []int
	Amounts []int
}

type generalTx struct {
	From    model.Address
	To      []model.Address
	Amounts []int
}

func convertInsideTxs(its []insideTx) []generalTx {
	ats := make([]generalTx, len(its))
	for i, it := range its {
		fromPri := keys[it.From]
		fromPub := &fromPri.PublicKey
		from1 := fromPub.X.String()
		from2 := fromPub.Y.String()
		from := model.Address{from1, from2}

		to := make([]model.Address, len(it.To))
		for j, v := range it.To {
			toPri := keys[v]
			toPub := &toPri.PublicKey
			t1 := toPub.X.String()
			t2 := toPub.Y.String()
			t := model.Address{t1, t2}
			to[j] = t
		}

		at := generalTx{
			From:    from,
			To:      to,
			Amounts: it.Amounts,
		}
		ats[i] = at
	}

	return ats
}

func findUTXOs(addr model.Address, amount int) ([]model.Output, int) {
	var candidates []model.Output
	address1 := addr.Address1
	address2 := addr.Address2

	// For mutual exclusion of db
	c <- true

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
		<-c
		return nil, 0
	}

	// Collect transactions so that the amount of them becomes equal to the arguement
	// FIXME: utxosをsortした方がいいかもしれない
	sum := 0
	for i, utxo := range utxos {
		sum += utxo.Amount
		if sum >= amount {
			utxos = utxos[:i+1]
			for _, utxo := range utxos {
				// FIXME: serverにこのUTXOが拒否されたら整合性が失われる
				db.Model(&utxo).Update("used", true)
			}
			<-c
			return utxos, sum
		}
	}

	<-c
	return nil, 0
}

func createJSONStr(tx generalTx) (string, error) {
	pubKey := tx.From
	want := 0
	for _, val := range tx.Amounts {
		want += val
	}
	utxos, sum := findUTXOs(pubKey, want)
	if sum == 0 {
		// TODO: このclient番号に紐づいている有効なUTXOがなかった場合
		// トランザクションの作成を保留にすべき？
		return "", fmt.Errorf("Error: there are no enough utxos of client %v", tx.From)
	}
	inputs := createInputStr(utxos)
	hash := getPreviousHash(inputs)

	ops := make([]model.Output, len(tx.To))
	for i, addr := range tx.To {
		ops[i] = model.Output{
			Amount: tx.Amounts[i],
			Address: model.Address{
				Address1: addr.Address1,
				Address2: addr.Address2,
			},
			PreviousHash: hash,
			Index:        uint(i),
		}
	}
	// Create a change transaction if the sum of utxos exceeds want
	if sum > want {
		change := model.Output{
			Amount: sum - want,
			Address: model.Address{
				Address1: tx.From.Address1,
				Address2: tx.From.Address2,
			},
			PreviousHash: hash,
			Index:        uint(len(ops)),
		}
		ops = append(ops, change)
	}
	outputs := createOutputStr(ops, hash)

	jsonStr := `
{
	"inputs": ` + inputs + `,
	"outputs": ` + outputs + `
}`

	return jsonStr, nil
}

// FIXME: この構造体いらない？
type genesisJSON struct {
	Message string       `json:"message"`
	Genesis model.Output `json:"genesis"`
}

func createGenesis(urls []string, owner model.Address, amount int) {
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
	sigs := make([]model.Signature, n)
	for i := 0; i < n; i++ {
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
	for _, baseURL := range urls {
		url := baseURL + "/genesis"
		body := post(url, jsonStr)
		var g genesisJSON
		err := json.Unmarshal(body, &g)
		if err != nil {
			log.Panicln(err)
		}
		log.Println(g.Message)
	}
}

func generateClients(num int) {
	keys = make([]*ecdsa.PrivateKey, num)
	pub2Pri = make(map[string]*ecdsa.PrivateKey)
	for i := 0; i < num; i++ {
		var err error
		keys[i], err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			log.Panicln(err)
		}
		pub := &keys[i].PublicKey
		pub2Pri[pub.X.String()+pub.Y.String()] = keys[i]
	}
}

// FIXME: サーバー側もstruct Addressを作ったらaddr1, addr2をまとめる
func updateOutputs(outputs []model.Output, addr1, addr2, sig1, sig2 string) {
	sigs := model.Signature{
		Address: model.Address{
			Address1: addr1,
			Address2: addr2,
		},
		Signature1: sig1,
		Signature2: sig2,
	}
	for _, output := range outputs {
		// Make a record if the output is not created yet
		c := 0
		db.Where("address1 = ? AND address2 = ? AND previous_hash = ? AND output_index = ?",
			output.Address1, output.Address2, output.PreviousHash, output.Index).
			First(&output).Count(&c)
		if c == 0 {
			db.Create(&output)
		}
		sigs.OutputID = output.ID
		db.Model(&output).Association("Signatures").Append(sigs)
	}
}

func executeTxs(baseURLs []string, txs []generalTx) {
	for i := 0; i < len(txs); i++ {
		fmt.Printf("tx %d is executing\n", i)
		jsonStr, err := createJSONStr(txs[i])
		if err != nil {
			log.Println(err)
			return
		}
		// fmt.Println(jsonStr)

		// FIXME: リクエストを送ってレスポンスを受け取る部分は並行処理の方がいい？
		for _, baseURL := range baseURLs {
			url := baseURL + "/transaction"
			body := post(url, jsonStr)
			// sendWs(baseURL, jsonStr)

			response := model.Response{}
			err := json.Unmarshal(body, &response)
			if err != nil {
				log.Panicln(err)
			}
			fmt.Println("Message: " + response.Message)
		}

		// FIXME: 全てのサーバーから署名が来るのを待つ
		time.Sleep(time.Millisecond * 250)

		// FIXME: 現実的にはずっと待ち続ける(実際はどれだけ待てばいいのか分からないため)
		// for j := 0; j < n; j++ {
		// 	<-received
		// }
	}
}

var conns = make(map[string]*websocket.Conn)

// FIXME: unused now
func sendWs(url, jsonStr string) {
	url = strings.Replace(url, "http", "ws", -1) + "/ws"
	conn := conns[url]
	transaction := model.Transaction{}
	err := json.Unmarshal([]byte(jsonStr), &transaction)
	if err != nil {
		log.Panicln(err)
	}
	err = conn.WriteJSON(transaction)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func setupWs(baseURLs []string) {
	for _, baseURL := range baseURLs {
		url := strings.Replace(baseURL, "http", "ws", -1) + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		conns[url] = conn

		// Send the addresses of this cluster to a server
		addrs := make([]model.Address, len(keys))
		for i, key := range keys {
			pub := &key.PublicKey
			addrs[i].Address1 = pub.X.String()
			addrs[i].Address2 = pub.Y.String()
		}
		conn.WriteJSON(addrs)

		resp := model.Response{}
		err = conn.ReadJSON(&resp)
		if err != nil {
			log.Println("cluster:", err)
			return
		}
		addr := conn.RemoteAddr().String()
		log.Printf("recv in %s: %s", addr, resp.Message)

		// Receive transactions whose outputs include addresses of this cluster
		go func() {
			for {
				response := model.Response{}
				err := conn.ReadJSON(&response)
				if err != nil {
					log.Println("read:", err)
					return
				}

				outputs := response.Transaction.Outputs
				addr1 := response.Address1
				addr2 := response.Address2
				sig1 := response.Signature1.String()
				sig2 := response.Signature2.String()
				updateOutputs(outputs, addr1, addr2, sig1, sig2)

				addr := conn.RemoteAddr().String()
				log.Printf("recv in %s:\n%v\n", addr, response)

				// received <- true
			}
		}()
	}
}

type respAddrs struct {
	Message string          `json:"message"`
	Addrs   []model.Address `json:"addresses"`
}

// Get all addresses of all clusters (including this cluster)
func getAddrs(baseURL string) []model.Address {
	url := baseURL + "/address"
	body := get(url)

	response := respAddrs{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Panicln(err)
	}
	log.Printf("Message: %s\n", response.Message)

	return response.Addrs
}

var db *gorm.DB

// FIXME: DBに秘密鍵と公開鍵を保存する
var keys []*ecdsa.PrivateKey

// Get a private key from a public key (PublicKey.X + PublicKey.Y)
var pub2Pri map[string]*ecdsa.PrivateKey

// n is the number of servers
const n = 2

// For mutual exclusion of db
var c = make(chan bool, 1)

// For synchronization of communication with servers
var received = make(chan bool, 200)

func main() {
	baseURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		// "http://localhost:8082",
		// "http://localhost:8083",
	}

	fmt.Printf("Input client number: ")
	var num int
	fmt.Scan(&num)
	db = initDB(num)
	defer db.Close()

	// Generate private keys
	const numClients = 4
	generateClients(numClients)

	setupWs(baseURLs)

	fmt.Println("Have all clusters been registered by servers?")
	var dummy string
	fmt.Scan(&dummy)

	addrs := getAddrs(baseURLs[0])
	log.Println(addrs)

	// Make a genesis transaction
L_FOR:
	for {
		fmt.Println("Have a genesis?")
		fmt.Println("1. yes")
		fmt.Println("2. no")
		var g int
		fmt.Scan(&g)
		switch g {
		case 1:
			owner := addrs[0]
			createGenesis(baseURLs, owner, 200)
			break L_FOR
		case 2:
			break L_FOR
		default:
			fmt.Println("Please input valid number.")
		}
	}

	// Can make transactions between different clusters with generalTxs
	// one cluster
	// atxs1 := []generalTx{
	// 	{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{200}},
	// }

	// two clusters
	// atxs1 := []generalTx{
	// 	{From: addrs[0], To: []model.Address{addrs[4]}, Amounts: []int{200}},
	// 	// {From: addrs[4], To: []model.Address{addrs[0]}, Amounts: []int{200}},
	// }

	// two clusters
	atxs1 := []generalTx{
		{From: addrs[0], To: []model.Address{addrs[0], addrs[4]}, Amounts: []int{100, 100}},
	}
	for i := 0; i < 10; i++ {
		tx := generalTx{From: addrs[0], To: []model.Address{addrs[4]}, Amounts: []int{10}}
		// tx := generalTx{From: addrs[4], To: []model.Address{addrs[0]}, Amounts: []int{10}}
		atxs1 = append(atxs1, tx)
	}

	executeTxs(baseURLs, atxs1)

	// Make sample transactions
	// case 1: 0 -> 1
	// itxs1 := []insideTx{
	// 	{From: 0, To: []int{1}, Amounts: []int{200}},
	// }
	// atxs1 := convertInsideTxs(itxs1)
	// executeTxs(baseURLs, atxs1)

	// case 2: 0 <-> 1
	// tx1 := insideTx{From: 0, To: []int{1}, Amounts: []int{200}}
	// tx2 := insideTx{From: 1, To: []int{0}, Amounts: []int{200}}
	// itxs1 := []insideTx{}
	// for i := 0; i < 10; i++ {
	// 	itxs1 = append(itxs1, tx1)
	// 	itxs1 = append(itxs1, tx2)
	// }
	// atxs1 := convertInsideTxs(itxs1)
	// executeTxs(baseURLs, atxs1)

	// case 3: 0 <-> 1 & 2 <-> 3 (parallelly)
	// tx0 := insideTx{From: 0, To: []int{0, 2}, Amounts: []int{50, 150}}
	// itxs0 := []insideTx{}
	// itxs0 = append(itxs0, tx0)
	// atxs0 := convertInsideTxs(itxs0)
	// executeTxs(baseURLs, atxs0)
	//
	// tx1 := insideTx{From: 0, To: []int{1}, Amounts: []int{50}}
	// tx2 := insideTx{From: 1, To: []int{0}, Amounts: []int{50}}
	// itxs1 := []insideTx{}
	// for i := 0; i < 10; i++ {
	// 	itxs1 = append(itxs1, tx1)
	// 	itxs1 = append(itxs1, tx2)
	// }
	// atxs1 := convertInsideTxs(itxs1)
	// executeTxs(baseURLs, atxs1)
	//
	// tx3 := insideTx{From: 2, To: []int{3}, Amounts: []int{150}}
	// tx4 := insideTx{From: 3, To: []int{2}, Amounts: []int{150}}
	// itxs2 := []insideTx{}
	// for i := 0; i < 10; i++ {
	// 	itxs2 = append(itxs2, tx3)
	// 	itxs2 = append(itxs2, tx4)
	// }
	// atxs2 := convertInsideTxs(itxs2)
	// executeTxs(baseURLs, atxs2)

	// case 4: random -> random
	mrand.Seed(time.Now().UnixNano())
	// tx0 := insideTx{From: 0, To: []int{0, 1, 2, 3}, Amounts: []int{50, 50, 50, 50}}
	// itxs0 := []insideTx{}
	// itxs0 = append(itxs0, tx0)
	// atxs0 := convertInsideTxs(itxs0)
	// executeTxs(baseURLs, atxs0)
	//
	// itxs1 := []insideTx{}
	// for i := 0; i < 25; i++ {
	// 	from := mrand.Intn(4)
	// 	to := mrand.Intn(4)
	// 	tx1 := insideTx{From: from, To: []int{to}, Amounts: []int{1}}
	// 	itxs1 = append(itxs1, tx1)
	// }
	// atxs1 := convertInsideTxs(itxs1)
	// executeTxs(baseURLs, atxs1)

	// Wait for receiving outputs of this cluster
	for {
	}
}
