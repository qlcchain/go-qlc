package consensus

import (
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
)

type BlockProcessor struct {
	blocks chan types.Block
	quitCh chan bool
	dp     *DposService
}

func NewBlockProcessor() *BlockProcessor {
	return &BlockProcessor{
		blocks: make(chan types.Block, 16384),
		quitCh: make(chan bool, 1),
	}
}

func (bp *BlockProcessor) SetDpos(dp *DposService) {
	bp.dp = dp
}

func (bp *BlockProcessor) Start() {
	bp.processBlocks()
}

func (bp *BlockProcessor) processBlocks() {
	for {
		select {
		case <-bp.quitCh:
			logger.Info("Stopped process blocks.")
			return
		case block := <-bp.blocks:
			result, _ := bp.dp.ledger.Process(block)
			bp.processResult(result, block)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (bp *BlockProcessor) processResult(result ledger.ProcessResult, block types.Block) error {
	switch result {
	case ledger.Progress:
		logger.Infof("Block %s basic info is correct,begin add it to roots", block.GetHash())
		bp.dp.actrx.addToRoots(block)
		bp.queueUnchecked(block.GetHash())
		break
	case ledger.BadSignature:
		logger.Infof("Bad signature for: %s", block.GetHash())
		break
	case ledger.BadWork:
		logger.Infof("Bad work for: %s", block.GetHash())
		break
	case ledger.BalanceMismatch:
		logger.Infof("Balance mismatch for: %s", block.GetHash())
		break
	case ledger.Old:
		logger.Infof("Old for: %s", block.GetHash())
		break
	case ledger.UnReceivable:
		logger.Infof("UnReceivable for: %s", block.GetHash())
		break
	case ledger.Other:
		logger.Infof("Unknow process result for: %s", block.GetHash())
		break
	case ledger.Fork:
		logger.Infof("Fork for: %s", block.GetHash())
		bp.processFork(block)
		break
	case ledger.GapPrevious:
		logger.Infof("Gap previous for: %s", block.GetHash())
		err := bp.dp.ledger.AddUncheckedBlock(block.Root(), block, types.UncheckedKindPrevious)
		if err != nil {
			return err
		}
		break
	case ledger.GapSource:
		logger.Infof("Gap source for: %s", block.GetHash())
		err := bp.dp.ledger.AddUncheckedBlock(block.Root(), block, types.UncheckedKindLink)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func (bp *BlockProcessor) processFork(block types.Block) {

}

func (bp *BlockProcessor) queueUnchecked(hash types.Hash) error {
	blklink, err := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink)
	if err != nil {
		return err
	}
	if blklink != nil {
		bp.blocks <- blklink
		err = bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindLink)
		if err != nil {
			return err
		}
	}
	blkpre, err := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious)
	if err != nil {
		return err
	}
	if blkpre != nil {
		bp.blocks <- blkpre
		bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindPrevious)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bp *BlockProcessor) Stop() {
	bp.quitCh <- true
}
