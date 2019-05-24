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
	msgCacheSize                      = 7000 * 5 * 60
	msgCacheExpirationTime            = 8 * time.Minute
	findOnlineRepresentativesInterval = 2 * time.Minute
	repTimeout                        = 5 * time.Minute
	blockCacheSize                    = 7000 * 5 * 60
	blockCacheExpirationTime          = 8 * time.Minute
	voteCacheSize                     = 7000 * 5 * 60
	voteCacheTimeout                  = 5 * time.Minute
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
	voteCache  gcache.Cache //vote blocks
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

func NewDPoS(cfg *config.Config, accounts []*types.Account) (*DPoS, error) {
	bp := NewBlockProcessor()
	acTrx := NewActiveTrx()
	l := ledger.NewLedger(cfg.LedgerDir())

	dps := &DPoS{
		ledger:    l,
		verifier:  process.NewLedgerVerifier(l),
		eb:        event.GetEventBus(cfg.LedgerDir()),
		bp:        bp,
		acTrx:     acTrx,
		accounts:  accounts,
		logger:    log.NewLogger("consensus"),
		cache:     gcache.New(msgCacheSize).LRU().Expiration(msgCacheExpirationTime).Build(),
		voteCache: gcache.New(voteCacheSize).LRU().Expiration(voteCacheTimeout).Build(),
		cfg:       cfg,
	}
	dps.bp.SetDpos(dps)
	dps.acTrx.SetDposService(dps)
	return dps, nil
}

func (dps *DPoS) setEvent() error {
	err := dps.eb.Subscribe(string(common.EventPublish), dps.ReceivePublish)
	if err != nil {
		return err
	}
	err = dps.eb.Subscribe(string(common.EventConfirmReq), dps.ReceiveConfirmReq)
	if err != nil {
		return err
	}
	err = dps.eb.Subscribe(string(common.EventConfirmAck), dps.ReceiveConfirmAck)
	if err != nil {
		return err
	}
	err = dps.eb.Subscribe(string(common.EventSyncBlock), dps.ReceiveSyncBlock)
	if err != nil {
		return err
	}
	return nil
}

func (dps *DPoS) ReceivePublish(blk *types.StateBlock, hash types.Hash, msgFrom string) {
	dps.logger.Debugf("receive publish block [%s] from [%s]", blk.GetHash(), msgFrom)
	bs := blockSource{
		block:     blk,
		blockFrom: types.UnSynchronized,
	}
	dps.onReceivePublish(hash, bs, msgFrom)
}

func (dps *DPoS) onReceivePublish(hash types.Hash, bs blockSource, msgFrom string) {
	blkHash := bs.block.GetHash()
	if !dps.cache.Has(hash) {
		if !dps.bp.blockCache.Has(blkHash) {
			_ = dps.bp.blockCache.Set(blkHash, "")
			dps.bp.blocks <- bs
		}
		dps.eb.Publish(string(common.EventSendMsgToPeers), p2p.PublishReq, bs.block, msgFrom)
		err := dps.cache.Set(hash, "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with publish message", err, bs.block.GetHash())
		}
	}
}

