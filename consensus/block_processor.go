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
	timer := time.NewTicker(findOnlineRepresentativesIntervalms)
	for {
		select {
		case <-bp.quitCh:
			bp.dp.logger.Info("Stopped process blocks.")
			return
		case block := <-bp.blocks:
			result, _ := bp.dp.ledger.Process(block)
			bp.processResult(result, block)
		case <-timer.C:
			bp.dp.logger.Info("begin Find Online Representatives.")
			bp.dp.findOnlineRepresentatives()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (bp *BlockProcessor) processResult(result ledger.ProcessResult, block types.Block) error {
	switch result {
	case ledger.Progress:
		bp.dp.logger.Infof("Block %s basic info is correct,begin add it to roots", block.GetHash())
		bp.dp.actrx.addToRoots(block)
		bp.queueUnchecked(block.GetHash())
		break
	case ledger.BadSignature:
		bp.dp.logger.Infof("Bad signature for: %s", block.GetHash())
		break
	case ledger.BadWork:
		bp.dp.logger.Infof("Bad work for: %s", block.GetHash())
		break
	case ledger.BalanceMismatch:
		bp.dp.logger.Infof("Balance mismatch for: %s", block.GetHash())
		break
	case ledger.Old:
		bp.dp.logger.Infof("Old for: %s", block.GetHash())
		break
	case ledger.UnReceivable:
		bp.dp.logger.Infof("UnReceivable for: %s", block.GetHash())
		break
	case ledger.Other:
		bp.dp.logger.Infof("Unknow process result for: %s", block.GetHash())
		break
	case ledger.Fork:
		bp.dp.logger.Infof("Fork for: %s", block.GetHash())
		bp.processFork(block)
		break
	case ledger.GapPrevious:
		bp.dp.logger.Infof("Gap previous for: %s", block.GetHash())
		err := bp.dp.ledger.AddUncheckedBlock(block.GetPrevious(), block, types.UncheckedKindPrevious, types.UnSynchronized)
		if err != nil {
			return err
		}
		break
	case ledger.GapSource:
		bp.dp.logger.Infof("Gap source for: %s", block.GetHash())
		err := bp.dp.ledger.AddUncheckedBlock(block.(*types.StateBlock).Link, block, types.UncheckedKindLink, types.UnSynchronized)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func (bp *BlockProcessor) processFork(block types.Block) {

}

func (bp *BlockProcessor) queueUnchecked(hash types.Hash) {
	blkLink, _, _ := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink)
	//if err != nil {
	//	bp.dp.logger.Infof("Get blkLink err [%s] for hash: %s", err,hash)
	//	return err
	//}
	if blkLink != nil {
		bp.dp.logger.Infof("Get blkLink for hash: [%s]", blkLink.GetHash())
		bp.blocks <- blkLink
		err := bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindLink)
		if err != nil {
			bp.dp.logger.Infof("Get err [%s] for hash: [%s] when delete UncheckedKindLink", err, blkLink.GetHash())
		}
	}
	blkPre, _, _ := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious)
	//if err != nil {
	//	bp.dp.logger.Infof("Get blkPre err [%s] for hash: %s", err,hash)
	//	return err
	//}
	if blkPre != nil {
		bp.blocks <- blkPre
		bp.dp.logger.Infof("Get blkPre for hash: %s", blkPre.GetHash())
		err := bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindPrevious)
		if err != nil {
			bp.dp.logger.Infof("Get err [%s] for hash: [%s] when delete UncheckedKindPrevious", err, blkPre.GetHash())

		}
	}
}

func (bp *BlockProcessor) Stop() {
	bp.quitCh <- true
}
