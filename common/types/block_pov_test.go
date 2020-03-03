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

func TestPovHeader_Serialize(t *testing.T) {
	hdr1 := NewPovHeader()
	hdr1.BasHdr.Version = uint32(ALGO_SCRYPT)
	hdr1.BasHdr.Timestamp = uint32(time.Now().Unix())
	hdr1.BasHdr.Nonce = uint32(rand.Int31())
	hdr1.BasHdr.Bits = uint32(rand.Int31())
	hdr1.BasHdr.Hash = hdr1.ComputeHash()
	hdr1.BasHdr.Height = uint64(rand.Int63())

	data, err := hdr1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	hdr2 := NewPovHeader()
	err = hdr2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if hdr1.ComputePowHash() != hdr2.ComputePowHash() {
		t.Fatalf("exp: %v, act: %v", hdr1.ComputePowHash(), hdr2.ComputePowHash())
	}

	if hdr1.GetHash() != hdr2.ComputeHash() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetHash(), hdr2.ComputeHash())
	}

	if hdr1.GetHash() != hdr2.GetHash() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetHash(), hdr2.GetHash())
	}
	if hdr1.GetHeight() != hdr2.GetHeight() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetHeight(), hdr2.GetHeight())
	}
	if hdr1.GetTimestamp() != hdr2.GetTimestamp() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetHeight(), hdr2.GetHeight())
	}
	if hdr1.GetVersion() != hdr2.GetVersion() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetVersion(), hdr2.GetVersion())
	}

	if hdr1.GetAlgoType() != hdr2.GetAlgoType() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetAlgoType(), hdr2.GetAlgoType())
	}
	if hdr1.GetAlgoEfficiency() != hdr2.GetAlgoEfficiency() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetAlgoEfficiency(), hdr2.GetAlgoEfficiency())
	}
	if hdr1.GetAlgoTargetInt().Cmp(hdr2.GetAlgoTargetInt()) != 0 {
		t.Fatalf("exp: %v, act: %v", hdr1.GetAlgoTargetInt(), hdr2.GetAlgoTargetInt())
	}
	if hdr1.GetNormBits() != hdr2.GetNormBits() {
		t.Fatalf("exp: %v, act: %v", hdr1.GetNormBits(), hdr2.GetNormBits())
	}
	if hdr1.GetNormTargetInt().Cmp(hdr2.GetNormTargetInt()) != 0 {
		t.Fatalf("exp: %v, act: %v", hdr1.GetNormTargetInt(), hdr2.GetNormTargetInt())
	}
}

func TestPovBody_Serialize(t *testing.T) {
	bd1 := NewPovBody()
	tx1Hash, _ := NewHash("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1")
	tx1Pov := &PovTransaction{Hash: tx1Hash}
	bd1.Txs = append(bd1.Txs, tx1Pov)

	data, err := bd1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	bd2 := NewPovBody()
	err = bd2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(bd1.Txs) != len(bd2.Txs) {
		t.Fatalf("exp: %v, act: %v", len(bd1.Txs), len(bd2.Txs))
	}
}

func TestPovAuxHeader_Serialize(t *testing.T) {
	aux1 := NewPovAuxHeader()
	aux1.ParBlockHeader.Timestamp = uint32(time.Now().Unix())
	aux1.ParBlockHeader.Nonce = uint32(rand.Int31())
	aux1.ParBlockHeader.Bits = uint32(rand.Int31())
	aux1.ParCoinBaseTx.Version = 1
	aux1.ParCoinBaseTx.TxIn = append(aux1.ParCoinBaseTx.TxIn, &PovBtcTxIn{})
	aux1.ParCoinBaseTx.TxOut = append(aux1.ParCoinBaseTx.TxOut, &PovBtcTxOut{})

	aux1Algo := NewPoVHashAlgoFromStr("SHA256D")

	data, err := aux1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	aux2 := NewPovAuxHeader()
	err = aux2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if aux1.ComputePowHash(aux1Algo) != aux2.ComputePowHash(aux1Algo) {
		t.Fatalf("exp: %v, act: %v", aux1.ComputePowHash(aux1Algo), aux2.ComputePowHash(aux1Algo))
	}
}

