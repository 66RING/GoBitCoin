package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
	blockchain *BlockChain
}

func (cli *CLI) createBlockChain(address string) {
	bc := createBlockChain(address)
	bc.db.Close()
	fmt.Println("Create Success")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	defer bc.db.Close()
	balance := 0
	UTXOs := bc.FindUTXO(address) // Find balace
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance %s: %d \n", address, balance)
}

// print usage
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("addblock: add block to blockchain")
	fmt.Println("showchain: show blockchain")
	fmt.Println("getbalance: get balance with address")
	fmt.Println("createblockchain: create BlockChain with address")
	fmt.Println("send -from From -to To -amount Amount: New transaction")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) showBlockChian() {
	bc := NewBlockChain("")
	defer bc.db.Close()

	bci := bc.Iterator()
	for {
		block := bci.next()
		fmt.Printf("PrevBlockHash: %x", block.PrevBlockHash)
		fmt.Printf("Current Hash: %x", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s \n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBLock([]*Transaction{tx})
	fmt.Println("Transaction Success!")
}

func (cli *CLI) Run() {
	cli.validateArgs()

	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchaincmd", flag.ExitOnError)

	getbalanceaddress := getbalancecmd.String("address", "", "get balance addree")
	createblockaddress := createblockchaincmd.String("address", "", "get block addree")
	sendfrom := sendcmd.String("from", "", "from who")
	sendto := sendcmd.String("to", "", "to who")
	sendamount := sendcmd.Int("amount", 0, "amount")

	switch os.Args[1] {
	case "getbalance":
		err := getbalancecmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createblockchaincmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "showchain":
		err := showchaincmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	if getbalancecmd.Parsed() {
		if *getbalanceaddress == "" {
			getbalancecmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getbalanceaddress)
	}

	if createblockchaincmd.Parsed() {
		if *createblockaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockaddress)
	}

	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount)

	}
	if showchaincmd.Parsed() {
		cli.showBlockChian()
	}
}
