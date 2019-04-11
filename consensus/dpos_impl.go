package consensus

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
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
	msgCacheExpirationTime            = 15 * time.Minute
	findOnlineRepresentativesInterval = 2 * time.Minute
	repTimeout                        = 5 * time.Minute
)

var (
	localRepAccount sync.Map
)

type DPoS struct {
	ledger     *ledger.Ledger
	verifier   *process.LedgerVerifier
	eb         event.EventBus
	bp         *BlockProcessor
	acTrx      *ActiveTrx
	accounts   []*types.Account
	onlineReps sync.Map
	logger     *zap.SugaredLogger
	cache      gcache.Cache
	cfg        *config.Config
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
			localRepAccount.Store(addr, v)
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
	err := dps.setEvent()
	if err != nil {
		return err
	}
	dps.logger.Info("start dpos service")
	go dps.bp.Start()
	go dps.acTrx.start()

	return nil
}

func (dps *DPoS) Stop() error {
	dps.bp.quitCh <- true
	dps.acTrx.quitCh <- true
	err := dps.eb.Unsubscribe(string(common.EventPublish), dps.ReceivePublish)
	if err != nil {
		return err
	}
	err = dps.eb.Unsubscribe(string(common.EventConfirmReq), dps.ReceiveConfirmReq)
	if err != nil {
		return err
	}
	err = dps.eb.Unsubscribe(string(common.EventConfirmAck), dps.ReceiveConfirmAck)
	if err != nil {
		return err
	}
	err = dps.eb.Unsubscribe(string(common.EventSyncBlock), dps.ReceiveSyncBlock)
	if err != nil {
		return err
	}
	return nil
}

func NewDPoS(cfg *config.Config, accounts []*types.Account, eb event.EventBus) (*DPoS, error) {
	bp := NewBlockProcessor()
	acTrx := NewActiveTrx()
	l := ledger.NewLedger(cfg.LedgerDir(), eb)

	dps := &DPoS{
		ledger:   l,
		verifier: process.NewLedgerVerifier(l),
		eb:       eb,
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

func (dps *DPoS) setEvent() error {
	err := dps.eb.SubscribeAsync(string(common.EventPublish), dps.ReceivePublish, false)
	if err != nil {
		return err
	}
	err = dps.eb.SubscribeAsync(string(common.EventConfirmReq), dps.ReceiveConfirmReq, false)
	if err != nil {
		return err
	}
	err = dps.eb.SubscribeAsync(string(common.EventConfirmAck), dps.ReceiveConfirmAck, false)
	if err != nil {
		return err
	}
	err = dps.eb.SubscribeAsync(string(common.EventSyncBlock), dps.ReceiveSyncBlock, false)
	if err != nil {
		return err
	}
	return nil
}

func (dps *DPoS) ReceivePublish(blk *types.StateBlock, hash types.Hash, msgFrom string) {
	dps.logger.Infof("receive publish block [%s] from [%s]", blk.GetHash(), msgFrom)
	bs := blockSource{
		block:     blk,
		blockFrom: types.UnSynchronized,
	}
	dps.bp.blocks <- bs
	dps.onReceivePublish(hash, blk, msgFrom)
}

func (dps *DPoS) onReceivePublish(hash types.Hash, blk *types.StateBlock, msgFrom string) {
	if !dps.cache.Has(hash) {
		//dps.ns.SendMessageToPeers(p2p.PublishReq, blk, msgFrom)
		dps.eb.Publish(string(common.EventSendMsgToPeers), common.PublishReq, blk, msgFrom)
		err := dps.cache.Set(hash, "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with publish message", err, blk.GetHash())
		}
	}
}

func (dps *DPoS) ReceiveConfirmReq(blk *types.StateBlock, hash types.Hash, msgFrom string) {
	dps.logger.Infof("receive ConfirmReq block [%s] from [%s]", blk.GetHash(), msgFrom)
	var address types.Address
	var count uint32
	bs := blockSource{
		block:     blk,
		blockFrom: types.UnSynchronized,
	}
	if !dps.cache.Has(hash) {
		localRepAccount.Range(func(key, value interface{}) bool {
			count++
			address = key.(types.Address)
			dps.saveOnlineRep(address)
			result, _ := dps.verifier.Process(bs.block)
			if result == process.Old {
				err := dps.sendAckIfResultIsOld(bs.block, address, value.(*types.Account))
				if err != nil {
					return true
				}
			} else {
				dps.bp.processResult(result, bs)
			}
			return true
		})
		if count == 0 {
			dps.bp.blocks <- bs
			//dps.ns.SendMessageToPeers(p2p.ConfirmReq, blk, msgFrom)
			dps.eb.Publish(string(common.EventSendMsgToPeers), common.ConfirmReq, blk, msgFrom)
			err := dps.cache.Set(hash, "")
			if err != nil {
				dps.logger.Errorf("Set cache error [%s] for block [%s] with confirmReq message", err, blk.GetHash())
			}
		}
	}
}

func (dps *DPoS) ReceiveConfirmAck(ack *protos.ConfirmAckBlock, hash types.Hash, msgFrom string) {
	dps.logger.Infof("receive ConfirmAck block [%s] from [%s]", ack.Blk.GetHash(), msgFrom)
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
	if !dps.cache.Has(hash) {
		dps.saveOnlineRep(ack.Account)
		localRepAccount.Range(func(key, value interface{}) bool {
			count++
			address = key.(types.Address)
			dps.saveOnlineRep(address)
			result, _ := dps.verifier.Process(bs.block)
			if result == process.Old {
				err := dps.sendAckIfResultIsOld(bs.block, address, value.(*types.Account))
				if err != nil {
					return true
				}
			} else {
				dps.bp.processResult(result, bs)
			}
			if result == process.Progress {
				dps.acTrx.vote(ack)
			}
			return true
		})
		if count == 0 {
			dps.bp.blocks <- bs
		}

		//dps.ns.SendMessageToPeers(p2p.ConfirmAck, ack, msgFrom)
		dps.eb.Publish(string(common.EventSendMsgToPeers), common.ConfirmAck, ack, msgFrom)
		err := dps.cache.Set(hash, "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with confirmAck message", err, ack.Blk.GetHash())
		}
	}
}