func TestPovCoinBaseTx_Serialize(t *testing.T) {
	cbTx1 := NewPovCoinBaseTx(1, 2)

	minerTx1 := cbTx1.GetMinerTxOut()
	minerTx1.Value = NewBalance(rand.Int63())

	repTx1 := cbTx1.GetRepTxOut()
	repTx1.Value = NewBalance(rand.Int63())

	cbTx1.Hash = cbTx1.ComputeHash()

	data, err := cbTx1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	cbTx2 := NewPovCoinBaseTx(1, 2)
	err = cbTx2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if cbTx1.ComputeHash() != cbTx2.ComputeHash() {
		t.Fatalf("exp: %v, act: %v", cbTx1.ComputeHash(), cbTx2.ComputeHash())
	}

	if cbTx1.GetHash() != cbTx2.GetHash() {
		t.Fatalf("exp: %v, act: %v", cbTx1.GetHash(), cbTx2.GetHash())
	}
	if cbTx1.GetMinerTxOut().Value.Compare(cbTx2.GetMinerTxOut().Value) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", cbTx1.GetMinerTxOut().Value, cbTx2.GetMinerTxOut().Value)
	}
	if cbTx1.GetRepTxOut().Value.Compare(cbTx2.GetRepTxOut().Value) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", cbTx1.GetRepTxOut().Value, cbTx2.GetRepTxOut().Value)
	}

	cbTxIn1 := new(PovCoinBaseTxIn)
	cbTxIn1.Sequence = uint32(rand.Int31())
	dataTxIn, err := cbTxIn1.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	cbTxIn2 := new(PovCoinBaseTxIn)
	err = cbTxIn2.Deserialize(dataTxIn)
	if err != nil {
		t.Fatal(err)
	}
	if cbTxIn1.Sequence != cbTxIn2.Sequence {
		t.Fatalf("exp: %v, act: %v", cbTxIn1.Sequence, cbTxIn2.Sequence)
	}

	cbTxOut1 := new(PovCoinBaseTxOut)
	cbTxOut1.Value = NewBalance(rand.Int63())
	dataTxOut, err := cbTxOut1.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	cbTxOut2 := new(PovCoinBaseTxOut)
	err = cbTxOut2.Deserialize(dataTxOut)
	if err != nil {
		t.Fatal(err)
	}
	if cbTxOut1.Value.Compare(cbTxOut2.Value) != BalanceCompEqual {
		t.Fatalf("exp: %v, act: %v", cbTxOut1.Value, cbTxOut2.Value)
	}
}

func TestPovTxLookup_Serialize(t *testing.T) {
	txl1 := &PovTxLookup{}
	txl1.BlockHeight = uint64(rand.Int63())
	txl1.TxIndex = uint64(rand.Int31())

	data, err := txl1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	txl2 := &PovTxLookup{}
	err = txl2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if txl1.BlockHeight != txl2.BlockHeight {
		t.Fatalf("exp: %v, act: %v", txl1.BlockHeight, txl2.BlockHeight)
	}
	if txl1.TxIndex != txl2.TxIndex {
		t.Fatalf("exp: %v, act: %v", txl1.TxIndex, txl2.TxIndex)
	}
}

