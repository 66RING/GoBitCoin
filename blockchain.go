package main

import (
	"bytes"
	"crypto/ecdsa"
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

func (blockchain *BlockChain) MineBLock(transactions []*Transaction) {
	var lastHash []byte
	for _, tx := range transactions {
		if blockchain.VerifyTransaction(tx) != true {
			log.Panic("Transaction invalid")
		}
	}
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
func (blockchain *BlockChain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := blockchain.Iterator()
	for {
		block := bci.next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spendoutidx := range spentTXOs[txID] {
						if spendoutidx == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// get unspent transaction list
func (blockchain *BlockChain) FindUnspendTransactions(pubkeyhash []byte) []Transaction {
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
				if out.IsLockWithKey(pubkeyhash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					if in.UsesKeyHash(pubkeyhash) {
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
func (blockchain *BlockChain) FindSpendableOutputs(pubkeyhash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := blockchain.FindUnspendTransactions(pubkeyhash)
	accmulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outindex, out := range tx.Vout {
			if out.IsLockWithKey(pubkeyhash) && accmulated < amount {
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

func CreateBlockChain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("Database exists there is not need to create")
		os.Exit(1)
	}

	var tip []byte                                      //
	db, err := bolt.Open(dbFile, 0600, nil)             //open db
	cbtx := NewCoinBaseTx(address, genesisCoinbaseData) // New genesis block
	genesis := NewGenesisBlock(cbtx)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
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

func (blockchain *BlockChain) SignTransaction(tx *Transaction, privatekey ecdsa.PrivateKey) {

	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		preTx, err := blockchain.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(preTx.ID)] = preTx
	}
	tx.Sign(privatekey, prevTXs)
}

func (blockchian *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := blockchian.Iterator()
	for {
		block := bci.next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, nil

}

func (blockchain *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTx, err := blockchain.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	return tx.Verify(prevTxs)
}
