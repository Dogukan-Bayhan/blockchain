package main

import (
	"bitcoin/network"
	"fmt"
)

// StartNetwork creates and starts the P2P manager for this blockchain server.
func (bcs *BlockchainServer) StartNetwork() {
	if bcs.network != nil {
		return
	}

	nodeID := fmt.Sprintf("node-%d", bcs.Port())

	bcs.network = network.NewManager(network.Config{
		NodeID:    nodeID,
		P2PAddr:   bcs.P2PAddr(),
		Bootstrap: bcs.Bootstrap(),
	})

	go bcs.network.Start()
}

// StopNetwork stops the P2P manager if it is running.
func (bcs *BlockchainServer) StopNetwork() {
	if bcs.network == nil {
		return
	}

	bcs.network.Stop()
}
