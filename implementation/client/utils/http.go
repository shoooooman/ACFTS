package utils

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

// Get is
func Get(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panicln(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	return body
}

// Post posts request asynchronously
func Post(url, jsonStr string, b chan []byte) {
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		log.Panicln(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	b <- body
}

// Delete deletes all data in database
func Delete(url string) []byte {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Panicln(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	return body
}
