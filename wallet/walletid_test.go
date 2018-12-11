/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"github.com/json-iterator/go"
	"reflect"
	"testing"
)

func TestWalletId(t *testing.T) {
	id := NewWalletId()
	bytes, err := jsoniter.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(id)

	var id2 WalletId
	err = jsoniter.Unmarshal(bytes, &id2)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(id2)
	if !reflect.DeepEqual(id, id2) {
		t.Fatal("id != id2")
	}
}

func TestWalletIdManager(t *testing.T) {
	id, err := String2WalletId("fd6548f2-d029-4d15-92b4-529bbfdf7bca")
	if err != nil {
		t.Fatal(err)
	}
	ids := []WalletId{NewWalletId(), NewWalletId(), id, NewWalletId(), NewWalletId()}

	for i, val := range ids {
		t.Log(i, val.String())
	}

	i, err := indexOf(ids, id)

	if err != nil {
		t.Fatal(err)
	}
	t.Log("id", id, "pos: ", i)

	ids, err = remove(ids, id)
	if err != nil {
		t.Fatal(err)
	}
	for i, val := range ids {
		t.Log(i, val.String())
	}
}
