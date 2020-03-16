package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {

	if !ValidateAddress(from) {
		log.Panic("Invalid address (from)")
	}

	if !ValidateAddress(to) {
		log.Panic("Invalid address (to)")
	}
	bc := NewBlockChain(nodeID)
	defer bc.db.Close()
	UTXOSet := UTXOSet{bc}

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := NewCoinBaseTx(from, "")
		txs := []*Transaction{cbTx, tx}
		newBlock := bc.MineBLock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knowNodes[0], tx)
	}

	fmt.Println("Transaction Success!")
}
