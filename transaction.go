package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10 // reward

type TXInput struct {
	Txid      []byte // Store id of transaction
	Vout      int
	ScriptSig string // Store address of wallet
}

type TXOutput struct {
	Value        int
	ScriptPubket string
}

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// set transaction id
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (input *TXInput) CanUnlockOutPutWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubket == unlockingData
}

func NewCoinBaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s", to)
	}
	txin := TXInput{[]byte{}, -1, data} // reward of input
	txout := TXOutput{subsidy, to}      // reward of output
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	return &tx
}

// 转账交易
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	acc, validaOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("Lack of Balance")
	}
	for txid, out := range validaOutputs {
		txID, err := hex.DecodeString(txid) // Travers invalid output
		if err != nil {
			log.Panic(err)
		}
		for _, out := range out {
			input := TXInput{txID, out, from} // inputs
			inputs = append(inputs, input)    // outputs
		}
	}
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
