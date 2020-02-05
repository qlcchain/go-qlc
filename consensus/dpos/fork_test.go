package dpos

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
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

	// fork block on node 1
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

	time.Sleep(10*time.Second)
	has1, _ := n1.ledger.HasStateBlockConfirmed(recv1.GetHash())
	has2, _ := n1.ledger.HasStateBlockConfirmed(recv2.GetHash())
	t.Log(has1, has2)
	if has1 == has2 {
		t.Fatal()
	}
}
