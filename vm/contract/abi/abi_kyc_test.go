package abi

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func addKYCTestAdmin(t *testing.T, l *ledger.Ledger, admin *KYCAdminAccount, povHeight uint64) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	trieKey := statedb.PovCreateContractLocalStateKey(KYCDataAdmin, admin.Account.Bytes())

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

func updateKYCStatus(t *testing.T, l *ledger.Ledger, status *KYCStatus, csdb *statedb.PovContractStateDB, gsdb *statedb.PovGlobalStateDB) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 10

	trieKey := statedb.PovCreateContractLocalStateKey(KYCDataStatus, status.ChainAddress.Bytes())

	data, err := status.MarshalMsg(nil)
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

func updateKYCTradeAddress(t *testing.T, l *ledger.Ledger, ka *KYCAddress, csdb *statedb.PovContractStateDB, gsdb *statedb.PovGlobalStateDB) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 10

	switch ka.Action {
	case KYCActionAdd:
		trieKey := statedb.PovCreateContractLocalStateKey(KYCDataAddress, ka.GetMixKey())

		ka.Valid = true
		data, err := ka.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			t.Fatal(err)
		}

		tradeTrieKey := statedb.PovCreateContractLocalStateKey(KYCDataTradeAddress, ka.GetKey())
		err = csdb.SetValue(tradeTrieKey, trieKey)
		if err != nil {
			t.Fatal(err)
		}
	case KYCActionRemove:
		trieKey := statedb.PovCreateContractLocalStateKey(KYCDataAddress, ka.GetMixKey())
		data, err := csdb.GetValue(trieKey)
		if err != nil {
			t.Fatal(err)
		}

		oka := new(KYCAddress)
		_, err = oka.UnmarshalMsg(data)
		if err != nil {
			t.Fatal(err)
		}

		oka.Valid = false
		data, err = oka.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := gsdb.CommitToTrie()
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

func updateKYCOperator(t *testing.T, l *ledger.Ledger, koa *KYCOperatorAccount, csdb *statedb.PovContractStateDB, gsdb *statedb.PovGlobalStateDB) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 10

	switch koa.Action {
	case KYCActionAdd:
		trieKey := statedb.PovCreateContractLocalStateKey(KYCDataOperator, koa.Account.Bytes())

		koa.Valid = true
		data, err := koa.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			t.Fatal(err)
		}
	case KYCActionRemove:
		trieKey := statedb.PovCreateContractLocalStateKey(KYCDataOperator, koa.Account.Bytes())
		data, err := csdb.GetValue(trieKey)
		if err != nil {
			t.Fatal(err)
		}

		okoa := new(KYCOperatorAccount)
		_, err = okoa.UnmarshalMsg(data)
		if err != nil {
			t.Fatal(err)
		}

		okoa.Valid = false
		data, err = okoa.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := gsdb.CommitToTrie()
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

func TestKYCIsAdmin(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	account := cfg.GenesisAddress()

	// can not get pov global state
	if KYCIsAdmin(ctx, account) {
		t.Fatal()
	}

	admin := &KYCAdminAccount{
		Account: mock.Address(),
		Comment: "a1",
		Valid:   false,
	}
	addKYCTestAdmin(t, l, admin, 10)

	if !KYCIsAdmin(ctx, account) {
		t.Fatal()
	}

	admins, err := KYCGetAdmin(ctx)
	if err != nil || len(admins) != 1 || admins[0].Account != account {
		t.Fatal()
	}

	account = mock.Address()
	if KYCIsAdmin(ctx, account) {
		t.Fatal()
	}

	if KYCIsAdmin(ctx, admin.Account) {
		t.Fatal()
	}

	admin = &KYCAdminAccount{
		Account: account,
		Comment: "a1",
		Valid:   true,
	}
	addKYCTestAdmin(t, l, admin, 10)
	if !KYCIsAdmin(ctx, account) {
		t.Fatal()
	}

	account = cfg.GenesisAddress()
	if KYCIsAdmin(ctx, account) {
		t.Fatal()
	}

	admins, err = KYCGetAdmin(ctx)
	if err != nil || len(admins) != 1 || admins[0].Account != admin.Account {
		t.Fatal()
	}
}

