package contract

import (
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"strings"
	"testing"
)

func addKYCTestAdmin(t *testing.T, l *ledger.Ledger, admin *abi.KYCAdminAccount, povHeight uint64) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataAdmin, admin.Account.Bytes())

	data, err := admin.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = csdb.SetValue(trieKey, data)
	if err != nil {
		t.Fatal(err)
	}

	err = gsdb.CommitToTrie()
	if err != nil {
		t.Fatal(err)
	}
	txn := l.DBStore().Batch(true)
	err = gsdb.CommitToDB(txn)
	if err != nil {
		t.Fatal(err)
	}
	err = l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	povBlk.Header.CbTx.StateHash = gsdb.GetCurHash()
	mock.UpdatePovHash(povBlk)

	err = l.AddPovBlock(povBlk, povTd)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(povBlk.GetHeight(), povBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.SetPovLatestHeight(povBlk.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
}

func TestKYCAdminHandOver_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	a := new(KYCAdminHandOver)

	blk.Token = mock.Hash()
	_, _, err := a.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	admin := new(abi.KYCAdminAccount)
	admin.Account = mock.Address()
	admin.Comment = strings.Repeat("x", abi.KYCCommentMaxLen+1)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCAdminHandOver, admin.Account, admin.Comment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidLen {
		t.Fatal(err)
	}

	admin.Comment = strings.Repeat("x", abi.KYCCommentMaxLen)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCAdminHandOver, admin.Account, admin.Comment)
	blk.SetFromSync()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Flag = 0
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	admin = &abi.KYCAdminAccount{
		Account: blk.Address,
		Comment: "right admin",
		Valid:   true,
	}
	addKYCTestAdmin(t, l, admin, 10)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKYCAdminHandOver_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	a := new(KYCAdminHandOver)

	err := a.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	admin := new(abi.KYCAdminAccount)
	admin.Account = mock.Address()
	admin.Comment = strings.Repeat("x", abi.KYCCommentMaxLen)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCAdminHandOver, admin.Account, admin.Comment)
	blk.Address = cfg.GenesisAddress()
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}

	admin.Account = cfg.GenesisAddress()
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCAdminHandOver, admin.Account, admin.Comment)
	blk.Address = cfg.GenesisAddress()
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}

	blk.Address = mock.Address()
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	admin = &abi.KYCAdminAccount{
		Account: blk.Address,
		Comment: "adminComment",
		Valid:   true,
	}
	addKYCTestAdmin(t, l, admin, 10)

	ph, err := l.GetLatestPovHeader()
	if err != nil {
		t.Fatal()
	}
	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), ph.GetStateHash())
	csdb, err = gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}
}

func TestKYCStatusUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	a := new(KYCStatusUpdate)

	blk.Token = mock.Hash()
	_, _, err := a.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	status := new(abi.KYCStatus)
	status.ChainAddress = mock.Address()
	status.Status = "wrong"
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCStatusUpdate, status.ChainAddress, status.Status)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	status.Status = "KYC_STATUS_NOT_STARTED"
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCStatusUpdate, status.ChainAddress, status.Status)
	blk.SetFromSync()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Flag = 0
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	admin := &abi.KYCAdminAccount{
		Account: blk.Address,
		Comment: "admin",
		Valid:   true,
	}
	addKYCTestAdmin(t, l, admin, 10)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKYCStatusUpdate_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	a := new(KYCStatusUpdate)

	err := a.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	status := new(abi.KYCStatus)
	status.Status = "KYC_STATUS_NOT_STARTED"
	status.ChainAddress = mock.Address()
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCStatusUpdate, status.ChainAddress, status.Status)
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}
}

func TestKYCTradeAddressUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	a := new(KYCTradeAddressUpdate)

	blk.Token = mock.Hash()
	_, _, err := a.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	ka := new(abi.KYCAddress)
	ka.ChainAddress = mock.Address()
	ka.TradeAddress = ""
	ka.Action = abi.KYCActionInvalid
	ka.Comment = strings.Repeat("x", abi.KYCCommentMaxLen+1)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, ka.ChainAddress, ka.Action, ka.TradeAddress, ka.Comment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	ka.Action = abi.KYCActionAdd
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, ka.ChainAddress, ka.Action, ka.TradeAddress, ka.Comment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	ka.Comment = strings.Repeat("x", abi.KYCCommentMaxLen)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, ka.ChainAddress, ka.Action, ka.TradeAddress, ka.Comment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	ka.TradeAddress = "0xcd2a3d9f938e13cd947ec05abc7fe734df8dd826"
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, ka.ChainAddress, ka.Action, ka.TradeAddress, ka.Comment)
	blk.SetFromSync()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Flag = 0
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	admin := &abi.KYCAdminAccount{
		Account: blk.Address,
		Comment: "admin",
		Valid:   true,
	}
	addKYCTestAdmin(t, l, admin, 10)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKYCTradeAddressUpdate_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	a := new(KYCTradeAddressUpdate)

	err := a.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	ka := new(abi.KYCAddress)
	ka.ChainAddress = mock.Address()
	ka.TradeAddress = "0xcd2a3d9f938e13cd947ec05abc7fe734df8dd826"
	ka.Action = abi.KYCActionAdd
	ka.Comment = strings.Repeat("x", abi.KYCCommentMaxLen)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, ka.ChainAddress, ka.Action, ka.TradeAddress, ka.Comment)
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}

	ka.Action = abi.KYCActionRemove
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCTradeAddressUpdate, ka.ChainAddress, ka.Action, ka.TradeAddress, ka.Comment)
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}
}

func TestKYCOperatorUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	a := new(KYCOperatorUpdate)

	blk.Token = mock.Hash()
	_, _, err := a.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	koa := new(abi.KYCOperatorAccount)
	koa.Action = abi.KYCActionInvalid
	koa.Comment = strings.Repeat("x", abi.KYCCommentMaxLen+1)
	koa.Account = mock.Address()
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCOperatorUpdate, koa.Account, koa.Action, koa.Comment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	koa.Action = abi.KYCActionAdd
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCOperatorUpdate, koa.Account, koa.Action, koa.Comment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	koa.Comment = strings.Repeat("x", abi.KYCCommentMaxLen)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCOperatorUpdate, koa.Account, koa.Action, koa.Comment)
	blk.SetFromSync()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Flag = 0
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	admin := &abi.KYCAdminAccount{
		Account: blk.Address,
		Comment: "admin",
		Valid:   true,
	}
	addKYCTestAdmin(t, l, admin, 10)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKYCOperatorUpdate_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	blk := mock.StateBlockWithoutWork()
	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	a := new(KYCOperatorUpdate)

	err := a.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	koa := new(abi.KYCOperatorAccount)
	koa.Account = mock.Address()
	koa.Action = abi.KYCActionAdd
	koa.Comment = strings.Repeat("x", abi.KYCCommentMaxLen)
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCOperatorUpdate, koa.Account, koa.Action, koa.Comment)
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}

	koa.Action = abi.KYCActionRemove
	blk.Data, err = abi.KYCStatusABI.PackMethod(abi.MethodNameKYCOperatorUpdate, koa.Account, koa.Action, koa.Comment)
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}
}
