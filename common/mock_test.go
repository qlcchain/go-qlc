/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"reflect"
	"strings"
	"testing"
)

func TestMockHash(t *testing.T) {
	hash := MockHash()
	if hash.IsZero() {
		t.Fatal("create hash failed.")
	} else {
		t.Log(hash.String())
	}
}

func TestMockAddress(t *testing.T) {
	address := MockAddress()
	if !types.IsValidHexAddress(address.String()) {
		t.Fatal("mock address failed")
	}
}

func TestAccountMeta_Token(t *testing.T) {
	addr := MockAddress()
	am := MockAccountMeta(addr)
	token := am.Token(types.Hash{})
	if token != nil {
		t.Fatal("get token failed")
	}

	if len(am.Tokens) > 0 {
		tt := am.Tokens[0].Type
		tm := am.Token(tt)
		if !reflect.DeepEqual(tm, am.Tokens[0]) {
			t.Fatal("get the first token failed")
		}
	}
}

func TestMockAccountMeta(t *testing.T) {
	addr := MockAddress()
	am := MockAccountMeta(addr)
	bytes, err := jsoniter.Marshal(am)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(bytes))

	tm := MockTokenMeta(addr)
	tm.Type = types.Hash{}

	am.Tokens[0] = tm

	t.Log(strings.Repeat("*", 20))
	bytes2, err2 := jsoniter.Marshal(am)
	if err2 != nil {
		t.Fatal(err2)
	}
	t.Log(string(bytes2))
}
