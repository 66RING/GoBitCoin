package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) showBlockChian() {
	bc := NewBlockChain("")
	defer bc.db.Close()

	bci := bc.Iterator()
	for {
		block := bci.next()
		fmt.Printf("PrevBlockHash: %x", block.PrevBlockHash)
		fmt.Printf("Current Hash: %x", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s \n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
