package main

import (
	"flag"
	"log"
)

// init configures the wallet server log prefix.
func init() {
	log.SetPrefix("Wallet Server: ")
}

// main parses server flags and starts the wallet HTTP process.
func main() {
	port := flag.Uint("port", 8080, "Tcp Port Number for Wallet Server")
	gateway := flag.String("gateway", "http://127.0.0.1:5005", "Blockchain Gateway")
	flag.Parse()

	app := NewWalletServer(uint16(*port), *gateway)
	app.Run()
}
