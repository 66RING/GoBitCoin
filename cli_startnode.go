package main

import (
	"fmt"
	"log"
)

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("START ONE NODE %s \n", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			log.Printf("Mining... Address:%s \n", minerAddress)
		} else {
			log.Panic("Invalid mind Address")
		}
	}

	StartServer(nodeID, minerAddress)

}
