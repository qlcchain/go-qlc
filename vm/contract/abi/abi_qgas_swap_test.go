package abi

import (
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestQGasPledgeParam(t *testing.T) {
	pledgeParam := QGasPledgeParam{
		FromAddress: mock.Address(),
		Amount:      big.NewInt(1000),
		ToAddress:   mock.Address(),
	}
	bs, err := pledgeParam.ToABI()
	if err != nil {
		t.Fatal(err)
	}
	param, err := ParseQGasPledgeParam(bs)
	if err != nil {
		t.Fatal(err)
	}
	if b, err := param.Verify(); !b || err != nil {
		t.Fatal(err)
	}
	if param.Amount.Cmp(pledgeParam.Amount) != 0 || param.FromAddress != pledgeParam.FromAddress {
		t.Fatal(err)
	}

	if _, err := ParseQGasPledgeParam([]byte{}); err == nil {
		t.Fatal(err)
	}
	if _, err := ParseQGasPledgeParam([]byte{10}); err == nil {
		t.Fatal(err)
	}
}

func TestQGasWithdrawParam(t *testing.T) {
	withdrawParam := QGasWithdrawParam{
		FromAddress: mock.Address(),
		Amount:      big.NewInt(1000),
		ToAddress:   mock.Address(),
		LinkHash:    mock.Hash(),
	}
	bs, err := withdrawParam.ToABI()
	if err != nil {
		t.Fatal(err)
	}
	param, err := ParseQGasWithdrawParam(bs)
	if err != nil {
		t.Fatal(err)
	}
	if b, err := param.Verify(); !b || err != nil {
		t.Fatal(err)
	}
	if param.Amount.Cmp(withdrawParam.Amount) != 0 || param.ToAddress != withdrawParam.ToAddress {
		t.Fatal(err)
	}

	if _, err := ParseQGasWithdrawParam([]byte{}); err == nil {
		t.Fatal(err)
	}
	if _, err := ParseQGasWithdrawParam([]byte{10}); err == nil {
		t.Fatal(err)
	}
}

func TestQGasSwapInfo(t *testing.T) {
	swapInfo := QGasSwapInfo{
		Amount:      big.NewInt(1000),
		FromAddress: mock.Address(),
		ToAddress:   mock.Address(),
		SendHash:    mock.Hash(),
		LinkHash:    mock.Hash(),
		SwapType:    0,
		Time:        time.Now().Unix(),
	}
	bs, err := swapInfo.ToABI()
	if err != nil {
		t.Fatal(err)
	}
	param, err := ParseQGasSwapInfo(bs)
	if err != nil {
		t.Fatal(err)
	}
	if param.Amount.Cmp(swapInfo.Amount) != 0 || param.FromAddress != swapInfo.FromAddress ||
		param.SendHash != swapInfo.SendHash {
		t.Fatal(err)
	}

	if _, err := ParseQGasSwapInfo([]byte{}); err == nil {
		t.Fatal(err)
	}
	if _, err := ParseQGasSwapInfo([]byte{10}); err == nil {
		t.Fatal(err)
	}

	clean, l := ledger.NewTestLedger()
	defer clean()

	ctx := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	if err != nil {
		t.Fatal(err)
	}
	data, err := swapInfo.ToABI()

	if err := ctx.SetStorage(nil, GetQGasSwapKey(swapInfo.FromAddress, swapInfo.SendHash),
		data); err != nil {
		t.Fatal(err)
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	// store swapinfo 2
	swapInfo2 := QGasSwapInfo{
		Amount:      big.NewInt(1000),
		FromAddress: mock.Address(),
		ToAddress:   mock.Address(),
		SendHash:    mock.Hash(),
		LinkHash:    mock.Hash(),
		SwapType:    1,
		Time:        time.Now().Unix(),
	}
	ctx2 := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	if err != nil {
		t.Fatal(err)
	}
	data2, err := swapInfo2.ToABI()
	if err := ctx2.SetStorage(nil, GetQGasSwapKey(swapInfo2.FromAddress, swapInfo2.SendHash),
		data2); err != nil {
		t.Fatal(err)
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx2)); err != nil {
		t.Fatal(err)
	}

	// check
	if r, err := GetQGasSwapInfos(l, types.ZeroAddress, QGasPledge); err != nil || len(r) != 1 {
		t.Fatal(err, len(r))
	}
	if r, err := GetQGasSwapInfos(l, types.ZeroAddress, QGasWithdraw); err != nil || len(r) != 1 {
		t.Fatal(err, len(r))
	}
	if r, err := GetQGasSwapInfos(l, types.ZeroAddress, QGasAll); err != nil || len(r) != 2 {
		t.Fatal(err, len(r))
	}
	if r, err := GetQGasSwapInfos(l, swapInfo.FromAddress, QGasAll); err != nil || len(r) != 1 {
		t.Fatal(err, len(r))
	}
	if r, err := GetQGasSwapInfos(l, swapInfo2.FromAddress, QGasAll); err != nil || len(r) != 1 {
		t.Fatal(err, len(r))
	}
	if r, err := GetQGasSwapInfos(l, mock.Address(), QGasAll); err != nil || len(r) != 0 {
		t.Fatal(err, len(r))
	}

	if _, err := GetQGasSwapAmount(l, types.ZeroAddress); err != nil {
		t.Fatal(err)
	}
	if _, err := GetQGasSwapAmount(l, swapInfo.FromAddress); err != nil {
		t.Fatal(err)
	}
	if _, err := GetQGasSwapAmount(l, swapInfo2.FromAddress); err != nil {
		t.Fatal(err)
	}
}
