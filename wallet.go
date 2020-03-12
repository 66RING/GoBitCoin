package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletfile = "wallet.dat"
const addressChecksumlen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public} // new a wallet
	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	publickey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, publickey
}

// Check public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	SecondSHA := sha256.Sum256(firstSHA[:])
	return SecondSHA[:addressChecksumlen]
}

// hash of public key
func HashPubkey(pubkey []byte) []byte {
	publicsha256 := sha256.Sum256(pubkey)
	R160Hash := ripemd160.New()
	_, err := R160Hash.Write(publicsha256[:]) // write and tackle
	if err != nil {
		log.Panic(err)
	}
	publicR160Hash := R160Hash.Sum(nil)
	return publicR160Hash
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubkey(w.PublicKey)
	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload) // check verion and public key
	fullpayload := append(versionPayload, checksum...)
	address := Base58Encode(fullpayload)
	return address
}

func ValidateAddress(address string) bool {
	publicHash := Base58Decode([]byte(address))
	actualchecksum := publicHash[len(publicHash)-addressChecksumlen:]
	version := publicHash[0]
	publicHash = publicHash[1 : len(publicHash)-addressChecksumlen]
	targetCheckSum := checksum(append([]byte{version}, publicHash...))

	return bytes.Compare(actualchecksum, targetCheckSum) == 0
}
