package dpos

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"testing"
	"time"
)

func TestFork1(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	acc := mock.Account()
	send1 := n1.GenerateSendBlock(testAccount, acc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(send1)

	recv1 := n1.GenerateReceiveBlock(send1, acc)
	n1.ProcessBlockLocal(recv1)

	// fork block on node 1, sleep to make different timestamp
	time.Sleep(2 * time.Second)
	send2 := n2.GenerateSendBlock(testAccount, acc.Address(), types.NewBalance(10), "QLC")

	n2.ProcessBlock(send2)
	time.Sleep(10 * time.Second)

	if has, _ := n1.ledger.HasStateBlockConfirmed(send2.GetHash()); has {
		t.Fatal()
	}
}

func TestFork2(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	acc1 := mock.Account()
	acc2 := mock.Account()

	s1, r1 := n1.TokenTransactionAndConfirmed(testAccount, acc1, types.NewBalance(10000), "QLC")
	n2.WaitBlockConfirmed(s1.GetHash())
	n2.WaitBlockConfirmed(r1.GetHash())

	send1 := n1.GenerateSendBlock(testAccount, acc2.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockAndWaitConfirmed(send1)

	send2 := n1.GenerateSendBlock(acc1, acc2.Address(), types.NewBalance(20), "QLC")
	n1.ProcessBlockAndWaitConfirmed(send2)
	n2.WaitBlockConfirmed(send2.GetHash())

	recv1 := n1.GenerateReceiveBlock(send1, acc2)
	recv2 := n2.GenerateReceiveBlock(send2, acc2)
	n1.ProcessBlock(recv1)
	n2.ProcessBlock(recv2)

	time.Sleep(10 * time.Second)
	has1, _ := n1.ledger.HasStateBlockConfirmed(recv1.GetHash())
	has2, _ := n1.ledger.HasStateBlockConfirmed(recv2.GetHash())
	t.Log(has1, has2)
	if has1 == has2 {
		t.Fatal()
	}
}

func TestTransaction(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	toAcc := mock.Account()
	send := n1.GenerateSendBlock(testAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockAndWaitConfirmed(send)
	n2.WaitBlockConfirmed(send.GetHash())
}

func TestBatchVoteDo(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	hashes := make([]types.Hash, 0)
	for i := 0; i < 2000; i++ {
		hash := mock.Hash()
		n2.dps.confirmedBlockInc(hash)
		hashes = append(hashes, hash)
	}

	for _, h := range hashes {
		n1.dps.batchVote <- h
	}

	time.Sleep(5 * time.Second)
	repOnline := make(map[uint64]*RepOnlinePeriod, 0)
	n2.cons.RPC(common.RpcDPosOnlineInfo, nil, repOnline)
	if repOnline[0].Stat[testAccount.Address()].VoteCount != 2000 {
		t.Fatal()
	}
}

func TestOnline(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	var prevPov *types.PovBlock
	for i := 0; i < 160; i++ {
		pb, _ := mock.GeneratePovBlock(prevPov, 0)
		prevPov = pb
		n1.ctx.EventBus().Publish(topic.EventPovConnectBestBlock, pb)
		time.Sleep(100 * time.Millisecond)
	}

	repOnline := make(map[uint64]*RepOnlinePeriod, 0)
	n1.cons.RPC(common.RpcDPosOnlineInfo, nil, repOnline)
	if repOnline[0].Stat[testAccount.Address()].HeartCount != 59 {
		t.Fatal()
	}

	time.Sleep(3 * time.Second)

	hasOnline := false
	err = n1.dps.ledger.GetStateBlocks(func(block *types.StateBlock) error {
		fmt.Println(block.Type)
		if block.Type == types.Online && block.Address == testAccount.Address() {
			hasOnline = true
		}
		return nil
	})

	if err != nil || !hasOnline {
		t.Fatal()
	}
}

func TestSynchronize(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	toAcc := mock.Account()
	var blocks types.StateBlockList
	for i := 0; i < 10; i++ {
		b := n1.GenerateSendBlock(testAccount, toAcc.Address(), types.NewBalance(10), "QLC")
		n1.ProcessBlockLocal(b)
		blocks = append(blocks, b)
	}

	var bl types.StateBlockList
	frontier := blocks[len(blocks)-1]
	bl = append(bl, frontier)

	n2.cons.RPC(common.RpcDPosOnSyncStateChange, topic.Syncing, nil)
	n2.cons.RPC(common.RpcDPosProcessFrontier, bl, nil)
	n2.ctx.EventBus().Publish(topic.EventSyncBlock, blocks)
	n2.cons.RPC(common.RpcDPosOnSyncStateChange, topic.SyncDone, nil)

	finishTimer := time.NewTimer(30 * time.Second)

	for {
		select {
		case <-finishTimer.C:
			t.Fatal(n2.dps.blockSyncState)
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
	b := n1.GenerateSendBlock(testAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(b)
	bs = append(bs, b)

	n2.cons.RPC(common.RpcDPosOnSyncStateChange, topic.Syncing, nil)
	n2.cons.RPC(common.RpcDPosProcessFrontier, bs, nil)
	n2.ctx.EventBus().Publish(topic.EventSyncBlock, bs)
	n2.cons.RPC(common.RpcDPosOnSyncStateChange, topic.SyncDone, nil)

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

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	toAcc := mock.Account()
	s, r := n1.TokenTransactionAndConfirmed(testAccount, toAcc, types.NewBalance(10), "QLC")

	if has, _ := n1.ledger.HasStateBlockConfirmed(s.GetHash()); !has {
		t.Fatal()
	}

	if has, _ := n1.ledger.HasStateBlockConfirmed(r.GetHash()); !has {
		t.Fatal()
	}

	var blks []*types.StateBlock
	blks = append(blks, s)
	n1.dps.acTrx.rollBack(blks)

	if has, _ := n1.ledger.HasStateBlockConfirmed(s.GetHash()); has {
		t.Fatal()
	}

	if has, _ := n1.ledger.HasStateBlockConfirmed(r.GetHash()); has {
		t.Fatal()
	}
}

func TestGap(t *testing.T) {
	nodes, err := InitNodes(2, t)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		for _, n := range nodes {
			n.StopNodeAndRemoveDir()
		}
	}()

	for _, n := range nodes {
		n.RunNode()
		n.InitLedger()
	}

	n1 := nodes[0]
	n2 := nodes[1]

	n1.InitStatus()
	n2.InitStatus()

	toAcc := mock.Account()
	// gap source/gap link
	s1 := n1.GenerateSendBlock(testAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(s1)
	s2 := n1.GenerateSendBlock(testAccount, toAcc.Address(), types.NewBalance(10), "QLC")
	n1.ProcessBlockLocal(s2)
	r1 := n1.GenerateReceiveBlock(s1, toAcc)
	n1.ProcessBlockLocal(r1)

	// gap token
	cs1 := n1.GenerateContractSendBlock(testAccount, toAcc, types.MintageAddress, abi.MethodNameMintage, s2.GetHash())
	n1.ProcessBlockLocal(cs1)
	cr1 := n1.GenerateContractReceiveBlock(toAcc, types.MintageAddress, abi.MethodNameMintage, cs1)
	n1.ProcessBlockLocal(cr1)
	time.Sleep(time.Second)
	cs2 := n1.GenerateContractSendBlock(testAccount, testAccount, types.MintageAddress, abi.MethodNameMintageWithdraw, cr1.Token)
	n1.ProcessBlockLocal(cs2)
	cr2 := n1.GenerateContractReceiveBlock(testAccount, types.MintageAddress, abi.MethodNameMintageWithdraw, cs2)
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
		Account:   testAccount.Address(),
		Signature: testAccount.Sign(hash),
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
