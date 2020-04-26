package chain

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
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

func TestPermissionService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	ps := NewPermissionService(cm.ConfigFile)
	cc := context.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	rs, err := NewRPCService(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	err = cc.Register(context.RPCService, rs)
	if err != nil {
		t.Fatal(err)
	}

	err = ps.Init()
	if err == nil {
		t.Fatal()
	}

	adminAccount := mock.Account()
	admin := new(abi.AdminAccount)
	admin.Account = adminAccount.Address()
	admin.Comment = "t1"
	admin.Valid = true
	addTestAdmin(t, l, admin, 1)

	err = ps.Init()
	if err != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(adminAccount.Address())
	am.Tokens[0].Type = config.ChainToken()
	ps.vmCtx.Ledger.AddAccountMeta(am, ps.vmCtx.Ledger.Cache().GetCache())

	cc.SetAccounts([]*types.Account{adminAccount})
	wli := &config.WhiteListInfo{
		PeerId:  "xxxxxxx",
		Addr:    "127.0.0.1:9734",
		Comment: "tn1",
	}
	ps.cfg.WhiteList.WhiteListInfos = append(ps.cfg.WhiteList.WhiteListInfos, wli)
	err = ps.Init()
	if err != nil {
		t.Fatal()
	}

	err = ps.cc.Init(func() error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ps.cc.Start()

	err = ps.Start()
	if err != nil {
		t.Fatal(err)
	}

	ps.Status()

	ps.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Second)

	err = ps.Stop()
	if err != nil {
		t.Fatal()
	}
	time.Sleep(3 * time.Second)
}
