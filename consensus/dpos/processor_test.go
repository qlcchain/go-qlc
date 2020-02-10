package dpos

import (
	"github.com/qlcchain/go-qlc/consensus"
	"math"
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
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
	sb1.Token = common.ChainToken()
	sb1.Previous = prev
	sb1.Representative = common.GenesisAddress()
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
	sb2.Token = common.ChainToken()
	sb2.Previous = prev
	sb2.Representative = common.GenesisAddress()
	addr2 := mock.Address()
	sb2.Link = addr2.ToHash()
	sb2.Signature = a.Sign(sb2.GetHash())
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, sb2.Root())
	sb2.Work = worker2.NewWork()

	return sb1, sb2
}

func TestProcessResult(t *testing.T) {
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
