/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ipsn/go-secp256k1"
)

const (
	gcmAdditionData = "qlc"
)

var (
	secp256k1N, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	// secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return ToECDSA(b)
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}

func AesGCMEncrypt(key, inText []byte) (outText, nonce []byte, err error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	stream, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, nil, err
	}

	nonce = GetEntropyCSPRNG(12)

	outText = stream.Seal(nil, nonce, inText, []byte(gcmAdditionData))
	return outText, nonce, err
}

func AesGCMDecrypt(key, cipherText, nonce []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	outText, err := stream.Open(nil, nonce, cipherText, []byte(gcmAdditionData))
	if err != nil {
		return nil, err
	}

	return outText, err
}

func GetEntropyCSPRNG(n int) []byte {
	buff := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, buff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return buff
}
