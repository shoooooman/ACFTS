package main

import (
	"acfts/api"
	"acfts/config"
	"fmt"
	"log"
	"strconv"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	fmt.Printf("Input port number: ")
	var port int
	fmt.Scan(&port)

	db := config.InitDB(port)
	defer db.Close()

	go func() {
		log.Println(http.ListenAndServe("localhost:8000", nil))
	}()

	r := api.SetRouter(db)
	r.Run(":" + strconv.Itoa(port))
}
