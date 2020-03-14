package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const subsidy = 1000 // reward

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// signature
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	if tx.IsCoinBase() {
		return // 挖矿返回无需签名
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("Previous Transaction Invalid ")
		}
	}
	txCopy := tx.TrimmedCopy() // copy trimmed copy, withouy private key
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		//datatoSign := fmt.Sprintf("%x\n", txCopy)
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}

}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

// Object -> string
func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Transaction: %x\n", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("input %d\n", i))
		lines = append(lines, fmt.Sprintf("TXID %d\n", input.Txid))
		lines = append(lines, fmt.Sprintf("OUT %d\n", input.Vout))
		lines = append(lines, fmt.Sprintf("Signature %d\n", input.Signature))
		lines = append(lines, fmt.Sprintf("PubKey %d\n", input.PubKey))
	}

	for i, output := range tx.Vout {

		lines = append(lines, fmt.Sprintf("out %d\n", i))
		lines = append(lines, fmt.Sprintf("value %d\n", output.Value))
		lines = append(lines, fmt.Sprintf("out %d\n", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			//log.Panic("Previous Transaction is Invalid")
		}
	}
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		siglen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(siglen / 2)])
		s.SetBytes(vin.Signature[(siglen / 2):])

		x := big.Int{}
		y := big.Int{}
		keylen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:keylen/2])
		y.SetBytes(vin.PubKey[keylen/2:])

		//datatoVerigy := fmt.Sprintf("%x\n", txCopy)

		rawPubkey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubkey, txCopy.ID, &r, &s) == false {
			return false
		}
		//txCopy.Vin[inID].PubKey = nil
	}

	return true
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

func NewCoinBaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %x", to)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)} // reward of input
	txout := NewTXOutput(subsidy, to)                // reward of output
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()
	return &tx
}

// Object -> Binary
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

// Binary -> Object
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}
	return transaction
}

func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	pubkeyhash := HashPubkey(wallet.PublicKey)
	acc, validOutputs := bc.FindSpendableOutputs(pubkeyhash, amount)
	if acc < amount {
		log.Panic("Don't have enough money")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	bc.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}
