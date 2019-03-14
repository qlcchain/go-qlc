package consensus

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"go.uber.org/zap"
)

const (
	msgCacheSize                      = 65536
	msgCacheExpirationTime            = 30 * time.Minute
	findOnlineRepresentativesInterval = 2 * time.Minute
	repTimeout                        = 5 * time.Minute
)

var (
	localRepAccount sync.Map
)

type DPoS struct {
	ns         p2p.Service
	ledger     *ledger.Ledger
	verifier   *process.LedgerVerifier
	eventMsg   map[p2p.EventType]p2p.EventSubscriber
	bp         *BlockProcessor
	acTrx      *ActiveTrx
	accounts   []*types.Account
	onlineReps sync.Map
	logger     *zap.SugaredLogger
	cache      gcache.Cache
	cfg        *config.Config
}

type repInfo struct {
	time  int64
	state bool
}

func (dps *DPoS) GetP2PService() p2p.Service {
	return dps.ns
}

func (dps *DPoS) Init() error {
	if len(dps.accounts) != 0 {
		dps.refreshAccount()
	}
	return nil
}

func (dps *DPoS) refreshAccount() {
	var b bool
	var addr types.Address
	for _, v := range dps.accounts {
		addr = v.Address()
		b = dps.isRepresentation(addr)
		if b {
			_, _ = localRepAccount.LoadOrStore(addr, v)
		}
	}
	var count uint32
	localRepAccount.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	if count > 1 {
		dps.logger.Error("it is very dangerous to run two or more representatives on one node")
	}
}

func (dps *DPoS) Start() error {
	dps.setEvent()
	dps.logger.Info("start dpos service")
	go dps.bp.Start()
	go dps.acTrx.start()

	return nil
}

func (dps *DPoS) Stop() error {
	dps.bp.quitCh <- true
	dps.acTrx.quitCh <- true
	for i, j := range dps.eventMsg {
		_ = dps.ns.MessageEvent().GetEvent("consensus").UnSubscribe(i, j)
	}
	return nil
}

func NewDPoS(cfg *config.Config, netService p2p.Service, accounts []*types.Account) (*DPoS, error) {
	bp := NewBlockProcessor()
	acTrx := NewActiveTrx()
	l := ledger.NewLedger(cfg.LedgerDir())

	dps := &DPoS{
		ns:       netService,
		ledger:   l,
		verifier: process.NewLedgerVerifier(l),
		eventMsg: make(map[p2p.EventType]p2p.EventSubscriber),
		bp:       bp,
		acTrx:    acTrx,
		accounts: accounts,
		logger:   log.NewLogger("consensus"),
		cache:    gcache.New(msgCacheSize).LRU().Expiration(msgCacheExpirationTime).Build(),
		cfg:      cfg,
	}
	dps.bp.SetDpos(dps)
	dps.acTrx.SetDposService(dps)
	return dps, nil
}

func (dps *DPoS) setEvent() {
	event1 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventPublish, dps.ReceivePublish)
	event2 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmReq, dps.ReceiveConfirmReq)
	event3 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmAck, dps.ReceiveConfirmAck)
	event4 := dps.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventSyncBlock, dps.ReceiveSyncBlock)
	dps.eventMsg[p2p.EventPublish] = event1
	dps.eventMsg[p2p.EventConfirmReq] = event2
	dps.eventMsg[p2p.EventConfirmAck] = event3
	dps.eventMsg[p2p.EventSyncBlock] = event4
}

func (dps *DPoS) ReceivePublish(v interface{}) {
	dps.logger.Info("Publish Event")
	e := v.(p2p.Message)
	p, err := protos.PublishBlockFromProto(e.Data())
	if err != nil {
		dps.logger.Info(err)
		return
	}
	bs := blockSource{
		block:     p.Blk,
		blockFrom: types.UnSynchronized,
	}
	dps.bp.blocks <- bs
	dps.onReceivePublish(e, p.Blk)
}

func (dps *DPoS) onReceivePublish(e p2p.Message, blk *types.StateBlock) {
	if !dps.cache.Has(e.Hash()) {
		dps.ns.SendMessageToPeers(p2p.PublishReq, blk, e.MessageFrom())
		err := dps.cache.Set(e.Hash(), "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with publish message", err, blk.GetHash())
		}
	}
}

func (dps *DPoS) ReceiveConfirmReq(v interface{}) {
	dps.logger.Info("ConfirmReq Event")
	e := v.(p2p.Message)
	r, err := protos.ConfirmReqBlockFromProto(e.Data())
	if err != nil {
		dps.logger.Error(err)
		return
	}
	dps.onReceiveConfirmReq(e, r.Blk)
}

