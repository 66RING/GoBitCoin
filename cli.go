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

// print usage
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("addblock: add block to blockchain")
	fmt.Println("showchain: show blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("add block success")
}

func (cli *CLI) showBlockChian() {
	bci := cli.blockchain.Iterator() // create iterator
	for {
		block := bci.next() // get next block
		fmt.Printf("Previous hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Current hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("Pow: %s", strconv.FormatBool(pow.Validate()))
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 { // if is the first block
			break
		}
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()
	addblockcmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	showchain := flag.NewFlagSet("showchain", flag.ExitOnError)

	addBlockData := addblockcmd.String("data", "", "Block data")
	switch os.Args[1] {
	case "addblock":
		err := addblockcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "showchain":
		err := showchain.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	if addblockcmd.Parsed() {
		if *addBlockData == "" {
			addblockcmd.Usage()
			os.Exit(1)
		} else {
			cli.addBlock(*addBlockData)
		}
	}

	if showchain.Parsed() {
		cli.showBlockChian()
	}

}
