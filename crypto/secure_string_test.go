/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package crypto

import (
	"testing"

	"github.com/awnumar/memguard"
)

func TestNewSecureString(t *testing.T) {
	memguard.CatchInterrupt()

	defer memguard.Purge()
	p := "2Q9ZranUQbvM9JWwUBKN"
	ss, err := NewSecureString(p)
	if err != nil {
		t.Fatal(err)
	}
	if p != string(ss.Bytes()) {
		t.Fatal("decode error")
	}

	b2 := ss.Bytes()
	t.Log("b2: [", string(b2), "] len:", len(b2))

	ss.Destroy()

	p2 := "mQuj2fdGCys3snvhS6AS"
	ss, _ = NewSecureString(p2)

	if p2 != string(ss.Bytes()) {
		t.Fatal("decode p2 error")
	}
}
