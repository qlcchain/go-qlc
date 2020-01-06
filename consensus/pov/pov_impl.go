package pov

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/bluele/gcache"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
)

const (
	blkCacheSize       = 10240
	blkCacheExpireTime = 3 * time.Minute
)

type PoVEngine struct {
	logger     *zap.SugaredLogger
	cfg        *config.Config
	ledger     *ledger.Ledger
	eb         event.EventBus
	feb        *event.FeedEventBus
	subscriber *event.ActorSubscriber
	accounts   []*types.Account

	blkRecvCache gcache.Cache

	bp       *PovBlockProcessor
	txpool   *PovTxPool
	chain    *PovBlockChain
	cs       ConsensusPov
	verifier *PovVerifier
	syncer   *PovSyncer

	quitCh         chan struct{}
	febRpcMsgCh    chan *topic.EventRPCSyncCallMsg
	febRpcMsgSubID event.FeedSubscription
}

func NewPovEngine(cfgFile string) (*PoVEngine, error) {
	cc := context.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	l := ledger.NewLedger(cfgFile)

	pov := &PoVEngine{
		logger:   log.NewLogger("pov_engine"),
		cfg:      cfg,
		eb:       cc.EventBus(),
		feb:      cc.FeedEventBus(),
		accounts: cc.Accounts(),
		ledger:   l,

		quitCh:      make(chan struct{}),
		febRpcMsgCh: make(chan *topic.EventRPCSyncCallMsg, 1),
	}

	pov.blkRecvCache = gcache.New(blkCacheSize).Simple().Expiration(blkCacheExpireTime).Build()

	pov.chain = NewPovBlockChain(cfg, pov.eb, pov.ledger)
	pov.txpool = NewPovTxPool(pov.eb, pov.ledger, pov.chain)
	pov.cs = NewPovConsensus(PovConsensusModePow, pov.chain)
	pov.verifier = NewPovVerifier(l, pov.chain, pov.cs)
	pov.syncer = NewPovSyncer(pov.eb, pov.ledger, pov.chain)

	pov.bp = NewPovBlockProcessor(pov.eb, pov.ledger, pov.chain, pov.verifier, pov.syncer)

	return pov, nil
}

func (pov *PoVEngine) Init() error {
	err := pov.bp.Init()
	if err != nil {
		return err
	}
	err = pov.chain.Init()
	if err != nil {
		return err
	}
	err = pov.cs.Init()
	if err != nil {
		return err
	}
	err = pov.txpool.Init()
	if err != nil {
		return err
	}

	return nil
}

func (pov *PoVEngine) Start() error {
	pov.logger.Info("start pov engine service")

	err := pov.txpool.Start()
	if err != nil {
		return err
	}

	err = pov.chain.Start()
	if err != nil {
		return err
	}

	err = pov.cs.Start()
	if err != nil {
		return err
	}

	err = pov.bp.Start()
	if err != nil {
		return err
	}

	err = pov.syncer.Start()
	if err != nil {
		return err
	}

	err = pov.setEvent()
	if err != nil {
		return err
	}

	common.Go(pov.loop)

	return nil
}

func (pov *PoVEngine) Stop() error {
	pov.logger.Info("stop pov engine service")

	close(pov.quitCh)

	pov.unsetEvent()

	pov.syncer.Stop()

	pov.txpool.Stop()

	_ = pov.cs.Stop()

	_ = pov.chain.Stop()

	_ = pov.bp.Stop()

	return nil
}

func (pov *PoVEngine) GetConfig() *config.Config {
	return pov.cfg
}

func (pov *PoVEngine) GetLogger() *zap.SugaredLogger {
	return pov.logger
}

func (pov *PoVEngine) GetEventBus() event.EventBus {
	return pov.eb
}

func (pov *PoVEngine) GetLedger() ledger.Store {
	return pov.ledger
}

func (pov *PoVEngine) GetChain() *PovBlockChain {
	return pov.chain
}

func (pov *PoVEngine) GetConsensus() ConsensusPov {
	return pov.cs
}

func (pov *PoVEngine) GetTxPool() *PovTxPool {
	return pov.txpool
}

