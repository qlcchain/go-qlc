package contract

import (
	"strings"
	"testing"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/ledger"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func addTestAdmin(t *testing.T, l *ledger.Ledger, admin *abi.AdminAccount, povHeight uint64) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PermissionAddress)
	if err != nil {
		t.Fatal(err)
	}

	trieKey := statedb.PovCreateContractLocalStateKey(abi.PermissionDataAdmin, admin.Account.Bytes())

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

func TestAdminUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	a := new(AdminHandOver)

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

	newAdminAddr := mock.Address()
	newAdminComment := strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminHandOver, newAdminAddr, newAdminComment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidLen {
		t.Fatal(err)
	}

	newAdminComment = strings.Repeat("x", abi.PermissionCommentMaxLen)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminHandOver, newAdminAddr, newAdminComment)
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

	admin := &abi.AdminAccount{
		Account: blk.Address,
		Comment: "right admin",
		Valid:   true,
	}
	addTestAdmin(t, l, admin, 10)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdminHandOver_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	a := new(AdminHandOver)

	err := a.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	adminAddr := mock.Address()
	adminComment := strings.Repeat("x", abi.PermissionCommentMaxLen)
	blk.Data, _ = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminHandOver, adminAddr, adminComment)
	blk.Address = cfg.GenesisAddress()
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal(err)
	}

	adminAddr = cfg.GenesisAddress()
	blk.Data, _ = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminHandOver, adminAddr, adminComment)
	blk.Address = cfg.GenesisAddress()
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Address = mock.Address()
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	admin := &abi.AdminAccount{
		Account: blk.Address,
		Comment: adminComment,
		Valid:   true,
	}
	addTestAdmin(t, l, admin, 10)

	ph, err := l.GetLatestPovHeader()
	if err != nil {
		t.Fatal()
	}
	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), ph.GetStateHash())
	csdb, err = gsdb.LookupContractStateDB(contractaddress.PermissionAddress)
	err = a.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}
}

func TestNodeUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	n := new(NodeUpdate)

	blk.Token = mock.Hash()
	_, _, err := n.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	nodeId := "n1"
	nodeUrl := "123:1:2"
	comment := strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, nodeId, nodeUrl, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	nodeUrl = "123"
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, nodeId, nodeUrl, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	nodeUrl = "1.1.1.1:8000000"
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, nodeId, nodeUrl, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	nodeUrl = "1.1.1.1:8000"
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, nodeId, nodeUrl, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	comment = strings.Repeat("x", abi.PermissionCommentMaxLen)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, nodeId, nodeUrl, comment)
	blk.SetFromSync()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Flag = 0
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	admin := &abi.AdminAccount{
		Account: blk.Address,
		Comment: "a1",
		Valid:   true,
	}
	addTestAdmin(t, l, admin, 10)

	_, _, err = n.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNodeUpdate_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	n := new(NodeUpdate)

	err := n.DoSendOnPov(ctx, csdb, 10, blk)
	if err == nil {
		t.Fatal()
	}

	blk.Data, _ = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, "n1", "", "")
	err = n.DoSendOnPov(ctx, csdb, 10, blk)
	if err != nil {
		t.Fatal()
	}
}
