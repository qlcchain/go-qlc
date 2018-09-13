/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"testing"
)

func TestNewSeed(t *testing.T) {
	_, err := NewSeed()
	if err != nil {
		t.Errorf("fail to generate new seed: %v", err)
	}
}

func TestSeed_String(t *testing.T) {
	seed, _ := NewSeed()
	s := seed.String()
	if len(s) == 0 {
		t.Error("failed to convert seed to string")
	}
}

func TestSeed_From(t *testing.T) {
	seed, _ := NewSeed()
	const expect = "1234567890123456789012345678901234567890123456789012345678901234"
	b, _ := hex.DecodeString(expect)
	seed.UnmarshalText(b)
	s := seed.String()
	if s != expect {
		t.Errorf("failed to seed from expect %s but %s", expect, s)
	}
}

func TestSeed_Key(t *testing.T) {
	const s = "1234567890123456789012345678901234567890123456789012345678901234"
	seed, _ := NewSeed()
	b, _ := hex.DecodeString(s)
	seed.UnmarshalText(b)
	priv, err := seed.Key(0)
	if err != nil {
		pub, _ := KeypairFromPrivateKey(hex.EncodeToString(priv))
		addr := PubToAddress(pub)
		if addr.String() != "qlc_3iwi45me3cgo9aza9wx5f7rder37hw11xtc1ek8psqxw5oxb8cujjad6qp9y" {
			t.Error("generate seed failed.")
		}
	}
}
