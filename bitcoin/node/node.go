package node

import (
	"bitcoin/block"
	"bitcoin/network"
	"sync"
	"time"
)

// Node describes the process-level identity and addresses of a blockchain node.
type Node struct {
	Config Config

	Network *network.Manager
	Chain   *block.Blockchain

	seenTx  map[string]time.Time
	knownTx map[string]block.TransactionRequest

	mu sync.RWMutex
}
