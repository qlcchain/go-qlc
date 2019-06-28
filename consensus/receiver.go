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

const (
	msgCacheSize           = 1000 * 2 * 60
	msgCacheExpirationTime = 2 * time.Minute
)

type Receiver struct {
	eb    event.EventBus
	c     *Consensus
	cache gcache.Cache
}

func NewReceiver(eb event.EventBus) *Receiver {
	r := &Receiver{
		eb:    eb,
		cache: gcache.New(msgCacheSize).LRU().Expiration(msgCacheExpirationTime).Build(),
	}

	return r
}

func (r *Receiver) init(c *Consensus) {
	r.c = c
}

func (r *Receiver) start() error {
	err := r.eb.Subscribe(string(common.EventPublish), r.ReceivePublish)
	if err != nil {
		return err
	}

	//err = r.eb.Subscribe(string(common.EventConfirmReq), r.ReceiveConfirmReq)
	//if err != nil {
	//	return err
	//}

	err = r.eb.Subscribe(string(common.EventConfirmAck), r.ReceiveConfirmAck)
	if err != nil {
		return err
	}

	err = r.eb.Subscribe(string(common.EventSyncBlock), r.ReceiveSyncBlock)
	if err != nil {
		return err
	}

	err = r.eb.Subscribe(string(common.EventGenerateBlock), r.ReceiveGenerateBlock)
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) stop() error {
	err := r.eb.Unsubscribe(string(common.EventPublish), r.ReceivePublish)
	if err != nil {
		return err
	}

	//err = r.eb.Unsubscribe(string(common.EventConfirmReq), r.ReceiveConfirmReq)
	//if err != nil {
	//	return err
	//}

	err = r.eb.Unsubscribe(string(common.EventConfirmAck), r.ReceiveConfirmAck)
	if err != nil {
		return err
	}

	err = r.eb.Unsubscribe(string(common.EventSyncBlock), r.ReceiveSyncBlock)
	if err != nil {
		return err
	}

	err = r.eb.Unsubscribe(string(common.EventGenerateBlock), r.ReceiveGenerateBlock)
	if err != nil {
		return err
	}

	return nil
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
	r.c.logger.Debugf("receive ConfirmAck block [%s] from [%s]", ack.Blk.GetHash(), msgFrom)
	if !r.processed(hash) {
		r.processedUpdate(hash)

		valid := IsAckSignValidate(ack)
		if !valid {
			return
		}

		bs := &BlockSource{
			Block:     ack.Blk,
			BlockFrom: types.UnSynchronized,
			Type:      MsgConfirmAck,
			Para:      ack,
			MsgFrom:   msgFrom,
		}
		r.c.ca.ProcessMsg(bs)
	}
}

func (r *Receiver) ReceiveSyncBlock(blk *types.StateBlock) {
	r.c.logger.Debugf("Sync Event for block:[%s]", blk.GetHash())
	bs := &BlockSource{
		Block:     blk,
		BlockFrom: types.Synchronized,
		Type:      MsgSync,
	}
	r.c.ca.ProcessMsg(bs)
}

func (r *Receiver) ReceiveGenerateBlock(result process.ProcessResult, blk *types.StateBlock) {
	r.c.logger.Debugf("GenerateBlock Event for block:[%s]", blk.GetHash())
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
