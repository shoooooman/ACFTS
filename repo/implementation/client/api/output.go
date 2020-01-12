package api

import (
	"acfts-client/model"
	"net/http"

	"github.com/Equanox/gotron"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func getBalance(db *gorm.DB, addr model.Address) int {
	var balance int
	db.Table("outputs").
		Where("address1 = ? AND address2 = ? AND used = false", addr.Address1, addr.Address2).
		Select("sum(amount)").Row().Scan(&balance)
	return balance
}

func getTotalBalance(db *gorm.DB, addrs []model.Address, numClients int) int {
	sum := 0
	for i := 0; i < numClients; i++ {
		sum += getBalance(db, addrs[i])
	}
	return sum
}

func getAllAddrs(db *gorm.DB) []model.Address {
	clients := []model.Client{}
	db.Find(&clients)
	addrs := make([]model.Address, len(clients))
	for i, client := range clients {
		addrs[i] = client.Address
	}

	return addrs
}

// ReceiveUTXO is
func ReceiveUTXO(db *gorm.DB, window *gotron.BrowserWindow, numClients int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// log.Println("received utxos")
		j := struct {
			UTXOs []model.Output `json:"outputs"`
		}{}
		c.BindJSON(&j)

		// bytes, _ := json.MarshalIndent(j, "", "    ")
		// log.Println("utxos")
		// log.Println(string(bytes))

		// FIXME: 1つ1つのUTXOの署名を検証する
		for _, utxo := range j.UTXOs {
			db.Create(&utxo)
		}

		addrs := getAllAddrs(db)
		sum := getTotalBalance(db, addrs, numClients)
		balances := make([]int, numClients)
		for i := 0; i < numClients; i++ {
			balances[i] = getBalance(db, addrs[i])
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

		c.JSON(http.StatusCreated, gin.H{
			"message": "Received a utxo.",
			"utxos":   j.UTXOs,
		})
	}
}

// ClearOutputs is
func ClearOutputs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		output := model.Output{}
		db.DropTable(&output)
		db.AutoMigrate(&output)

		c.JSON(http.StatusOK, gin.H{
			"message": "all data is deleted.",
		})
	}
}
