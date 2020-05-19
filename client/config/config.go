package config

import (
	"crypto/ecdsa"

	"github.com/jinzhu/gorm"
)

const (
	// IsGUI should be true when using GUI
	IsGUI = true

	// NumServers is the number of servers
	NumServers = 2

	// BasePort is
	BasePort = 3000
	// CBase is
	CBase = "http://localhost"
)

var (
	// ServerURLs is URLs of servers
	ServerURLs = []string{
		"http://localhost:8080",
		"http://localhost:8081",
		// "http://localhost:8082",
		// "http://localhost:8083",
	}

	db *gorm.DB

	// Keys is
	// FIXME: DBに秘密鍵と公開鍵を保存する
	Keys []*ecdsa.PrivateKey

	// Pub2Pri is for getting a private key from a public key (PublicKey.X + PublicKey.Y)
	Pub2Pri map[string]*ecdsa.PrivateKey

	// Num is a ID among clients
	Num int
	// NumClients is the number of clients
	NumClients = 4
	// NumClusters is the number of clusters
	NumClusters = 2
	// HasGenesis is true if this cluster has the genesis
	HasGenesis bool
	// GAmount is the amount of the genesis
	GAmount = 200
)
