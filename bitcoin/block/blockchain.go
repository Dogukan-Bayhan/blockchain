package block

import (
	"bitcoin/utils"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	// MINING_DIFFICULTY is the number of leading zeroes required by proof of work.
	MINING_DIFFICULTY = 3
	// MINING_SENDER marks coinbase-style reward transactions.
	MINING_SENDER = "THE BLOCKCHAIN"
	// MINING_REWARD is the fixed block reward used by the current toy chain.
	MINING_REWARD = 1.0
	// MINING_TIMER_SEC controls the interval used by automatic mining.
	MINING_TIMER_SEC = 20
)

// Blockchain holds the in-memory chain and pending transaction pool.
type Blockchain struct {
	transactionPool  map[string]*Transaction
	chain            []*Block
	blockchainAddres string
	port             uint16
	mux              sync.RWMutex
}

// NewBlockchain initializes a blockchain with a genesis block.
func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.blockchainAddres = blockchainAddress
	bc.transactionPool = make(map[string]*Transaction)
	bc.CreateBlock(0, b.Hash())
	bc.port = port
	return bc
}

// TransactionPool returns the current pending transactions.
func (bc *Blockchain) TransactionPool() []*Transaction {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.transactionPoolSnapshotLocked()
}

// HasTransaction reports whether a pending transaction exists in the pool.
func (bc *Blockchain) HasTransaction(id string) bool {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	_, ok := bc.transactionPool[id]
	return ok
}

// TransactionByID returns a pending transaction by its transaction ID.
func (bc *Blockchain) TransactionByID(id string) (*Transaction, bool) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	tx, ok := bc.transactionPool[id]
	return tx, ok
}

// MarshalJSON serializes the chain for HTTP responses.
func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return json.Marshal(struct {
		Blocks []*Block `json:"chains"`
	}{
		Blocks: bc.chain,
	})
}

// CreateBlock appends a new block and clears the transaction pool.
func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	return bc.createBlockLocked(nonce, previousHash)
}

// createBlockLocked appends a block while the blockchain mutex is already held.
func (bc *Blockchain) createBlockLocked(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPoolSnapshotLocked())
	bc.chain = append(bc.chain, b)
	bc.transactionPool = make(map[string]*Transaction)
	return b
}

// LastBlock returns the current chain tip.
func (bc *Blockchain) LastBlock() *Block {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.chain[len(bc.chain)-1]
}

// Print writes all blocks in the chain to stdout.
func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}

	fmt.Printf("%s\n", strings.Repeat("*", 25))
}

// CreateTransaction validates and adds a user transaction to the pool.
func (bc *Blockchain) CreateTransaction(sender string, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, s)

	// TODO
	// Sync

	return isTransacted
}

// AddTransaction appends a transaction after applying mining or signature rules.
func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	if sender == MINING_SENDER {
		bc.mux.Lock()
		defer bc.mux.Unlock()

		bc.transactionPool[t.ID()] = t
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		// if bc.CalculateTotalAmount(sender) < value {
		// 	log.Println("Error: not enough balance in a wallet")
		// 	return false
		// }
		bc.mux.Lock()
		defer bc.mux.Unlock()

		bc.transactionPool[t.ID()] = t
		return true
	} else {
		log.Println("Error: Verify Transaction")
	}

	return false
}

// VerifyTransactionSignature verifies that a transaction was signed by the sender key.
func (bc *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	if senderPublicKey == nil || s == nil || t == nil {
		log.Println("Error: invalid transaction signature input")
		return false
	}
	h := sha256.Sum256(t.signingBytes())
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

// CopyTransactionPool returns a detached copy used during proof-of-work search.
func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	return bc.copyTransactionPoolLocked()
}

// copyTransactionPoolLocked copies pending transactions while the mutex is already held.
func (bc *Blockchain) copyTransactionPoolLocked() []*Transaction {
	transactions := make([]*Transaction, 0, len(bc.transactionPool))
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderBlockchainAddress,
				t.recipientBlockchainAddress,
				t.value))
	}
	return transactions
}

// transactionPoolSnapshotLocked returns pending transactions without copying their values.
func (bc *Blockchain) transactionPoolSnapshotLocked() []*Transaction {
	transactions := make([]*Transaction, 0, len(bc.transactionPool))
	for _, tx := range bc.transactionPool {
		transactions = append(transactions, tx)
	}

	return transactions
}

// ValidProof checks whether a nonce satisfies the configured difficulty.
func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousHash, transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHashStr[:difficulty] == zeros
}

// ProofOfWork searches for a nonce that satisfies the current difficulty.
func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}

	return nonce
}

// Mining creates a reward transaction and appends a mined block if work exists.
func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	if len(bc.transactionPool) == 0 {
		return false
	}

	rewardTx := NewTransaction(MINING_SENDER, bc.blockchainAddres, MINING_REWARD)
	bc.transactionPool[rewardTx.ID()] = rewardTx

	transactions := bc.copyTransactionPoolLocked()
	previousHash := bc.chain[len(bc.chain)-1].Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}

	bc.createBlockLocked(nonce, previousHash)
	log.Println("action=mining, status=success")
	return true
}

// StartMining schedules repeated mining attempts.
func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

// CalculateTotalAmount computes a wallet balance by scanning all chain transactions.
func (bc *Blockchain) CalculateTotalAmount(blockchainAddress string) float32 {
	bc.mux.RLock()
	defer bc.mux.RUnlock()

	var totalAmount float32 = 0.0
	for _, b := range bc.chain {
		for _, t := range b.transactions {
			value := t.value
			if blockchainAddress == t.recipientBlockchainAddress {
				totalAmount += value
			}

			if blockchainAddress == t.senderBlockchainAddress {
				totalAmount -= value
			}
		}
	}

	return totalAmount
}
