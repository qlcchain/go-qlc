package dpos

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

var (
	TestPrivateKey = "194908c480fddb6e66b56c08f0d55d935681da0b3c9c33077010bf12a91414576c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7"
	TestAddress    = "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7"
	priByte, _     = hex.DecodeString(TestPrivateKey)
	TestAccount    = types.NewAccount(priByte)
)

func TestFork(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}
	defer StopNodes(nodes)

	n1 := nodes[0]
	n2 := nodes[1]

	// open fork
	acc1 := mock.Account()
	acc2 := mock.Account()

	s1, r1 := n1.TokenTransactionAndConfirmed(TestAccount, acc1, types.NewBalance(10000), "QLC")
	n2.ProcessBlockAndWaitConfirmed(s1)
	n2.ProcessBlockAndWaitConfirmed(r1)

	s2 := n1.GenerateSendBlock(TestAccount, acc2.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockAndWaitConfirmed(s2)
	n2.ProcessBlockAndWaitConfirmed(s2)

	s3 := n1.GenerateSendBlock(acc1, acc2.Address(), types.NewBalance(20), "QLC")
	n1.ProcessBlockAndWaitConfirmed(s3)
	n2.ProcessBlockAndWaitConfirmed(s3)

	r2 := n1.GenerateReceiveBlock(s2, acc2)
	r3 := n2.GenerateReceiveBlock(s3, acc2)
	n1.ProcessBlock(r2)
	n2.ProcessBlock(r2)
	n1.ProcessBlock(r3)
	n2.ProcessBlock(r3)

	time.Sleep(time.Second)
	has1, _ := n1.ledger.HasStateBlockConfirmed(r2.GetHash())
	has2, _ := n1.ledger.HasStateBlockConfirmed(r3.GetHash())
	if has1 == has2 {
		t.Fatal()
	}

	// common fork
	acc3 := mock.Account()
	s4 := n1.GenerateSendBlock(TestAccount, acc3.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(s4)

	r4 := n1.GenerateReceiveBlock(s4, acc3)
	n1.ProcessBlockLocal(r4)

	// fork block on node 1, sleep to make different timestamp
	time.Sleep(time.Second)
	s5 := n2.GenerateSendBlock(TestAccount, acc3.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlock(s5)

	time.Sleep(time.Second)
	if has, _ := n1.ledger.HasStateBlockConfirmed(s5.GetHash()); has {
		t.Fatal("block not found")
	}
}

func TestBatchVoteDo(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}
	defer StopNodes(nodes)

	n1 := nodes[0]

	hashes := make([]types.Hash, 0)
	for i := 0; i < 5000; i++ {
		hash := mock.Hash()
		n1.dps.confirmedBlockInc(hash)
		hashes = append(hashes, hash)
	}

	for _, h := range hashes {
		n1.dps.batchVote <- h
	}

	for len(n1.dps.batchVote) > 0 {
		time.Sleep(time.Second)
	}
}

func TestOnline(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}
	defer StopNodes(nodes)

	n1 := nodes[0]

	var prevPov *types.PovBlock
	for i := 0; i < 160; i++ {
		pb, _ := mock.GeneratePovBlock(prevPov, 0)
		prevPov = pb
		n1.ctx.EventBus().Publish(topic.EventPovConnectBestBlock, pb)
		time.Sleep(10 * time.Millisecond)
	}

	n1.TestWithTimeout(30*time.Second, func() bool {
		repOnline := make(map[uint64]*RepOnlinePeriod)
		n1.cons.RPC(common.RpcDPoSOnlineInfo, nil, repOnline)
		return repOnline[0].Stat[TestAccount.Address()].HeartCount == 59
	})

	n1.TestWithTimeout(30*time.Second, func() bool {
		hasOnline := false
		err = n1.dps.ledger.GetStateBlocksConfirmed(func(block *types.StateBlock) error {
			if block.Type == types.Online && block.Address == TestAccount.Address() {
				hasOnline = true
			}
			return nil
		})

		if err != nil || !hasOnline {
			return false
		}

		return true
	})
}

func TestSynchronize(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}
	defer StopNodes(nodes)

	n1 := nodes[0]
	n2 := nodes[1]

	toAcc := mock.Account()
	var blocks types.StateBlockList
	for i := 0; i < 3; i++ {
		b := n1.GenerateSendBlock(TestAccount, toAcc.Address(), types.NewBalance(10), "QLC")
		n1.ProcessBlockLocal(b)
		blocks = append(blocks, b)
	}

	var bl types.StateBlockList
	frontier := blocks[len(blocks)-1]
	bl = append(bl, frontier)

	n2.cons.RPC(common.RpcDPoSOnSyncStateChange, topic.Syncing, nil)
	n2.cons.RPC(common.RpcDPoSProcessFrontier, bl, nil)
	n2.ctx.EventBus().Publish(topic.EventSyncBlock, blocks)
	n2.cons.RPC(common.RpcDPoSOnSyncStateChange, topic.SyncDone, nil)
	n2.VoteBlock(TestAccount, frontier)

	finishTimer := time.NewTimer(30 * time.Second)

	for {
		select {
		case <-finishTimer.C:
			t.Fatal(n2.dps.blockSyncState, n2.dps.povSyncState)
		default:
			if n2.dps.blockSyncState == topic.SyncFinish {
				n2.WaitBlockConfirmed(frontier.GetHash())
				finishTimer.Stop()
				goto SyncOneBlockTest
			}
			time.Sleep(time.Second)
		}
	}

