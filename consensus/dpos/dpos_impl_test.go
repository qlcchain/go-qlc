package dpos

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"

	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func getTestDpos() *DPoS {
	dir := filepath.Join(config.QlcTestDataDir(), "dpos", uuid.New().String())
	cm := config.NewCfgManager(dir)
	return NewDPoS(cm.ConfigFile)
}

func TestGetSeq(t *testing.T) {
	dps := getTestDpos()

	seq1 := dps.getSeq(ackTypeCommon)
	if seq1 != 0 {
		fmt.Printf("expect:0   get:%d", seq1)
		//		t.Fail()
	}

	seq2 := dps.getSeq(ackTypeCommon)
	if seq2 != 1 {
		fmt.Printf("expect:1   get:%d", seq2)
		//		t.Fail()
	}

	seq3 := dps.getSeq(ackTypeFindRep)
	if seq3 != 0x10000002 {
		fmt.Printf("expect:%0X   get:%0X", 0x10000002, seq3)
		//		t.Fail()
	}

	seq4 := dps.getSeq(ackTypeFindRep)
	if seq4 != 0x10000003 {
		fmt.Printf("expect:%0X   get:%0X", 0x10000003, seq4)
		//		t.Fail()
	}
}

func TestGetAckType(t *testing.T) {
	dps := getTestDpos()

	type1 := dps.getAckType(0x10000003)
	if type1 != ackTypeFindRep {
		t.Errorf("expect:%d   get:%d", ackTypeFindRep, type1)
		t.Fail()
	}

	type2 := dps.getAckType(3)
	if type2 != ackTypeCommon {
		t.Errorf("expect:%d   get:%d", ackTypeCommon, type2)
		t.Fail()
	}
}

func TestOnFrontierConfirmed(t *testing.T) {
	dps := getTestDpos()
	block := mock.StateBlockWithoutWork()
	hash := block.GetHash()

	var confirmed bool
	dps.onFrontierConfirmed(hash, &confirmed)
	if confirmed {
		t.Fatal()
	}

	dps.frontiersStatus.Store(hash, frontierChainConfirmed)
	dps.onFrontierConfirmed(hash, &confirmed)
	if !confirmed {
		t.Fatal()
	}

	dps.frontiersStatus.Store(hash, frontierWaitingForVote)
	dps.onFrontierConfirmed(hash, &confirmed)
	if confirmed {
		t.Fatal()
	}
}

func TestBatchVote(t *testing.T) {
	acc := mock.Account()

	hashes := make([]types.Hash, 0)
	hashBytes := make([]byte, 0)

	for i := 0; i < 1024; i++ {
		h := types.HashData([]byte(fmt.Sprintf("%d", i)))
		hashes = append(hashes, h)
		hashBytes = append(hashBytes, h[:]...)
	}

	hash, _ := types.HashBytes(hashBytes)
	sign1 := acc.Sign(hash)

	hashBytes2 := make([]byte, 0)
	for _, h := range hashes {
		hashBytes2 = append(hashBytes2, h[:]...)
	}

	signHash, _ := types.HashBytes(hashBytes2)
	valid := acc.Address().Verify(signHash[:], sign1[:])
	if !valid {
		t.Fatal("sign err")
	}
}

func TestCacheAck(t *testing.T) {
	dps := getTestDpos()
	hash := mock.Hash()

	vi1 := &voteInfo{
		hash:    hash,
		account: mock.Address(),
	}
	dps.cacheAck(vi1)
	_, err := dps.voteCache.Get(hash)
	if err != nil {
		t.Fatal(err)
	}

	vi2 := &voteInfo{
		hash:    hash,
		account: mock.Address(),
	}
	dps.cacheAck(vi2)
	v, err := dps.voteCache.Get(hash)
	if err != nil {
		t.Fatal(err)
	}

	vc := v.(*sync.Map)
	if _, ok := vc.Load(vi1.account); !ok {
		t.Fatal()
	}
	if _, ok := vc.Load(vi2.account); !ok {
		t.Fatal()
	}
}

func TestHasLocalValidRep(t *testing.T) {
	dps := getTestDpos()
	dps.minVoteWeight = types.NewBalance(1000)

	account := mock.Account()
	dps.localRepAccount.Store(account.Address(), account)
	if dps.hasLocalValidRep() {
		t.Fatal()
	}

	benefit := types.ZeroBenefit
	benefit.Vote = dps.minVoteWeight
	benefit.Total = dps.minVoteWeight
	dps.ledger.AddRepresentation(account.Address(), benefit, dps.ledger.Cache().GetCache())
	if !dps.hasLocalValidRep() {
		t.Fatal()
	}
}

