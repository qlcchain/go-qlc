package dpos

import (
	"math"
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
)

func generateForkBlock() (block1, block2 *types.StateBlock) {
	a := mock.Account()
	prev := mock.Hash()

	b1, _ := types.NewBlock(types.State)
	sb1 := b1.(*types.StateBlock)
	i1, _ := random.Intn(math.MaxInt16)
	sb1.Type = types.State
	sb1.Balance = types.Balance{Int: big.NewInt(int64(i1))}
	sb1.Vote = types.ZeroBalance
	sb1.Oracle = types.ZeroBalance
	sb1.Network = types.ZeroBalance
	sb1.Storage = types.ZeroBalance
	sb1.Address = a.Address()
	sb1.Token = config.ChainToken()
	sb1.Previous = prev
	sb1.Representative = config.GenesisAddress()
	addr1 := mock.Address()
	sb1.Link = addr1.ToHash()
	sb1.Signature = a.Sign(sb1.GetHash())
	var w1 types.Work
	worker1, _ := types.NewWorker(w1, sb1.Root())
	sb1.Work = worker1.NewWork()

	b2, _ := types.NewBlock(types.State)
	sb2 := b2.(*types.StateBlock)
	i2, _ := random.Intn(math.MaxInt16)
	sb2.Type = types.State
	sb1.Balance = types.Balance{Int: big.NewInt(int64(i2))}
	sb2.Vote = types.ZeroBalance
	sb2.Oracle = types.ZeroBalance
	sb2.Network = types.ZeroBalance
	sb2.Storage = types.ZeroBalance
	sb2.Address = a.Address()
	sb2.Token = config.ChainToken()
	sb2.Previous = prev
	sb2.Representative = config.GenesisAddress()
	addr2 := mock.Address()
	sb2.Link = addr2.ToHash()
	sb2.Signature = a.Sign(sb2.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, sb2.Root())
	sb2.Work = worker2.NewWork()

	return sb1, sb2
}

func TestProcessor_processResultProgress(t *testing.T) {
	dps := getTestDpos()
	blk1, blk2 := generateForkBlock()

	bs1 := &consensus.BlockSource{
		Block:     blk1,
		BlockFrom: types.UnSynchronized,
		Type:      consensus.MsgPublishReq,
		Para:      nil,
	}
	bs2 := &consensus.BlockSource{
		Block:     blk2,
		BlockFrom: types.UnSynchronized,
		Type:      consensus.MsgConfirmReq,
		Para:      nil,
	}

	if dps.acTrx.addToRoots(blk1) && bs1.Type != consensus.MsgConfirmReq {
		dps.acTrx.setVoteHash(blk1)
	}

	if dps.acTrx.addToRoots(blk2) && bs2.Type != consensus.MsgConfirmReq {
		t.Fatal()
	}

	if el := dps.acTrx.getVoteInfo(blk2); el != nil {
		if el.voteHash == types.ZeroHash {
			t.Fatal("vote for block2")
		} else if blk2.GetHash() == el.voteHash {
			t.Fatal("vote for block2")
		}
	}
}

func TestProcessor_localRepVoteFrontier(t *testing.T) {
	dps := getTestDpos()
	dps.Init()
	processor := dps.processors[0]

	account := mock.Account()
	dps.localRepAccount.Store(account.Address(), account)

	benefit := types.ZeroBenefit
	err := dps.ledger.AddRepresentation(account.Address(), benefit, dps.ledger.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	blk := mock.StateBlockWithoutWork()
	processor.localRepVoteFrontier(blk)
	if len(processor.acks) > 0 {
		t.Fatal()
	}

	benefit.Vote = dps.minVoteWeight
	benefit.Total = dps.minVoteWeight
	err = dps.ledger.AddRepresentation(account.Address(), benefit, dps.ledger.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	processor.localRepVoteFrontier(blk)
	if len(processor.acks) != 1 {
		t.Fatal()
	}

	blk.Type = types.Online
	processor.localRepVoteFrontier(blk)
	if len(processor.acks) != 1 {
		t.Fatal()
	}
}

func TestProcessor_processResult(t *testing.T) {
	dps := getTestDpos()
	processor := dps.processors[0]
	blk := mock.StateBlockWithoutWork()
	blk.Previous = mock.Hash()
	blk.Link = mock.Hash()
	bs := &consensus.BlockSource{
		Block:     blk,
		BlockFrom: 0,
		Type:      0,
		Para:      nil,
	}

	results := []process.ProcessResult{
		process.BadSignature,
		process.BadWork,
		process.BalanceMismatch,
		process.Old,
		process.UnReceivable,
		process.GapSmartContract,
		process.InvalidData,
		process.Other,
		process.Fork,
		process.GapPrevious,
		process.GapSource,
		process.GapTokenInfo,
		process.GapPovHeight,
		process.GapPublish,
	}

	for _, pr := range results {
		processor.processResult(pr, bs)
	}
}

func TestProcessor_confirmBlock(t *testing.T) {
	dps := getTestDpos()
	processor := dps.processors[0]

	blk1 := mock.StateBlockWithoutWork()
	blk2 := blk1.Clone()
	blk2.Timestamp++

	vk := getVoteKey(blk1)
	el := newElection(dps, blk1)
	dps.acTrx.roots.Store(vk, el)

	processor.confirmBlock(blk2)
	if has, _ := dps.ledger.HasStateBlockConfirmed(blk2.GetHash()); !has {
		t.Fatal()
	}
}

func TestProcessor_dequeueGapPublish(t *testing.T) {
	dps := getTestDpos()

	blk := mock.StateBlockWithoutWork()
	dps.ledger.AddStateBlock(blk)

	blk1 := mock.StateBlockWithoutWork()
	blk1.Previous = blk.GetHash()
	dps.ledger.AddStateBlock(blk1)

	blk2 := mock.StateBlockWithoutWork()
	dps.ledger.AddGapPublishBlock(blk1.Previous, blk2, types.UnSynchronized)

	index := dps.getProcessorIndex(blk2.Address)
	processor := dps.processors[index]
	processor.dequeueGapPublish(blk1.GetHash())
	dps.ledger.GetGapPublishBlock(blk1.Previous, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		t.Fatal()
		return nil
	})
}

func TestProcessor_processFrontier(t *testing.T) {
	dps := getTestDpos()
	processor := dps.processors[0]
	blk1 := mock.StateBlockWithoutWork()
	blk2 := mock.StateBlockWithoutWork()

	blk1.Previous = mock.Hash()
	blk2.Previous = blk1.Previous
	dps.acTrx.addToRoots(blk1)
	processor.processFrontier(blk2)

	if _, ok := dps.hash2el.Load(blk2.GetHash()); !ok {
		t.Fatal()
	}
}

func TestProcessor_enqueueUncheckedSync(t *testing.T) {
	dps := getTestDpos()
	processor := dps.processors[0]
	blk := mock.StateBlockWithoutWork()
	blk.Previous = mock.Hash()

	processor.enqueueUncheckedSync(blk)
	if has, _ := dps.ledger.HasUncheckedSyncBlock(blk.Previous); !has {
		t.Fatal()
	}
}
