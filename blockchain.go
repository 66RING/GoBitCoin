package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blockBucket = "blocks"

type BlockChain struct {
	tip []byte // address of block
	db  *bolt.DB
}

type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// Add a block
func (block *BlockChain) AddBlock(data string) {
	var lastHash []byte
	err := block.db.View(func(tx *bolt.Tx) error {
		block := tx.Bucket([]byte(blockBucket)) // get data
		lastHash = block.Get([]byte("1"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(data, lastHash) // New block
	err = block.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))               // take out
		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) // comprass into data
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("1"), newBlock.Hash) // comprass into data
		if err != nil {
			log.Panic(err)
		}
		block.tip = newBlock.Hash

		return nil
	})
}

// An Iterator
func (block *BlockChain) Iterator() *BlockChainIterator {
	bcit := &BlockChainIterator{block.tip, block.db}
	return bcit // create interator
}

// get next block
func (it *BlockChainIterator) next() *Block {
	var block *Block
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		encodedBlock := bucket.Get(it.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	it.currentHash = block.PrevBlockHash
	return block
}

func NewBlockChain() *BlockChain {
	var tip []byte                          //
	db, err := bolt.Open(dbFile, 0600, nil) //open db
	if err != nil {
		log.Panic(err)
	}
	// handle update
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			fmt.Println("Null BlockChain here")
			genesis := NewGenesisBlock()
			bucket, err := tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put([]byte("1"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			tip = bucket.Get([]byte("1"))
		}
		return nil

	})
	if err != nil {
		log.Panic(err)
	}
	bc := BlockChain{tip, db}
	return &bc
}
