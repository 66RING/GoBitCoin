package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blockBucket = "blocks"
const genesisCoinbaseData = "###THE GENESIS BLOCK###"

type BlockChain struct {
	tip []byte // address of block
	db  *bolt.DB
}

type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (blockchain *BlockChain) MineBLock(transactions []*Transaction) {
	var lastHash []byte
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) // show data
		lastHash = bucket.Get([]byte("1"))       // get last block
		return nil
	})
	if err != nil {

		log.Panic(err)
	}
	newBlock := NewBlock(transactions, lastHash) // create a new block
	err = blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))               // get index
		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) // store into database
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("1"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		blockchain.tip = newBlock.Hash // store last hash
		return nil
	})
}

// get all unspent transaction
func (blockchain *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := blockchain.FindUnspendTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// get unspent transaction list
func (blockchain *BlockChain) FindUnspendTransactions(address string) []Transaction {
	var unspentTXs []Transaction // all transaction
	spentTXOS := make(map[string][]int)
	bci := blockchain.Iterator()
	for {
		block := bci.next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outindex, out := range tx.Vout {
				if spentTXOS[txID] != nil {
					for _, spentOut := range spentTXOS[txID] {
						if spentOut == outindex {
							continue Outputs // for until unequalt
						}
					}
				}
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutPutWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOS[inTxID] = append(spentTXOS[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// get all unspent output to know what to input
func (blockchain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := blockchain.FindUnspendTransactions(address)
	accmulated := 0
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outindex, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accmulated < amount {
				accmulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outindex)
				if accmulated >= amount {
					break Work
				}
			}
		}
	}

	return accmulated, unspentOutputs
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//// Add a block
//func (block *BlockChain) AddBlock(transaction []*Transaction) {
//	var lastHash []byte
//	err := block.db.View(func(tx *bolt.Tx) error {
//		block := tx.Bucket([]byte(blockBucket)) // get data
//		lastHash = block.Get([]byte("1"))
//		return nil
//	})
//	if err != nil {
//		log.Panic(err)
//	}
//	newBlock := NewBlock(transaction, lastHash) // New block
//	err = block.db.Update(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte(blockBucket))               // take out
//		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) // comprass into data
//		if err != nil {
//			log.Panic(err)
//		}
//		err = bucket.Put([]byte("1"), newBlock.Hash) // comprass into data
//		if err != nil {
//			log.Panic(err)
//		}
//		block.tip = newBlock.Hash
//
//		return nil
//	})
//}

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

func NewBlockChain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("Database do not exist")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil) //open db
	if err != nil {
		log.Panic(err)
	}

	// handle update
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		tip = bucket.Get([]byte("1"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := BlockChain{tip, db}
	return &bc
}

func createBlockChain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("Database exists there is not need to create")
		os.Exit(1)
	}

	var tip []byte                          //
	db, err := bolt.Open(dbFile, 0600, nil) //open db
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinBaseTx(address, genesisCoinbaseData) // New genesis block
		genesis := NewGenesisBlock(cbtx)
		bucket, err := tx.CreateBucket([]byte(blockBucket))
		if err != nil {
			log.Panic(err)
		}

		err = bucket.Put(genesis.Hash, genesis.Serialize()) // store into database
		if err != nil {
			log.Panic(err)
		}

		err = bucket.Put([]byte("1"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash
		return nil
	})

	bc := BlockChain{tip, db}
	return &bc
}
