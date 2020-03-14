package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int) {

	if !ValidateAddress(from) {
		log.Panic("Invalid address (from)")
	}

	if !ValidateAddress(to) {
		log.Panic("Invalid address (to)")
	}
	bc := NewBlockChain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBLock([]*Transaction{tx})
	fmt.Println("Transaction Success!")
}
