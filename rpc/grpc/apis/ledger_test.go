package apis

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

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/mock/mocks"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

func setupDefaultLedgerAPI(t *testing.T) (func(t *testing.T), ledger.Store, *LedgerAPI) {
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
			wantReturn: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ledgerApi.AccountBlocksCount(context.Background(), &pbtypes.Address{
				Address: tt.arg.String(),
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountBlocksCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if r.GetValue() != tt.wantReturn {
				t.Errorf("AccountBlocksCount() value = %v, want %v", r, tt.wantReturn)
			}
		})
	}
}

func TestLedgerAPI_AccountHistoryTopn(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.AccountHistoryTopn(context.Background(), &pb.AccountHistoryTopnReq{
		Address: account1.Address().String(),
		Count:   100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetBlocks()) != 4 {
		t.Fatal()
	}
}

func TestLedgerAPI_AccountInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.AccountInfo(context.Background(), &pbtypes.Address{
		Address: account1.Address().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.GetAddress() != account1.Address().String() {
		t.Fatal()
	}
}

func TestLedgerAPI_ConfirmedAccountInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.ConfirmedAccountInfo(context.Background(), &pbtypes.Address{
		Address: account1.Address().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Address != account1.Address().String() {
		t.Fatal()
	}
}

func TestLedgerAPI_AccountRepresentative(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.AccountRepresentative(context.Background(), &pbtypes.Address{
		Address: account1.Address().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.GetAddress() != account1.Address().String() {
		t.Fatal()
	}
	r, err = ledgerApi.AccountRepresentative(context.Background(), &pbtypes.Address{
		Address: mock.Address().String(),
	})
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
			r, err := ledgerApi.AccountVotingWeight(context.Background(), &pbtypes.Address{
				Address: tt.arg.String(),
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountVotingWeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !toOriginBalanceByValue(r.GetBalance()).Equal(tt.wantReturn) {
				t.Errorf("AccountVotingWeight() value = %v, want %v", r, tt.wantReturn)
			}
		})
	}
}

func TestLedgerAPI_AccountsBalance(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	addrs := []string{account1.Address().String(), account2.Address().String()}
	r, err := ledgerApi.AccountsBalance(context.Background(), &pbtypes.Addresses{
		Addresses: addrs,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetAccountsBalances()) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_AccountsFrontiers(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	addrs := []string{account1.Address().String(), account2.Address().String()}
	r, err := ledgerApi.AccountsFrontiers(context.Background(), &pbtypes.Addresses{
		Addresses: addrs,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetAccountsFrontiers()) != 2 {
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
	r, err := ledgerApi.AccountsPending(context.Background(), &pb.AccountsPendingReq{
		Addresses: []string{sAdd.String()},
		Count:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetAccountsPendings()) != 1 {
		t.Fatal(len(r.GetAccountsPendings()))
	}

	ps, err := ledgerApi.Pendings(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps.GetPendings()) != 1 {
		t.Fatal(len(ps.GetPendings()))
	}

}

func TestLedgerAPI_AccountsCount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	expect := uint64(10)
	l.On("CountAccountMetas").Return(expect, nil)
	r, err := ledgerApi.AccountsCount(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.GetValue() != expect {
		t.Fatal(err)
	}
}

func TestLedgerAPI_Accounts(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.Accounts(context.Background(), &pb.Offset{
		Count: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	if len(r.GetAddresses()) != 5 {
		t.Fatalf("invalid len %d", len(r.GetAddresses()))
	}
}

func TestLedgerAPI_BlocksCount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	l.On("BlocksCount").Return(uint64(10), nil)
	l.On("CountSmartContractBlocks").Return(uint64(5), nil)
	l.On("CountUncheckedBlocks").Return(uint64(1), nil)

	c, err := ledgerApi.BlocksCount(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.GetCount()["count"] != 15 || c.GetCount()["unchecked"] != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_BlocksCount2(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	l.On("CountStateBlocks").Return(uint64(10), nil)
	l.On("CountSmartContractBlocks").Return(uint64(5), nil)
	l.On("CountUncheckedBlocks").Return(uint64(1), nil)

	c, err := ledgerApi.BlocksCount2(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.GetCount()["count"] != 15 || c.GetCount()["unchecked"] != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_BlockAccount(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	blk := mock.StateBlockWithoutWork()
	l.On("GetStateBlockConfirmed", blk.GetHash()).Return(blk, nil)
	r, err := ledgerApi.BlockAccount(context.Background(), &pbtypes.Hash{
		Hash: blk.GetHash().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.GetAddress() != blk.Address.String() {
		t.Fatal()
	}
}

func TestLedgerAPI_BlockConfirmedStatus(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	blk := mock.StateBlockWithoutWork()
	l.On("HasStateBlockConfirmed", blk.GetHash()).Return(true, nil)
	r, err := ledgerApi.BlockConfirmedStatus(context.Background(), &pbtypes.Hash{
		Hash: blk.GetHash().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !r.GetValue() {
		t.Fatal()
	}
}

func TestLedgerAPI_BlockHash(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)
	blk := mock.StateBlockWithoutWork()
	r, err := ledgerApi.BlockHash(context.Background(), toStateBlock(blk))
	if err != nil {
		t.Fatal(err)
	}
	if r.GetHash() != blk.GetHash().String() {
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
	r, err := ledgerApi.BlocksCountByType(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetCount()) != 8 || r.GetCount()[types.Open.String()] != 10 || r.GetCount()[types.Receive.String()] != 5 {
		t.Fatal()
	}
}

func TestLedgerAPI_BlocksInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	hashes := []string{config.GenesisBlockHash().String(), config.GenesisMintageHash().String()}
	r, err := ledgerApi.BlocksInfo(context.Background(), &pbtypes.Hashes{
		Hashes: hashes,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	if len(r.GetBlocks()) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_ConfirmedBlocksInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	hashes := []string{config.GenesisBlockHash().String(), config.GenesisMintageHash().String()}
	r, err := ledgerApi.ConfirmedBlocksInfo(context.Background(), &pbtypes.Hashes{
		Hashes: hashes,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
	if len(r.GetBlocks()) != 2 {
		t.Fatal()
	}
}

func TestLedgerAPI_Blocks(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.Blocks(context.Background(), &pb.Offset{
		Count: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetBlocks()) != 12 {
		t.Fatalf("invalid len %d", len(r.GetBlocks()))
	}
}

func TestLedgerAPI_Chain(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}
	r, err := ledgerApi.Chain(context.Background(), &pb.ChainReq{
		Hash:  blocks[6].GetHash().String(),
		Count: -1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetHashes()) != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_Delegators(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.Delegators(context.Background(), &pbtypes.Address{
		Address: account1.Address().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetBalances()) != 3 {
		t.Fatalf("invalid len %d", len(r.GetBalances()))
	}
	for _, v := range r.GetBalances() {
		t.Log(v.GetAddress())
		t.Log(v.GetBalance())
	}
	a, err := json.Marshal(r.GetBalances())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(a))
}

func TestLedgerAPI_DelegatorsCount(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.DelegatorsCount(context.Background(), &pbtypes.Address{
		Address: account1.Address().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.GetValue() != 3 {
		t.Fatalf("invalid len %d", r.GetValue())
	}
}

func TestLedgerAPI_Representatives(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	r, err := ledgerApi.Representatives(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetRepresentatives()) != 2 {
		t.Fatal(len(r.GetRepresentatives()))
	}
}

func TestLedgerAPI_Tokens(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	ts, err := ledgerApi.Tokens(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ts.GetTokenInfos()) != 2 {
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
	r, err := ledgerApi.TransactionsCount(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.GetCount()["count"] != expert1 || r.GetCount()["unchecked"] != expert2 {
		t.Fatal()
	}
}

func TestLedgerAPI_TokenInfoById(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	ts, err := ledgerApi.TokenInfoById(context.Background(), &pbtypes.Hash{
		Hash: config.ChainToken().String(),
	})
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

	ts, err := ledgerApi.TokenInfoByName(context.Background(), toString("QLC"))
	if err != nil {
		t.Fatal(err)
	}
	if ts.TokenId != config.ChainToken().String() {
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
	r, err := ledgerApi.GetAccountOnlineBlock(context.Background(), &pbtypes.Address{
		Address: addr.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.GetStateBlocks()) != 1 {
		t.Fatal()
	}
}

func TestLedgerAPI_GenesisBlock(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupMockLedgerAPI(t)
	defer teardownTestCase(t)

	if r, _ := ledgerApi.GenesisAddress(context.Background(), nil); r.GetAddress() != config.GenesisAddress().String() {
		t.Fatal()
	}
	if r, _ := ledgerApi.GasAddress(context.Background(), nil); r.GetAddress() != config.GasAddress().String() {
		t.Fatal()
	}
	if r, _ := ledgerApi.ChainToken(context.Background(), nil); r.GetHash() != config.ChainToken().String() {
		t.Fatal()
	}
	if r, _ := ledgerApi.GasToken(context.Background(), nil); r.GetHash() != config.GasToken().String() {
		t.Fatal()
	}
	if r, _ := ledgerApi.GenesisMintageHash(context.Background(), nil); r.GetHash() != config.GenesisMintageHash().String() {
		t.Fatal()
	}
	if r, err := ledgerApi.GenesisBlock(context.Background(), nil); err != nil {
		t.Fatal()
	} else {
		if b, _ := toOriginStateBlock(r); b.GetHash() != config.GenesisBlockHash() {
			t.Fatal(b.GetHash(), config.GenesisBlockHash())
		}
	}
	if r, err := ledgerApi.GenesisMintageBlock(context.Background(), nil); err != nil {
		t.Fatal()
	} else {
		if b, _ := toOriginStateBlock(r); b.GetHash() != config.GenesisMintageHash() {
			t.Fatal(b.GetHash(), config.GenesisMintageHash())
		}
	}

	if r, err := ledgerApi.GasBlock(context.Background(), nil); err != nil {
		t.Fatal()
	} else {
		if b, _ := toOriginStateBlock(r); b.GetHash() != config.GasBlockHash() {
			t.Fatal(b.GetHash(), config.GasBlockHash())
		}
	}
	if r, err := ledgerApi.GasMintageBlock(context.Background(), nil); err != nil {
		t.Fatal()
	} else {
		bm := config.GasMintageBlock()
		if b, _ := toOriginStateBlock(r); b.GetHash() != bm.GetHash() {
			t.Fatal()
		}
	}

	if r, _ := ledgerApi.GenesisBlockHash(context.Background(), nil); r.GetHash() != config.GenesisBlockHash().String() {
		t.Fatal()
	}
	if r, _ := ledgerApi.GasBlockHash(context.Background(), nil); r.GetHash() != config.GasBlockHash().String() {
		t.Fatal()
	}
	if r, _ := ledgerApi.IsGenesisToken(context.Background(), &pbtypes.Hash{
		Hash: config.ChainToken().String(),
	}); !r.GetValue() {
		t.Fatal()
	}
	gb := config.GenesisBlock()
	if r, _ := ledgerApi.IsGenesisBlock(context.Background(), toStateBlock(&gb)); !r.GetValue() {
		t.Fatal()
	}
	if bs, _ := ledgerApi.AllGenesisBlocks(context.Background(), nil); len(bs.GetStateBlocks()) != 4 {
		t.Fatal()
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

	sendBlk1, err := ledgerApi.GenerateSendBlock(context.Background(), &pb.GenerateSendBlockReq{
		Param: &pb.APISendBlockPara{
			From:      toAddressValue(ac1.Address()),
			TokenName: "QLC",
			To:        toAddressValue(ac2Addr),
			Amount:    toBalanceValue(balance2),
		},
		PrkStr: ac1PrkStr,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sendBlk1)

	if _, err := ledgerApi.Process(context.Background(), sendBlk1); err != nil {
		t.Fatal(err)
	}
	sb, err := toOriginStateBlock(sendBlk1)
	if err != nil {
		t.Fatal(err)
	}
	// GenerateReceiveBlock  ac2 open block
	pendingKey := &types.PendingKey{
		Address: ac2.Address(),
		Hash:    sb.GetHash(),
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
	receBlk1, err := ledgerApi.GenerateReceiveBlock(context.Background(), &pb.GenerateReceiveBlockReq{
		Block:  sendBlk1,
		PrkStr: ac2PrkStr,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(receBlk1)
	if _, err := ledgerApi.Process(context.Background(), receBlk1); err != nil {
		t.Fatal(err)
	}

	receBlk2, err := ledgerApi.GenerateReceiveBlockByHash(context.Background(), &pb.GenerateReceiveBlockByHashReq{
		Hash:   toHashValue(sb.GetHash()),
		PrkStr: ac2PrkStr,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(receBlk2)

	if _, err := ledgerApi.Process(context.Background(), receBlk2); err == nil {
		t.Fatal(err)
	}

	// GenerateChangeBlock
	ac3 := mock.AccountMeta(mock.Address())
	if err := l.AddAccountMeta(ac3, l.Cache().GetCache()); err != nil {
		t.Fatal(err)
	}
	changeBlk, err := ledgerApi.GenerateChangeBlock(context.Background(), &pb.GenerateChangeBlockReq{
		Account:        toAddressValue(ac1.Address()),
		Representative: toAddressValue(ac2.Address()),
		PrkStr:         ac1PrkStr,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(changeBlk)
	if _, err := ledgerApi.Process(context.Background(), changeBlk); err != nil {
		t.Fatal(err)
	}
}

func TestLedgerAPI_NewBlock(t *testing.T) {
	teardownTestCase, l, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)

	ac1 := initAccount(l, t)

	// GenerateSendBlock
	ac2 := mock.Account()
	go func() {
		time.Sleep(1 * time.Second)
		amount := types.Balance{Int: big.NewInt(int64(100000))}
		sendBlk1, err := ledgerApi.GenerateSendBlock(context.Background(), &pb.GenerateSendBlockReq{
			Param: &pb.APISendBlockPara{
				From:      toAddressValue(ac1.Address()),
				TokenName: "QLC",
				To:        toAddressValue(ac2.Address()),
				Amount:    toBalanceValue(amount),
			},
			PrkStr: "",
		})
		sendBlk, err := toOriginStateBlock(sendBlk1)
		if err != nil {
			t.Fatal(err)
		}
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

	err := ledgerApi.NewBlock(new(empty.Empty), &ledgerAPINewBlockServer{ServerStream: new(baseStream)})
	if err != nil {
		t.Fatal(err)
	}
	err = ledgerApi.NewAccountBlock(toAddress(ac1.Address()), &ledgerAPINewBlockServer{ServerStream: new(baseStream)})
	if err != nil {
		t.Fatal(err)
	}

	err = ledgerApi.BalanceChange(toAddress(ac1.Address()), &ledgerAPIBalanceChangeServer{ServerStream: new(baseStream)})
	if err != nil {
		t.Fatal(err)
	}

	err = ledgerApi.NewPending(toAddress(ac2.Address()), &ledgerAPINewPendingServer{ServerStream: new(baseStream)})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)
}

type ledgerAPINewBlockServer struct {
	grpc.ServerStream
}

func (x *ledgerAPINewBlockServer) Send(m *pb.APIBlock) error {
	return x.ServerStream.SendMsg(m)
}

type ledgerAPIBalanceChangeServer struct {
	grpc.ServerStream
}

func (x *ledgerAPIBalanceChangeServer) Send(m *pb.APIAccount) error {
	return x.ServerStream.SendMsg(m)
}

type ledgerAPINewPendingServer struct {
	grpc.ServerStream
}

func (x *ledgerAPINewPendingServer) Send(m *pb.APIPending) error {
	return x.ServerStream.SendMsg(m)
}

type baseStream struct {
}

func (s *baseStream) SetHeader(metadata.MD) error {
	panic("implement me")
}

func (s *baseStream) SendHeader(metadata.MD) error {
	panic("implement me")
}

func (s *baseStream) SetTrailer(metadata.MD) {
	panic("implement me")
}

func (s *baseStream) Context() context.Context {
	return context.Background()
}

func (s *baseStream) SendMsg(m interface{}) error {
	fmt.Println(m)
	return nil
}

func (s *baseStream) RecvMsg(m interface{}) error {
	panic("implement me")
}

//
//func TestLedgerAPI_AddChan(t *testing.T) {
//	cs := make(map[string]int)
//	for i := 0; i < 200; i++ {
//		c := util.RandomFixedString(32)
//		if _, ok := cs[c]; ok {
//			t.Fatal("id exist already", i, cs[c])
//		}
//		cs[c] = i
//	}
//}