func (dps *DPoS) ReceiveConfirmReq(blk *types.StateBlock, hash types.Hash, msgFrom string) {
	//dps.logger.Infof("receive ConfirmReq block [%s] from [%s]", blk.GetHash(), msgFrom)
	var address types.Address
	var count uint32
	bs := blockSource{
		block:     blk,
		blockFrom: types.UnSynchronized,
	}
	blkHash := bs.block.GetHash()
	if !dps.cache.Has(hash) {
		localRepAccount.Range(func(key, value interface{}) bool {
			count++
			address = key.(types.Address)
			dps.saveOnlineRep(address)
			result, _ := dps.verifier.BlockCheck(bs.block)
			if result == process.Old {
				err := dps.sendAckIfResultIsOld(bs.block, address, value.(*types.Account))
				if err != nil {
					return true
				}
			}
			if !dps.bp.blockCache.Has(blkHash) {
				if result == process.Progress {
					_ = dps.verifier.BlockProcess(bs.block)
				}
				_ = dps.bp.processResult(result, bs)
				_ = dps.bp.blockCache.Set(blkHash, "")
			}
			return true
		})
		if count == 0 {
			if !dps.bp.blockCache.Has(blkHash) {
				dps.bp.blocks <- bs
				_ = dps.bp.blockCache.Set(blkHash, "")
			}
			dps.eb.Publish(string(common.EventSendMsgToPeers), p2p.ConfirmReq, blk, msgFrom)
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
	blkHash := bs.block.GetHash()
	valid := IsAckSignValidate(ack)
	if !valid {
		return
	}

	dps.acTrx.vote(ack)
	dps.saveOnlineRep(ack.Account)
	if !dps.cache.Has(hash) {
		result, _ := dps.verifier.BlockCheck(bs.block)
		//cache the ack messages
		if result == process.GapPrevious || result == process.GapSource {
			if dps.voteCache.Has(hash) {
				v, err := dps.voteCache.Get(hash)
				if err != nil {
					dps.logger.Error("get vote cache err")
				}
				vc := v.(*sync.Map)
				vc.Store(ack.Account, ack)
			} else {
				vc := new(sync.Map)
				vc.Store(ack.Account, ack)
				err := dps.voteCache.Set(hash, vc)
				if err != nil {
					dps.logger.Error("set vote cache err")
				}
			}
		}
		localRepAccount.Range(func(key, value interface{}) bool {
			count++
			address = key.(types.Address)
			dps.saveOnlineRep(address)
			if result == process.Old {
				err := dps.sendAckIfResultIsOld(bs.block, address, value.(*types.Account))
				if err != nil {
					dps.logger.Error(err)
				}
			}
			if !dps.bp.blockCache.Has(blkHash) {
				if result == process.Progress {
					_ = dps.verifier.BlockProcess(bs.block)
					_ = dps.bp.processResult(result, bs)
					dps.acTrx.vote(ack)
				}
				_ = dps.bp.blockCache.Set(blkHash, "")
			}
			return true
		})
		if count == 0 {
			if !dps.bp.blockCache.Has(blkHash) {
				dps.bp.blocks <- bs
				_ = dps.bp.blockCache.Set(blkHash, "")
			}
		}
		dps.eb.Publish(string(common.EventSendMsgToPeers), p2p.ConfirmAck, ack, msgFrom)
		err := dps.cache.Set(hash, "")
		if err != nil {
			dps.logger.Errorf("Set cache error [%s] for block [%s] with confirmAck message", err, ack.Blk.GetHash())
		}
	}
}

func (dps *DPoS) ReceiveSyncBlock(blk *types.StateBlock) {
	//	dps.logger.Info("Sync Event")
	bs := blockSource{
		block:     blk,
		blockFrom: types.Synchronized,
	}
	dps.logger.Debugf("Sync Event for block:[%s]", bs.block.GetHash())
	hash := bs.block.GetHash()
	if !dps.bp.blockCache.Has(hash) {
		dps.bp.blocks <- bs
		_ = dps.bp.blockCache.Set(hash, "")
	}
}

func (dps *DPoS) sendConfirmAck(block *types.StateBlock, account types.Address, acc *types.Account) error {
	va, err := dps.voteGenerate(block, account, acc)
	if err != nil {
		dps.logger.Error("vote generate error")
		return err
	}
	//dps.ns.Broadcast(p2p.ConfirmAck, va)
	dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, va)
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
		blk, err := dps.ledger.GetRandomStateBlock()
		if err != nil {
			return true
		}
		va, err := dps.voteGenerate(blk, address, value.(*types.Account))
		if err != nil {
			return true
		}
		dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, va)
		return true
	})
	//dps.ns.Broadcast(p2p.ConfirmReq, blk)

	//dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, blk)
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
		dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, va)
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
	message := p2p.NewQlcMessage(data, byte(version), p2p.ConfirmAck)
	hash, err := types.HashBytes(message)
	if err != nil {
		return types.ZeroHash, err
	}
	return hash, nil
}
