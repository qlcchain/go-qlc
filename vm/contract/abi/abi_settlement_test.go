/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

import (
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/mock"
)

func TestCreateContractParam(t *testing.T) {
	param := CreateContractParam{
		PartyA:      mock.Address(),
		PartyAName:  "c1",
		PartyB:      mock.Address(),
		PartyBName:  "c2",
		PreHash:     mock.Hash(),
		ServiceId:   mock.Hash().String(),
		MCC:         1,
		MNC:         2,
		TotalAmount: 100,
		UnitPrice:   1,
		Currency:    "USD",
		SignDate:    time.Now().Unix(),
		SignatureA:  nil,
	}
	a1 := mock.Account()
	if err := param.Sign(a1); err != nil {
		t.Fatal(err)
	}
	t.Log(param.String())
	addr1, err := param.Address()
	if err != nil {
		t.Fatal(err)
	}

	cp := ContractParam{
		CreateContractParam: param,
		ConfirmDate:         time.Now().Unix() + 100,
	}

	a2 := mock.Account()
	if err = cp.Sign(a2); err != nil {
		t.Fatal(err)
	}
	addr2, err := cp.Address()
	if err != nil {
		t.Fatal(err)
	}
	if addr1 != addr2 {
		t.Fatalf("invalid addr,%s==>%s", addr1, addr2)
	}
	t.Log(cp.String())

	msg, err := param.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	param2 := &CreateContractParam{}
	_, err = param2.UnmarshalMsg(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(param, *param2) {
		t.Fatalf("invaid param,%s,%s", param.String(), param2.String())
	}

	msg, err = cp.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	cp2 := &ContractParam{}
	_, err = cp2.UnmarshalMsg(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cp, *cp2) {
		t.Fatalf("invaid param,%s,%s", cp.String(), cp2.String())
	}
}
