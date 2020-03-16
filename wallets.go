package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat"

type Wallets struct {
	Wallets map[string]*Wallet // a string for a wallet
}

// New a wallet or get a wallet
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeID)
	return &wallets, err
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet // save wallet
	return address
}

// get all wallets
func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses // return all addresses
}

// get a wallet
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// load wallet from file
func (ws *Wallets) LoadFromFile(nodeID string) error {
	mywalletfile := fmt.Sprintf(walletFile, nodeID) // gen file address
	if _, err := os.Stat(mywalletfile); os.IsNotExist(err) {
		return err
	}
	fileContent, err := ioutil.ReadFile(mywalletfile)
	if err != nil {
		log.Panic(err)
	}
	// load file binary and parse
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}
	ws.Wallets = wallets.Wallets

	return nil
}

// save wallet to file
func (ws *Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	mywalletfile := fmt.Sprintf(walletFile, nodeID) // gen file address
	gob.Register(elliptic.P256())                   // registe a crypto
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(mywalletfile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}

}
