package consensus

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
	"time"
)

const (
	blockChanSize = 1024
	maxOrphanBlocks = 1000
)

type PovBlockResult struct {
	err error
}

type PovBlockSource struct {
	block *types.PovBlock
	from  types.PovBlockFrom
	replyCh chan PovBlockResult
}

type PovOrphanBlock struct {
	blockSrc *PovBlockSource
	expiration time.Time
}

type PovBlockProcessor struct {
	povEngine *PoVEngine

	orphanBlocks  map[types.Hash]*PovOrphanBlock
	parentOrphans map[types.Hash][]*PovOrphanBlock
	oldestOrphan  *PovOrphanBlock

	blockCh chan *PovBlockSource
	quitCh  chan struct{}
}

func NewPovBlockProcessor(povEngine *PoVEngine) *PovBlockProcessor {
	bp := &PovBlockProcessor{
		povEngine: povEngine,
	}

	bp.orphanBlocks = make(map[types.Hash]*PovOrphanBlock)
	bp.parentOrphans = make(map[types.Hash][]*PovOrphanBlock)

	bp.blockCh = make(chan *PovBlockSource, blockChanSize)
	bp.quitCh = make(chan struct{})

	return bp
}

func (bp *PovBlockProcessor) Start() error {
	common.Go(bp.loop)
	return nil
}

func (bp *PovBlockProcessor) Init() error {
	return nil
}

func (bp *PovBlockProcessor) Stop() error {
	close(bp.quitCh)
	return nil
}

func (bp *PovBlockProcessor) AddBlock(block *types.PovBlock, from types.PovBlockFrom) error {
	bp.blockCh <- &PovBlockSource{block: block, from: from}
	return nil
}

func (bp *PovBlockProcessor) AddMinedBlock(block *types.PovBlock) error {
	replyCh := make(chan PovBlockResult)
	bp.blockCh <- &PovBlockSource{block: block, from: types.PovBlockFromLocal, replyCh: replyCh}
	result := <- replyCh
	close(replyCh)
	return result.err
}

func (bp *PovBlockProcessor) loop() {
	for {
		select {
		case block := <-bp.blockCh:
			bp.processBlock(block)
		case <-bp.quitCh:
			bp.povEngine.GetLogger().Info("Exiting process blocks")
			return
		}
	}
}

func (bp *PovBlockProcessor) processBlock(blockSrc *PovBlockSource) error {
	block := blockSrc.block
	blockHash := blockSrc.block.GetHash()
	bp.povEngine.GetLogger().Debugf("process block, %d/%s", blockSrc.block.GetHeight(), blockHash)

	chain := bp.povEngine.GetChain()

	// duplicate block
	if bp.orphanBlocks[blockHash] != nil {
		bp.povEngine.GetLogger().Debugf("duplicate block %s exist in orphans", blockHash)
		return nil
	}
	if chain.HasBlock(blockHash, block.GetHeight()) {
		bp.povEngine.GetLogger().Debugf("duplicate block %s exist in chain", blockHash)
		return nil
	}

	prevBlock := chain.GetBlockByHash(block.GetPrevious())
	if prevBlock == nil {
		bp.addOrphanBlock(blockSrc)
		return nil
	}

	// check block
	result, err := bp.povEngine.GetVerifier().BlockCheck(block)
	if err != nil {
		bp.povEngine.GetLogger().Errorf("failed to verify block %s, err %s", block.GetHash(), err)
		return err
	}

	// orphan block
	if result == process.GapPrevious {
		bp.addOrphanBlock(blockSrc)
		return nil
	}

	err = bp.povEngine.GetChain().InsertBlock(block)

	if blockSrc.replyCh != nil {
		blockSrc.replyCh <- PovBlockResult{err: err}
	}

	if err != nil {
		bp.processOrphanBlock(blockSrc)
	}

	return err
}

func (bp *PovBlockProcessor) addOrphanBlock(blockSrc *PovBlockSource) {
	blockHash := blockSrc.block.GetHash()

	for _, oBlock := range bp.orphanBlocks {
		if time.Now().After(oBlock.expiration) {
			bp.removeOrphanBlock(oBlock)
			continue
		}

		if bp.oldestOrphan == nil || oBlock.expiration.Before(bp.oldestOrphan.expiration) {
			bp.oldestOrphan = oBlock
		}
	}

	if len(bp.orphanBlocks)+1 > maxOrphanBlocks {
		bp.removeOrphanBlock(bp.oldestOrphan)
		bp.oldestOrphan = nil
	}

	expiration := time.Now().Add(time.Hour)
	oBlock := &PovOrphanBlock{
		blockSrc:      blockSrc,
		expiration: expiration,
	}
	bp.orphanBlocks[blockHash] = oBlock

	prevHash := blockSrc.block.GetPrevious()
	bp.parentOrphans[prevHash] = append(bp.parentOrphans[prevHash], oBlock)

	bp.povEngine.GetLogger().Debugf("add orphan block %s prev %s", blockHash, prevHash)
}

func (bp *PovBlockProcessor) removeOrphanBlock(orphanBlock *PovOrphanBlock) {
	orphanHash := orphanBlock.blockSrc.block.GetHash()
	delete(bp.orphanBlocks, orphanHash)

	prevHash := orphanBlock.blockSrc.block.GetPrevious()
	orphans := bp.parentOrphans[prevHash]
	for i:=0; i<len(orphans); i++ {
		orphans := bp.parentOrphans[prevHash]
		for i := 0; i < len(orphans); i++ {
			hash := orphans[i].blockSrc.block.GetHash()
			if hash == orphanHash {
				copy(orphans[i:], orphans[i+1:])
				orphans[len(orphans)-1] = nil
				orphans = orphans[:len(orphans)-1]
				i--
			}
		}
	}

	bp.parentOrphans[prevHash] = orphans

	if len(bp.parentOrphans[prevHash]) == 0 {
		delete(bp.parentOrphans, prevHash)
	}
}

func (bp *PovBlockProcessor) processOrphanBlock(blockSrc *PovBlockSource) error {
	blockHash := blockSrc.block.GetHash()
	orphans, ok := bp.parentOrphans[blockHash]
	if !ok {
		return nil
	}
	if len(orphans) <= 0 {
		delete(bp.parentOrphans, blockHash)
		return nil
	}

	processHashes := make([]*types.Hash, 0, 10)
	processHashes = append(processHashes, &blockHash)
	for len(processHashes) > 0 {
		processHash := processHashes[0]
		processHashes[0] = nil
		processHashes = processHashes[1:]

		orphans := bp.parentOrphans[*processHash]

		bp.povEngine.GetLogger().Debugf("parent %s has %d orphan blocks", processHash, len(orphans))

		for i := 0; i < len(orphans); i++ {
			orphan := orphans[i]
			if orphan == nil {
				continue
			}

			orphanHash := orphan.blockSrc.block.GetHash()
			bp.removeOrphanBlock(orphan)
			i--

			for _, orphanBlock := range orphans {
				bp.blockCh <- orphanBlock.blockSrc
			}

			processHashes = append(processHashes, &orphanHash)
		}
	}

	return nil
}