func TestPovBlock_Serialize(t *testing.T) {
	hdr1 := NewPovHeader()
	hdr1.BasHdr.Version = uint32(ALGO_SHA256D)
	hdr1.BasHdr.Timestamp = uint32(time.Now().Unix())
	hdr1.BasHdr.Nonce = uint32(rand.Int31())
	hdr1.BasHdr.Bits = uint32(rand.Int31())
	hdr1.BasHdr.Hash = hdr1.ComputeHash()
	hdr1.BasHdr.Height = uint64(rand.Int63())

	bd1 := NewPovBody()
	tx1Hash, _ := NewHash("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1")
	tx1Pov := &PovTransaction{Hash: tx1Hash}
	bd1.Txs = append(bd1.Txs, tx1Pov)

	blk1 := NewPovBlockWithBody(hdr1, bd1)

	data, err := blk1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	blk2 := NewPovBlock()
	err = blk2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	hdr2 := blk2.GetHeader()
	//bd2 := blk2.GetBody()

	if hdr1.ComputePowHash() != hdr2.ComputePowHash() {
		t.Fatalf("exp: %v, act: %v", hdr1.ComputePowHash(), hdr2.ComputePowHash())
	}

	if blk2.GetHash() != blk2.ComputeHash() {
		t.Fatalf("exp: %v, act: %v", blk2.GetHash(), blk2.ComputeHash())
	}

	if blk1.GetHash() != blk2.GetHash() {
		t.Fatalf("exp: %v, act: %v", blk1.GetHash(), blk2.GetHash())
	}
	if blk1.GetHeight() != blk2.GetHeight() {
		t.Fatalf("exp: %v, act: %v", blk1.GetHeight(), blk2.GetHeight())
	}
	if blk1.GetTimestamp() != blk2.GetTimestamp() {
		t.Fatalf("exp: %v, act: %v", blk1.GetHeight(), blk2.GetHeight())
	}
	if blk1.GetVersion() != blk2.GetVersion() {
		t.Fatalf("exp: %v, act: %v", blk1.GetVersion(), blk2.GetVersion())
	}
	if blk1.GetPrevious() != blk2.GetPrevious() {
		t.Fatalf("exp: %v, act: %v", blk1.GetPrevious(), blk2.GetPrevious())
	}
	if blk1.GetMerkleRoot() != blk2.GetMerkleRoot() {
		t.Fatalf("exp: %v, act: %v", blk1.GetMerkleRoot(), blk2.GetMerkleRoot())
	}

	if blk1.GetTxNum() != blk2.GetTxNum() {
		t.Fatalf("exp: %v, act: %v", blk1.GetTxNum(), blk2.GetTxNum())
	}
	if len(blk1.GetAllTxs()) != len(blk2.GetAllTxs()) {
		t.Fatalf("exp: %v, act: %v", len(blk1.GetAllTxs()), len(blk2.GetAllTxs()))
	}
	if len(blk1.GetAccountTxs()) != len(blk2.GetAccountTxs()) {
		t.Fatalf("exp: %v, act: %v", len(blk1.GetAccountTxs()), len(blk2.GetAccountTxs()))
	}

	if blk1.GetAlgoType() != blk2.GetAlgoType() {
		t.Fatalf("exp: %v, act: %v", blk1.GetAlgoType(), blk2.GetAlgoType())
	}
	if blk1.GetAlgoEfficiency() != blk2.GetAlgoEfficiency() {
		t.Fatalf("exp: %v, act: %v", blk1.GetAlgoEfficiency(), blk2.GetAlgoEfficiency())
	}

	blk3 := blk1.Copy()
	if blk1.GetHash() != blk3.GetHash() {
		t.Fatalf("exp: %v, act: %v", blk1.GetHash(), blk3.GetHash())
	}

	blk4 := blk1.Clone()
	if blk1.GetHash() != blk4.GetHash() {
		t.Fatalf("exp: %v, act: %v", blk1.GetHash(), blk4.GetHash())
	}

	var blkList1 PovBlocks
	blkList1 = append(blkList1, blk1)
	dataBs, err := blkList1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	var blkList2 PovBlocks
	err = blkList2.Deserialize(dataBs)
	if err != nil {
		t.Fatal(err)
	}

	if len(blkList1) != len(blkList2) {
		t.Fatalf("exp: %v, act: %v", len(blkList1), len(blkList2))
	}
}

func TestPovTD_Serialize(t *testing.T) {
	td1 := NewPovTD()
	td1.Chain.Add(NewBigNumFromInt(0), NewBigNumFromInt(rand.Int63()))

	data, err := td1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	td2 := NewPovTD()
	err = td2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if td1.Chain.Uint64() != td2.Chain.Uint64() {
		t.Fatalf("exp: %v, act: %v", td1.Chain.Uint64(), td2.Chain.Uint64())
	}

	td3 := td1.Copy()
	if td1.Chain.Uint64() != td3.Chain.Uint64() {
		t.Fatalf("exp: %v, act: %v", td1.Chain.Uint64(), td3.Chain.Uint64())
	}
}

func TestPovMinerDayStat_Serialize(t *testing.T) {
	ds1 := NewPovMinerDayStat()
	ds1.DayIndex = uint32(rand.Int31())
	ds1.MinerStats["test1"] = NewPovMinerStatItem()
	ds1.MinerStats["test1"].BlockNum = uint32(rand.Int31())

	data, err := ds1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	ds2 := NewPovMinerDayStat()
	err = ds2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if ds1.DayIndex != ds2.DayIndex {
		t.Fatalf("exp: %v, act: %v", ds1.DayIndex, ds2.DayIndex)
	}
	if ds1.MinerStats["test1"].BlockNum != ds2.MinerStats["test1"].BlockNum {
		t.Fatalf("exp: %v, act: %v", ds1.MinerStats["test1"].BlockNum, ds2.MinerStats["test1"].BlockNum)
	}
}

func TestPovDiffDayStat_Serialize(t *testing.T) {
	ds1 := NewPovDiffDayStat()
	ds1.DayIndex = uint32(rand.Int31())
	ds1.AvgDiffRatio = uint64(rand.Int63())

	data, err := ds1.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	ds2 := NewPovDiffDayStat()
	err = ds2.Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}

	if ds1.DayIndex != ds2.DayIndex {
		t.Fatalf("exp: %v, act: %v", ds1.DayIndex, ds2.DayIndex)
	}
	if ds1.AvgDiffRatio != ds2.AvgDiffRatio {
		t.Fatalf("exp: %v, act: %v", ds1.AvgDiffRatio, ds2.AvgDiffRatio)
	}
}
