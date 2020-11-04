package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	rpc "github.com/qlcchain/jsonrpc2"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/mock/mocks"
)

func setupDefaultLedgerAPI(t *testing.T) (func(t *testing.T), ledger.Store, *LedgerAPI) {
	//t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)
	fmt.Println("case: ", t.Name())

	ledgerApi := NewLedgerApi(context.Background(), l, cc.EventBus(), cc)
	verifier := process.NewLedgerVerifier(l)

	var blocks []*types.StateBlock

	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}
	for i := range blocks {
		block := blocks[i]
		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
	}, l, ledgerApi
}

func setupMockLedgerAPI(t *testing.T) (func(t *testing.T), *mocks.Store, *LedgerAPI) {
	//t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)

	l := new(mocks.Store)
	ledgerApi := NewLedgerApi(context.Background(), l, cc.EventBus(), cc)
	return func(t *testing.T) {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}, l, ledgerApi
}

func TestLedger_GetBlockCacheLock(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	eb := cc.EventBus()

	ledgerApi := NewLedgerApi(context.Background(), l, eb, cc)

	defer func() {
		//err := l.DBStore.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}()

	ledgerApi.ledger.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	chainToken := config.ChainToken()
	gasToken := config.GasToken()
	addr, _ := types.HexToAddress("qlc_361j3uiqdkjrzirttrpu9pn7eeussymty4rz4gifs9ijdx1p46xnpu3je7sy")
	_ = ledgerApi.getProcessLock(addr, chainToken)
	if ledgerApi.processLockLen() != 1 {
		t.Fatal("lock len error for addr")
	}
	_ = ledgerApi.getProcessLock(addr, gasToken)
	if ledgerApi.processLockLen() != 2 {
		t.Fatal("lock error for different token")
	}

	for i := 0; i < 998; i++ {
		a := mock.Address()
		ledgerApi.getProcessLock(a, chainToken)
	}
	if ledgerApi.processLockLen() != 1000 {
		t.Fatal("lock len error for 1000 addresses")
	}
	sb := mock.StateBlockWithAddress(addr)
	_, _ = ledgerApi.Process(sb)
	addr2, _ := types.HexToAddress("qlc_1gnggt8b6cwro3b4z9gootipykqd6x5gucfd7exsi4xqkryiijciegfhon4u")
	_ = ledgerApi.getProcessLock(addr2, chainToken)
	if ledgerApi.processLockLen() != 1001 {
		t.Fatal("get error when delete idle lock")
	}
}

func TestLedgerApi_Subscription(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	// init ac1
	ac1 := initAccount(l, t)

	// GenerateSendBlock
	ac2 := mock.Account()
	go func() {
		amount := types.Balance{Int: big.NewInt(int64(100000))}
		sendBlk, err := ledgerApi.GenerateSendBlock(&APISendBlockPara{
			From:      ac1.Address(),
			TokenName: "QLC",
			To:        ac2.Address(),
			Amount:    amount,
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		pendingKey := &types.PendingKey{
			Address: ac2.Address(),
			Hash:    sendBlk.GetHash(),
		}
		pendingInfo := &types.PendingInfo{
			Source: ac1.Address(),
			Type:   config.ChainToken(),
			Amount: amount,
		}
		if err := l.AddPending(pendingKey, pendingInfo, l.Cache().GetCache()); err != nil {
			t.Fatal(err)
		}

		if err := l.AddStateBlock(sendBlk); err != nil {
			t.Fatal(err)
		}
	}()

	r, err := ledgerApi.NewBlock(rpc.SubscriptionContextRandom())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.ID)
	r, err = ledgerApi.NewBlock(rpc.SubscriptionContextRandom())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.ID)
	r, err = ledgerApi.BalanceChange(rpc.SubscriptionContextRandom(), ac1.Address())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.ID)
	r, err = ledgerApi.BalanceChange(rpc.SubscriptionContextRandom(), ac1.Address())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.ID)
	r, err = ledgerApi.NewAccountBlock(rpc.SubscriptionContextRandom(), ac1.Address())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.ID)
	r, err = ledgerApi.NewPending(rpc.SubscriptionContextRandom(), ac2.Address())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.ID)
	time.Sleep(3 * time.Second)

}

func TestLedgerAPI_AccountBlocksCount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	addr1 := mock.Address()
	addr2 := mock.Address()
	addr3 := mock.Address()
	l.On("GetAccountMetaConfirmed", addr1).Return(mock.AccountMeta(addr1), nil)
	l.On("GetAccountMetaConfirmed", addr2).Return(nil, ledger.ErrAccountNotFound)
	l.On("GetAccountMetaConfirmed", addr3).Return(nil, errors.New("nil pointer"))
	tests := []struct {
		name       string
		arg        types.Address
		wantReturn int64
		wantErr    bool
	}{
		{
			name:       "f1",
			arg:        addr1,
			wantReturn: 5,
			wantErr:    false,
		},
		{
			name:       "f2",
			arg:        addr2,
			wantReturn: 0,
			wantErr:    false,
		},
		{
			name:       "f3",
			arg:        addr3,
			wantReturn: -1,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ledgerApi.AccountBlocksCount(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBlocksCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if r != tt.wantReturn {
				t.Errorf("AccountBlocksCount() value = %v, want %v", r, tt.wantReturn)
			}
		})
	}
}

