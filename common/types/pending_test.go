/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"math/big"
	"reflect"
	"testing"
)

func TestPendingKey_Serialize(t *testing.T) {
	d1 := make([]byte, AddressSize)
	a, _ := BytesToAddress(d1)
	d2 := make([]byte, HashSize)
	h, _ := BytesToHash(d2)
	pk := &PendingKey{
		Address: a,
		Hash:    h,
	}

	if data, err := pk.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		pk2 := &PendingKey{}
		if err := pk2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(pk, pk2) {
				t.Fatalf("exp: %v, act: %v", pk, pk2)
			}
		}
	}
}

func TestPendingInfo_Serialize(t *testing.T) {
	d1 := make([]byte, AddressSize)
	a, _ := BytesToAddress(d1)
	d2 := make([]byte, HashSize)
	h, _ := BytesToHash(d2)
	pi := &PendingInfo{
		Source: a,
		Amount: Balance{Int: big.NewInt(10)},
		Type:   h,
	}

	if data, err := pi.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		pi2 := &PendingInfo{}
		if err := pi2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(pi, pi2) {
				t.Fatalf("exp: %v, act: %v", pi, pi2)
			}
		}
	}
}
