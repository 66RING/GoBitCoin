package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte) {
	pubkeyhash := Base58Decode(address)
	pubkeyhash = pubkeyhash[1 : len(pubkeyhash)-4]
	out.PubKeyHash = pubkeyhash
}

func (out *TXOutput) IsLockWithKey(pubkeyHash []byte) bool {

	return bytes.Compare(out.PubKeyHash, pubkeyHash) == 0
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address)) // record pubkeyHash with address
	return txo
}

type TXOutputs struct {
	Outputs []TXOutput
}

// object to binary
func (out *TXOutputs) Serialize() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(out)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// binary to object
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}
	return outputs
}
