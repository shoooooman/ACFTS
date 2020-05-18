package gui

import (
	"acfts-client/boot"
	"acfts-client/boot/gui/event"
	"acfts-client/config"
	"log"
	"strconv"

	"github.com/Equanox/gotron"
)

// InitGUI is
func InitGUI() {
	window, err := gotron.New("./boot/gui/webapp")
	if err != nil {
		panic(err)
	}

	window.WindowOptions.Width = windowWidth
	window.WindowOptions.Height = windowHeight
	window.WindowOptions.Title = windowTitle

	done, err := window.Start()
	if err != nil {
		panic(err)
	}

	if openDev {
		window.OpenDevTools()
	}

	// set handlers
	c := make(chan bool)
	window.On(&gotron.Event{Event: "config"}, event.HandleConfig(c))

	rdy := make(chan bool)
	window.On(&gotron.Event{Event: "ready"}, event.HandleReady(rdy))

	br := make(chan bool)
	window.On(&gotron.Event{Event: "begin-req"}, event.HandleBegin(br))

	// wait for finishing config page
	<-c

	db := config.SetDB(config.Num)
	defer db.Close()

	port := config.BasePort + config.Num
	myurl := config.CBase + ":" + strconv.Itoa(port)
	boot.GenerateClients(myurl)

	r := boot.SetRouter(db, window)
	go r.Run(":" + strconv.Itoa(port))

	otherClients := boot.GetOtherCURLs(config.Num)

	// 他のクライアントが設定情報を入れるのを待機
	<-rdy

	boot.CollectOtherAddrs(otherClients)

	<-br
	addrs := boot.GetAllAddrs()
	log.Println(addrs)

	event.SendAddress(window, addrs)

	if config.HasGenesis {
		owner := addrs[0]
		boot.CreateGenesis(owner, config.GAmount)
	}

	event.SendBalance(window, db, addrs)

	window.On(&gotron.Event{Event: "request"}, event.HandleRequest(window, db, addrs))

	<-done
}
