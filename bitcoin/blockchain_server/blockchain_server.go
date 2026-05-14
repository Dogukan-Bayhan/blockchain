package main

import (
	"bitcoin/block"
	"bitcoin/network"
	"bitcoin/wallet"
	"log"
)

var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

// BlockchainServer exposes blockchain operations over HTTP and starts P2P networking.
type BlockchainServer struct {
	port      uint16
	p2pAddr   string
	bootstrap string
	network   *network.Manager
}

// NewBlockchainServer creates a blockchain server with HTTP and P2P settings.
func NewBlockchainServer(port uint16, p2pAddr string, bootstrap string) *BlockchainServer {
	return &BlockchainServer{
		port:      port,
		p2pAddr:   p2pAddr,
		bootstrap: bootstrap,
	}
}

// Port returns the HTTP listen port.
func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

// P2PAddr returns the local P2P listen address.
func (bcs *BlockchainServer) P2PAddr() string {
	return bcs.p2pAddr
}

// Bootstrap returns the optional initial peer address.
func (bcs *BlockchainServer) Bootstrap() string {
	return bcs.bootstrap
}

// GetBlockchain returns the singleton in-memory blockchain for this process.
func (bcs *BlockchainServer) GetBlockchain() *block.Blockchain {
	bc, ok := cache["blockchain"]
	if !ok {
		minersWallet := wallet.NewWallet()
		bc = block.NewBlockchain(minersWallet.BlockchainAddress(), bcs.Port())
		cache["blockchain"] = bc
		log.Printf("private_key %v", minersWallet.PrivateKeyStr())
		log.Printf("public_key %v", minersWallet.PublicKeyStr())
		log.Printf("blockchain_address %v", minersWallet.BlockchainAddress())
	}
	return bc
}
