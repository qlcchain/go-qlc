package consensus

import (
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
)

type ConsensusAlgorithm interface {
	Init()
	Start()
	Stop()
	ProcessMsg(bs *BlockSource)
	RPC(kind uint, in, out interface{})
}

type Consensus struct {
	ca       ConsensusAlgorithm
	recv     *Receiver
	logger   *zap.SugaredLogger
	ledger   ledger.Store
	verifier *process.LedgerVerifier
}

var GlobalUncheckedBlockNum atomic.Uint64

func NewConsensus(ca ConsensusAlgorithm, cfgFile string) *Consensus {
	cc := context.NewChainContext(cfgFile)
	l := ledger.NewLedger(cfgFile)
	return &Consensus{
		ca:       ca,
		recv:     NewReceiver(cc.EventBus()),
		logger:   log.NewLogger("consensus"),
		ledger:   l,
		verifier: process.NewLedgerVerifier(l),
	}
}

func (c *Consensus) Init() {
	c.ca.Init()
	c.recv.init(c)
}

func (c *Consensus) Start() {
	go c.ca.Start()

	err := c.recv.start()
	if err != nil {
		c.logger.Error("receiver start failed")
	}
}

func (c *Consensus) Stop() {
	err := c.recv.stop()
	if err != nil {
		c.logger.Error("receiver stop failed")
	}

	c.ca.Stop()
}

func (c *Consensus) RPC(kind uint, in, out interface{}) {
	c.ca.RPC(kind, in, out)
}