func TestKYCGetAllStatus(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	chainAddress := mock.Address()
	status := "KYC_STATUS_APPROVED"

	ks := &KYCStatus{ChainAddress: chainAddress, Status: status, Valid: true}

	_, err = KYCGetAllStatus(ctx)
	if err == nil {
		t.Fatal()
	}

	_, err = KYCGetStatusByChainAddress(ctx, chainAddress)
	if err == nil {
		t.Fatal()
	}

	updateKYCStatus(t, l, ks, csdb, gsdb)

	kss, err := KYCGetAllStatus(ctx)
	if err != nil || len(kss) != 1 || kss[0].ChainAddress != chainAddress {
		t.Fatal()
	}

	ks, err = KYCGetStatusByChainAddress(ctx, chainAddress)
	if err != nil || ks.ChainAddress != chainAddress {
		t.Fatal()
	}

	ks.Valid = false
	updateKYCStatus(t, l, ks, csdb, gsdb)

	kss, err = KYCGetAllStatus(ctx)
	if err != nil || len(kss) != 0 {
		t.Fatal()
	}

	ks, err = KYCGetStatusByChainAddress(ctx, chainAddress)
	if err != nil || ks != nil {
		t.Fatal()
	}
}

func TestKYCGetTradeAddress(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	chainAddress := mock.Address()
	tradeAddress := "0xcd2a3d9f938e13cd947ec05abc7fe734df8dd826"
	status := "KYC_STATUS_APPROVED"

	_, err = KYCGetTradeAddress(ctx, chainAddress)
	if err == nil {
		t.Fatal()
	}

	ka := &KYCAddress{ChainAddress: chainAddress, TradeAddress: tradeAddress, Comment: "a1", Action: KYCActionAdd}
	updateKYCTradeAddress(t, l, ka, csdb, gsdb)

	kas, err := KYCGetTradeAddress(ctx, chainAddress)
	if err != nil || len(kas) != 1 || kas[0].TradeAddress != tradeAddress {
		t.Fatal()
	}

	ks := &KYCStatus{ChainAddress: chainAddress, Status: status, Valid: true}
	updateKYCStatus(t, l, ks, csdb, gsdb)

	ks, err = KYCGetStatusByTradeAddress(ctx, tradeAddress)
	if err != nil || ks == nil || ks.ChainAddress != chainAddress {
		t.Fatal()
	}

	ka = &KYCAddress{ChainAddress: chainAddress, TradeAddress: tradeAddress, Comment: "a1", Action: KYCActionRemove}
	updateKYCTradeAddress(t, l, ka, csdb, gsdb)

	kas, err = KYCGetTradeAddress(ctx, chainAddress)
	if err != nil || len(kas) != 0 {
		t.Fatal()
	}
}

func TestKYCTradeAddressActionFromString(t *testing.T) {
	action, _ := KYCActionFromString("add")
	if action != KYCActionAdd {
		t.Fatal()
	}

	action, _ = KYCActionFromString("remove")
	if action != KYCActionRemove {
		t.Fatal()
	}

	action, err := KYCActionFromString("wrong")
	if err == nil {
		t.Fatal()
	}

	actions := KYCActionToString(KYCActionAdd)
	if actions != "add" {
		t.Fatal()
	}

	actions = KYCActionToString(KYCActionRemove)
	if actions != "remove" {
		t.Fatal()
	}

	actions = KYCActionToString(KYCActionInvalid)
	if actions != "wrong action" {
		t.Fatal()
	}
}

func TestKYCGetOperator(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.KYCAddress)
	koa := &KYCOperatorAccount{Account: mock.Address(), Action: KYCActionAdd, Comment: "op1"}

	_, err = KYCGetOperator(ctx)
	if err == nil {
		t.Fatal()
	}

	if KYCIsOperator(ctx, koa.Account) {
		t.Fatal()
	}

	updateKYCOperator(t, l, koa, csdb, gsdb)

	ops, err := KYCGetOperator(ctx)
	if err != nil || len(ops) != 1 || ops[0].Account != koa.Account {
		t.Fatal()
	}

	if !KYCIsOperator(ctx, koa.Account) {
		t.Fatal()
	}
}
