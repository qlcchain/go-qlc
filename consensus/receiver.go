package consensus

import (
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type Receiver struct {
	eb           event.EventBus
	c            *Consensus
	lastSyncTime time.Time
}

func NewReceiver(eb event.EventBus) *Receiver {
	r := &Receiver{
		eb:           eb,
		lastSyncTime: time.Now(),
	}
	return r
}

func (r *Receiver) init(c *Consensus) {
	r.c = c
}

func (r *Receiver) start() error {
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

func (r *Receiver) ReceivePublish(blk *types.StateBlock, hash types.Hash, msgFrom string) {
	r.c.logger.Debugf("receive publish block [%s] from [%s]", blk.GetHash(), msgFrom)

	bs := &BlockSource{
		Block:     blk,
		BlockFrom: types.UnSynchronized,
		Type:      MsgPublishReq,
	}
	r.c.ca.ProcessMsg(bs)
	//	}
}

func (r *Receiver) ReceiveConfirmReq(blk []*types.StateBlock, hash types.Hash, msgFrom string) {

	for _, b := range blk {
		r.c.logger.Debugf("receive ConfirmReq block [%s] from [%s]", b.GetHash(), msgFrom)

		bs := &BlockSource{
			Block:     b,
			BlockFrom: types.UnSynchronized,
			Type:      MsgConfirmReq,
		}
		r.c.ca.ProcessMsg(bs)
	}
}

func (r *Receiver) ReceiveConfirmAck(ack *protos.ConfirmAckBlock, hash types.Hash, msgFrom string) {
	r.c.logger.Debugf("receive ConfirmAck for %d blocks [%s] from [%s]", len(ack.Hash), ack.Hash, msgFrom)

	valid := IsAckSignValidate(ack)
	if !valid {
		r.c.logger.Error("ack sign err")
		return
	}

	if _, err := r.c.ledger.GetRepresentation(ack.Account); err != nil {
		return
	}

	bs := &BlockSource{
		Type: MsgConfirmAck,
		Para: ack,
	}
	r.c.ca.ProcessMsg(bs)
}

func (r *Receiver) ReceiveSyncBlock(blocks types.StateBlockList) {
	for _, blk := range blocks {
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
