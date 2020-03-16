package main

import "fmt"

func (cli *CLI) reindexUTXOP(nodeID string) {
	blockchain := NewBlockChain(nodeID)
	UTXOSet := UTXOSet{blockchain}
	UTXOSet.Reindex()
	count := UTXOSet.CountTransaction()
	fmt.Printf("Have %d transaction in UTXO Set", count)
}
