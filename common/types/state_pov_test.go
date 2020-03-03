/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"math/rand"
	"testing"
)

func TestPovAccountState_Serialize(t *testing.T) {
	as1 := NewPovAccountState()
	as1.Balance = NewBalance(rand.Int63())

	t.Log(as1.String())

	data, err := as1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	as2 := NewPovAccountState()
	err = as2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(as1.String())

	if as1.Balance.Compare(as2.Balance) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", as1.Balance, as2.Balance)
	}

	as3 := as1.Clone()
	if as1.Balance.Compare(as3.Balance) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", as1.Balance, as3.Balance)
	}
}

func TestPovTokenState_Serialize(t *testing.T) {
	tHash, _ := NewHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4321")
	ts1 := NewPovTokenState(tHash)
	ts1.Balance = NewBalance(rand.Int63())

	t.Log(ts1.String())

	data, err := ts1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	ts2 := NewPovTokenState(tHash)
	err = ts2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ts2.String())

	if ts1.Balance.Compare(ts2.Balance) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", ts1.Balance, ts2.Balance)
	}

	ts3 := ts1.Clone()
	if ts1.Balance.Compare(ts3.Balance) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", ts1.Balance, ts3.Balance)
	}
}

func TestPovRepState_Serialize(t *testing.T) {
	rs1 := NewPovRepState()
	rs1.Vote = NewBalance(rand.Int63())

	data, err := rs1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	rs2 := NewPovRepState()
	err = rs2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if rs1.Vote.Compare(rs2.Vote) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", rs1.Vote, rs2.Vote)
	}

	rs3 := rs1.Clone()
	if rs1.Vote.Compare(rs3.Vote) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", rs1.Vote, rs3.Vote)
	}
}

func TestPovContractState_Serialize(t *testing.T) {
	cs1 := NewPovContractState()
	cs1.StateHash, _ = NewHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4321")

	data, err := cs1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	cs2 := NewPovContractState()
	err = cs2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if cs1.StateHash != cs2.StateHash {
		t.Fatalf("exp: %v, act: %v", cs1.StateHash, cs2.StateHash)
	}
}

func TestPovPublishState_Serialize(t *testing.T) {
	ps1 := NewPovPublishState()
	ps1.PublishHeight = uint64(rand.Int63())

	data, err := ps1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	ps2 := NewPovPublishState()
	err = ps2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if ps1.PublishHeight != ps2.PublishHeight {
		t.Fatalf("exp: %v, act: %v", ps1.PublishHeight, ps2.PublishHeight)
	}
}

func TestPovVerifierState_Serialize(t *testing.T) {
	vs1 := NewPovVerifierState()
	vs1.TotalVerify = uint64(rand.Int63())

	data, err := vs1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	vs2 := NewPovVerifierState()
	err = vs2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if vs1.TotalVerify != vs2.TotalVerify {
		t.Fatalf("exp: %v, act: %v", vs1.TotalVerify, vs2.TotalVerify)
	}
}
