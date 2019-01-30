/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import (
	"encoding/hex"
	"testing"
	"time"
)

const (
	passphrase  = "98qUb5Ud"
	seed        = "a7bcc2785e93226699618087528c4fbc8990fc247f12743e2a9caea8590756a0"
	encryptSeed = `{
  "crypto": {
    "ciphertext": "cd423f79ab838549eb9bc6996bc04267ed06b154d1f3768b8a50e8362f9bd735ccc5858ddafef6e643037f51d56975bb",
    "nonce": "1dfd8e5bf0e609a014836c46",
    "scryptparams": {
      "n": 262144,
      "r": 8,
      "p": 1,
      "keylen": 32,
      "salt": "3abc0f46dc00d8b044c4c08e9c748a5d1c837ccb51ff16c0203d8f7d99e7dd5e"
    }
  },
  "timestamp": 1545220913
}`
)

func TestEncryptSeed(t *testing.T) {
	t.Parallel()
	start := time.Now()
	s, err := hex.DecodeString(seed)

	if err != nil {
		t.Fatal(err)
	}
	encryptSeed, err := EncryptBytes(s, []byte(passphrase))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(encryptSeed))

	t.Logf("EncryptBytes cost %s", time.Since(start))
}

func TestDecryptSeed(t *testing.T) {
	t.Parallel()
	start := time.Now()
	bytes, err := DecryptBytes([]byte(encryptSeed), []byte(passphrase))
	if err != nil {
		t.Fatal(err)
	}

	if hex.EncodeToString(bytes) != seed {
		t.Fatal("decrypt err.")
	}
	t.Logf("DecryptBytes cost %s", time.Since(start))
}
