package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockchainAddress string
}

func NewWallet() *Wallet {
	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%x %x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}
