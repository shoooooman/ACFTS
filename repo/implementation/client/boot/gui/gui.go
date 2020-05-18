package gui

import (
	"acfts-client/boot"
	"acfts-client/config"
	"acfts-client/model"
	"acfts-client/transaction"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/Equanox/gotron"
	"github.com/jinzhu/gorm"
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

	resp := struct {
		Cluster int  `json:"cluster"`
		Client  int  `json:"address"`
		All     int  `json:"all"`
		Genesis bool `json:"genesis"`
		GAmount int  `json:"gamount"`
	}{}

	c := make(chan bool)
	window.On(&gotron.Event{Event: "config"}, func(bin []byte) {
		buf := bytes.NewBuffer(bin)
		fmt.Println(buf)
		err := json.Unmarshal(bin, &resp)
		if err != nil {
			fmt.Println(err)
		}
		config.Num = resp.Cluster
		config.NumClients = resp.Client
		config.NumClusters = resp.All
		config.HasGenesis = resp.Genesis
		if config.HasGenesis {
			config.GAmount = resp.GAmount
		}
		c <- true
	})

	rdy := make(chan bool)
	window.On(&gotron.Event{Event: "ready"}, func(bin []byte) {
		rdy <- true
	})

	br := make(chan bool)
	window.On(&gotron.Event{Event: "begin-req"}, func(bin []byte) {
		br <- true
	})

	<-c
	log.Println(config.Num)

	config.DB = boot.SetDB(config.Num)
	defer config.DB.Close()

	boot.DeleteAll(config.DB)

	port := config.BasePort + config.Num

	myurl := config.CBase + ":" + strconv.Itoa(port)
	boot.GenerateClients(myurl)

	r := boot.SetRouter(config.DB, window)
	go r.Run(":" + strconv.Itoa(port))

	otherClients := boot.GetOtherCURLs(config.Num)

	// 他のクライアントが設定情報を入れるのを待機
	<-rdy

	boot.CollectOtherAddrs(otherClients)

	<-br
	addrs := boot.GetAllAddrs()
	log.Println(addrs)

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

	/* Make the genesis */
	if config.HasGenesis {
		owner := addrs[0]
		boot.CreateGenesis(owner, config.GAmount)
	}

	sum := getTotalBalance(addrs)
	balances := make([]int, config.NumClients)
	for i := 0; i < config.NumClients; i++ {
		balances[i] = getBalance(config.DB, addrs[i])
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

	/* Make sample transactions */
	// Create a custom event struct that has a pointer to gotron.Event
	type CustomEvent struct {
		*gotron.Event
		CustomAttribute string `json:"AtrNameInFrontend"`
	}

	req := struct {
		From   int `json:"from"`
		To     int `json:"to"`
		Amount int `json:"coin"`
	}{}

	window.On(&gotron.Event{Event: "request"}, func(bin []byte) {
		buf := bytes.NewBuffer(bin)
		fmt.Println(buf)
		err := json.Unmarshal(bin, &req)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(req)

		from := addrs[req.From]
		to := addrs[req.To]
		amount := req.Amount
		atxs := []transaction.GeneralTx{
			{From: from, To: []model.Address{to}, Amounts: []int{amount}},
		}
		transaction.Execute(atxs)

		sum := getTotalBalance(addrs)
		balances := make([]int, config.NumClients)
		for i := 0; i < config.NumClients; i++ {
			balances[i] = getBalance(config.DB, addrs[i])
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
	})

	<-done
}

func getBalance(db *gorm.DB, addr model.Address) int {
	var balance int
	db.Table("outputs").
		Where("address1 = ? AND address2 = ? AND used = false", addr.Address1, addr.Address2).
		Select("sum(amount)").Row().Scan(&balance)
	return balance
}

func getTotalBalance(addrs []model.Address) int {
	sum := 0
	for i := 0; i < config.NumClients; i++ {
		sum += getBalance(config.DB, addrs[i])
	}
	return sum
}
