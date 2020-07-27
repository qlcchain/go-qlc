package api

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
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

func updateKYCStatus(t *testing.T, l *ledger.Ledger, status *abi.KYCStatus, csdb *statedb.PovContractStateDB, gsdb *statedb.PovGlobalStateDB) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 10

	trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataStatus, status.ChainAddress.Bytes())

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

func updateKYCTradeAddress(t *testing.T, l *ledger.Ledger, ka *abi.KYCAddress, csdb *statedb.PovContractStateDB, gsdb *statedb.PovGlobalStateDB) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 10

	switch ka.Action {
	case abi.KYCActionAdd:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataAddress, ka.GetMixKey())

		ka.Valid = true
		data, err := ka.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			t.Fatal(err)
		}

		tradeTrieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataTradeAddress, ka.GetKey())
		err = csdb.SetValue(tradeTrieKey, trieKey)
		if err != nil {
			t.Fatal(err)
		}
	case abi.KYCActionRemove:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataAddress, ka.GetMixKey())
		data, err := csdb.GetValue(trieKey)
		if err != nil {
			t.Fatal(err)
		}

		oka := new(abi.KYCAddress)
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

func updateKYCOperator(t *testing.T, l *ledger.Ledger, koa *abi.KYCOperatorAccount, csdb *statedb.PovContractStateDB, gsdb *statedb.PovGlobalStateDB) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 10

	switch koa.Action {
	case abi.KYCActionAdd:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataOperator, koa.Account.Bytes())

		koa.Valid = true
		data, err := koa.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			t.Fatal(err)
		}
	case abi.KYCActionRemove:
		trieKey := statedb.PovCreateContractLocalStateKey(abi.KYCDataOperator, koa.Account.Bytes())
		data, err := csdb.GetValue(trieKey)
		if err != nil {
			t.Fatal(err)
		}

		okoa := new(abi.KYCOperatorAccount)
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

func TestKYCApi_GetAdminHandoverBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)
	param := new(KYCAdminUpdateParam)

	_, err := p.GetAdminHandoverBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	_, _ = p.GetAdminHandoverBlock(param)
}

func TestKYCApi_GetUpdateStatusBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)
	param := new(KYCUpdateStatusParam)

	_, err := p.GetUpdateStatusBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	_, _ = p.GetUpdateStatusBlock(param)
}

func TestKYCApi_GetUpdateTradeAddressBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)
	param := new(KYCUpdateTradeAddressParam)

	_, err := p.GetUpdateTradeAddressBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	_, err = p.GetUpdateTradeAddressBlock(param)
	if err == nil {
		t.Fatal()
	}

	param.Action = "add"
	_, _ = p.GetUpdateTradeAddressBlock(param)
}

func TestKYCApi_GetAdmin(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)
	_, err := p.GetAdmin()
	if err == nil {
		t.Fatal()
	}

	admin := new(abi.KYCAdminAccount)
	admin.Account = mock.Address()
	admin.Valid = true
	addKYCTestAdmin(t, l, admin, 10)

	am, err := p.GetAdmin()
	if err != nil || am.Account != admin.Account {
		t.Fatal()
	}
}

func TestKYCApi_GetStatus(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)
	ks := new(abi.KYCStatus)
	ks.ChainAddress = mock.Address()
	ks.Valid = true
	ks.Status = "KYC_STATUS_APPROVED"

	if p.GetStatusCount() != 0 {
		t.Fatal()
	}

	_, err := p.GetStatus(10, 0)
	if err == nil {
		t.Fatal()
	}

	_, err = p.GetStatusByChainAddress(ks.ChainAddress)
	if err == nil {
		t.Fatal()
	}

	_, err = p.GetStatusByTradeAddress("invalid")
	if err == nil {
		t.Fatal()
	}

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	updateKYCStatus(t, l, ks, csdb, gsdb)

	if p.GetStatusCount() != 1 {
		t.Fatal()
	}

	ksis, err := p.GetStatus(10, 0)
	if err != nil || len(ksis) != 1 || ksis[0].ChainAddress != ks.ChainAddress {
		t.Fatal()
	}

	ksi, err := p.GetStatusByChainAddress(ks.ChainAddress)
	if err != nil || ksi == nil || ksi.ChainAddress != ks.ChainAddress {
		t.Fatal()
	}
}

func TestKYCApi_GetTradeAddress(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)
	ka := new(abi.KYCAddress)
	ka.ChainAddress = mock.Address()
	ka.TradeAddress = "0xcd2a3d9f938e13cd947ec05abc7fe734df8dd826"
	ka.Action = abi.KYCActionAdd

	_, err := p.GetTradeAddress(ka.ChainAddress)
	if err == nil {
		t.Fatal()
	}

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	updateKYCTradeAddress(t, l, ka, csdb, gsdb)

	kas, err := p.GetTradeAddress(ka.ChainAddress)
	if err != nil || kas == nil || kas.ChainAddress != ka.ChainAddress || kas.TradeAddress[0].Address != ka.TradeAddress {
		t.Fatal()
	}
}

func TestKYCApi_GetOperator(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)

	if p.GetOperatorCount() != 0 {
		t.Fatal()
	}

	_, err := p.GetOperator(10, 0)
	if err == nil {
		t.Fatal()
	}

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.KYCAddress)
	if err != nil {
		t.Fatal(err)
	}

	koa := &abi.KYCOperatorAccount{Account: mock.Address(), Action: abi.KYCActionAdd, Comment: "op1"}
	updateKYCOperator(t, l, koa, csdb, gsdb)

	if p.GetOperatorCount() != 1 {
		t.Fatal()
	}

	koas, err := p.GetOperator(10, 0)
	if err != nil || len(koas) != 1 || koas[0].Operator != koa.Account {
		t.Fatal()
	}
}

func TestKYCApi_GetUpdateOperatorBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewKYCApi(cfgFile, l)
	param := new(KYCUpdateOperatorParam)

	_, err := p.GetUpdateOperatorBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	_, err = p.GetUpdateOperatorBlock(param)
	if err == nil {
		t.Fatal()
	}

	param.Action = "add"
	_, _ = p.GetUpdateOperatorBlock(param)
}
