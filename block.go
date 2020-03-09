package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// genesis block
func NewGenesisBlock() *Block {
	return NewBlock("The First Blcok: Travis Turing", []byte{})
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
