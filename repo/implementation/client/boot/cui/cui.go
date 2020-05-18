package cui

import (
	"acfts-client/boot"
	"acfts-client/config"
	"fmt"
	"log"
	"strconv"
)

// InitCUI is
func InitCUI() {
	fmt.Printf("Input cluster number: ")
	fmt.Scan(&config.Num)

	config.DB = boot.SetDB(config.Num)
	defer config.DB.Close()

	boot.DeleteAll(config.DB)

	port := config.BasePort + config.Num

	// Generate private keys
	config.NumClients = 4
	myurl := config.CBase + ":" + strconv.Itoa(port)
	boot.GenerateClients(myurl)

	r := boot.SetRouter(config.DB, nil)
	go r.Run(":" + strconv.Itoa(port))

	config.NumClusters = 1
	otherClients := boot.GetOtherCURLs(config.Num)

	// 他のクライアントが設定情報を入れるのを待機
	fmt.Println("Input something when all clusters have been registered by all servers.")
	var dummy string
	fmt.Scan(&dummy)

	boot.CollectOtherAddrs(otherClients)

	addrs := boot.GetAllAddrs()
	log.Println(addrs)

	/* Make the genesis */
L_FOR:
	for {
		fmt.Println("Have a genesis?")
		fmt.Println("1. yes")
		fmt.Println("2. no")
		var cmd int
		fmt.Scan(&cmd)
		switch cmd {
		case 1:
			owner := addrs[0]
			boot.CreateGenesis(owner, 200)
			break L_FOR
		case 2:
			break L_FOR
		default:
			fmt.Println("Please input valid number.")
		}
	}
}
