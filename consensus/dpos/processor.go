package dpos

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

type syncCacheInfo struct {
	kind byte
	hash types.Hash
}

type chainKey struct {
	addr  types.Address
	token types.Hash
}

type chainOrderKey struct {
	chainKey
	order uint64
}

type Processor struct {
	index           int
	dps             *DPoS
	uncheckedCache  gcache.Cache //gap blocks
	quitCh          chan bool
	blocks          chan *consensus.BlockSource
	blocksAcked     chan types.Hash
	syncBlock       chan *types.StateBlock
	syncBlockAcked  chan types.Hash
	acks            chan *voteInfo
	frontiers       chan *types.StateBlock
	syncStateChange chan common.SyncState
	syncState       common.SyncState
	syncCache       chan *syncCacheInfo
	orderedChain    map[chainOrderKey]types.Hash
	chainHeight     map[chainKey]uint64
}

func newProcessors(num int) []*Processor {
	processors := make([]*Processor, 0)

	for i := 0; i < num; i++ {
		p := &Processor{
			index:           i,
			quitCh:          make(chan bool, 1),
			blocks:          make(chan *consensus.BlockSource, common.DPoSMaxBlocks),
			blocksAcked:     make(chan types.Hash, common.DPoSMaxBlocks),
			syncBlock:       make(chan *types.StateBlock, common.DPoSMaxBlocks),
			syncBlockAcked:  make(chan types.Hash, common.DPoSMaxBlocks),
			acks:            make(chan *voteInfo, common.DPoSMaxBlocks),
			frontiers:       make(chan *types.StateBlock, common.DPoSMaxBlocks),
			syncCache:       make(chan *syncCacheInfo, common.DPoSMaxBlocks),
			syncStateChange: make(chan common.SyncState, 1),
			syncState:       common.SyncNotStart,
			orderedChain:    make(map[chainOrderKey]types.Hash),
			chainHeight:     make(map[chainKey]uint64),
		}
		processors = append(processors, p)
	}

	return processors
}

func (p *Processor) setDposService(dps *DPoS) {
	p.dps = dps
}

func (p *Processor) start() {
	go p.processMsg()
}

func (p *Processor) stop() {
	p.quitCh <- true
}

func (p *Processor) syncBlockCheck(block *types.StateBlock) {
	dps := p.dps
	hash := block.GetHash()
	checked := false

	if block.IsOpen() {
		checked = true
	} else {
		if has, _ := dps.ledger.HasUnconfirmedSyncBlock(block.Previous); has {
			checked = true
		} else {
			if has, _ := dps.ledger.HasStateBlockConfirmed(block.Previous); has {
				checked = true
			} else {
				p.enqueueUncheckedSync(block)
			}
		}
	}

	if checked {
		if err := dps.ledger.AddUnconfirmedSyncBlock(hash, block); err != nil {
			dps.logger.Errorf("add unconfirmed sync block err", err)
		}
		p.processConfirmedSync(hash, block)
	}
}

func (p *Processor) processMsg() {
	getTimeout := time.NewTicker(10 * time.Millisecond)

	for {
	PriorityOut:
		for {
			select {
			case p.syncState = <-p.syncStateChange:
				p.dps.syncStateNotifyWait.Done()
			case hash := <-p.syncBlockAcked:
				if p.dps.isConfirmedFrontier(hash) {
					p.dps.frontiersStatus.Store(hash, frontierChainConfirmed)
					p.confirmChain(hash)
				}
				p.dequeueUncheckedSync(hash)
			case hash := <-p.blocksAcked:
				p.dequeueUnchecked(hash)
			case ack := <-p.acks:
				p.processAck(ack)
			case frontier := <-p.frontiers:
				p.processFrontier(frontier)
			default:
				break PriorityOut
			}
		}

		select {
		case <-p.quitCh:
			return
		case bs := <-p.blocks:
			p.processMsgDo(bs)
		case block := <-p.syncBlock:
			p.syncBlockCheck(block)
		case cache := <-p.syncCache:
			if cache.kind == common.SyncCacheUnconfirmed {
				_ = p.dps.ledger.DeleteUnconfirmedSyncBlock(cache.hash)
			} else {
				_ = p.dps.ledger.DeleteUncheckedSyncBlock(cache.hash)
			}
		case <-getTimeout.C:
			//
		}
	}
}

