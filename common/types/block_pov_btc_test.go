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
	"time"
)

func TestPovBtcHeader_Serialize(t *testing.T) {
	hdr1 := new(PovBtcHeader)
	hdr1.Version = uint32(rand.Int31())
	hdr1.Timestamp = uint32(time.Now().Unix())
	hdr1.Nonce = uint32(rand.Int31())
	hdr1.Bits = uint32(rand.Int31())

	data, err := hdr1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	hdr2 := new(PovBtcHeader)
	err = hdr2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	algo1 := NewPoVHashAlgoFromStr("X11")
	if hdr1.ComputePowHash(algo1) != hdr2.ComputePowHash(algo1) {
		t.Fatalf("exp: %v, act: %v", hdr1.ComputePowHash(algo1), hdr2.ComputePowHash(algo1))
	}

	algo2 := NewPoVHashAlgoFromStr("SCRYPT")
	if hdr1.ComputePowHash(algo2) != hdr2.ComputePowHash(algo2) {
		t.Fatalf("exp: %v, act: %v", hdr1.ComputePowHash(algo2), hdr2.ComputePowHash(algo2))
	}

	algo3 := NewPoVHashAlgoFromStr("SHA256D")
	if hdr1.ComputePowHash(algo3) != hdr2.ComputePowHash(algo3) {
		t.Fatalf("exp: %v, act: %v", hdr1.ComputePowHash(algo3), hdr2.ComputePowHash(algo3))
	}

	if hdr1.ComputeHash() != hdr2.ComputeHash() {
		t.Fatalf("exp: %v, act: %v", hdr1.ComputeHash(), hdr2.ComputeHash())
	}
}

func TestPovBtcTx_Serialize(t *testing.T) {
	txInList := make([]*PovBtcTxIn, 0)
	txIn1 := new(PovBtcTxIn)
	txIn1.Sequence = uint32(rand.Int31())

	dataTxIn, err := txIn1.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	txIn11 := new(PovBtcTxIn)
	err = txIn11.Deserialize(dataTxIn)
	if err != nil {
		t.Fatal(err)
	}
	if txIn1.Sequence != txIn11.Sequence {
		t.Fatalf("exp: %v, act: %v", txIn1.Sequence, txIn11.Sequence)
	}

	txInList = append(txInList, txIn1)

	txOutList := make([]*PovBtcTxOut, 0)
	txOut1 := new(PovBtcTxOut)
	txOut1.Value = rand.Int63()

	dataTxOut, err := txOut1.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	txOut11 := new(PovBtcTxOut)
	err = txOut11.Deserialize(dataTxOut)
	if err != nil {
		t.Fatal(err)
	}
	if txOut1.Value != txOut11.Value {
		t.Fatalf("exp: %v, act: %v", txOut1.Value, txOut11.Value)
	}

	txOutList = append(txOutList, txOut1)

	tx1 := NewPovBtcTx(txInList, txOutList)

	data, err := tx1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	tx2 := NewPovBtcTx(txInList, txOutList)
	err = tx2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if tx1.ComputeHash() != tx2.ComputeHash() {
		t.Fatalf("exp: %v, act: %v", tx1.ComputeHash(), tx2.ComputeHash())
	}
}
