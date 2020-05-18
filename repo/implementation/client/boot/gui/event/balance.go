package event

import (
	"acfts-client/config"
	"acfts-client/model"

	"github.com/Equanox/gotron"
	"github.com/jinzhu/gorm"
)

// SendBalance sends the balance to frontent
func SendBalance(window *gotron.BrowserWindow, db *gorm.DB, addrs []model.Address) {
	sum := 0
	balances := make([]int, config.NumClients)
	for i := 0; i < config.NumClients; i++ {
		balances[i] = getBalance(db, addrs[i])
		sum += balances[i]
	}
	bal := struct {
		*gotron.Event
		Total    int   `json:"total"`
		Balances []int `json:"balances"`
	}{
		Event:    &gotron.Event{Event: "balance"},
		Total:    sum,
		Balances: balances,
	}
	window.Send(&bal)
}
