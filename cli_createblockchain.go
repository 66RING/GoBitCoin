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
	bc.db.Close()
	fmt.Println("Create SUCCESS")
}
