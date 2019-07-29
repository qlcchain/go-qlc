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

	"github.com/qlcchain/go-qlc/crypto"
	"golang.org/x/crypto/scrypt"
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

func Encrypt(raw string, passphrase string) (string, error) {
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
func EncryptBytes(raw []byte, passphrase []byte) ([]byte, error) {
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

func Decrypt(cryptograph string, passphrase string) (string, error) {
	encryptedJSON, err := base64.StdEncoding.DecodeString(cryptograph)
	if err != nil {
		return "", nil
	}
	r, err := DecryptBytes(encryptedJSON, []byte(passphrase))
	if err != nil {
		return "", nil
	}
	return hex.EncodeToString(r), nil
}

//DecryptBytes decrypt raw json to raw
func DecryptBytes(encryptedJSON []byte, passphrase []byte) ([]byte, error) {
	cryptograph := cryptoSeedJSON{}
	err := json.Unmarshal(encryptedJSON, &cryptograph)
	if err != nil {
		return nil, errors.New("invalid cryptograph json")
	}

	cipherData, err := hex.DecodeString(cryptograph.Crypto.CipherText)
	if err != nil {
		return nil, errors.New("invalid cryptograph cipher text")
	}

	nonce, err := hex.DecodeString(cryptograph.Crypto.Nonce)
	if err != nil {
		return nil, errors.New("invalid cryptograph nonce")
	}

	scryptParams := cryptograph.Crypto.ScryptParams
	salt, err := hex.DecodeString(scryptParams.Salt)
	if err != nil {
		return nil, errors.New("invalid cryptograph salt")
	}
	// begin decrypt
	derivedKey, err := scrypt.Key(passphrase, salt, scryptParams.N, scryptParams.R, scryptParams.P, scryptParams.KeyLen)
	if err != nil {
		return nil, err
	}

	s, err := crypto.AesGCMDecrypt(derivedKey[:32], cipherData, nonce)
	if err != nil {
		return nil, errors.New("error decrypt cryptograph")
	}

	return s, nil
}
