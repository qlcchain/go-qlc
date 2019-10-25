/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"golang.org/x/crypto/scrypt"

	"github.com/qlcchain/go-qlc/crypto"
)

const (
	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = 1 << 18

	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptP = 1

	scryptR      = 8
	scryptKeyLen = 32
)

type scryptParams struct {
	N      int    `json:"n"`
	R      int    `json:"r"`
	P      int    `json:"p"`
	KeyLen int    `json:"keylen"`
	Salt   string `json:"salt"`
}

type cryptoJSON struct {
	CipherText   string       `json:"ciphertext"`
	Nonce        string       `json:"nonce"`
	ScryptParams scryptParams `json:"scryptparams"`
}

type cryptoSeedJSON struct {
	Crypto    cryptoJSON `json:"crypto"`
	Timestamp int64      `json:"timestamp"`
}

func Encrypt(raw, passphrase string) (string, error) {
	s, err := hex.DecodeString(raw)
	if err != nil {
		return "", err
	}
	encryptSeed, err := EncryptBytes(s, []byte(passphrase))
	if err != nil {
		return "", err
	}
	r := base64.StdEncoding.EncodeToString(encryptSeed)
	return r, nil
}

//EncryptBytes encrypts raw by passphrase to json binary
func EncryptBytes(raw, passphrase []byte) ([]byte, error) {
	n := StandardScryptN
	p := StandardScryptP
	salt := crypto.GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(passphrase, salt, n, scryptR, p, scryptKeyLen)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:32]

	cipherData, nonce, err := crypto.AesGCMEncrypt(encryptKey, raw)
	if err != nil {
		return nil, err
	}

	ScryptParams := scryptParams{
		N:      n,
		R:      scryptR,
		P:      p,
		KeyLen: scryptKeyLen,
		Salt:   hex.EncodeToString(salt),
	}

	cryptoJSON := cryptoJSON{
		CipherText:   hex.EncodeToString(cipherData),
		Nonce:        hex.EncodeToString(nonce),
		ScryptParams: ScryptParams,
	}

	encryptedJSON := cryptoSeedJSON{
		Crypto:    cryptoJSON,
		Timestamp: time.Now().Unix(),
	}

	return json.Marshal(encryptedJSON)
}

func Decrypt(cryptoGraph, passphrase string) (string, error) {
	encryptedJSON, err := base64.StdEncoding.DecodeString(cryptoGraph)
	if err != nil {
		return "", nil
	}
	r, err := DecryptBytes(encryptedJSON, []byte(passphrase))
	if err != nil {
		return "", nil
	}
	return hex.EncodeToString(r), nil
}

// DecryptBytes decrypt raw json to raw
func DecryptBytes(encryptedJSON, passphrase []byte) ([]byte, error) {
	cryptoGraph := cryptoSeedJSON{}
	err := json.Unmarshal(encryptedJSON, &cryptoGraph)
	if err != nil {
		return nil, errors.New("invalid crypto graph json")
	}

	cipherData, err := hex.DecodeString(cryptoGraph.Crypto.CipherText)
	if err != nil {
		return nil, errors.New("invalid crypto graph cipher text")
	}

	nonce, err := hex.DecodeString(cryptoGraph.Crypto.Nonce)
	if err != nil {
		return nil, errors.New("invalid crypto graph nonce")
	}

	scryptParams := cryptoGraph.Crypto.ScryptParams
	salt, err := hex.DecodeString(scryptParams.Salt)
	if err != nil {
		return nil, errors.New("invalid crypto graph salt")
	}
	// begin decrypt
	derivedKey, err := scrypt.Key(passphrase, salt, scryptParams.N, scryptParams.R, scryptParams.P, scryptParams.KeyLen)
	if err != nil {
		return nil, err
	}

	s, err := crypto.AesGCMDecrypt(derivedKey[:32], cipherData, nonce)
	if err != nil {
		return nil, errors.New("error decrypt crypto graph")
	}

	return s, nil
}
