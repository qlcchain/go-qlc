/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

const seed = "5a32b2325437cc10c07e36161fcda24f01ec0038969ecaaa709a133372bf4b94"

func TestAccount_Address(t *testing.T) {
	pub, priv, err := KeypairFromSeed(seed, 1)
	if err != nil {
		t.Fatal(err)
	}
	account := NewAccount(priv)
	t.Log(account.String())
	address := account.Address()
	if address != PubToAddress(pub) {
		t.Fatal("invalid address")
	}
	t.Log(address.String())

	h := Hash{}
	err = random.Bytes(h[:])
	if err != nil {
		t.Fatal(err)
	}

	sign := account.Sign(h)
	t.Log(sign.String())

	if !address.Verify(h[:], sign[:]) {
		t.Fatal("sign failed")
	}
}
