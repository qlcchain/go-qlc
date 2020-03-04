/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"reflect"
	"strings"
	"testing"
)

func TestHexBytes_Serialize(t *testing.T) {
	hb1 := NewHexBytesFromHex("A1B2C3D4")
	hb1Str := strings.ToUpper(hb1.String())
	if hb1Str != "A1B2C3D4" {
		t.Fatalf("exp: %v, act: %v", "A1B2C3D4", hb1Str)
	}
	if hb1.Len() != 4 {
		t.Fatalf("exp: %v, act: %v", 4, hb1.Len())
	}

	hb2 := NewHexBytesFromData(hb1.Bytes())

	if !reflect.DeepEqual(hb1, hb2) {
		t.Fatalf("exp: %v, act: %v", hb1, hb2)
	}

	if msgData, err := hb1.MarshalMsg(nil); err != nil {
		t.Fatal(err)
	} else {
		hb3 := &HexBytes{}
		if _, err := hb3.UnmarshalMsg(msgData); err != nil {
			t.Fatal(err)
		} else {
			if hb1.String() != hb3.String() {
				t.Fatalf("exp: %v, act: %v", hb1, hb3)
			}
		}
	}

	binData := make([]byte, 0)
	if err := hb1.MarshalBinaryTo(binData); err != nil {
		t.Fatal(err)
	} else {
		hb4 := &HexBytes{}
		if err := hb1.UnmarshalBinary(binData); err != nil {
			t.Fatal(err)
		} else {
			if hb1.String() != hb4.String() {
				t.Fatalf("exp: %v, act: %v", hb1, hb4)
			}
		}
	}

	if textData, err := hb1.MarshalText(); err != nil {
		t.Fatal(err)
	} else {
		hb5 := &HexBytes{}
		if err := hb5.UnmarshalText(textData); err != nil {
			t.Fatal(err)
		} else {
			if hb1.String() != hb5.String() {
				t.Fatalf("exp: %v, act: %v", hb1, hb5)
			}
		}
	}
}
