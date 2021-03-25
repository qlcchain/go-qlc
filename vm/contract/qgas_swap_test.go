package contract

import (
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	mAccount1    = mock.Account3
	mAddress1    = mAccount1.Address()
	mAccount2    = mock.Account2
	mAddress2    = mAccount2.Address()
	pledgeAmount = types.Balance{Int: big.NewInt(1000000)}
)

func TestQGasPledge_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	if err := updateBlock(l, &mock.TestQGasOpenBlock); err != nil {
		t.Fatal(err)
	}
	l.Flush()

	_, err := l.GetTokenMeta(mAddress1, cfg.GasToken())
	if err != nil {
		t.Fatal(err)
	}
	pledgeParam := abi.QGasPledgeParam{
		FromAddress: mAddress1,
		Amount:      pledgeAmount.Int,
		ToAddress:   mAddress2,
	}
	data, err := pledgeParam.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	q := new(QGasPledge)
	ctx := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	sendBlk := &types.StateBlock{
		Type:      types.ContractSend,
		Token:     cfg.GasToken(),
		Address:   pledgeParam.FromAddress,
		Balance:   types.Balance{Int: pledgeParam.Amount},
		Data:      data,
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}

	if _, _, err := q.ProcessSend(ctx, sendBlk); err != nil {
		t.Fatal(err)
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	// receive - reward addr qgas not found
	ctx1 := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	receiveBlk := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     cfg.GasToken(),
		Link:      sendBlk.GetHash(),
		Balance:   types.Balance{Int: pledgeParam.Amount},
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}
	if _, err := q.DoReceive(ctx1, receiveBlk, sendBlk); err != nil {
		t.Fatal(err)
	}

	// receive - reward addr found, but not qgas
	am2 := mock.AccountMeta(pledgeParam.ToAddress)
	if err := l.AddAccountMeta(am2, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	l.Flush()
	ctx2 := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	receiveBlk2 := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     cfg.GasToken(),
		Link:      sendBlk.GetHash(),
		Balance:   types.Balance{Int: pledgeParam.Amount},
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}
	if _, err := q.DoReceive(ctx2, receiveBlk2, sendBlk); err != nil {
		t.Fatal(err)
	}

	// receive - reward addr found, but not qgas
	tm3 := mock.TokenMeta2(pledgeParam.ToAddress, cfg.GasToken())
	am3 := mock.AccountMeta(pledgeParam.ToAddress)
	am3.Tokens = append(am3.Tokens, tm3)
	if err := l.UpdateAccountMeta(am3, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	l.Flush()
	ctx3 := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	receiveBlk3 := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     cfg.GasToken(),
		Link:      sendBlk.GetHash(),
		Balance:   types.Balance{Int: pledgeParam.Amount},
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}
	if _, err := q.DoReceive(ctx3, receiveBlk3, sendBlk); err != nil {
		t.Fatal(err)
	}
}

func TestQGasWithdraw_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	if err := updateBlock(l, &mock.TestQGasOpenBlock); err != nil {
		t.Fatal(err)
	}
	l.Flush()

	_, err := l.GetTokenMeta(mAddress1, cfg.GasToken())
	if err != nil {
		t.Fatal(err)
	}
	withdrawParam := abi.QGasWithdrawParam{
		ToAddress:   mAddress2,
		Amount:      pledgeAmount.Int,
		FromAddress: mAddress1,
		LinkHash:    mock.Hash(),
	}
	data, err := withdrawParam.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	q := new(QGasWithdraw)
	ctx := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	sendBlk := &types.StateBlock{
		Type:      types.ContractSend,
		Token:     cfg.GasToken(),
		Address:   withdrawParam.FromAddress,
		Balance:   types.Balance{Int: withdrawParam.Amount},
		Data:      data,
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}

	if _, _, err := q.ProcessSend(ctx, sendBlk); err != nil {
		t.Fatal(err)
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	// receive - reward addr qgas not found
	ctx1 := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	receiveBlk := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     cfg.GasToken(),
		Link:      sendBlk.GetHash(),
		Balance:   types.Balance{Int: withdrawParam.Amount},
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}
	if _, err := q.DoReceive(ctx1, receiveBlk, sendBlk); err != nil {
		t.Fatal(err)
	}

	// receive - reward addr found, but not qgas
	am2 := mock.AccountMeta(withdrawParam.ToAddress)
	if err := l.AddAccountMeta(am2, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	l.Flush()
	ctx2 := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	receiveBlk2 := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     cfg.GasToken(),
		Link:      sendBlk.GetHash(),
		Balance:   types.Balance{Int: withdrawParam.Amount},
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}
	if _, err := q.DoReceive(ctx2, receiveBlk2, sendBlk); err != nil {
		t.Fatal(err)
	}

	// receive - reward addr found, but not qgas
	tm3 := mock.TokenMeta2(withdrawParam.ToAddress, cfg.GasToken())
	am3 := mock.AccountMeta(withdrawParam.ToAddress)
	am3.Tokens = append(am3.Tokens, tm3)
	if err := l.UpdateAccountMeta(am3, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	l.Flush()
	ctx3 := vmstore.NewVMContext(l, &contractaddress.QGasSwapAddress)
	receiveBlk3 := &types.StateBlock{
		Type:      types.ContractReward,
		Token:     cfg.GasToken(),
		Link:      sendBlk.GetHash(),
		Balance:   types.Balance{Int: withdrawParam.Amount},
		PoVHeight: 0,
		Timestamp: time.Now().Unix(),
	}
	if _, err := q.DoReceive(ctx3, receiveBlk3, sendBlk); err != nil {
		t.Fatal(err)
	}
}
