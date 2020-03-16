package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("Invalid address")
	}
	bc := NewBlockChain(nodeID)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	balance := 0
	pubkeyhash := Base58Decode([]byte(address))
	pubkeyhash = pubkeyhash[1 : len(pubkeyhash)-4]

	UTXOs := UTXOSet.FindUTXO(pubkeyhash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance %s: %d \n", address, balance)
}
