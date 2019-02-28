package consensus

import (
	"errors"
	"time"

	"github.com/qlcchain/go-qlc/p2p"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
)

type blockSource struct {
	block     types.Block
	blockFrom types.SynchronizedKind
}

type BlockProcessor struct {
	blocks chan blockSource
	quitCh chan bool
	dp     *DposService
}

func NewBlockProcessor() *BlockProcessor {
	return &BlockProcessor{
		blocks: make(chan blockSource, 16384),
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
	timer := time.NewTicker(findOnlineRepresentativesIntervals)
	for {
		select {
		case <-bp.quitCh:
			bp.dp.logger.Info("Stopped process blocks.")
			return
		case bs := <-bp.blocks:
			result, _ := bp.dp.ledger.Process(bs.block)
			bp.processResult(result, bs)
		case <-timer.C:
			bp.dp.logger.Info("begin Find Online Representatives.")
			go bp.dp.findOnlineRepresentatives()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (bp *BlockProcessor) processResult(result ledger.ProcessResult, bs blockSource) error {
	blk := bs.block
	hash := blk.GetHash()
	switch result {
	case ledger.Progress:
		if bs.blockFrom == types.Synchronized {
			bp.dp.logger.Debugf("Block %s from sync,no need consensus", hash)
		} else if bs.blockFrom == types.UnSynchronized {
			bp.dp.logger.Debugf("Block %s basic info is correct,begin add it to roots", hash)
			bp.dp.acTrx.addToRoots(blk)
		} else {
			bp.dp.logger.Errorf("Block %s UnKnow from", hash)
			return errors.New("UnKnow block from")
		}
		bp.queueUnchecked(hash)
		break
	case ledger.BadSignature:
		bp.dp.logger.Errorf("Bad signature for: %s", hash)
		break
	case ledger.BadWork:
		bp.dp.logger.Errorf("Bad work for: %s", hash)
		break
	case ledger.BalanceMismatch:
		bp.dp.logger.Errorf("Balance mismatch for: %s", hash)
		break
	case ledger.Old:
		bp.dp.logger.Infof("Old for: %s", hash)
		break
	case ledger.UnReceivable:
		bp.dp.logger.Errorf("UnReceivable for: %s", hash)
		break
	case ledger.Other:
		bp.dp.logger.Errorf("UnKnow process result for: %s", hash)
		break
	case ledger.Fork:
		bp.dp.logger.Errorf("Fork for: %s", hash)
		bp.processFork(blk)
		break
	case ledger.GapPrevious:
		bp.dp.logger.Debugf("Gap previous for: %s", hash)
		err := bp.dp.ledger.AddUncheckedBlock(blk.GetPrevious(), blk, types.UncheckedKindPrevious, bs.blockFrom)
		if err != nil {
			return err
		}
		break
	case ledger.GapSource:
		bp.dp.logger.Debugf("Gap source for: %s", hash)
		err := bp.dp.ledger.AddUncheckedBlock(blk.(*types.StateBlock).Link, blk, types.UncheckedKindLink, bs.blockFrom)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func (bp *BlockProcessor) processFork(block types.Block) {
	blk := bp.findAnotherForkedBlock(block)
	if _, ok := bp.dp.acTrx.roots.Load(blk.Root()); !ok {
		bp.dp.acTrx.addToRoots(blk)
		bp.dp.ns.Broadcast(p2p.ConfirmReq, blk)
	}
	//count := 0
	//bp.dp.priInfos.Range(func(key, value interface{}) bool {
	//	count++
	//	isRep := bp.dp.isThisAccountRepresentation(key.(types.Address))
	//	if isRep {
	//		bp.dp.putRepresentativesToOnline(key.(types.Address))
	//		blk = bp.findAnotherForkedBlock(block)
	//
	//	} else {
	//		blk = block
	//	}
	//	if _, ok := bp.dp.acTrx.roots.Load(blk.Root()); !ok {
	//		bp.dp.acTrx.addToRoots(blk)
	//		bp.dp.ns.Broadcast(p2p.ConfirmReq, blk)
	//	}
	//
	//	return true
	//})
	//if count == 0 {
	//	if _, ok := bp.dp.acTrx.roots.Load(blk.Root()); !ok {
	//		bp.dp.acTrx.addToRoots(block)
	//		bp.dp.ns.Broadcast(p2p.ConfirmReq, blk)
	//	}
	//}
}

func (bp *BlockProcessor) findAnotherForkedBlock(block types.Block) types.Block {
	hash := block.Root()
	forkedHash, err := bp.dp.ledger.GetPosterior(hash)
	if err != nil {
		bp.dp.logger.Error(err)
		return block
	}
	forkedBlock, err := bp.dp.ledger.GetStateBlock(forkedHash)
	if err != nil {
		bp.dp.logger.Error(err)
		return block
	}
	return forkedBlock
}

func (bp *BlockProcessor) queueUnchecked(hash types.Hash) {
	blkLink, bf, _ := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink)
	if blkLink != nil {
		//bp.dp.logger.Debugf("Get blkLink for hash: [%s]", blkLink.GetHash())
		bs := blockSource{
			block:     blkLink,
			blockFrom: bf,
		}
		bp.blocks <- bs
		err := bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindLink)
		if err != nil {
			bp.dp.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindLink", err, blkLink.GetHash())
		}
	}

	blkPre, bf, _ := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious)
	if blkPre != nil {
		//bp.dp.logger.Infof("Get blkPre for hash: %s", blkPre.GetHash())
		bs := blockSource{
			block:     blkPre,
			blockFrom: bf,
		}
		bp.blocks <- bs
		err := bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindPrevious)
		if err != nil {
			bp.dp.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindPrevious", err, blkPre.GetHash())

		}
	}
}

func (bp *BlockProcessor) Stop() {
	bp.quitCh <- true
}
