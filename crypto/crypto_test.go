/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestAesGCMEncrypt_And_Decrypt(t *testing.T) {
	gcmDummyKey32 := "6368616e676520746869732070617373776f726420746f206120736563726574"
	gcmDummyPlainText := hex.EncodeToString([]byte("exampleplaintext"))
	key, _ := hex.DecodeString(gcmDummyKey32)
	plain, _ := hex.DecodeString(gcmDummyPlainText)
	out, nonce, err := AesGCMEncrypt(key, plain)
	t.Log("key:", hex.EncodeToString(key))
	t.Log("origin(Hex):", hex.EncodeToString(plain))
	t.Log("origin:", string(plain))
	t.Log("cipher:", hex.EncodeToString(out))
	t.Log("nonce:", hex.EncodeToString(nonce))
	if err != nil {
		t.Fatal(err)
	}

	plain1, err := AesGCMDecrypt(key, out, nonce)
	t.Log("origin: ", string(plain1))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plain1, plain) {
		t.Fatal("Mis content")
	}
}
