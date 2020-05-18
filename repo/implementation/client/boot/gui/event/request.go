package event

import (
	"acfts-client/config"
	"acfts-client/model"
	"acfts-client/transaction"
	"bytes"
	"encoding/json"
	"log"

	"github.com/Equanox/gotron"
	"github.com/jinzhu/gorm"
)

// HandleRequest is called when receiving "request" message from frontend
func HandleRequest(window *gotron.BrowserWindow, db *gorm.DB, addrs []model.Address) func([]byte) {
	return func(bin []byte) {
		buf := bytes.NewBuffer(bin)
		log.Println(buf)

		req := struct {
			From   int `json:"from"`
			To     int `json:"to"`
			Amount int `json:"coin"`
		}{}
		if err := json.Unmarshal(bin, &req); err != nil {
			log.Fatal(err)
		}

		from := addrs[req.From]
		to := addrs[req.To]
		amount := req.Amount
		atxs := []transaction.GeneralTx{
			{From: from, To: []model.Address{to}, Amounts: []int{amount}},
		}
		transaction.Execute(atxs)

		sum := 0
		balances := make([]int, config.NumClients)
		for i := 0; i < config.NumClients; i++ {
			balances[i] = getBalance(db, addrs[i])
			sum += balances[i]
		}
		b := struct {
			*gotron.Event
			Total    int   `json:"total"`
			Balances []int `json:"balances"`
		}{
			Event:    &gotron.Event{Event: "balance"},
			Total:    sum,
			Balances: balances,
		}
		window.Send(&b)
	}
}