SyncOneBlockTest:
	var bs types.StateBlockList
	b := n1.GenerateSendBlock(TestAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(b)
	bs = append(bs, b)

	n2.cons.RPC(common.RpcDPoSOnSyncStateChange, topic.Syncing, nil)
	n2.cons.RPC(common.RpcDPoSProcessFrontier, bs, nil)
	n2.ctx.EventBus().Publish(topic.EventSyncBlock, bs)
	n2.cons.RPC(common.RpcDPoSOnSyncStateChange, topic.SyncDone, nil)

	finishTimer.Reset(30 * time.Second)

	for {
		select {
		case <-finishTimer.C:
			t.Fatal(n2.dps.blockSyncState)
		default:
			if n2.dps.blockSyncState == topic.SyncFinish {
				n2.WaitBlockConfirmed(frontier.GetHash())
				return
			}
			time.Sleep(time.Second)
		}
	}
}

func TestRollback(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}
	defer StopNodes(nodes)

	n1 := nodes[0]

	toAcc := mock.Account()
	s := n1.GenerateSendBlock(TestAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(s)
	r := n1.GenerateReceiveBlock(s, toAcc)
	n1.ProcessBlockLocal(r)

	n1.WaitBlockConfirmed(s.GetHash())
	n1.WaitBlockConfirmed(r.GetHash())

	var blks []*types.StateBlock
	blks = append(blks, s)
	n1.dps.acTrx.rollBack(blks)

	n1.TestWithTimeout(30*time.Second, func() bool {
		if has, _ := n1.ledger.HasStateBlockConfirmed(s.GetHash()); has {
			return false
		}
		return true
	})

	n1.TestWithTimeout(30*time.Second, func() bool {
		if has, _ := n1.ledger.HasStateBlockConfirmed(r.GetHash()); has {
			return false
		}
		return true
	})
}

func TestGap(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}
	defer StopNodes(nodes)

	n1 := nodes[0]
	n2 := nodes[1]

	toAcc := mock.Account()
	// gap source/gap link
	s1 := n1.GenerateSendBlock(TestAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(s1)
	s2 := n1.GenerateSendBlock(TestAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(s2)
	r1 := n1.GenerateReceiveBlock(s1, toAcc)
	n1.ProcessBlockLocal(r1)

	// gap token
	cs1 := n1.GenerateContractSendBlock(TestAccount, toAcc, contractaddress.MintageAddress, mintage.MethodNameMintage, s2.GetHash())
	n1.ProcessBlockLocal(cs1)
	cr1 := n1.GenerateContractReceiveBlock(toAcc, contractaddress.MintageAddress, mintage.MethodNameMintage, cs1)
	n1.ProcessBlockLocal(cr1)
	time.Sleep(3 * time.Second)
	cs2 := n1.GenerateContractSendBlock(TestAccount, TestAccount, contractaddress.MintageAddress, mintage.MethodNameMintageWithdraw, cr1.Token)
	n1.ProcessBlockLocal(cs2)
	cr2 := n1.GenerateContractReceiveBlock(TestAccount, contractaddress.MintageAddress, mintage.MethodNameMintageWithdraw, cs2)
	n1.ProcessBlockLocal(cr2)

	hashes := make([]types.Hash, 0)
	hashes = append(hashes, s1.GetHash(), s2.GetHash(), r1.GetHash(), cs1.GetHash(), cr1.GetHash(), cs2.GetHash(), cr2.GetHash())

	n1.CheckBlocksConfirmed(hashes)

	hashBytes := make([]byte, 0)
	for _, h := range hashes {
		hashBytes = append(hashBytes, h[:]...)
	}
	hash, _ := types.HashBytes(hashBytes)
	vote := &protos.ConfirmAckBlock{
		Sequence:  n2.dps.getSeq(ackTypeCommon),
		Hash:      hashes,
		Account:   TestAccount.Address(),
		Signature: TestAccount.Sign(hash),
	}

	n2.ctx.EventBus().Publish(topic.EventConfirmAck, &p2p.EventConfirmAckMsg{Block: vote, From: "123"})
	n2.ctx.EventBus().Publish(topic.EventPublish, &topic.EventPublishMsg{Block: r1, From: "123"})
	n2.ctx.EventBus().Publish(topic.EventPublish, &topic.EventPublishMsg{Block: s2, From: "123"})
	n2.ctx.EventBus().Publish(topic.EventPublish, &topic.EventPublishMsg{Block: s1, From: "123"})
	n2.ctx.EventBus().Publish(topic.EventPublish, &topic.EventPublishMsg{Block: cs2, From: "123"})
	n2.ctx.EventBus().Publish(topic.EventPublish, &topic.EventPublishMsg{Block: cr2, From: "123"})
	n2.ctx.EventBus().Publish(topic.EventPublish, &topic.EventPublishMsg{Block: cs1, From: "123"})
	n2.ctx.EventBus().Publish(topic.EventPublish, &topic.EventPublishMsg{Block: cr1, From: "123"})

	n2.WaitBlocksConfirmed(hashes)
}
