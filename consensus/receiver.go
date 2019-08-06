package consensus

import (
	"time"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

var (
	msgCacheExpirationTime = 2 * time.Minute
)

type Receiver struct {
	eb           event.EventBus
	c            *Consensus
	cache        gcache.Cache
	lastSyncTime time.Time
	quitchClean  chan bool
}

func NewReceiver(eb event.EventBus) *Receiver {
	r := &Receiver{
		eb:           eb,
		cache:        gcache.New(common.ConsensusMsgCacheSize).LRU().Expiration(msgCacheExpirationTime).Build(),
		lastSyncTime: time.Now(),
		quitchClean:  make(chan bool, 1),
	}

	return r
}

func (r *Receiver) init(c *Consensus) {
	r.c = c
}

func (r *Receiver) start() error {
	//go r.cleanCacheStart()

	err := r.eb.Subscribe(common.EventPublish, r.ReceivePublish)
	if err != nil {
		return err
	}

	err = r.eb.Subscribe(common.EventConfirmReq, r.ReceiveConfirmReq)
	if err != nil {
		return err
	}

	err = r.eb.Subscribe(common.EventConfirmAck, r.ReceiveConfirmAck)
	if err != nil {
		return err
	}

	err = r.eb.Subscribe(common.EventSyncBlock, r.ReceiveSyncBlock)
	if err != nil {
		return err
	}

	err = r.eb.Subscribe(common.EventGenerateBlock, r.ReceiveGenerateBlock)
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) stop() error {
	//r.cleanCacheStop()

	err := r.eb.Unsubscribe(common.EventPublish, r.ReceivePublish)
	if err != nil {
		return err
	}

	err = r.eb.Unsubscribe(common.EventConfirmReq, r.ReceiveConfirmReq)
	if err != nil {
		return err
	}

	err = r.eb.Unsubscribe(common.EventConfirmAck, r.ReceiveConfirmAck)
	if err != nil {
		return err
	}

	err = r.eb.Unsubscribe(common.EventSyncBlock, r.ReceiveSyncBlock)
	if err != nil {
		return err
	}

	err = r.eb.Unsubscribe(common.EventGenerateBlock, r.ReceiveGenerateBlock)
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) cleanCacheStart() {
	timerClean := time.NewTicker(msgCacheExpirationTime)

	for {
		select {
		case <-r.quitchClean:
			return
		case <-timerClean.C:
			r.cleanCache()
		}
	}
}

func (r *Receiver) cleanCacheStop() {
	r.quitchClean <- true
}

func (r *Receiver) cleanCache() {

}

func (r *Receiver) ReceivePublish(blk *types.StateBlock, hash types.Hash, msgFrom string) {
	r.c.logger.Debugf("receive publish block [%s] from [%s]", blk.GetHash(), msgFrom)
	if !r.processed(hash) {
		r.processedUpdate(hash)

		bs := &BlockSource{
			Block:     blk,
			BlockFrom: types.UnSynchronized,
			Type:      MsgPublishReq,
			MsgFrom:   msgFrom,
		}
		r.c.ca.ProcessMsg(bs)
	}
}

func (r *Receiver) ReceiveConfirmReq(blk *types.StateBlock, hash types.Hash, msgFrom string) {
	r.c.logger.Debugf("receive ConfirmReq block [%s] from [%s]", blk.GetHash(), msgFrom)
	if !r.processed(hash) {
		r.processedUpdate(hash)

		bs := &BlockSource{
			Block:     blk,
			BlockFrom: types.UnSynchronized,
			Type:      MsgConfirmReq,
			MsgFrom:   msgFrom,
		}
		r.c.ca.ProcessMsg(bs)
	}
}

func (r *Receiver) ReceiveConfirmAck(ack *protos.ConfirmAckBlock, hash types.Hash, msgFrom string) {
	r.c.logger.Debugf("receive ConfirmAck block [%s] from [%s]", ack.Hash, msgFrom)
	if !r.processed(hash) {
		r.processedUpdate(hash)

		valid := IsAckSignValidate(ack)
		if !valid {
			return
		}

		bs := &BlockSource{
			Type:    MsgConfirmAck,
			Para:    ack,
			MsgFrom: msgFrom,
		}
		r.c.ca.ProcessMsg(bs)
	}
}

func (r *Receiver) ReceiveSyncBlock(blk *types.StateBlock) {
	r.c.logger.Debugf("Sync Event for block:[%s]", blk.GetHash())
	now := time.Now()
	if now.Sub(r.lastSyncTime) > time.Second {
		r.eb.Publish(common.EventSyncing, now)
		r.lastSyncTime = now
	}

	bs := &BlockSource{
		Block:     blk,
		BlockFrom: types.Synchronized,
		Type:      MsgSync,
	}
	r.c.ca.ProcessMsg(bs)
}

func (r *Receiver) ReceiveGenerateBlock(result process.ProcessResult, blk *types.StateBlock) {
	r.c.logger.Infof("GenerateBlock Event for block:[%s]", blk.GetHash())
	bs := &BlockSource{
		Block:     blk,
		BlockFrom: types.UnSynchronized,
		Type:      MsgGenerateBlock,
	}
	r.c.ca.ProcessMsg(bs)
}

func (r *Receiver) processed(hash types.Hash) bool {
	return r.cache.Has(hash)
}

func (r *Receiver) processedUpdate(hash types.Hash) {
	err := r.cache.Set(hash, "")
	if err != nil {
		r.c.logger.Errorf("Set cache error [%s] for block [%s] with publish message", err, hash)
	}
}
