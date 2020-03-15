package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockChain(address string) {
	if !ValidateAddress(address) {
		log.Panic("Address Invalid")
	}
	bc := CreateBlockChain(address)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc} // creata a UTXOSet
	UTXOSet.Reindex()
	fmt.Println("Create SUCCESS")
}
