package event

import (
	"acfts-client/config"
	"acfts-client/model"

	"github.com/Equanox/gotron"
)

// SendAddress sends addresses to frontend
func SendAddress(window *gotron.BrowserWindow, addrs []model.Address) {
	data := struct {
		*gotron.Event
		Addrs   []model.Address `json:"addresses"`
		MyAddrs int             `json:"myAddresses"`
	}{
		Event:   &gotron.Event{Event: "addrs"},
		Addrs:   addrs,
		MyAddrs: config.NumClients,
	}
	window.Send(&data)
}
