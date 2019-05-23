/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSignature(t *testing.T) {
	const s = "148AA79F002D747E4E262B0CC2F7B5FAB121C9362C8DB5906DC40B91147A57DAA827DF4321D0D8DED972C2469C72B4191E3AF9A69A67FC893462DCE19E9E7005"
	var sign Signature
	err := sign.Of(s)
	upper := strings.ToUpper(sign.String())
	if err != nil || upper != s {
		t.Errorf("sign missmatch. expect: %s but: %s", s, upper)
	}
}

func TestSignature_UnmarshalJSON(t *testing.T) {
	s := `"5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600"`
	var sign2 Signature
	err := json.Unmarshal([]byte(s), &sign2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sign2)
}

func TestBytesToSignature(t *testing.T) {
	s := `5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600`
	bytes, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	sign1, err := BytesToSignature(bytes)
	if err != nil {
		t.Fatal(err)
	}
	sign := new(Signature)
	err = sign.Of(s)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(sign[:], sign1[:]) {
		t.Fatal(sign.String(), sign1.String())
	}
}
