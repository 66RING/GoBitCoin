package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

//Secondary packaging
type UTXOSet struct {
	blockchain *BlockChain
}

func (utxo UTXOSet) FindSpendableOutputs(publickeyhash []byte, amount int) (int, map[string][]int) {
	unspendOutputs := make(map[string][]int)
	accumulated := 0
	db := utxo.blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()

		for key, value := cur.First(); key != nil; key, value = cur.Next() {
			txID := hex.EncodeToString(key)
			outs := DeserializeOutputs(value)
			for outIdx, out := range outs.Outputs {
				if out.IsLockWithKey(publickeyhash) && accumulated < amount {
					accumulated += out.Value
					unspendOutputs[txID] = append(unspendOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspendOutputs

}

// Find UTXO with public key
func (utxo UTXOSet) FindUTXO(publickeyhash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := utxo.blockchain.db
	err := db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()

		for key, value := cur.First(); key != nil; key, value = cur.Next() {
			outs := DeserializeOutputs(value)
			for _, out := range outs.Outputs {
				if out.IsLockWithKey(publickeyhash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

// Statistical transaction
func (utxo UTXOSet) CountTransaction() int {
	db := utxo.blockchain.db
	counter := 0
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()
		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			counter++
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return counter
}

// Rebuilding the index
func (utxo UTXOSet) Reindex() {
	db := utxo.blockchain.db
	bucketname := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketname)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_, err = tx.CreateBucket(bucketname)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	UTXO := utxo.blockchain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketname)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

}

// Update data
func (utxo UTXOSet) Update(block *Block) {
	db := utxo.blockchain.db
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions {
			if tx.IsCoinBase() == false {
				for _, vin := range tx.Vin {
					updateOuts := TXOutputs{}
					outsBytes := bucket.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}
					if len(updateOuts.Outputs) == 0 {
						err := bucket.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := bucket.Put(vin.Txid, updateOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := bucket.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

}