func (p *Processor) processConfirmedSync(hash types.Hash, block *types.StateBlock) {
	p.syncBlockAcked <- hash

	ck := chainKey{
		addr:  block.Address,
		token: block.Token,
	}
	if _, ok := p.chainHeight[ck]; ok {
		p.chainHeight[ck]++
	} else {
		p.chainHeight[ck] = 1
	}

	cok := chainOrderKey{
		chainKey: ck,
		order:    p.chainHeight[ck],
	}
	p.orderedChain[cok] = hash
}

func (p *Processor) processFrontier(block *types.StateBlock) {
	hash := block.GetHash()

	if has, _ := p.dps.ledger.HasStateBlockConfirmed(hash); has {
		return
	}

	p.dps.acTrx.addToRoots(block)
	if !p.dps.isReceivedFrontier(hash) {
		p.dps.subAckDo(p.index, hash)
		p.dps.frontiersStatus.Store(hash, frontierWaitingForVote)
		p.syncBlock <- block
	}
}

func (p *Processor) confirmChain(hash types.Hash) {
	dps := p.dps

	if block, err := dps.ledger.GetUnconfirmedSyncBlock(hash); err != nil {
		dps.logger.Debugf("get unconfirmed sync block err", err)
		return
	} else {
		cok := chainOrderKey{
			chainKey: chainKey{block.Address, block.Token},
			order:    1,
		}

		for {
			h, ok := p.orderedChain[cok]
			if ok {
				if blk, err := dps.ledger.GetUnconfirmedSyncBlock(h); err != nil {
					break
				} else {
					bs := &consensus.BlockSource{
						Block:     blk,
						BlockFrom: types.Synchronized,
						Type:      consensus.MsgSync,
					}
					p.blocks <- bs

					if err := dps.ledger.DeleteUnconfirmedSyncBlock(h); err != nil {
						dps.logger.Errorf("delete unconfirmed sync block err", err)
					}
				}
			} else {
				break
			}

			cok.order++
		}
	}
}

func (p *Processor) processAck(vi *voteInfo) {
	dps := p.dps
	dps.logger.Infof("processor recv confirmAck block[%s]", vi.hash)

	if !p.dps.isWaitingFrontier(vi.hash) {
		if has, _ := dps.ledger.HasStateBlockConfirmed(vi.hash); !has {
			dps.acTrx.vote(vi)
		} else {
			dps.heartAndVoteInc(vi.hash, vi.account, onlineKindVote)
		}
	} else {
		if dps.acTrx.voteFrontier(vi) {
			p.dps.frontiersStatus.Store(vi.hash, frontierConfirmed)
			p.syncBlockAcked <- vi.hash
			dps.logger.Infof("frontier %s confirmed", vi.hash)
		}
	}
}

