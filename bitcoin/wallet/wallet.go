package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"

	"bitcoin/utils"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

// Wallet owns an ECDSA key pair and its derived blockchain address.
type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockchainAddress string
}

// NewWallet generates a key pair and derives a Base58Check-style address.
func NewWallet() *Wallet {
	w := new(Wallet)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Error: generate private key: %v", err)
	}
	w.privateKey = privateKey
	w.publicKey = &w.privateKey.PublicKey

	// 2. Perform SHA-256 hashing on the public key (32 bytes).
	h2 := sha256.New()
	h2.Write(w.publicKey.X.Bytes())
	h2.Write(w.publicKey.Y.Bytes())
	digest2 := h2.Sum(nil)

	// 3. Perform RIPEMD-160 hashing on the result of SHA-256 (20 bytes).
	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)

	// 4. Add version byte in front of RIPEMD-160 hash (0x00 for main network).
	vd5 := make([]byte, 21)
	vd5[0] = 0x00
	copy(vd5[1:], digest3[:])

	// 5. Perform SHA-256 hash on the extended RIPEMD-160 result.
	h6 := sha256.New()
	h6.Write(vd5)
	digest6 := h6.Sum(nil)

	// 6. Perform SHA-256 hash on the result of the previous SHA-256 hash.
	h8 := sha256.New()
	h8.Write(digest6)
	digest8 := h8.Sum(nil)

	// 7. Take the first 4 bytes of the second SHA-256 hash for checksum.
	chsum := digest8[:4]

	// 8. Add the 4 checksum bytes at the end of extended RIPEMD-160 hash.
	dc10 := make([]byte, 25)
	copy(dc10[:21], vd5[:])
	copy(dc10[21:], chsum[:])

	// 9. Convert the result from a byte string into base58.
	address := base58.Encode(dc10)
	w.blockchainAddress = address

	return w
}

// PrivateKey returns the wallet's ECDSA private key.
func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

// PrivateKeyStr returns the private key scalar as a hexadecimal string.
func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

// PublicKey returns the wallet's ECDSA public key.
func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

// PublicKeyStr returns the public key coordinates as a fixed-width hex string.
func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}

// BlockchainAddress returns the wallet's derived address.
func (w *Wallet) BlockchainAddress() string {
	return w.blockchainAddress
}

// MarshalJSON serializes wallet credentials for the wallet API response.
func (w *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey        string `json:"private_key"`
		PublicKey         string `json:"public_key"`
		BlockchainAddress string `json:"blockchain_address"`
	}{
		PrivateKey:        w.PrivateKeyStr(),
		PublicKey:         w.PublicKeyStr(),
		BlockchainAddress: w.BlockchainAddress(),
	})
}

// Transaction is a wallet-side transaction prepared for signing.
type Transaction struct {
	senderPrivateKey           *ecdsa.PrivateKey
	senderPublicKey            *ecdsa.PublicKey
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

// NewTransaction creates a wallet transaction ready for signature generation.
func NewTransaction(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey,
	sender string, recipient string, value float32) *Transaction {
	return &Transaction{privateKey, publicKey, sender, recipient, value}
}

// GenerateSignature signs the JSON representation of the transaction.
func (t *Transaction) GenerateSignature() *utils.Signature {
	m, err := json.Marshal(t)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil
	}
	h := sha256.Sum256([]byte(m))
	r, s, err := ecdsa.Sign(rand.Reader, t.senderPrivateKey, h[:])
	if err != nil {
		log.Printf("Error: %v", err)
		return nil
	}
	return &utils.Signature{r, s}
}

// MarshalJSON serializes the fields that are signed and submitted to the chain.
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}

// TransactionRequest is the wallet server payload for creating a signed transfer.
type TransactionRequest struct {
	SenderPrivateKey           *string `json:"sender_private_key"`
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	SenderPublicKey            *string `json:"sender_public_key"`
	Value                      *string `json:"value"`
}

// Validate reports whether all required wallet transaction fields are present.
func (tr *TransactionRequest) Validate() bool {
	if tr.SenderPrivateKey == nil ||
		tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil {
		return false
	}
	return true
}
