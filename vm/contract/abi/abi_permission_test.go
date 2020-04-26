package abi

import (
	"strings"
	"testing"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func addTestAdmin(t *testing.T, l *ledger.Ledger, admin *AdminAccount, povHeight uint64) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PermissionAddress)
	if err != nil {
		t.Fatal(err)
	}

	trieKey := statedb.PovCreateContractLocalStateKey(PermissionDataAdmin, admin.Account.Bytes())

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

func addTestNode(t *testing.T, l *ledger.Ledger, pn *PermNode, povHeight uint64) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PermissionAddress)
	if err != nil {
		t.Fatal(err)
	}

	trieKey := statedb.PovCreateContractLocalStateKey(PermissionDataNode, []byte(pn.NodeId))

	data, err := pn.MarshalMsg(nil)
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

func TestPermissionABI(t *testing.T) {
	_, err := abi.JSONToABIContract(strings.NewReader(JsonPermission))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPermissionIsAdmin(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.PermissionAddress)
	account := cfg.GenesisAddress()

	if PermissionIsAdmin(ctx, account) {
		t.Fatal()
	}

	admin := &AdminAccount{
		Account: mock.Address(),
		Comment: "a1",
		Valid:   false,
	}
	addTestAdmin(t, l, admin, 10)

	if !PermissionIsAdmin(ctx, account) {
		t.Fatal()
	}

	account = mock.Address()
	if PermissionIsAdmin(ctx, account) {
		t.Fatal()
	}

	if PermissionIsAdmin(ctx, admin.Account) {
		t.Fatal()
	}

	admin = &AdminAccount{
		Account: account,
		Comment: "a1",
		Valid:   true,
	}
	addTestAdmin(t, l, admin, 10)
	if !PermissionIsAdmin(ctx, account) {
		t.Fatal()
	}

	account = cfg.GenesisAddress()
	if PermissionIsAdmin(ctx, account) {
		t.Fatal()
	}
}

func TestPermissionGetAdmin(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.PermissionAddress)

	_, err := PermissionGetAdmin(ctx)
	if err == nil {
		t.Fatal()
	}

	admin := &AdminAccount{
		Account: mock.Address(),
		Comment: "a1",
		Valid:   false,
	}
	addTestAdmin(t, l, admin, 10)

	ac, err := PermissionGetAdmin(ctx)
	if err != nil || len(ac) != 1 || ac[0].Account != cfg.GenesisAddress() {
		t.Fatal()
	}

	admin = &AdminAccount{
		Account: mock.Address(),
		Comment: "a1",
		Valid:   true,
	}
	addTestAdmin(t, l, admin, 10)

	ac, err = PermissionGetAdmin(ctx)
	if err != nil || len(ac) != 1 || ac[0].Account != admin.Account {
		t.Fatal()
	}
}

func TestPermissionUpdateNode(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.PermissionAddress)
	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PermissionAddress)
	if err != nil {
		t.Fatal(err)
	}

	nodeId := "id1"
	nodeUrl := "1.1.1.1:8000"
	comment := "222"
	err = PermissionUpdateNode(csdb, nodeId, nodeUrl, comment)
	if err != nil {
		t.Fatal()
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

	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
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

	pn, err := PermissionGetNode(ctx, nodeId)
	if err != nil || pn.NodeId != nodeId || pn.NodeUrl != nodeUrl || pn.Comment != comment {
		t.Fatal()
	}
}

func TestPermissionGetNode(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.PermissionAddress)
	nodeId := "123"
	_, err := PermissionGetNode(ctx, nodeId)
	if err == nil {
		t.Fatal()
	}

	pn := &PermNode{
		NodeId:  "000",
		NodeUrl: "1.1.1.1:8000",
		Comment: "222",
		Valid:   false,
	}
	addTestNode(t, l, pn, 10)

	_, err = PermissionGetNode(ctx, nodeId)
	if err == nil {
		t.Fatal()
	}

	pn = &PermNode{
		NodeId:  nodeId,
		NodeUrl: "1.1.1.1:8000",
		Comment: "222",
		Valid:   false,
	}
	addTestNode(t, l, pn, 10)

	_, err = PermissionGetNode(ctx, nodeId)
	if err == nil {
		t.Fatal()
	}

	pn = &PermNode{
		NodeId:  nodeId,
		NodeUrl: "1.1.1.1:8000",
		Comment: "222",
		Valid:   true,
	}
	addTestNode(t, l, pn, 10)

	_, err = PermissionGetNode(ctx, nodeId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPermissionGetAllNodes(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	_, err := PermissionGetAllNodes(l)
	if err == nil {
		t.Fatal()
	}

	pn := &PermNode{
		NodeId:  "n1",
		NodeUrl: "1.1.1.1:8000",
		Comment: "222",
		Valid:   false,
	}
	addTestNode(t, l, pn, 10)

	pns, err := PermissionGetAllNodes(l)
	if err != nil || len(pns) != 0 {
		t.Fatal()
	}

	pn = &PermNode{
		NodeId:  "n1",
		NodeUrl: "1.1.1.1:8000",
		Comment: "222",
		Valid:   true,
	}
	addTestNode(t, l, pn, 10)

	pns, err = PermissionGetAllNodes(l)
	if err != nil || len(pns) != 1 {
		t.Fatal()
	}
}