func (dps *DPoS) ReceiveSyncBlock(blk *types.StateBlock) {
	dps.logger.Info("Sync Event")
	bs := blockSource{
		block:     blk,
		blockFrom: types.Synchronized,
	}
	dps.logger.Infof("Sync Event for block:[%s]", bs.block.GetHash())
	dps.bp.blocks <- bs
}

func (dps *DPoS) sendConfirmAck(block *types.StateBlock, account types.Address, acc *types.Account) error {
	va, err := dps.voteGenerate(block, account, acc)
	if err != nil {
		dps.logger.Error("vote generate error")
		return err
	}
	//dps.ns.Broadcast(p2p.ConfirmAck, va)
	dps.eb.Publish(string(common.EventBroadcast), common.ConfirmAck, va)
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

func (dps *DPoS) saveOnlineRep(addr types.Address) {
	now := time.Now().Add(repTimeout).UTC().Unix()
	dps.onlineReps.Store(addr, now)
}

func (dps *DPoS) GetOnlineRepresentatives() []types.Address {
	var repAddresses []types.Address
	dps.onlineReps.Range(func(key, value interface{}) bool {

		addr := key.(types.Address)
		repAddresses = append(repAddresses, addr)
		return true
	})
	return repAddresses
}

func (dps *DPoS) findOnlineRepresentatives() error {
	var address types.Address
	localRepAccount.Range(func(key, value interface{}) bool {
		address = key.(types.Address)
		dps.saveOnlineRep(address)
		return true
	})
	blk, err := dps.ledger.GetRandomStateBlock()
	if err != nil {
		return err
	}
	//dps.ns.Broadcast(p2p.ConfirmReq, blk)
	dps.eb.Publish(string(common.EventBroadcast), common.ConfirmReq, blk)
	return nil
}

func (dps *DPoS) cleanOnlineReps() {
	var repAddresses []*types.Address
	now := time.Now().UTC().Unix()
	dps.onlineReps.Range(func(key, value interface{}) bool {
		addr := key.(types.Address)
		v := value.(int64)
		if v < now {
			dps.onlineReps.Delete(addr)
		} else {
			repAddresses = append(repAddresses, &addr)
		}
		return true
	})
	_ = dps.ledger.SetOnlineRepresentations(repAddresses)
}

func (dps *DPoS) sendAckIfResultIsOld(block *types.StateBlock, account types.Address, acc *types.Account) error {
	va, err := dps.voteGenerate(block, account, acc)
	if err != nil {
		return err
	}
	msgHash, err := dps.calculateAckHash(va)
	if err != nil {
		return err
	}
	if !dps.cache.Has(msgHash) {
		dps.logger.Infof("send confirm ack for hash %s,previous hash is %s", block.GetHash(), block.Parent())
		//dps.ns.Broadcast(p2p.ConfirmAck, va)
		dps.eb.Publish(string(common.EventBroadcast), common.ConfirmAck, va)
		err := dps.cache.Set(msgHash, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func (dps *DPoS) calculateAckHash(va *protos.ConfirmAckBlock) (types.Hash, error) {
	data, err := protos.ConfirmAckBlockToProto(va)
	if err != nil {
		return types.ZeroHash, err
	}
	version := dps.cfg.Version
	message := p2p.NewQlcMessage(data, byte(version), common.ConfirmAck)
	hash, err := types.HashBytes(message)
	if err != nil {
		return types.ZeroHash, err
	}
	return hash, nil
}
