package statedb

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
)

type povStateDBMockData struct {
	l ledger.Store
}

func setupPovStateDBTestCase(t *testing.T) (func(t *testing.T), *povStateDBMockData) {
	t.Parallel()

	md := &povStateDBMockData{}

	uid := uuid.New().String()
	rootDir := filepath.Join(config.QlcTestDataDir(), uid)

	lDir := filepath.Join(rootDir, "ledger")
	_ = os.RemoveAll(lDir)
	cm := config.NewCfgManager(lDir)
	_, _ = cm.Load()
	md.l = ledger.NewLedger(cm.ConfigFile)

	genBlk, genTd := mock.GenerateGenesisPovBlock()
	_ = md.l.AddPovBlock(genBlk, genTd)

	return func(t *testing.T) {
		err := md.l.DBStore().Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll(rootDir)
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovStateDB_GlobalDB(t *testing.T) {
	teardownTestCase, md := setupPovStateDBTestCase(t)
	defer teardownTestCase(t)

	gsdb := NewPovGlobalStateDB(md.l.DBStore(), types.ZeroHash)

	ac1 := mock.Account()
	as1 := types.NewPovAccountState()
	as1.Balance = types.NewBalance(rand.Int63())
	err := gsdb.SetAccountState(ac1.Address(), as1)
	if err != nil {
		t.Fatal(err)
	}
	retAs1, err := gsdb.GetAccountState(ac1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if retAs1.Balance.Compare(as1.Balance) != types.BalanceCompEqual {
		t.Fatal("account state not equal", retAs1.Balance, as1.Balance)
	}

	rep1 := mock.Account()
	rs1 := types.NewPovRepState()
	rs1.Balance = types.NewBalance(rand.Int63())
	err = gsdb.SetRepState(rep1.Address(), rs1)
	if err != nil {
		t.Fatal(err)
	}
	retRs1, err := gsdb.GetRepState(rep1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if retRs1.Balance.Compare(rs1.Balance) != types.BalanceCompEqual {
		t.Fatal("rep state not equal", retRs1.Balance, rs1.Balance)
	}

	key1 := util.RandomFixedString(16)
	val1 := util.RandomFixedString(32)
	err = gsdb.SetValue([]byte(key1), []byte(val1))
	if err != nil {
		t.Fatal(err)
	}
	retVal1, err := gsdb.GetValue([]byte(key1))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retVal1, []byte(val1)) {
		t.Fatal("key value not equal")
	}

	err = gsdb.CommitToTrie()
	if err != nil {
		t.Fatal(err)
	}

	curHash := gsdb.GetCurHash()

	txn := md.l.DBStore().Batch(true)
	err = gsdb.CommitToDB(txn)
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	gsdb = NewPovGlobalStateDB(md.l.DBStore(), curHash)
	if gsdb.GetPrevTrie() == nil {
		t.Fatal("prev tire is nil")
	}
	if gsdb.GetCurTrie() == nil {
		t.Fatal("cur tire is nil")
	}

	if curHash != gsdb.GetPrevHash() {
		t.Fatal("hash not equal", curHash, gsdb.GetPrevHash())
	}

	retAs1, err = gsdb.GetAccountState(ac1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if retAs1.Balance.Compare(as1.Balance) != types.BalanceCompEqual {
		t.Fatal("account state not equal", retAs1.Balance, as1.Balance)
	}

	retRs1, err = gsdb.GetRepState(rep1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if retRs1.Balance.Compare(rs1.Balance) != types.BalanceCompEqual {
		t.Fatal("rep state not equal", retRs1.Balance, rs1.Balance)
	}

	retVal1, err = gsdb.GetValue([]byte(key1))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retVal1, []byte(val1)) {
		t.Fatal("key value not equal")
	}
}

func TestPovStateDB_ContractDB(t *testing.T) {
	teardownTestCase, md := setupPovStateDBTestCase(t)
	defer teardownTestCase(t)

	cs := types.NewPovContractState()
	csdb := NewPovContractStateDB(md.l.DBStore(), cs)

	var allKeys []string
	var allVals []string

	// set contract kv by csdb
	for idx := 0; idx < 3; idx++ {
		key1 := util.RandomFixedString(32)
		val1 := util.RandomFixedString(64)
		err := csdb.SetValue([]byte(key1), []byte(val1))
		if err != nil {
			t.Fatal(err)
		}
		retVal1, err := csdb.GetValue([]byte(key1))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retVal1, []byte(val1)) {
			t.Fatal("key value not equal", "idx", idx, key1, val1, string(retVal1))
		}
		allKeys = append(allKeys, key1)
		allVals = append(allVals, val1)
	}

	_, err := csdb.CommitToTrie()
	if err != nil {
		t.Fatal(err)
	}

	curHash := csdb.GetCurHash()
	t.Log("csdb state hash", curHash)

	txn := md.l.DBStore().Batch(true)
	err = csdb.CommitToDB(txn)
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	cs2 := types.NewPovContractState()
	cs2.StateHash = curHash

	csdb2 := NewPovContractStateDB(md.l.DBStore(), cs2)
	if csdb2.GetPrevTrie() == nil {
		t.Fatal("prev tire is nil")
	}
	if csdb2.GetCurTrie() == nil {
		t.Fatal("cur tire is nil")
	}

	if curHash != csdb2.GetPrevHash() {
		t.Fatal("hash not equal", curHash, csdb2.GetPrevHash())
	}

	// get contract kv by csdb
	for idx, key1 := range allKeys {
		retVal1, err := csdb2.GetValue([]byte(key1))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retVal1, []byte(allVals[idx])) {
			t.Fatal("key value not equal", "idx", idx)
		}
	}
}

func TestPovStateDB_ContractDB2(t *testing.T) {
	teardownTestCase, md := setupPovStateDBTestCase(t)
	defer teardownTestCase(t)

	gsdb := NewPovGlobalStateDB(md.l.DBStore(), types.ZeroHash)

	cc1 := mock.Account()

	var allKeys []string
	var allVals []string

	// set contract kv by gsdb
	for idx := 0; idx < 3; idx++ {
		key1 := util.RandomFixedString(32)
		val1 := util.RandomFixedString(64)
		err := gsdb.SetContractValue(cc1.Address(), []byte(key1), []byte(val1))
		if err != nil {
			t.Fatal(err)
		}
		retVal1, err := gsdb.GetContractValue(cc1.Address(), []byte(key1))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retVal1, []byte(val1)) {
			t.Fatal("key value not equal", "idx", idx, key1, val1, string(retVal1))
		}
		allKeys = append(allKeys, key1)
		allVals = append(allVals, val1)
	}

	err := gsdb.CommitToTrie()
	if err != nil {
		t.Fatal(err)
	}

	curHash := gsdb.GetCurHash()

	txn := md.l.DBStore().Batch(true)
	err = gsdb.CommitToDB(txn)
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	// get contract kv by gsdb
	gsdb2 := NewPovGlobalStateDB(md.l.DBStore(), curHash)

	for idx, key1 := range allKeys {
		retVal1, err := gsdb2.GetContractValue(cc1.Address(), []byte(key1))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retVal1, []byte(allVals[idx])) {
			t.Fatal("key value not equal", "idx", idx)
		}
	}

	// get contract kv by csdb
	cs2, err := gsdb2.GetContractState(cc1.Address())
	if err != nil {
		t.Fatal(err)
	}
	csdb2 := NewPovContractStateDB(md.l.DBStore(), cs2)

	for idx, key1 := range allKeys {
		retVal1, err := csdb2.GetValue([]byte(key1))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retVal1, []byte(allVals[idx])) {
			t.Fatal("key value not equal", "idx", idx)
		}
	}
}
