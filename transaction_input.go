package main

import "bytes"

type TXInput struct {
	Txid      []byte // Store id of transaction
	Vout      int
	Signature []byte
	PubKey    []byte
}

// check address and transaction
func (in *TXInput) UsesKeyHash(pubKeyHash []byte) bool {
	lockinghash := HashPubkey(in.PubKey)
	return bytes.Compare(lockinghash, pubKeyHash) == 0
}
