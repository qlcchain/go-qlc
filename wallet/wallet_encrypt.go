/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/qlcchain/go-qlc/crypto"
	"golang.org/x/crypto/scrypt"
	"time"
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

//EncryptSeed encrypt seed by passphrase to json binary
func EncryptSeed(seed []byte, passphrase []byte) ([]byte, error) {
	n := StandardScryptN
	p := StandardScryptP
	salt := crypto.GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(passphrase, salt, n, scryptR, p, scryptKeyLen)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:32]

	cipherData, nonce, err := crypto.AesGCMEncrypt(encryptKey, seed)
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
		Timestamp: time.Now().UTC().Unix(),
	}

	return json.Marshal(encryptedJSON)
}

//DecryptSeed decrypt seed json to seed
func DecryptSeed(encryptedJSON []byte, passphrase []byte) ([]byte, error) {
	encryptSeed := cryptoSeedJSON{}
	err := json.Unmarshal(encryptedJSON, &encryptSeed)
	if err != nil {
		return nil, errors.New("invalid encryptSeed json")
	}

	cipherData, err := hex.DecodeString(encryptSeed.Crypto.CipherText)
	if err != nil {
		return nil, errors.New("invalid encryptSeed cipher text")
	}

	nonce, err := hex.DecodeString(encryptSeed.Crypto.Nonce)
	if err != nil {
		return nil, errors.New("invalid encryptSeed nonce")
	}

	scryptParams := encryptSeed.Crypto.ScryptParams
	salt, err := hex.DecodeString(scryptParams.Salt)
	if err != nil {
		return nil, errors.New("invalid encryptSeed salt")
	}
	// begin decrypt
	derivedKey, err := scrypt.Key(passphrase, salt, scryptParams.N, scryptParams.R, scryptParams.P, scryptParams.KeyLen)
	if err != nil {
		return nil, err
	}

	s, err := crypto.AesGCMDecrypt(derivedKey[:32], cipherData, nonce)
	if err != nil {
		return nil, errors.New("error decrypt encryptSeed")
	}

	return s, nil
}
