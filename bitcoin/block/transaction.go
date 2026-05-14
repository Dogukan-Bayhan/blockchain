package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// Transaction is the current account-style transaction model.
type Transaction struct {
	id                         string
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

// NewTransaction creates a transaction between two blockchain addresses.
func NewTransaction(sender string, recipient string, value float32) *Transaction {
	t := &Transaction{
		senderBlockchainAddress:    sender,
		recipientBlockchainAddress: recipient,
		value:                      value,
	}
	t.id = t.Hash()
	return t
}

// ID returns the transaction hash used by the P2P inventory protocol.
func (t *Transaction) ID() string {
	return t.id
}

// Hash returns a deterministic SHA-256 hash for the transaction content.
func (t *Transaction) Hash() string {
	h := sha256.Sum256(t.signingBytes())
	return hex.EncodeToString(h[:])
}

// signingBytes returns the payload covered by wallet transaction signatures.
func (t *Transaction) signingBytes() []byte {
	m, err := json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
	if err != nil {
		return nil
	}
	return m
}

// Print writes a transaction to stdout.
func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" id                             %s\n", t.id)
	fmt.Printf(" sender_blockchain_address      %s\n", t.senderBlockchainAddress)
	fmt.Printf(" recipient_blockchain_address   %s\n", t.recipientBlockchainAddress)
	fmt.Printf(" value                          %.1f\n", t.value)
}

// MarshalJSON serializes private transaction fields into the public API shape.
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        string  `json:"id"`
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		ID:        t.id,
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}

// TransactionRequest is the HTTP payload accepted by the blockchain server.
type TransactionRequest struct {
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublicKey            *string  `json:"sender_public_key"`
	Value                      *float32 `json:"value"`
	Signature                  *string  `json:"signature"`
}

// Validate reports whether all required transaction request fields are present.
func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}
