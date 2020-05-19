package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
)

const (
	// NumServers is the number of servers
	NumServers = 2
)

var (
	// key is a private key of this server
	key *ecdsa.PrivateKey
)

// GetKey returns PrivateKey (including PublicKey)
func GetKey() *ecdsa.PrivateKey {
	if key == nil {
		log.Fatal("key is not set")
	}
	return key
}

func init() {
	var err error
	key, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
}