func (p *Processor) processMsgDo(bs *consensus.BlockSource) {
	dps := p.dps
	hash := bs.Block.GetHash()
	var result process.ProcessResult
	var err error

	if bs.BlockFrom == types.Synchronized {
		result, err = dps.lv.BlockSyncCheck(bs.Block)
		if err != nil {
			dps.logger.Infof("block[%s] check err[%s]", hash, err.Error())
			return
		}
	} else {
		result, err = dps.lv.BlockCheck(bs.Block)
		if err != nil {
			dps.logger.Infof("block[%s] check err[%s]", hash, err.Error())
			return
		}
	}
	p.processResult(result, bs)

	switch bs.Type {
	case consensus.MsgPublishReq:
		dps.logger.Infof("dps recv publishReq block[%s]", hash)
	case consensus.MsgConfirmReq:
		dps.logger.Infof("dps recv confirmReq block[%s]", hash)

		//vote if the result is progress and old
		if result == process.Progress {
			if el := dps.acTrx.getVoteInfo(bs.Block); el != nil {
				if el.voteHash == types.ZeroHash {
					dps.localRepVote(bs.Block)
				} else if hash == el.voteHash {
					dps.localRepVote(bs.Block)
				}
			}
		} else if result == process.Old {
			dps.localRepVote(bs.Block)
		}
	case consensus.MsgSync:
		//do nothing
	case consensus.MsgGenerateBlock:
		if dps.getPovSyncState() != common.SyncDone {
			dps.logger.Errorf("pov is syncing, can not send tx!")
			return
		}

		//cache fork
		if result == process.Progress {
			el := dps.acTrx.getVoteInfo(bs.Block)
			if el != nil && el.status.winner.GetHash() != hash {
				_ = dps.lv.Rollback(hash)
				return
			}
		}

		dps.acTrx.updatePerfTime(hash, time.Now().UnixNano(), false)
	default:
		//
	}
}

func (p *Processor) processResult(result process.ProcessResult, bs *consensus.BlockSource) {
	blk := bs.Block
	hash := blk.GetHash()
	dps := p.dps

	switch result {
	case process.Progress:
		if bs.BlockFrom == types.Synchronized {
			p.confirmBlock(bs.Block)
			dps.logger.Infof("Block %s from sync", hash)
		} else if bs.BlockFrom == types.UnSynchronized {
			dps.logger.Infof("Add block %s to roots", hash)
			//make sure we only vote one of the forked blocks
			if dps.acTrx.addToRoots(blk) && bs.Type != consensus.MsgConfirmReq {
				dps.localRepVote(bs.Block)
			}
			dps.subAckDo(p.index, hash)
		} else {
			dps.logger.Errorf("Block %s UnKnow from", hash)
		}
	case process.BadSignature:
		dps.logger.Errorf("Bad signature for block: %s", hash)
	case process.BadWork:
		dps.logger.Errorf("Bad work for block: %s", hash)
	case process.BalanceMismatch:
		dps.logger.Errorf("Balance mismatch for block: %s", hash)
	case process.Old:
		dps.logger.Debugf("Old for block: %s", hash)
	case process.UnReceivable:
		dps.logger.Errorf("UnReceivable for block: %s", hash)
	case process.GapSmartContract:
		dps.logger.Errorf("GapSmartContract for block: %s", hash)
		//dps.processGapSmartContract(blk)
	case process.InvalidData:
		dps.logger.Errorf("InvalidData for block: %s", hash)
	case process.Other:
		dps.logger.Errorf("UnKnow process result for block: %s", hash)
	case process.Fork:
		dps.logger.Errorf("Fork for block: %s", hash)
		p.processFork(bs)
	case process.GapPrevious:
		dps.logger.Infof("block:[%s] Gap previous:[%s]", hash, blk.Previous.String())
		p.enqueueUnchecked(result, bs)
	case process.GapSource:
		dps.logger.Infof("block:[%s] Gap source:[%s]", hash, blk.Link.String())
		p.enqueueUnchecked(result, bs)
	case process.GapTokenInfo:
		dps.logger.Infof("block:[%s] Gap tokenInfo", hash)
		p.enqueueUnchecked(result, bs)
	}
}

