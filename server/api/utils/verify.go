package utils

import (
	"acfts/config"
	"acfts/db/model"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"log"
	"math/big"
)

// UnlockUTXO unlocks the utxo using the signature
func UnlockUTXO(utxo model.Output, signature1, signature2 string) bool {
	// Convert transaction struct to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v%v%v", utxo.Address1, utxo.Address2, utxo.PreviousHash))

	// Get a hash value using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	// Convert string to big.Int
	var ok bool
	n1 := new(big.Int)
	n1, ok = n1.SetString(signature1, 10)
	if !ok {
		fmt.Println("Signature1 is not valid.")
		return false
	}

	n2 := new(big.Int)
	n2, ok = n2.SetString(signature2, 10)
	if !ok {
		fmt.Println("Signature2 is not valid.")
		return false
	}

	// model.Output.Address1 and Address2 represent a public key
	// Convert Address1 and Address2 to ecdsa.PublicKey
	address1, ok1 := new(big.Int).SetString(utxo.Address1, 10)
	address2, ok2 := new(big.Int).SetString(utxo.Address2, 10)
	if !ok1 || !ok2 {
		log.Println("Error: an address of a server is not valid.")
		return false
	}
	publicKey := ecdsa.PublicKey{elliptic.P521(), address1, address2}

	if ecdsa.Verify(&publicKey, hashed, n1, n2) {
		fmt.Println("Verifyed!")
		return true
	}
	return false
}

// VerifyUTXO verifies the utxo if it has enough valid server signatures
func VerifyUTXO(utxo model.Output, siblings []model.Output) bool {
	// Genesis is approved without verification
	if utxo.PreviousHash == "genesis" {
		fmt.Println("genesis is approved without verification.")
		return true
	}

	// Create the same array as when creating a signature
	outputs := make([]SimpleOutput, 1+len(siblings))
	for i := range outputs {
		if i < int(utxo.Index) {
			outputs[i] = ConvertOutput(siblings[i])
		} else if i > int(utxo.Index) {
			outputs[i] = ConvertOutput(siblings[i-1])
		} else {
			outputs[i] = ConvertOutput(utxo)
		}
	}

	// Convert transaction struct to bytes to get its hash
	buf := []byte(fmt.Sprintf("%v", outputs))

	// Get a hash value using SHA256
	h := crypto.Hash.New(crypto.SHA256)
	h.Write(buf)
	hashed := h.Sum(nil)

	fmt.Println("hash when verification")
	fmt.Println(hashed)

	valid := 0
	for _, signature := range utxo.Signatures {
		// FIXME: Should get public keys of other servers independently of clients
		address1, ok1 := new(big.Int).SetString(signature.Address1, 10)
		address2, ok2 := new(big.Int).SetString(signature.Address2, 10)
		if !ok1 || !ok2 {
			log.Println("VerifyUTXO: an address of a server is not valid")
			return false
		}
		serverPubKey := ecdsa.PublicKey{elliptic.P521(), address1, address2}
		signature1, ok1 := new(big.Int).SetString(signature.Signature1, 10)
		signature2, ok2 := new(big.Int).SetString(signature.Signature2, 10)
		if !ok1 || !ok2 {
			log.Println("VerifyUTXO: a signature of server is not valid")
			return false
		}
		if ecdsa.Verify(&serverPubKey, hashed, signature1, signature2) {
			valid++
			fmt.Println("one valid signature")
			if float64(valid) >= 2.0*float64(config.NumServers)/3.0 {
				fmt.Println("Server Verifyed!")
				return true
			}
		} else {
			fmt.Println("not valid")
		}
	}
	return false
}
