package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

func (block *Block) HashTransactions() []byte {
	var transacions [][]byte
	for _, tx := range block.Transactions {
		transacions = append(transacions, tx.Serialize())
	}
	mTree := NewMerkleTree(transacions)
	return mTree.RootNode.data
}

func NewBlock(transacions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(),
		transacions,
		prevBlockHash,
		[]byte{},
		0, height}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// genesis block
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

// Object to binary byte set, and write into file
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encode := gob.NewEncoder(&result) // new encoder object
	err := encode.Encode(block)       //  encode
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// Read file, binary byte set to object
func DeserializeBlock(data []byte) *Block {
	var block Block                                  // object to save the object that byte turn to
	decoder := gob.NewDecoder(bytes.NewReader(data)) // decode
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
