package consensus

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type algTest struct{}

var flagTest int

func (a algTest) Init() {
	flagTest = 1
}

func (a algTest) Start() {
	flagTest = 2
}

func (a algTest) Stop() {
	flagTest = 3
}

func (a algTest) ProcessMsg(bs *BlockSource) {
	flagTest++
}

func (a algTest) RPC(kind uint, in, out interface{}) {
	flagTest = 5
}

func setupEnv(t *testing.T) (func(t *testing.T), string) {
	dir := filepath.Join(config.QlcTestDataDir(), "consensus", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}

	return func(t *testing.T) {
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, cm.ConfigFile
}

func TestConsensus(t *testing.T) {
	teardown, cfgFile := setupEnv(t)
	defer teardown(t)

	alg := new(algTest)
	c := NewConsensus(alg, cfgFile)
	c.Init()
	if flagTest != 1 {
		t.Fatal()
	}

	c.Start()
	time.Sleep(time.Second)
	if flagTest != 2 {
		t.Fatal()
	}

	c.RPC(0, nil, nil)
	if flagTest != 5 {
		t.Fatal()
	}

	block := mock.StateBlockWithoutWork()
	c.recv.eb.Publish(topic.EventPublish, &topic.EventPublishMsg{Block: block, From: "123"})
	time.Sleep(time.Second)
	if flagTest != 6 {
		t.Fatal()
	}

	c.recv.eb.Publish(topic.EventConfirmReq, &topic.EventConfirmReqMsg{Blocks: []*types.StateBlock{block}, From: "123"})
	time.Sleep(time.Second)
	if flagTest != 7 {
		t.Fatal()
	}

	c.recv.eb.Publish(topic.EventGenerateBlock, block)
	time.Sleep(time.Second)
	if flagTest != 8 {
		t.Fatal()
	}

	c.recv.eb.Publish(topic.EventSyncBlock, types.StateBlockList{block})
	time.Sleep(time.Second)
	if flagTest != 9 {
		t.Fatal()
	}

	account := mock.Account()
	hash := mock.Hash()
	signHash, _ := types.HashBytes(hash.Bytes())
	ack := &protos.ConfirmAckBlock{
		Account:   account.Address(),
		Signature: account.Sign(signHash),
		Sequence:  0,
		Hash:      []types.Hash{hash},
	}
	c.recv.eb.Publish(topic.EventConfirmAck, &p2p.EventConfirmAckMsg{Block: ack, From: "123"})
	time.Sleep(time.Second)
	if flagTest != 10 {
		t.Fatal()
	}

	c.Stop()
	if flagTest != 3 {
		t.Fatal()
	}
}