func (pov *PoVEngine) GetVerifier() *PovVerifier {
	return pov.verifier
}

func (pov *PoVEngine) GetAccounts() []*types.Account {
	return pov.accounts
}

func (pov *PoVEngine) AddMinedBlock(block *types.PovBlock) error {
	_ = pov.blkRecvCache.Set(block.GetHash(), struct{}{})
	err := pov.bp.AddMinedBlock(block)
	if err == nil {
		pov.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.PovPublishReq, Message: block})
	}
	return err
}

func (pov *PoVEngine) AddBlock(block *types.PovBlock, from types.PovBlockFrom, peerID string) error {
	blockHash := block.GetHash()

	stat := pov.verifier.VerifyNet(block)
	if stat.Result != process.Progress {
		pov.logger.Infof("block %s verify net err %s", blockHash, stat.ErrMsg)
		return fmt.Errorf("block %s verify net err %s", blockHash, stat.ErrMsg)
	}

	err := pov.bp.AddBlock(block, from, peerID)
	return err
}

func (pov *PoVEngine) setEvent() error {
	pov.febRpcMsgSubID = pov.feb.Subscribe(topic.EventRpcSyncCall, pov.febRpcMsgCh)
	if pov.febRpcMsgSubID == nil {
		pov.logger.Error("failed to subscribe EventRpcSyncCall")
	}

	pov.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *topic.EventPovRecvBlockMsg:
			if err := pov.onRecvPovBlock(msg.Block, msg.From, msg.MsgPeer); err != nil {
				pov.logger.Error(err)
			}
		}
	}), pov.eb)

	if err := pov.subscriber.Subscribe(topic.EventPovRecvBlock); err != nil {
		pov.logger.Error("failed to subscribe events")
		return err
	}

	return nil
}

func (pov *PoVEngine) unsetEvent() {
	pov.febRpcMsgSubID.Unsubscribe()

	if err := pov.subscriber.UnsubscribeAll(); err != nil {
		pov.logger.Errorf("failed to unsubscribe events %s", err)
	}
}

func (pov *PoVEngine) onRecvPovBlock(block *types.PovBlock, from types.PovBlockFrom, msgPeer string) error {
	if from == types.PovBlockFromLocal {
		return pov.AddMinedBlock(block)
	}

	if from == types.PovBlockFromRemoteBroadcast {
		blockHash := block.GetHash()

		if pov.blkRecvCache.Has(blockHash) {
			return nil
		}
		_ = pov.blkRecvCache.Set(blockHash, struct{}{})

		pov.logger.Infof("receive broadcast block %d/%s from %s", block.GetHeight(), blockHash, msgPeer)
	}

	err := pov.AddBlock(block, from, msgPeer)
	if err == nil {
		if from == types.PovBlockFromRemoteBroadcast {
			pov.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.PovPublishReq, Message: block})
		}
	}

	return err
}

func (pov *PoVEngine) loop() {
	for {
		select {
		case <-pov.quitCh:
			return
		case msg := <-pov.febRpcMsgCh:
			pov.onEventRPCSyncCall(msg)
		}
	}
}

func (pov *PoVEngine) onEventRPCSyncCall(msg *topic.EventRPCSyncCallMsg) {
	needRsp := false
	if msg.Name == "Debug.PovInfo" {
		pov.getDebugInfo(msg.In, msg.Out)
		needRsp = true
	}
	if needRsp && msg.ResponseChan != nil {
		msg.ResponseChan <- msg.Out
	}
}

func (pov *PoVEngine) getDebugInfo(in interface{}, out interface{}) {
	outArgs := out.(map[string]interface{})

	outArgs["err"] = nil

	if pov.syncer != nil {
		outArgs["syncInfo"] = pov.syncer.GetDebugInfo()
	}

	if pov.bp != nil {
		outArgs["procInfo"] = pov.bp.GetDebugInfo()
	}

	if pov.chain != nil {
		outArgs["chainInfo"] = pov.chain.GetDebugInfo()
	}

	if pov.txpool != nil {
		outArgs["poolInfo"] = pov.txpool.GetDebugInfo()
	}
}
