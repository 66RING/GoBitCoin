package main

import (
	"fmt"
	"log"
)

func (cli *CLI) listAddress(nodeID string) {
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	for _, addr := range addresses {
		fmt.Println(addr)
	}
}
