package main

import (
	"acfts-client/model"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

// delete all data in DB
func deleteAll(db *gorm.DB) {
	output := model.Output{}
	signature := model.Signature{}
	db.Unscoped().Delete(&output)
	db.Unscoped().Delete(&signature)
}

func benchmarkSetup(serverURLs []string, numClients int, genesisOwner int) []model.Address {
	db = initDB(0)
	// db = initDB(3)

	// Delete all data in a client
	deleteAll(db)

	// Delete all data in servers
	for _, serverURL := range serverURLs {
		url := serverURL + "/all"
		body := _delete(url)
		log.Printf("Delete: %v\n", string(body))
	}

	// Generate private keys
	generateClients(numClients)

	setupWs(serverURLs)

	addrs := getAddrs(serverURLs[0])
	log.Println(addrs)

	// Make a genesis transaction
	owner := addrs[genesisOwner]
	createGenesis(serverURLs, owner, 1000000)

	return addrs
}

// Scenario1: 0 -> 1
func BenchmarkScenario1(b *testing.B) {
	serverURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	addrs := benchmarkSetup(serverURLs, 4, 0)
	defer db.Close()

	atxs1 := []generalTx{
		{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{1}},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		executeTxs(serverURLs, atxs1)
	}
}

// Scenario2: 0 <-> 1
func BenchmarkScenario2(b *testing.B) {
	serverURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	addrs := benchmarkSetup(serverURLs, 4, 0)
	defer db.Close()

	atxs1 := []generalTx{
		// {From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{1000000}},
		// {From: addrs[1], To: []model.Address{addrs[0]}, Amounts: []int{1000000}},
		{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{1}},
		{From: addrs[1], To: []model.Address{addrs[0]}, Amounts: []int{1}},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		executeTxs(serverURLs, atxs1)
	}
}

// Scenario3: 0 <-> 1, 2 <-> 3
func BenchmarkScenario3(b *testing.B) {
	serverURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	addrs := benchmarkSetup(serverURLs, 4, 0)
	defer db.Close()

	atx := generalTx{
		From:    addrs[0],
		To:      []model.Address{addrs[0], addrs[2]},
		Amounts: []int{500000, 500000},
	}
	executeTxs(serverURLs, []generalTx{atx})

	atxs1 := []generalTx{
		{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{500000}},
		{From: addrs[1], To: []model.Address{addrs[0]}, Amounts: []int{500000}},
		{From: addrs[2], To: []model.Address{addrs[3]}, Amounts: []int{500000}},
		{From: addrs[3], To: []model.Address{addrs[2]}, Amounts: []int{500000}},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		executeTxs(serverURLs, atxs1)
	}
}

// Scenario4: random -> random
func BenchmarkScenario4(b *testing.B) {
	serverURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	addrs := benchmarkSetup(serverURLs, 4, 0)
	defer db.Close()

	atx := generalTx{
		From:    addrs[0],
		To:      []model.Address{addrs[0], addrs[1], addrs[2], addrs[3]},
		Amounts: []int{250000, 250000, 250000, 250000},
	}
	executeTxs(serverURLs, []generalTx{atx})

	rand.Seed(time.Now().UnixNano())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		from := addrs[rand.Intn(4)]
		to := addrs[rand.Intn(4)]
		atxs1 := []generalTx{
			{From: from, To: []model.Address{to}, Amounts: []int{1}},
		}
		b.StartTimer()

		executeTxs(serverURLs, atxs1)
	}
}

// Scenario5: 4 -> 0
// func BenchmarkScenario5(b *testing.B) {
// 	serverURLs := []string{
// 		// "http://localhost:8080",
// 		// "http://localhost:8081",
// 		"http://localhost:8082",
// 		"http://localhost:8083",
// 	}
//
// 	addrs := benchmarkSetup(serverURLs, 4, 4)
// 	defer db.Close()
//
// 	atxs1 := []generalTx{
// 		{From: addrs[4], To: []model.Address{addrs[0]}, Amounts: []int{1}},
// 	}
//
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		executeTxs(serverURLs, atxs1)
// 	}
// }
