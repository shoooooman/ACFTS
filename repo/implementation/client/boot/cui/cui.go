package cui

import (
	"acfts-client/boot"
	"acfts-client/config"
	"acfts-client/model"
	"acfts-client/transaction"
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

	/* Use generalTx */

	// Inside one cluster
	atxs := []transaction.GeneralTx{
		{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{200}},
	}
	transaction.Execute(atxs)

	// atxs := make([]transaction.GeneralTx, 200)
	// tx := transaction.GeneralTx{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{1}}
	// for i := 0; i < 200; i++ {
	// 	atxs[i] = tx
	// }
	// transaction.Execute(atxs)

	// Between two clusters
	// atxs := []transaction.GeneralTx{
	// 	{From: addrs[0], To: []model.Address{addrs[4]}, Amounts: []int{200}},
	// }
	// transaction.Execute(atxs)

	// atxs := make([]transaction.GeneralTx, 200)
	// tx := transaction.GeneralTx{From: addrs[0], To: []model.Address{addrs[4]}, Amounts: []int{1}}
	// for i := 0; i < 200; i++ {
	// 	atxs[i] = tx
	// }
	// transaction.Execute(atxs)

	/* Use insideTx */

	// case 1: 0 -> 1
	// itxs := []transaction.InsideTx{
	// 	{From: 0, To: []int{1}, Amounts: []int{200}},
	// }
	// atxs := convertInsideTxs(itxs)
	// transaction.Execute(atxs)

	// case 2: 0 <-> 1
	// tx1 := transaction.InsideTx{From: 0, To: []int{1}, Amounts: []int{200}}
	// tx2 := transaction.InsideTx{From: 1, To: []int{0}, Amounts: []int{200}}
	// itxs := []transaction.InsideTx{}
	// for i := 0; i < 10; i++ {
	// 	itxs = append(itxs, tx1)
	// 	itxs = append(itxs, tx2)
	// }
	// atxs := convertInsideTxs(itxs)
	// transaction.Execute(atxs)

	// case 3: 0 <-> 1 & 2 <-> 3
	// tx0 := transaction.InsideTx{From: 0, To: []int{0, 2}, Amounts: []int{50, 150}}
	// itxs0 := []transaction.InsideTx{}
	// itxs0 = append(itxs0, tx0)
	// atxs0 := convertInsideTxs(itxs0)
	// transaction.Execute(atxs0)
	//
	// tx1 := transaction.InsideTx{From: 0, To: []int{1}, Amounts: []int{50}}
	// tx2 := transaction.InsideTx{From: 1, To: []int{0}, Amounts: []int{50}}
	// tx3 := transaction.InsideTx{From: 2, To: []int{3}, Amounts: []int{150}}
	// tx4 := transaction.InsideTx{From: 3, To: []int{2}, Amounts: []int{150}}
	// itxs1 := []transaction.InsideTx{}
	// for i := 0; i < 10; i++ {
	// 	itxs1 = append(itxs1, tx1)
	// 	itxs1 = append(itxs1, tx2)
	// 	itxs1 = append(itxs1, tx3)
	// 	itxs1 = append(itxs1, tx4)
	// }
	// atxs1 := convertInsideTxs(itxs1)
	// transaction.Execute(atxs1)

	// case 4: random -> random
	// mrand.Seed(time.Now().UnixNano())
	// tx0 := transaction.InsideTx{From: 0, To: []int{0, 1, 2, 3}, Amounts: []int{50, 50, 50, 50}}
	// itxs0 := []transaction.InsideTx{}
	// itxs0 = append(itxs0, tx0)
	// atxs0 := convertInsideTxs(itxs0)
	// transaction.Execute(atxs0)
	//
	// itxs1 := []transaction.InsideTx{}
	// for i := 0; i < 25; i++ {
	// 	from := mrand.Intn(4)
	// 	to := mrand.Intn(4)
	// 	tx1 := transaction.InsideTx{From: from, To: []int{to}, Amounts: []int{1}}
	// 	itxs1 = append(itxs1, tx1)
	// }
	// atxs1 := convertInsideTxs(itxs1)
	// transaction.Execute(atxs1)

	// Wait for receiving outputs of this cluster
	// for {
	// }
}