func (dps *DPoS) onReceiveConfirmReq(e p2p.Message, blk *types.StateBlock) {
	var address types.Address
	var count uint32
	bs := blockSource{
		block:     blk,
		blockFrom: types.UnSynchronized,
	}
	if !dps.cache.Has(e.Hash()) {
		localRepAccount.Range(func(key, value interface{}) bool {
			count++
			address = key.(types.Address)
			dps.saveOnlineRep(&address)
			result, _ := dps.verifier.Process(bs.block)
			if result == process.Old {
				dps.logger.Infof("send confirm ack for hash %s,previous hash is %s", bs.block.GetHash(), bs.block.Root())
				dps.sendConfirmAck(bs.block, address, value.(*types.Account))
			}
			dps.bp.processResult(result, bs)
			return true
		})
		if count == 0 {
			dps.bp.blocks <- bs
		}
		dps.ns.SendMessageToPeers(p2p.ConfirmReq, blk, e.MessageFrom())
		err := dps.cache.Set(e.Hash(), "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with confirmReq message", err, blk.GetHash())
		}
	}
}

func (dps *DPoS) ReceiveConfirmAck(v interface{}) {
	dps.logger.Info("ConfirmAck Event")
	e := v.(p2p.Message)
	ack, err := protos.ConfirmAckBlockFromProto(e.Data())
	if err != nil {
		dps.logger.Info(err)
		return
	}
	dps.onReceiveConfirmAck(e, ack)
}

func (dps *DPoS) onReceiveConfirmAck(e p2p.Message, ack *protos.ConfirmAckBlock) {
	var address types.Address
	var count uint32
	bs := blockSource{
		block:     ack.Blk,
		blockFrom: types.UnSynchronized,
	}
	valid := IsAckSignValidate(ack)
	if !valid {
		return
	}
	dps.acTrx.vote(ack)
	if !dps.cache.Has(e.Hash()) {
		dps.saveOnlineRep(&ack.Account)
		localRepAccount.Range(func(key, value interface{}) bool {
			count++
			address = key.(types.Address)
			dps.saveOnlineRep(&address)
			result, _ := dps.verifier.Process(bs.block)
			if result == process.Old {
				dps.logger.Infof("send confirm ack for hash %s,previous hash is %s", bs.block.GetHash(), bs.block.Root())
				dps.sendConfirmAck(bs.block, address, value.(*types.Account))
			}
			dps.bp.processResult(result, bs)
			if result == process.Progress {
				dps.acTrx.vote(ack)
			}
			return true
		})
		if count == 0 {
			dps.bp.blocks <- bs
		}

		dps.ns.SendMessageToPeers(p2p.ConfirmAck, ack, e.MessageFrom())
		err := dps.cache.Set(e.Hash(), "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with confirmAck message", err, ack.Blk.GetHash())
		}
	}
}

func (dps *DPoS) ReceiveSyncBlock(v interface{}) {
	dps.logger.Info("Sync Event")
	bs := blockSource{
		block:     v.(*types.StateBlock),
		blockFrom: types.Synchronized,
	}
	dps.bp.blocks <- bs
}

func (dps *DPoS) sendConfirmAck(block *types.StateBlock, account types.Address, acc *types.Account) error {
	va, err := dps.voteGenerate(block, account, acc)
	if err != nil {
		dps.logger.Error("vote generate error")
		return err
	}
	dps.ns.Broadcast(p2p.ConfirmAck, va)
	return nil
}

func (dps *DPoS) voteGenerate(block *types.StateBlock, account types.Address, acc *types.Account) (*protos.ConfirmAckBlock, error) {
	va := &protos.ConfirmAckBlock{
		Sequence:  0,
		Blk:       block,
		Account:   account,
		Signature: acc.Sign(block.GetHash()),
	}
	return va, nil
}

func (dps *DPoS) isRepresentation(address types.Address) bool {
	if _, err := dps.ledger.GetRepresentation(address); err != nil {
		return false
	}
	return true
}

func (dps *DPoS) saveOnlineRep(addr *types.Address) {
	now := time.Now().Add(repTimeout).UTC().Unix()
	_, _ = dps.onlineReps.LoadOrStore(addr, now)
}

func (dps *DPoS) GetOnlineRepresentatives() []*types.Address {
	var repAddresses []*types.Address
	dps.onlineReps.Range(func(key, value interface{}) bool {

		addr := key.(*types.Address)
		repAddresses = append(repAddresses, addr)
		return true
	})
	return repAddresses
}

func (dps *DPoS) findOnlineRepresentatives() error {
	var address types.Address
	localRepAccount.Range(func(key, value interface{}) bool {
		address = key.(types.Address)
		dps.saveOnlineRep(&address)
		return true
	})
	blk, err := dps.ledger.GetRandomStateBlock()
	if err != nil {
		return err
	}
	dps.ns.Broadcast(p2p.ConfirmReq, blk)
	return nil
}

func (dps *DPoS) cleanOnlineReps() {
	var repAddresses []*types.Address
	now := time.Now().UTC().Unix()
	dps.onlineReps.Range(func(key, value interface{}) bool {
		addr := key.(*types.Address)
		v := value.(int64)
		if v < now {
			dps.onlineReps.Delete(addr)
		} else {
			repAddresses = append(repAddresses, addr)
		}
		return true
	})
	_ = dps.ledger.SetOnlineRepresentations(repAddresses)
}
