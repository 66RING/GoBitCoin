package main

func main() {
	block := NewBlockChain()
	defer block.db.Close()
	cli := CLI{block}
	cli.Run()
}
