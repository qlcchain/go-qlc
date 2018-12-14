package consensus

import (
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p/protos"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
)

var logger = log.NewLogger("consensus")

type Dpos struct {
	ns       p2p.Service
	ledger   *ledger.Ledger
	eventMsg map[p2p.EventType]p2p.EventSubscriber
	quitCh   chan bool
	bp       *BlockProcessor
}

func NewDpos(netService p2p.Service, ledger *ledger.Ledger) *Dpos {
	bp := NewBlockProcessor()
	dp := &Dpos{
		ns:       netService,
		ledger:   ledger,
		eventMsg: make(map[p2p.EventType]p2p.EventSubscriber),
		quitCh:   make(chan bool, 1),
		bp:       bp,
	}
	dp.bp.SetDpos(dp)
	return dp
}
func (dp *Dpos) Start() {
	event1 := dp.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventPublish, dp.ReceivePublish)
	event2 := dp.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmReq, dp.ReceiveConfirmReq)
	event3 := dp.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmAck, dp.ReceiveConfirmAck)
	event4 := dp.ns.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventSyncBlock, dp.ReceiveSyncBlock)
	dp.eventMsg[p2p.EventPublish] = event1
	dp.eventMsg[p2p.EventConfirmReq] = event2
	dp.eventMsg[p2p.EventConfirmAck] = event3
	dp.eventMsg[p2p.EventSyncBlock] = event4
	dp.start()
}
func (dp *Dpos) start() {
	dp.bp.Start()
}
func (dp *Dpos) ReceivePublish(v interface{}) {
	logger.Info("Receive Publish block")
	dp.bp.blocks <- v.(types.Block)

}
func (dp *Dpos) ReceiveConfirmReq(v interface{}) {
	logger.Info("Receive ConfirmReq")
	dp.bp.blocks <- v.(types.Block)
}
func (dp *Dpos) ReceiveConfirmAck(v interface{}) {
	logger.Info("Receive ConfirmAck")
	vote := v.(protos.ConfirmAckBlock)
	dp.bp.blocks <- vote.Blk
}
func (dp *Dpos) ReceiveSyncBlock(v interface{}) {
	logger.Info("Receive Sync Block")
	dp.bp.blocks <- v.(types.Block)
}
func (dp *Dpos) Stop() {
	dp.quitCh <- true
	for i, j := range dp.eventMsg {
		dp.ns.MessageEvent().GetEvent("consensus").UnSubscribe(i, j)
	}
}
