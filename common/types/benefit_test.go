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

func TestBenefit(t *testing.T) {
	benefit := &Benefit{
		Balance: Balance{Int: big.NewInt(1)},
		Vote:    Balance{Int: big.NewInt(11)},
		Network: Balance{Int: big.NewInt(21)},
		Storage: Balance{Int: big.NewInt(12)},
		Oracle:  Balance{Int: big.NewInt(133)},
		Total:   Balance{Int: big.NewInt(112)},
	}
	if data, err := benefit.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		b2 := &Benefit{}
		if err := b2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(benefit, b2) {
				t.Fatalf("exp: %v,act: %v", benefit, b2)
			} else {
				t.Log(benefit.String())
			}
		}
	}

	b3 := benefit.Clone()
	if !reflect.DeepEqual(benefit, b3) {
		t.Fatalf("exp: %v,act: %v", benefit, b3)
	}
}
