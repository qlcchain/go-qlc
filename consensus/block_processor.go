package consensus

import (
	"errors"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/p2p"

	"github.com/qlcchain/go-qlc/p2p/protos"

	"github.com/bluele/gcache"

	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
)

var (
	errFoundBlock = errors.New("state block not found")
)

type blockSource struct {
	block     *types.StateBlock
	blockFrom types.SynchronizedKind
}

type BlockProcessor struct {
	blocks              chan blockSource
	blocksFromUnchecked chan blockSource
	quitCh              chan bool
	dp                  *DPoS
	blockCache          gcache.Cache
}

type cacheInfo struct {
	b             blockSource
	uncheckedKind types.UncheckedKind
	time          int64
	votes         []*protos.ConfirmAckBlock
}

func NewBlockProcessor() *BlockProcessor {
	return &BlockProcessor{
		blocks:              make(chan blockSource, blockCacheSize),
		blocksFromUnchecked: make(chan blockSource, blockCacheSize),
		quitCh:              make(chan bool, 1),
		blockCache:          gcache.New(blockCacheSize).LRU().Expiration(blockCacheExpirationTime).Build(),
	}
}

func (bp *BlockProcessor) SetDpos(dp *DPoS) {
	bp.dp = dp
}

func (bp *BlockProcessor) Start() {
	bp.processBlocks()
}

func (bp *BlockProcessor) processBlocks() {
	timer := time.NewTicker(findOnlineRepresentativesInterval)
	for {
		select {
		case <-bp.quitCh:
			bp.dp.logger.Info("Stopped process blocks.")
			return
		case bs := <-bp.blocks:
			result, err := bp.dp.verifier.Process(bs.block)
			if err != nil {
				bp.dp.logger.Errorf("error: [%s] when verify block:[%s]", err, bs.block.GetHash())
			} else {
				err = bp.processResult(result, bs)
				if err != nil {
					bp.dp.logger.Error(err)
				}
			}
		case bs := <-bp.blocksFromUnchecked:
			result, err := bp.dp.verifier.Process(bs.block)
			if err != nil {
				bp.dp.logger.Errorf("error: [%s] when verify block:[%s]", err, bs.block.GetHash())
			} else {
				err = bp.processResult(result, bs)
				if err != nil {
					bp.dp.logger.Error(err)
				}
			}
		case <-timer.C:
			bp.dp.logger.Info("begin Find Online Representatives.")
			go func() {
				err := bp.dp.findOnlineRepresentatives()
				if err != nil && err.Error() != errFoundBlock.Error() {
					bp.dp.logger.Error(err)
				}
				bp.dp.cleanOnlineReps()
			}()
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (bp *BlockProcessor) processResult(result process.ProcessResult, bs blockSource) error {
	blk := bs.block
	hash := blk.GetHash()
	switch result {
	case process.Progress:
		if bs.blockFrom == types.Synchronized {
			bp.dp.logger.Debugf("Block %s from sync,no need consensus", hash)
			now := time.Now()
			bp.dp.eb.Publish(string(common.EventSyncing), now)
			bp.dp.eb.Publish(string(common.EventConfirmedBlock), bs.block)
		} else if bs.blockFrom == types.UnSynchronized {
			bp.dp.logger.Infof("Block %s basic info is correct,begin add it to roots", hash)
			bp.dp.acTrx.addToRoots(blk)
			bp.checkVoteCache(blk)
		} else {
			bp.dp.logger.Errorf("Block %s UnKnow from", hash)
			return errors.New("UnKnow block from")
		}
		bp.queueUnchecked(hash)
	case process.BadSignature:
		bp.dp.logger.Errorf("Bad signature for block: %s", hash)
	case process.BadWork:
		bp.dp.logger.Errorf("Bad work for block: %s", hash)
	case process.BalanceMismatch:
		bp.dp.logger.Errorf("Balance mismatch for block: %s", hash)
	case process.Old:
		bp.dp.logger.Debugf("Old for block: %s", hash)
	case process.UnReceivable:
		bp.dp.logger.Errorf("UnReceivable for block: %s", hash)
	case process.GapSmartContract:
		bp.dp.logger.Errorf("GapSmartContract for block: %s", hash)
		bp.processGapSmartContract(blk)
	case process.InvalidData:
		bp.dp.logger.Errorf("InvalidData for block: %s", hash)
	case process.Other:
		bp.dp.logger.Errorf("UnKnow process result for: %s", hash)
	case process.Fork:
		bp.dp.logger.Errorf("Fork for block: %s", hash)
		bp.processFork(blk)
	case process.GapPrevious:
		bp.dp.logger.Debugf("Gap previous for block: %s", hash)
		err := bp.dp.ledger.AddUncheckedBlock(blk.GetPrevious(), blk, types.UncheckedKindPrevious, bs.blockFrom)
		if err != nil {
			return err
		}
	case process.GapSource:
		bp.dp.logger.Debugf("Gap source for block: %s", hash)
		err := bp.dp.ledger.AddUncheckedBlock(blk.Link, blk, types.UncheckedKindLink, bs.blockFrom)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bp *BlockProcessor) processGapSmartContract(block *types.StateBlock) {

}

func (bp *BlockProcessor) checkVoteCache(block *types.StateBlock) {
	v, e := bp.dp.voteCache.Get(block.GetHash())
	if e == nil {
		vc := v.(*sync.Map)
		vc.Range(func(key, value interface{}) bool {
			bp.dp.acTrx.vote(value.(*protos.ConfirmAckBlock))
			return true
		})

		bp.dp.voteCache.Remove(block.GetHash())
	}
}

func (bp *BlockProcessor) processFork(block *types.StateBlock) {
	blk := bp.findAnotherForkedBlock(block)
	vk := getVoteKey(blk)
	if _, ok := bp.dp.acTrx.roots.Load(vk); !ok {
		bp.dp.acTrx.addToRoots(blk)
		bp.dp.eb.Publish(string(common.EventBroadcast), p2p.ConfirmReq, blk)
	}
}

func (bp *BlockProcessor) findAnotherForkedBlock(block *types.StateBlock) *types.StateBlock {
	hash := block.Parent()
	forkedHash, err := bp.dp.ledger.GetChild(hash, block.Address)
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
		bp.dp.logger.Debugf("Get blkLink for hash: [%s]", blkLink.GetHash())
		bs := blockSource{
			block:     blkLink,
			blockFrom: bf,
		}
		bp.blocksFromUnchecked <- bs
		err := bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindLink)
		if err != nil {
			bp.dp.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindLink", err, blkLink.GetHash())
		}
	}

	blkPre, bf, _ := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious)
	if blkPre != nil {
		bp.dp.logger.Debugf("Get blkPre for hash: %s", blkPre.GetHash())
		bs := blockSource{
			block:     blkPre,
			blockFrom: bf,
		}
		bp.blocksFromUnchecked <- bs
		err := bp.dp.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindPrevious)
		if err != nil {
			bp.dp.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindPrevious", err, blkPre.GetHash())

		}
	}
}

func (bp *BlockProcessor) queueUncheckedFromLedger(hash types.Hash) {
	blkLink, bf, _ := bp.dp.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink)
	if blkLink != nil {
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
