package consensus

import (
	"github.com/qlcchain/go-qlc/common/types"
)

type process_result byte

const (
	progress process_result = iota
	bad_signature
	old
)

type BlockProcessor struct {
	blocks chan types.Block
	quitCh chan bool
	dp     *Dpos
}

func NewBlockProcessor() *BlockProcessor {
	return &BlockProcessor{
		blocks: make(chan types.Block, 16384),
		quitCh: make(chan bool, 1),
	}
}
func (bp *BlockProcessor) SetDpos(dp *Dpos) {
	bp.dp = dp
}
func (bp *BlockProcessor) Start() {
	go bp.process_blocks()
}
func (bp *BlockProcessor) process_blocks() {
	for {
		select {
		case <-bp.quitCh:
			logger.Info("Stopped process_blocks.")
			return
		case block := <-bp.blocks:
			bp.process_receive_one(block)
		}
	}
}
func (bp *BlockProcessor) process_receive_one(block types.Block) process_result {
	result := bp.process(block)

	return result
}
func (bp *BlockProcessor) process(block types.Block) process_result {
	return progress
}
func (bp *BlockProcessor) Stop() {
	bp.quitCh <- true
}
