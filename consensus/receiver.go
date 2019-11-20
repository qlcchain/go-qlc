package consensus

import (
	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type Receiver struct {
	eb         event.EventBus
	subscriber *event.ActorSubscriber
	c          *Consensus
}

func NewReceiver(eb event.EventBus) *Receiver {
	r := &Receiver{
		eb: eb,
	}
	return r
}

func (r *Receiver) init(c *Consensus) {
	r.c = c
}

func (r *Receiver) start() error {
	r.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *topic.EventPublishMsg:
			r.ReceivePublish(msg.Block, msg.From)
		case *topic.EventConfirmReqMsg:
			r.ReceiveConfirmReq(msg.Blocks, msg.From)
		case *p2p.EventConfirmAckMsg:
			r.ReceiveConfirmAck(msg.Block, msg.From)
		case types.StateBlockList:
			r.ReceiveSyncBlock(msg)
		case *types.StateBlock:
			r.ReceiveGenerateBlock(msg)
		}
	}), r.eb)

	if err := r.subscriber.SubscribeSync(topic.EventPublish, topic.EventConfirmReq, topic.EventConfirmAck, topic.EventSyncBlock,
		topic.EventGenerateBlock); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) stop() error {
	//r.cleanCacheStop()
	return r.subscriber.UnsubscribeAll()
}

func (r *Receiver) ReceivePublish(blk *types.StateBlock, msgFrom string) {
	r.c.logger.Debugf("receive publish block [%s] from [%s]", blk.GetHash(), msgFrom)

	bs := &BlockSource{
		Block:     blk,
		BlockFrom: types.UnSynchronized,
		Type:      MsgPublishReq,
	}
	r.c.ca.ProcessMsg(bs)
}

func (r *Receiver) ReceiveConfirmReq(blk []*types.StateBlock, msgFrom string) {
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

func (r *Receiver) ReceiveConfirmAck(ack *protos.ConfirmAckBlock, msgFrom string) {
	r.c.logger.Debugf("receive ConfirmAck for %d blocks [%s] from [%s]", len(ack.Hash), ack.Hash, msgFrom)

	valid := IsAckSignValidate(ack)
	if !valid {
		r.c.logger.Error("ack sign err")
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
		bs := &BlockSource{
			Block:     blk,
			BlockFrom: types.Synchronized,
			Type:      MsgSync,
		}
		r.c.ca.ProcessMsg(bs)
	}
}

func (r *Receiver) ReceiveGenerateBlock(blk *types.StateBlock) {
	r.c.logger.Infof("GenerateBlock Event for block:[%s]", blk.GetHash())
	bs := &BlockSource{
		Block:     blk,
		BlockFrom: types.UnSynchronized,
		Type:      MsgGenerateBlock,
	}
	r.c.ca.ProcessMsg(bs)
}