func (p *Processor) confirmBlock(blk *types.StateBlock) {
	hash := blk.GetHash()
	vk := getVoteKey(blk)
	dps := p.dps

	if v, ok := dps.acTrx.roots.Load(vk); ok {
		el := v.(*Election)

		if !el.ifValidAndSetInvalid() {
			return
		}

		dps.acTrx.roots.Delete(el.vote.id)
		dps.acTrx.updatePerfTime(hash, time.Now().UnixNano(), true)

		if el.status.winner.GetHash() != hash {
			dps.logger.Infof("hash:%s ...is loser", el.status.winner.GetHash())
			el.status.loser = append(el.status.loser, el.status.winner)
		}

		t := el.tally(false)
		for _, value := range t {
			thash := value.block.GetHash()
			if thash != hash {
				el.status.loser = append(el.status.loser, value.block)
			}
		}

		el.cleanBlockInfo()
		dps.acTrx.rollBack(el.status.loser)
		dps.acTrx.addSyncBlock2Ledger(blk)
		p.blocksAcked <- hash
		dps.dispatchAckedBlock(blk, hash, p.index)
		dps.eb.Publish(common.EventConfirmedBlock, blk)
	} else {
		dps.acTrx.addSyncBlock2Ledger(blk)
		p.blocksAcked <- hash
		dps.dispatchAckedBlock(blk, hash, p.index)
		dps.eb.Publish(common.EventConfirmedBlock, blk)
	}
}

func (p *Processor) processFork(bs *consensus.BlockSource) {
	dps := p.dps
	newBlock := bs.Block
	confirmedBlock := p.findAnotherForkedBlock(newBlock)
	confirmedHash := confirmedBlock.GetHash()
	newHash := newBlock.GetHash()

	//cache block forked with confirmed block
	if bs.Type == consensus.MsgGenerateBlock {
		_ = dps.lv.Rollback(bs.Block.GetHash())
		dps.logger.Errorf("fork:[new:%s]--[local:%s]--cache forked", newHash, confirmedHash)
		return
	}

	// Block has been packed by pov, can not be rolled back
	if dps.ledger.HasPovTxLookup(confirmedHash) {
		dps.logger.Errorf("fork:[new:%s]--[local:%s]--pov packed", newHash, confirmedHash)
		return
	}

	dps.logger.Errorf("fork:[new:%s]--[local:%s]--pov not packed", newHash, confirmedHash)

	if bs.Type == consensus.MsgGenerateBlock {
		_ = dps.lv.Rollback(bs.Block.GetHash())
		return
	}

	if dps.acTrx.addToRoots(confirmedBlock) {
		confirmReqBlocks := make([]*types.StateBlock, 0)
		confirmReqBlocks = append(confirmReqBlocks, confirmedBlock)
		dps.eb.Publish(common.EventBroadcast, p2p.ConfirmReq, confirmReqBlocks)
		dps.subAckDo(p.index, confirmedBlock.GetHash())

		dps.subAckDo(p.index, newHash)
		if el := dps.acTrx.getVoteInfo(confirmedBlock); el != nil {
			el.blocks.Store(newHash, newBlock)
			dps.hash2el.Store(newHash, el)
		}

		dps.localRepVote(confirmedBlock)
	} else {
		dps.subAckDo(p.index, newHash)
		if el := dps.acTrx.getVoteInfo(confirmedBlock); el != nil {
			el.blocks.Store(newHash, newBlock)
			dps.hash2el.Store(newHash, el)
		}
	}
}

func (p *Processor) findAnotherForkedBlock(block *types.StateBlock) *types.StateBlock {
	dps := p.dps

	var forkedHash types.Hash
	if block.IsOpen() {
		tm, err := dps.ledger.GetTokenMeta(block.GetAddress(), block.GetToken())
		if err != nil {
			dps.logger.Error(err)
			return block
		}
		forkedHash = tm.OpenBlock
	} else {
		hash := block.Parent()
		var err error
		forkedHash, err = dps.ledger.GetChild(hash)
		if err != nil {
			dps.logger.Error(err)
			return block
		}
	}
	forkedBlock, err := dps.ledger.GetStateBlock(forkedHash)
	if err != nil {
		dps.logger.Error(err)
		return block
	}

	return forkedBlock
}

func (p *Processor) processUncheckedBlock(bs *consensus.BlockSource) {
	dps := p.dps
	var result process.ProcessResult

	if bs.BlockFrom == types.Synchronized {
		result, _ = dps.lv.BlockSyncCheck(bs.Block)
	} else {
		result, _ = dps.lv.BlockCheck(bs.Block)
	}

	p.processResult(result, bs)
}

