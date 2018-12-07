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
	"crypto/rand"
	"io"
)

const (
	gcmAdditionData = "qlc"
)

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