func TestLedgerAPI_AccountHistoryTopn(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	r, err := ledgerApi.AccountHistoryTopn(account1.Address(), 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 4 {
		t.Fatal()
	}
}

func TestLedgerAPI_AccountInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.AccountInfo(account1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if r.Address != account1.Address() {
		t.Fatal()
	}
}

func TestLedgerAPI_ConfirmedAccountInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.ConfirmedAccountInfo(account1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if r.Address != account1.Address() {
		t.Fatal()
	}
}

func TestLedgerAPI_AccountRepresentative(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.AccountRepresentative(account1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if r != account1.Address() {
		t.Fatal()
	}
	r, err = ledgerApi.AccountRepresentative(mock.Address())
	if err == nil {
		t.Fatal(err)
	}
}

func TestLedgerAPI_AccountVotingWeight(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	addr1 := mock.Address()
	addr2 := mock.Address()
	addr3 := mock.Address()
	balance := types.Balance{Int: big.NewInt(int64(100000000000))}

	l.On("GetRepresentation", addr1).Return(&types.Benefit{
		Balance: balance,
		Total:   balance,
	}, nil)
	l.On("GetRepresentation", addr2).Return(types.ZeroBenefit, ledger.ErrRepresentationNotFound)
	l.On("GetRepresentation", addr3).Return(nil, errors.New("nil pointer"))
	tests := []struct {
		name       string
		arg        types.Address
		wantReturn types.Balance
		wantErr    bool
	}{
		{
			name:       "f1",
			arg:        addr1,
			wantReturn: balance,
			wantErr:    false,
		},
		{
			name:       "f2",
			arg:        addr2,
			wantReturn: types.ZeroBalance,
			wantErr:    false,
		},
		{
			name:       "f3",
			arg:        addr3,
			wantReturn: types.ZeroBalance,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ledgerApi.AccountVotingWeight(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountVotingWeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !r.Equal(tt.wantReturn) {
				t.Errorf("AccountVotingWeight() value = %v, want %v", r, tt.wantReturn)
			}
		})
	}
}

func TestLedgerAPI_AccountsBalance(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	addrs := []types.Address{account1.Address(), account2.Address()}
	r, err := ledgerApi.AccountsBalance(addrs)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_AccountsFrontiers(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	addrs := []types.Address{account1.Address(), account2.Address()}
	r, err := ledgerApi.AccountsFrontiers(addrs)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_AccountsPending(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	blk := mock.StateBlockWithoutWork()

	sAdd := mock.Address()
	pendingKey := &types.PendingKey{
		Address: sAdd,
		Hash:    blk.GetHash(),
	}
	pendingInfo := &types.PendingInfo{
		Source: blk.Address,
		Type:   config.ChainToken(),
		Amount: types.Balance{Int: big.NewInt(int64(100000000000))},
	}
	if err := l.AddPending(pendingKey, pendingInfo, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}

	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}

	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	r, err := ledgerApi.AccountsPending([]types.Address{sAdd}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 1 {
		t.Fatal(len(r))
	}

	ps, err := ledgerApi.Pendings()
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 {
		t.Fatal(len(r))
	}

}

func TestLedgerAPI_AccountsCount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	expect := uint64(10)
	l.On("CountAccountMetas").Return(expect, nil)
	r, err := ledgerApi.AccountsCount()
	if err != nil {
		t.Fatal(err)
	}
	if r != expect {
		t.Fatal(err)
	}
}

func TestLedgerAPI_Accounts(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.Accounts(10, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	if len(r) != 5 {
		t.Fatalf("invalid len %d", len(r))
	}
}

func TestLedgerAPI_BlocksCount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	l.On("BlocksCount").Return(uint64(10), nil)
	l.On("CountSmartContractBlocks").Return(uint64(5), nil)
	l.On("CountUncheckedBlocksStore").Return(uint64(1), nil)

	c, err := ledgerApi.BlocksCount()
	if err != nil {
		t.Fatal(err)
	}
	if c["count"] != 15 || c["unchecked"] != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_BlocksCount2(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	l.On("CountStateBlocks").Return(uint64(10), nil)
	l.On("CountSmartContractBlocks").Return(uint64(5), nil)
	l.On("CountUncheckedBlocks").Return(uint64(1), nil)

	c, err := ledgerApi.BlocksCount2()
	if err != nil {
		t.Fatal(err)
	}
	if c["count"] != 15 || c["unchecked"] != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_BlockAccount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	blk := mock.StateBlockWithoutWork()
	l.On("GetStateBlockConfirmed", blk.GetHash()).Return(blk, nil)
	r, err := ledgerApi.BlockAccount(blk.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if r != blk.Address {
		t.Fatal()
	}
}

func TestLedgerAPI_BlockConfirmedStatus(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	blk := mock.StateBlockWithoutWork()
	l.On("HasStateBlockConfirmed", blk.GetHash()).Return(true, nil)
	r, err := ledgerApi.BlockConfirmedStatus(blk.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	if !r {
		t.Fatal()
	}
}

func TestLedgerAPI_BlockHash(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	blk := mock.StateBlockWithoutWork()
	r := ledgerApi.BlockHash(*blk)
	if r != blk.GetHash() {
		t.Fatal()
	}
}

func TestLedgerAPI_BlocksCountByType(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	l.On("BlocksCountByType").Return(map[string]uint64{
		types.Open.String():    10,
		types.Receive.String(): 5,
		types.Send.String():    15,
		types.Change.String():  1,
	}, nil)
	r, err := ledgerApi.BlocksCountByType()
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 8 || r[types.Open.String()] != 10 || r[types.Receive.String()] != 5 {
		t.Fatal()
	}
}

func TestLedgerAPI_BlocksInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	hashes := []types.Hash{config.GenesisBlockHash(), config.GenesisMintageHash()}
	r, err := ledgerApi.BlocksInfo(hashes)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	if len(r) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_ConfirmedBlocksInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	hashes := []types.Hash{config.GenesisBlockHash(), config.GenesisMintageHash()}
	r, err := ledgerApi.ConfirmedBlocksInfo(hashes)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	if len(r) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_Blocks(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	l.Flush()
	r, err := ledgerApi.Blocks(100, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 12 {
		t.Fatalf("invalid len %d", len(r))
	}
}

func TestLedgerAPI_Chain(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}
	r, err := ledgerApi.Chain(blocks[6].GetHash(), -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_Delegators(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.Delegators(account1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 3 {
		t.Fatalf("invalid len %d", len(r))
	}
}

func TestLedgerAPI_DelegatorsCount(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.DelegatorsCount(account1.Address())
	if err != nil {
		t.Fatal(err)
	}
	if r != 3 {
		t.Fatalf("invalid len %d", r)
	}
}

func initAccount(l ledger.Store, t *testing.T) *types.Account {
	ac1 := mock.Account()
	balance := types.Balance{Int: big.NewInt(int64(100000000000))}
	blk := new(types.StateBlock)
	blk.Type = types.Open
	blk.Address = ac1.Address()
	blk.Previous = types.ZeroHash
	blk.Token = config.ChainToken()
	blk.Balance = balance
	blk.Timestamp = common.TimeNow().Unix()
	blk.Link = mock.Hash()
	blk.Representative = ac1.Address()

	am := mock.AccountMeta(ac1.Address())
	tm := &types.TokenMeta{
		Type:           config.ChainToken(),
		Header:         blk.GetHash(),
		OpenBlock:      types.ZeroHash,
		Representative: ac1.Address(),
		Balance:        balance,
		BelongTo:       ac1.Address(),
	}
	am.Tokens = []*types.TokenMeta{tm}

	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	if err := l.AddAccountMeta(am, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	return ac1
}

func TestLedgerAPI_Process(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	// init ac1
	ac1 := initAccount(l, t)

	// GenerateSendBlock
	ac2 := mock.Account()
	ac2Addr := ac2.Address()
	balance2 := types.Balance{Int: big.NewInt(int64(100000))}

	ac1PrkStr := hex.EncodeToString(ac1.PrivateKey()[:])

	sendBlk1, err := ledgerApi.GenerateSendBlock(&APISendBlockPara{
		From:      ac1.Address(),
		TokenName: "QLC",
		To:        ac2Addr,
		Amount:    balance2,
	}, &ac1PrkStr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sendBlk1)

	if _, err := ledgerApi.Process(sendBlk1); err != nil {
		t.Fatal(err)
	}

	// GenerateReceiveBlock  ac2 open block
	pendingKey := &types.PendingKey{
		Address: ac2.Address(),
		Hash:    sendBlk1.GetHash(),
	}
	pendingInfo := &types.PendingInfo{
		Source: ac1.Address(),
		Type:   config.ChainToken(),
		Amount: balance2,
	}
	if err := l.AddPending(pendingKey, pendingInfo, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}

	ac2PrkStr := hex.EncodeToString(ac2.PrivateKey()[:])
	receBlk1, err := ledgerApi.GenerateReceiveBlock(sendBlk1, &ac2PrkStr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(receBlk1)
	if _, err := ledgerApi.Process(receBlk1); err != nil {
		t.Fatal(err)
	}

	receBlk2, err := ledgerApi.GenerateReceiveBlockByHash(sendBlk1.GetHash(), &ac2PrkStr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(receBlk2)

	if _, err := ledgerApi.Process(receBlk2); err == nil {
		t.Fatal(err)
	}

	// GenerateChangeBlock
	ac3 := mock.AccountMeta(mock.Address())
	if err := l.AddAccountMeta(ac3, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	changeBlk, err := ledgerApi.GenerateChangeBlock(ac1.Address(), ac2Addr, &ac1PrkStr)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(changeBlk)
	if _, err := ledgerApi.Process(changeBlk); err != nil {
		t.Fatal(err)
	}
}

func TestLedgerAPI_Representatives(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.Representatives(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_Tokens(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	ts, err := ledgerApi.Tokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_TransactionsCount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	expert1 := uint64(10)
	expert2 := uint64(5)
	l.On("BlocksCount").Return(expert1, nil)
	l.On("CountUncheckedBlocks").Return(expert2, nil)
	r, err := ledgerApi.TransactionsCount()
	if err != nil {
		t.Fatal(err)
	}
	if r["count"] != expert1 || r["unchecked"] != expert2 {
		t.Fatal()
	}
}

func TestLedgerAPI_TokenInfoById(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	ts, err := ledgerApi.TokenInfoById(config.ChainToken())
	if err != nil {
		t.Fatal(err)
	}
	if ts.TokenName != "QLC" {
		t.Fatal()
	}
}

func TestLedgerAPI_TokenInfoByName(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	ts, err := ledgerApi.TokenInfoByName("QLC")
	if err != nil {
		t.Fatal(err)
	}
	if ts.TokenId != config.ChainToken() {
		t.Fatal()
	}
}

func TestLedgerAPI_GetAccountOnlineBlock(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	addr := mock.Address()
	blk := mock.StateBlockWithoutWork()
	blk.Type = types.Online

	l.On("BlocksByAccount", addr, -1, -1).Return([]types.Hash{blk.GetHash()}, nil)
	l.On("GetStateBlockConfirmed", blk.GetHash()).Return(blk, nil)
	r, err := ledgerApi.GetAccountOnlineBlock(addr)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_GenesisBlock(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	if ledgerApi.GenesisAddress() != config.GenesisAddress() {
		t.Fatal()
	}
	if ledgerApi.GasAddress() != config.GasAddress() {
		t.Fatal()
	}
	if ledgerApi.ChainToken() != config.ChainToken() {
		t.Fatal()
	}
	if ledgerApi.GasToken() != config.GasToken() {
		t.Fatal()
	}
	if ledgerApi.GenesisMintageHash() != config.GenesisMintageHash() {
		t.Fatal()
	}
	if ledgerApi.GenesisBlockHash() != config.GenesisBlockHash() {
		t.Fatal()
	}
	if ledgerApi.GasBlockHash() != config.GasBlockHash() {
		t.Fatal()
	}
	if !ledgerApi.IsGenesisToken(config.ChainToken()) {
		t.Fatal()
	}
	if bs := ledgerApi.AllGenesisBlocks(); len(bs) != 4 {
		t.Fatal()
	}
}

func TestLedgerAPI_PubSub(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	blk1 := mock.StateBlock()

	blkRpcCtx := rpc.SubscriptionContextRandom()
	blkSub, err := ledgerApi.NewBlock(blkRpcCtx)
	if err != nil {
		t.Fatal(err)
	}
	if blkSub == nil {
		t.Fatal("blkSub is nil")
	}

	pendRpcCtx := rpc.SubscriptionContextRandom()
	pendSub, err := ledgerApi.NewPending(pendRpcCtx, blk1.GetAddress())
	if err != nil {
		t.Fatal(err)
	}
	if pendSub == nil {
		t.Fatal("blkSub is nil")
	}

	accRpcCtx := rpc.SubscriptionContextRandom()
	accSub, err := ledgerApi.NewAccountBlock(accRpcCtx, blk1.GetAddress())
	if err != nil {
		t.Fatal(err)
	}
	if accSub == nil {
		t.Fatal("accSub is nil")
	}

	blRpcCtx := rpc.SubscriptionContextRandom()
	blSub, err := ledgerApi.BalanceChange(blRpcCtx, blk1.GetAddress())
	if err != nil {
		t.Fatal(err)
	}
	if blSub == nil {
		t.Fatal("blSub is nil")
	}

	ledgerApi.blockSubscription.setBlocks(blk1)
	time.Sleep(10 * time.Millisecond)
}