func (p *Processor) enqueueUnchecked(result process.ProcessResult, bs *consensus.BlockSource) {
	//frontier confirmReq may cause problem
	if bs.Type != consensus.MsgConfirmReq {
		p.enqueueUncheckedToDb(result, bs)
	}
}

func (p *Processor) enqueueUncheckedToDb(result process.ProcessResult, bs *consensus.BlockSource) {
	blk := bs.Block
	dps := p.dps

	switch result {
	case process.GapPrevious:
		err := dps.ledger.AddUncheckedBlock(blk.Previous, blk, types.UncheckedKindPrevious, bs.BlockFrom)
		if err != nil && err != ledger.ErrUncheckedBlockExists {
			dps.logger.Errorf("add unchecked block to ledger err %s", err)
		}
	case process.GapSource:
		err := dps.ledger.AddUncheckedBlock(blk.Link, blk, types.UncheckedKindLink, bs.BlockFrom)
		if err != nil && err != ledger.ErrUncheckedBlockExists {
			dps.logger.Errorf("add unchecked block to ledger err %s", err)
		}
	case process.GapTokenInfo:
		input, err := dps.ledger.GetStateBlock(bs.Block.GetLink())
		if err != nil {
			dps.logger.Errorf("get contract send block err %s", err)
			return
		}

		tokenId := new(types.Hash)
		err = cabi.MintageABI.UnpackMethod(tokenId, cabi.MethodNameMintageWithdraw, input.GetData())
		if err != nil {
			dps.logger.Errorf("get token info err %s", err)
			return
		}
		err = dps.ledger.AddUncheckedBlock(*tokenId, blk, types.UncheckedKindTokenInfo, bs.BlockFrom)
		if err != nil && err != ledger.ErrUncheckedBlockExists {
			dps.logger.Errorf("add unchecked block to ledger err %s", err)
		}
	}
}

func (p *Processor) enqueueUncheckedToMem(hash types.Hash, depHash types.Hash, bs *consensus.BlockSource) {
	if !p.uncheckedCache.Has(depHash) {
		consensus.GlobalUncheckedBlockNum.Inc()
		blocks := new(sync.Map)
		blocks.Store(hash, bs)

		err := p.uncheckedCache.Set(depHash, blocks)
		if err != nil {
			p.dps.logger.Errorf("Gap previous set cache err for block:%s", hash)
		}
	} else {
		c, err := p.uncheckedCache.Get(depHash)
		if err != nil {
			p.dps.logger.Errorf("Gap previous get cache err for block:%s", hash)
		}

		blocks := c.(*sync.Map)
		blocks.Store(hash, bs)
	}
}

func (p *Processor) enqueueUncheckedSync(block *types.StateBlock) {
	p.enqueueUncheckedSyncToDb(block)
}

func (p *Processor) enqueueUncheckedSyncToDb(block *types.StateBlock) {
	if err := p.dps.ledger.AddUncheckedSyncBlock(block.Previous, block); err != nil {
		p.dps.logger.Errorf("add unchecked sync block to ledger err %s", err)
	}
}

func (p *Processor) dequeueUnchecked(hash types.Hash) {
	p.dequeueUncheckedFromDb(hash)
}

