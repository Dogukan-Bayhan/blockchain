package main

import (
	"bitcoin/network"
	"fmt"
)

func (bcs *BlockchainServer) StartNetwork() {
	nodeID := fmt.Sprintf("node-%d", bcs.Port())

	manager := network.NewManager(network.Config{
		NodeID:    nodeID,
		P2PAddr:   bcs.P2PAddr(),
		Bootstrap: bcs.Bootstrap(),
	})

	go manager.Start()
}
