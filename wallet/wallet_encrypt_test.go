/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/hex"
	"testing"
)

const (
	passphrase = "98qUb5Ud"
	seed       = "a7bcc2785e93226699618087528c4fbc8990fc247f12743e2a9caea8590756a0"
)


func TestEncryptSeed(t *testing.T) {
	s, err := hex.DecodeString(seed)

	if err != nil {
		t.Fatal(err)
	}

	encryptSeed, err := EncryptSeed(s, []byte(passphrase))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(encryptSeed))

	bytes, err := DecryptSeed(encryptSeed, []byte(passphrase))
	if err != nil {
		t.Fatal(err)
	}

	if hex.EncodeToString(bytes) != seed {
		t.Fatal("decrypt err.")
	}
}