func (p *Processor) dequeueUncheckedFromDb(hash types.Hash) {
	dps := p.dps

	//hash is token id
	if blkToken, bf, _ := dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindTokenInfo); blkToken != nil {
		if dps.getProcessorIndex(blkToken.Address) == p.index {
			dps.logger.Debugf("dequeue gap token info[%s] block[%s]", hash, blkToken.GetHash())
			bs := &consensus.BlockSource{
				Block:     blkToken,
				BlockFrom: bf,
			}

			p.processUncheckedBlock(bs)
			err := dps.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindTokenInfo)
			if err != nil {
				dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindTokenInfo", err, blkToken.GetHash())
			}
		}
		return
	}

	if blkLink, bf, _ := dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink); blkLink != nil {
		if dps.getProcessorIndex(blkLink.Address) == p.index {
			dps.logger.Debugf("dequeue gap link[%s] block[%s]", hash, blkLink.GetHash())
			bs := &consensus.BlockSource{
				Block:     blkLink,
				BlockFrom: bf,
			}

			p.processUncheckedBlock(bs)

			err := dps.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindLink)
			if err != nil {
				dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindLink", err, blkLink.GetHash())
			}
		}
	}

	if blkPre, bf, _ := dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious); blkPre != nil {
		if dps.getProcessorIndex(blkPre.Address) == p.index {
			dps.logger.Debugf("dequeue gap previous[%s] block[%s]", hash, blkPre.GetHash())
			bs := &consensus.BlockSource{
				Block:     blkPre,
				BlockFrom: bf,
			}

			p.processUncheckedBlock(bs)

			err := dps.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindPrevious)
			if err != nil {
				dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindPrevious", err, blkPre.GetHash())
			}
		}
	}
}

func (p *Processor) dequeueUncheckedFromMem(hash types.Hash) {
	dps := p.dps
	dps.logger.Debugf("dequeue gap[%s]", hash.String())

	if !p.uncheckedCache.Has(hash) {
		return
	}

	m, err := p.uncheckedCache.Get(hash)
	if err != nil {
		dps.logger.Errorf("dequeue unchecked err [%s] for hash [%s]", err, hash)
		return
	}

	cm := m.(*sync.Map)
	cm.Range(func(key, value interface{}) bool {
		bs := value.(*consensus.BlockSource)
		dps.logger.Debugf("dequeue gap[%s] block[%s]", hash, bs.Block.GetHash())

		result, _ := dps.lv.BlockCheck(bs.Block)
		p.processResult(result, bs)

		if p.isResultValid(result) {
			if bs.BlockFrom == types.Synchronized {
				p.confirmBlock(bs.Block)
				return true
			}

			dps.localRepVote(bs.Block)
		}

		return true
	})

	r := p.uncheckedCache.Remove(hash)
	if !r {
		dps.logger.Error("remove cache for unchecked fail")
	}

	if consensus.GlobalUncheckedBlockNum.Load() > 0 {
		consensus.GlobalUncheckedBlockNum.Dec()
	}
}

func (p *Processor) dequeueUncheckedSync(hash types.Hash) {
	p.dequeueUncheckedSyncFromDb(hash)
}

func (p *Processor) dequeueUncheckedSyncFromDb(hash types.Hash) {
	dps := p.dps

	if block, err := dps.ledger.GetUncheckedSyncBlock(hash); err != nil {
		return
	} else {
		confirmedHash := block.GetHash()
		if err := dps.ledger.AddUnconfirmedSyncBlock(confirmedHash, block); err != nil {
			dps.logger.Errorf("add unconfirmed sync block err", err)
		}
		p.processConfirmedSync(confirmedHash, block)

		if err := dps.ledger.DeleteUncheckedSyncBlock(hash); err != nil {
			dps.logger.Errorf("del uncheck sync block err", err)
		}
	}
}

func (p *Processor) rollbackUncheckedFromMem(hash types.Hash) {
	if !p.uncheckedCache.Has(hash) {
		return
	}

	_, err := p.uncheckedCache.Get(hash)
	if err != nil {
		p.dps.logger.Errorf("dequeue unchecked err [%s] for hash [%s]", err, hash)
		return
	}

	r := p.uncheckedCache.Remove(hash)
	if !r {
		p.dps.logger.Error("remove cache for unchecked fail")
	}

	if consensus.GlobalUncheckedBlockNum.Load() > 0 {
		consensus.GlobalUncheckedBlockNum.Dec()
	}
}

func (p *Processor) isResultValid(result process.ProcessResult) bool {
	if result == process.Progress || result == process.Old {
		return true
	} else {
		return false
	}
}

func (p *Processor) isResultGap(result process.ProcessResult) bool {
	if result == process.GapPrevious || result == process.GapSource || result == process.GapTokenInfo {
		return true
	} else {
		return false
	}
}
