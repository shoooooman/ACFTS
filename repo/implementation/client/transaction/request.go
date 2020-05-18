package transaction

import (
	"acfts-client/config"
	"acfts-client/model"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"strconv"
)

// InsideTx is
// From, To: Number of clients
type InsideTx struct {
	From    int
	To      []int
	Amounts []int
}

// GeneralTx is
type GeneralTx struct {
	From    model.Address
	To      []model.Address
	Amounts []int
}

// ConvertInsideTxs converts InsideTxs to GeneralTxs
func ConvertInsideTxs(its []InsideTx) []GeneralTx {
	ats := make([]GeneralTx, len(its))
	for i, it := range its {
		fromPri := config.Keys[it.From]
		fromPub := &fromPri.PublicKey
		from1 := fromPub.X.String()
		from2 := fromPub.Y.String()
		from := model.Address{from1, from2}

		to := make([]model.Address, len(it.To))
		for j, v := range it.To {
			toPri := config.Keys[v]
			toPub := &toPri.PublicKey
			t1 := toPub.X.String()
			t2 := toPub.Y.String()
			t := model.Address{t1, t2}
			to[j] = t
		}

		at := GeneralTx{
			From:    from,
			To:      to,
			Amounts: it.Amounts,
		}
		ats[i] = at
	}

	return ats
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
	pri, ok := config.Pub2Pri[pub]
	if !ok {
		log.Panicf("A public key %s is not in this cluster.\n", pub)
	}
	r, s, err := ecdsa.Sign(rand.Reader, pri, hashed)
	if err != nil {
		log.Panicln(err)
	}

	return r.String(), s.String()
}

func getServerSigs(utxo model.Output) string {
	db := config.DB
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
	if len(str) > 0 {
		str = "[" + str[:len(str)-1] + "]"
	} else {
		str = "[]"
	}

	return str
}

func getSiblings(utxo model.Output) string {
	db := config.DB
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
					"server_signatures": []
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

func createSigStr(sigs []model.Signature) string {
	str := ""
	for _, sig := range sigs {
		s := `
					{
						"address1": "` + sig.Address1 + `",
						"address2": "` + sig.Address2 + `",
						"signature1": "` + sig.Signature1 + `",
						"signature2": "` + sig.Signature2 + `"
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

func createOutputStr(ops []model.Output, hash string, sigs []model.Signature) string {
	outputs := ""
	for _, op := range ops {
		outputStr := `
		{
			"amount": ` + strconv.Itoa(op.Amount) + `,
			"address1": "` + op.Address1 + `",
			"address2": "` + op.Address2 + `",
			"previous_hash": "` + hash + `",
			"index": ` + strconv.Itoa(int(op.Index)) + `,
			"server_signatures": ` + createSigStr(sigs) + `
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

func findUTXOs(addr model.Address, amount int) ([]model.Output, int) {
	db := config.DB
	var candidates []model.Output
	address1 := addr.Address1
	address2 := addr.Address2

	// TODO: signatures > 2/3のUTXOをSQLで見つける
	db.Where("address1 = ? AND address2 = ? AND used = false", address1, address2).
		Find(&candidates)

	// Select outputs which have enough signatures
	utxos := make([]model.Output, 0)
	for _, candidate := range candidates {
		count := 0
		db.Table("signatures").Where("output_id = ?", candidate.ID).Count(&count)
		if float64(count) >= 2.0*float64(config.NumServers)/3.0 {
			utxos = append(utxos, candidate)
		}
	}
	if len(utxos) == 0 {
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
			return utxos, sum
		}
	}

	return nil, 0
}

func createJSONStr(tx GeneralTx) (string, error) {
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
	outputs := createOutputStr(ops, hash, nil)

	jsonStr := `
{
	"inputs": ` + inputs + `,
	"outputs": ` + outputs + `
}`

	return jsonStr, nil
}
