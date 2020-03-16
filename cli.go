package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
	blockchain *BlockChain
}

// print usage
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Printf("\tcreatewallet: Create a wallet\n")
	fmt.Printf("\tlistaddress: show all address(account)\n")
	fmt.Printf("\taddblock: add block to blockchain\n")
	fmt.Printf("\tshowchain: show blockchain\n")
	fmt.Printf("\tgetbalance: get balance with address\n")
	fmt.Printf("\tcreateblockchain: create BlockChain with address\n")
	fmt.Printf("\tsend -from From -to To -amount Amount -mine: New transaction (-mine) now or not\n")
	fmt.Printf("\treindexutxo: Reindex UTXO\n")
	fmt.Printf("\tstartnode -miner ADDR: Open a node\n")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	nodeID := "3002"
	// export NODE_ID= ?
	//nodeID := os.Getenv("NODE_ID")
	//if nodeID == "" {
	//	fmt.Println("Must have a running port num")
	//	os.Exit(1)
	//}

	listaddressescmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	createwalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchaincmd", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	reindexutxocmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startnodecmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getbalanceaddress := getbalancecmd.String("address", "", "get balance addree")
	createblockaddress := createblockchaincmd.String("address", "", "get block addree")
	sendfrom := sendcmd.String("from", "", "from who")
	sendto := sendcmd.String("to", "", "to who")
	sendamount := sendcmd.Int("amount", 0, "amount")
	sendmine := sendcmd.Bool("mine", false, "mine now?")
	startnodeminer := startnodecmd.String("miner", "", "mine?")

	switch os.Args[1] {
	case "createwallet":
		err := createwalletcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listaddressescmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
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
	case "reindexutxo":
		err := reindexutxocmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startnodecmd.Parse(os.Args[2:])
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
		cli.getBalance(*getbalanceaddress, nodeID)
	}

	if createblockchaincmd.Parsed() {
		if *createblockaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockaddress, nodeID)
	}

	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount, nodeID, *sendmine)

	}
	if showchaincmd.Parsed() {
		cli.showBlockChian(nodeID)
	}
	if createwalletcmd.Parsed() {
		cli.createWallet(nodeID)
	}
	if listaddressescmd.Parsed() {
		cli.listAddress(nodeID)
	}
	if reindexutxocmd.Parsed() {
		cli.reindexUTXOP(nodeID)
	}
	if startnodecmd.Parsed() {
		cli.startNode(nodeID, *startnodeminer)
	}
}
