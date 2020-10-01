package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandlength = 12

var nodeAddress string
var miningAddress string
var knowNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type addr struct {
	Addrlist []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}
type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddrFrom    string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func commandToBytes(command string) []byte {
	var bytes [commandlength]byte
	for index, char := range command {
		bytes[index] = byte(char)
	}
	return bytes[:]
}

func extractCommand(request []byte) []byte {
	return request[:commandlength]
}

func requestBlocks() {
	for _, node := range knowNodes { // send request to all knowNodes
		sendGetBlocks(node)
	}
}

// send block
func sendBlock(addr string, bc *Block) {
	data := block{nodeAddress, bc.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	sendData(addr, request)

}

func sendaddr(address string) {
	nodes := addr{knowNodes}
	nodes.Addrlist = append(nodes.Addrlist, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)
	sendData(address, request)

}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr) // Establish tcp connection object
	if err != nil {
		fmt.Printf("%s Address can't be reach.\n", addr)
		var updateNodes []string
		for _, node := range knowNodes {
			if node != addr {
				updateNodes = append(updateNodes, node)
			}
		}
		knowNodes = updateNodes // update list
		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}

}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items} // inventory data
	payload := gobEncode(inventory)            // history data

	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)

}

// send request block
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress}) // parse address
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)

}

func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{address, kind, id}) // parse address
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request)

}

func sendTx(addr string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)
	sendData(addr, request)

}

func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandlength:]) // get data
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)
	fmt.Println("Receive a new block")
	bc.AddBlock(block)
	fmt.Printf("Add a block %x \n", block.Hash)
	if len(blocksInTransit) > 0 {
		blockhash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockhash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

// handle network address
func handleaddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandlength:]) // get data
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	knowNodes = append(knowNodes, payload.Addrlist...)
	fmt.Printf("Have %d nodes now\n", len(knowNodes))
	requestBlocks()
}

func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandlength:]) // get data
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Receive inventory %d %s \n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockhash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockhash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockhash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit // sync block
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID) // send request transaction
		}
	}

}

func handleGetBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getdata
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, &block)
	}
	if payload.Type == "block" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddrFrom, &tx)
	}

}

func handleTx(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload tx
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx
	fmt.Println(nodeAddress, knowNodes[0])

	if nodeAddress == knowNodes[0] {

		for _, node := range knowNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {

		if len(mempool) >= 2 && len(miningAddress) >= 0 {

		MineTransactions:
			var txs []*Transaction
			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}
			if len(txs) == 0 {
				fmt.Println("No transaction, waiting to transaction")
				return
			}
			cbTx := NewCoinBaseTx(miningAddress, "")
			txs = append(txs, cbTx)
			newBlock := bc.MineBLock(txs)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()
			fmt.Println("GET NEW BLOCK!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}
			for _, node := range knowNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}
			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload verzion
	buff.Write(request[commandlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	mybestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if mybestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if mybestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKnow(payload.AddrFrom) {
		knowNodes = append(knowNodes, payload.AddrFrom)
	}
}

func handleConnection(conn net.Conn, bc *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := bytesToCommand(request[:commandlength])
	fmt.Printf("Receive command %s \n", command)

	switch command {
	case "addr":
		handleaddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Invalid command")
	}

	conn.Close()

}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	In, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer In.Close()

	bc := NewBlockChain(nodeID)
	if nodeAddress != knowNodes[0] {
		sendVersion(knowNodes[0], bc)
	}

	for {
		conn, err := In.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func nodeIsKnow(addr string) bool {
	for _, node := range knowNodes {
		if node == addr {
			return true
		}
	}
	return false
}
