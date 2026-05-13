package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	port := flag.Uint("port", 5005, "HTTP port")
	p2pAddr := flag.String("p2p", ":6005", "P2P listen address")
	bootstrap := flag.String("bootstrap", "", "Bootstrap peer address")
	flag.Parse()
	app := NewBlockchainServer(uint16(*port), *p2pAddr, *bootstrap)
	app.Run()
}
