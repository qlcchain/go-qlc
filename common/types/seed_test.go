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
	_ = seed.UnmarshalBinary(b)
	s := seed.String()
	if s != expect {
		t.Errorf("failed to seed from expect %s but %s", expect, s)
	}
}

func TestSeed_Account(t *testing.T) {
	const s = "1234567890123456789012345678901234567890123456789012345678901234"
	seed, _ := NewSeed()
	b, _ := hex.DecodeString(s)
	_ = seed.UnmarshalBinary(b)
	account, err := seed.Account(0)
	if err != nil {
		pub, _ := KeypairFromPrivateKey(hex.EncodeToString(account.privKey))
		addr := PubToAddress(pub)
		if addr.String() != "qlc_3iwi45me3cgo9aza9wx5f7rder37hw11xtc1ek8psqxw5oxb8cujjad6qp9y" {
			t.Error("generate seed failed.")
		}
	}
}

func TestSeed_MasterAddress(t *testing.T) {
	seed, err := NewSeed()
	if err != nil {
		t.Fatal(err)
	}

	pub, _, err := KeypairFromSeed(seed.String(), 0)
	if err != nil {
		t.Fatal(err)
	}
	addr := PubToAddress(pub)
	t.Log(addr.String())
	addr2 := seed.MasterAddress()

	if addr != addr2 {
		t.Fatal("addr != addr2")
	}
}

func TestSeed_IsZero(t *testing.T) {
	const s = "1234567890123456789012345678901234567890123456789012345678901234"
	const s0 = "0000000000000000000000000000000000000000000000000000000000000000"
	sByte, _ := hex.DecodeString(s)
	seed, err := BytesToSeed(sByte)
	if err != nil {
		t.Fatal(err)
	}
	b := seed.IsZero()
	if b {
		t.Fatal("seed should not be zero")
	}
	sByte0, _ := hex.DecodeString(s0)
	seed0, err := BytesToSeed(sByte0)
	if err != nil {
		t.Fatal(err)
	}
	b0 := seed0.IsZero()
	if !b0 {
		t.Fatal("seed should be zero")
	}
}
