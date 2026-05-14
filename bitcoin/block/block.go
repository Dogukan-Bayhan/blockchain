package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Block stores the minimal data required to link transactions into a chain.
type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byte
	transactions []*Transaction
}

// NewBlock creates a block from a nonce, previous hash, and transaction set.
func NewBlock(nonce int, previousHash [32]byte, transanctions []*Transaction) *Block {
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.previousHash = previousHash
	b.transactions = transanctions
	return b
}

// Print writes a human-readable block representation to stdout.
func (b *Block) Print() {
	fmt.Printf("timestamp       %d\n", b.timestamp)
	fmt.Printf("nonce           %d\n", b.nonce)
	fmt.Printf("previous_hash   %x\n", b.previousHash)

	for _, t := range b.transactions {
		t.Print()
	}
}

// Hash returns the SHA-256 hash of the block's JSON representation.
func (b *Block) Hash() [32]byte {
	m, err := json.Marshal(b)
	if err != nil {
		log.Printf("Error: %v", err)
		return [32]byte{}
	}

	return sha256.Sum256(m)
}

// MarshalJSON serializes private block fields into the public API shape.
func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}