func TestRefreshAccount(t *testing.T) {
	dps := getTestDpos()
	dps.minVoteWeight = types.NewBalance(1000)

	account := mock.Account()
	dps.accounts = []*types.Account{account}

	benefit := types.ZeroBenefit
	benefit.Vote = dps.minVoteWeight
	benefit.Total = dps.minVoteWeight
	dps.ledger.AddRepresentation(account.Address(), benefit, dps.ledger.Cache().GetCache())
	dps.refreshAccount()

	if _, ok := dps.localRepAccount.Load(account.Address()); !ok {
		t.Fatal()
	}
}

func TestCleanOnlineReps(t *testing.T) {
	dps := getTestDpos()

	addr := mock.Address()
	dps.onlineReps.Store(addr, time.Now().Unix()-int64(repTimeout.Seconds()))
	dps.cleanOnlineReps()
	if _, ok := dps.onlineReps.Load(addr); ok {
		t.Fatal()
	}
}

func TestOnRollback(t *testing.T) {
	dps := getTestDpos()
	dps.minVoteWeight = types.NewBalance(1000)

	account := mock.Account()
	dps.accounts = []*types.Account{account}
	blk := mock.StateBlockWithoutWork()
	blk.Type = types.Online
	blk.Address = account.Address()

	benefit := types.ZeroBenefit
	benefit.Vote = dps.minVoteWeight
	benefit.Total = dps.minVoteWeight
	dps.ledger.AddRepresentation(account.Address(), benefit, dps.ledger.Cache().GetCache())
	dps.ledger.AddStateBlock(blk)

	header := mock.StateBlockWithoutWork()
	header.Address = mock.Address()
	dps.ledger.AddStateBlock(header)
	am := mock.AccountMeta(account.Address())
	am.Tokens[0].Type = config.ChainToken()
	am.Tokens[0].Header = header.GetHash()
	dps.ledger.AddAccountMeta(am, dps.ledger.Cache().GetCache())

	dps.onRollback(blk.GetHash())
}

func TestIsWaitingFrontier(t *testing.T) {
	dps := getTestDpos()
	hash := mock.Hash()
	dps.frontiersStatus.Store(hash, frontierWaitingForVote)
	ok, s := dps.isWaitingFrontier(hash)
	if !ok || s != frontierWaitingForVote {
		t.Fatal()
	}
}

func TestDequeueGapPovBlocksFromDb(t *testing.T) {
	dps := getTestDpos()

	blk := mock.StateBlockWithoutWork()
	dps.ledger.AddGapPovBlock(2880, blk, types.UnSynchronized)

	ds := types.NewPovMinerDayStat()
	it := types.NewPovMinerStatItem()
	it.FirstHeight = 1440
	it.LastHeight = 2880
	it.BlockNum = 120
	it.RewardAmount = types.NewBalance(10)
	ds.DayIndex = uint32(common.PovMinerRewardHeightStart / uint64(common.POVChainBlocksPerDay))
	ds.MinerStats[blk.Address.String()] = it
	dps.ledger.AddPovMinerStat(ds)

	dps.dequeueGapPovBlocksFromDb(2880)

	dps.ledger.WalkGapPovBlocksWithHeight(2880, func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		t.Fatal()
		return nil
	})
}

func TestDispatchAckedBlock(t *testing.T) {
	dps := getTestDpos()

	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = 100
	err := dps.ledger.AddPovBlock(povBlk, povTd)
	if err != nil {
		t.Fatal(err)
	}

	err = dps.ledger.AddPovBestHash(povBlk.GetHeight(), povBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = dps.ledger.SetPovLatestHeight(povBlk.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	blk1 := mock.StateBlockWithoutWork()
	blk1.Type = types.ContractSend
	blk1.Link = contractaddress.DoDSettlementAddress.ToHash()
	blk1.PoVHeight = 100

	param := new(cabi.DoDSettleUpdateOrderInfoParam)
	blk1.Data, _ = param.ToABI()
	dps.dispatchAckedBlock(blk1, mock.Hash(), 0)

	blk2 := mock.StateBlockWithoutWork()
	blk2.Type = types.ContractSend
	blk2.Link = contractaddress.DoDSettlementAddress.ToHash()
	blk2.PoVHeight = 100

	param2 := new(cabi.DoDSettleCreateOrderParam)
	blk2.Data, _ = param2.ToABI()
	err = dps.ledger.AddStateBlock(blk2)
	if err != nil {
		t.Fatal(err)
	}

	blk3 := mock.StateBlockWithoutWork()
	blk3.Type = types.ContractReward
	blk3.Link = blk2.GetHash()
	blk3.PoVHeight = 100
	dps.dispatchAckedBlock(blk3, mock.Hash(), 0)
}
