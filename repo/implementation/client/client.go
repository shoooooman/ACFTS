package main

import (
	"acfts-client/boot/cui"
	"acfts-client/boot/gui"
	"acfts-client/config"
	"log"
	"net/http"

	_ "net/http/pprof"
)

func main() {
	/* For pprof */
	go func() {
		log.Println(http.ListenAndServe("localhost:7000", nil))
	}()

	if config.IsGUI {
		gui.InitGUI()
	} else {
		cui.InitCUI()
	}
}
