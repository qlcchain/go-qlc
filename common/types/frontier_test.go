/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"reflect"
	"sort"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestFrontier(t *testing.T) {
	h1 := Hash{}
	_ = random.Bytes(h1[:])
	h2 := Hash{}
	_ = random.Bytes(h2[:])
	f := &Frontier{
		HeaderBlock: h1,
		OpenBlock:   h2,
	}

	if f.IsZero() {
		t.Fatal()
	}
	if data, err := f.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		f2 := &Frontier{}
		if err := f2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(f, f2) {
				t.Fatalf("exp: %v, act: %v", f, f2)
			}
		}
	}
}

func TestFrontierBlock(t *testing.T) {
	var data []*Frontier

	for i := 0; i < 4; i++ {
		h1 := Hash{}
		_ = random.Bytes(h1[:])
		h2 := Hash{}
		_ = random.Bytes(h2[:])
		f := &Frontier{
			HeaderBlock: h1,
			OpenBlock:   h2,
		}

		data = append(data, f)
	}

	if len(data) != 4 {
		t.Fatalf("invalid len %d, exp: 4", len(data))
	}

	sort.Sort(Frontiers(data))
	for _, b := range data {
		t.Log(b)
	}

	fbs := &FrontierBlock{
		Fr:        data[0],
		HeaderBlk: &StateBlock{},
	}

	if data, err := fbs.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		f2 := &FrontierBlock{}
		if err := f2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(fbs.Fr, f2.Fr) {
				t.Fatalf("exp: %v, act: %v", fbs, f2)
			}
		}
	}

}
