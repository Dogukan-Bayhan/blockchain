package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

// Signature stores the two ECDSA signature scalars.
type Signature struct {
	R *big.Int
	S *big.Int
}

// String returns the signature as a fixed-width hexadecimal string.
func (s *Signature) String() string {
	return fmt.Sprintf("%064x%064x", s.R, s.S)
}

// String2BigIntTuple decodes two concatenated 64-character hex integers.
func String2BigIntTuple(s string) (big.Int, big.Int) {
	var bix big.Int
	var biy big.Int

	if len(s) < 128 {
		log.Printf("Error: invalid string length: %d", len(s))
		return bix, biy
	}

	bx, err := hex.DecodeString(s[:64])
	if err != nil {
		log.Printf("Error: %v", err)
		return bix, biy
	}

	by, err := hex.DecodeString(s[64:])
	if err != nil {
		log.Printf("Error: %v", err)
		return bix, biy
	}

	_ = bix.SetBytes(bx)
	_ = biy.SetBytes(by)

	return bix, biy
}

// SignatureFromString decodes an ECDSA signature from its hex string form.
func SignatureFromString(s string) *Signature {
	if len(s) < 128 {
		log.Printf("Error: invalid signature length: %d", len(s))
		return nil
	}
	x, y := String2BigIntTuple(s)
	return &Signature{&x, &y}
}

// PublicKeyFromString decodes an ECDSA public key from concatenated hex coordinates.
func PublicKeyFromString(s string) *ecdsa.PublicKey {
	if len(s) < 128 {
		log.Printf("Error: invalid public key length: %d", len(s))
		return nil
	}
	x, y := String2BigIntTuple(s)
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: &x, Y: &y}
}

// PrivateKeyFromString decodes an ECDSA private key and attaches its public key.
func PrivateKeyFromString(s string, publicKey *ecdsa.PublicKey) *ecdsa.PrivateKey {
	if publicKey == nil {
		log.Println("Error: public key is nil")
		return nil
	}

	b, err := hex.DecodeString(s[:])
	if err != nil {
		log.Printf("Error: %v", err)
		return nil
	}
	var bi big.Int
	_ = bi.SetBytes(b)
	return &ecdsa.PrivateKey{PublicKey: *publicKey, D: &bi}
}
