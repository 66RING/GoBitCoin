package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("Invalid address")
	}
	bc := NewBlockChain(address)
	defer bc.db.Close()
	balance := 0
	pubkeyhash := Base58Decode([]byte(address))
	pubkeyhash = pubkeyhash[1 : len(pubkeyhash)-4]
	//UTXOs := bc.FindUTXO(pubkeyhash) // Find balace
	//for _, out := range UTXOs {
	//	balance += out.Value
	//}
	fmt.Printf("Balance %s: %d \n", address, balance)
}
