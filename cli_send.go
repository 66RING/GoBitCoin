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
	bc := NewBlockChain()
	defer bc.db.Close()
	UTXOSet := UTXOSet{bc}

	tx := NewUTXOTransaction(from, to, amount, &UTXOSet)
	cbTx := NewCoinBaseTx(from, "")
	txs := []*Transaction{cbTx, tx}

	newBlock := bc.MineBLock(txs)
	UTXOSet.Update(newBlock)

	fmt.Println("Transaction Success!")
}
