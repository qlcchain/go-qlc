package dpos

import (
	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"sync"
	"time"
)

type Processor struct {
	dps            *DPoS
	uncheckedCache gcache.Cache //gap blocks
	voteCache      gcache.Cache //vote blocks
	quitCh         chan bool
	blocks         chan *consensus.BlockSource
	blocksAcked    chan types.Hash
}

func newProcessors(num int) []*Processor {
	processors := make([]*Processor, 0)

	for i := 0; i < num; i++ {
		p := &Processor{
			voteCache:   gcache.New(voteCacheSize).LRU().Build(),
			quitCh:      make(chan bool, 1),
			blocks:      make(chan *consensus.BlockSource, maxBlocks),
			blocksAcked: make(chan types.Hash, maxBlocks),
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

func (p *Processor) processMsg() {
	getTimeout := time.NewTicker(10 * time.Millisecond)

	for {
	DequeueOut:
		for {
			select {
			case hash := <-p.blocksAcked:
				p.dequeueUnchecked(hash)
			default:
				break DequeueOut
			}
		}

		select {
		case <-p.quitCh:
			return
		case bs := <-p.blocks:
			p.processMsgDo(bs)
		case <-getTimeout.C:
			//
		}
	}
}

func (p *Processor) processMsgDo(bs *consensus.BlockSource) {
	var result process.ProcessResult
	var err error
	hash := bs.Block.GetHash()
	dps := p.dps

	result, err = dps.lv.BlockCheck(bs.Block)
	if err != nil {
		dps.logger.Infof("block[%s] check err[%s]", hash, err.Error())
		return
	}
	p.processResult(result, bs)

	switch bs.Type {
	case consensus.MsgPublishReq:
		dps.logger.Infof("dps recv publishReq block[%s]", hash)
		if result != process.Old && result != process.Fork {
			//if send ack, there's no need to send publish
			if dps.hasLocalValidRep() {
				dps.localRepVote(bs)
			} else {
				dps.eb.Publish(string(common.EventSendMsgToPeers), p2p.PublishReq, bs.Block, bs.MsgFrom)
			}
		}
	case consensus.MsgConfirmReq:
		dps.logger.Infof("dps recv confirmReq block[%s]", hash)
		if result != process.Fork {
			if !dps.hasLocalValidRep() {
				dps.eb.Publish(string(common.EventSendMsgToPeers), p2p.ConfirmReq, bs.Block, bs.MsgFrom)
			}

			if p.isResultValid(result) {
				dps.localRepVote(bs)
			}
		}
	case consensus.MsgConfirmAck:
		dps.logger.Infof("dps recv confirmAck block[%s]", hash)
		ack := bs.Para.(*protos.ConfirmAckBlock)
		dps.saveOnlineRep(ack.Account)

		//retransmit if the block has not reached a consensus or seq is not 0(for finding reps)
		dps.eb.Publish(string(common.EventSendMsgToPeers), p2p.ConfirmAck, ack, bs.MsgFrom)

		//cache the ack messages
		if p.isResultGap(result) {
			if p.voteCache.Has(hash) {
				v, err := p.voteCache.Get(hash)
				if err != nil {
					dps.logger.Error("get vote cache err")
					return
				}

				vc := v.(*sync.Map)
				vc.Store(ack.Account, ack)
			} else {
				vc := new(sync.Map)
				vc.Store(ack.Account, ack)
				err := p.voteCache.Set(hash, vc)
				if err != nil {
					dps.logger.Error("set vote cache err")
					return
				}
			}
		} else if p.isResultValid(result) { //local send will be old
			dps.acTrx.vote(ack)

			if result == process.Progress {
				dps.localRepVote(bs)
			}
		}
	case consensus.MsgSync:
		if result == process.Progress {
			p.confirmBlock(bs.Block)
		}
	case consensus.MsgGenerateBlock:
		if dps.getPovSyncState() != common.Syncdone {
			dps.logger.Errorf("pov is syncing, can not send tx!")
			return
		}

		dps.acTrx.updatePerfTime(hash, time.Now().UnixNano(), false)
		//dps.acTrx.addToRoots(bs.Block)

		if p.isResultValid(result) {
			dps.localRepVote(bs)
		}
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
			dps.logger.Infof("Block %s from sync,no need consensus", hash)
		} else if bs.BlockFrom == types.UnSynchronized {
			dps.logger.Infof("Block %s basic info is correct,begin add it to roots", hash)
			dps.acTrx.addToRoots(blk)
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
		p.processFork(blk)
	case process.GapPrevious:
		dps.logger.Infof("block:[%s] Gap previous:[%s]", hash, blk.Previous.String())
		p.enqueueUnchecked(result, bs)
	case process.GapSource:
		dps.logger.Infof("block:[%s] Gap source:[%s]", hash, blk.Link.String())
		p.enqueueUnchecked(result, bs)
	}
}

func (p *Processor) confirmBlock(blk *types.StateBlock) {
	hash := blk.GetHash()
	vk := getVoteKey(blk)
	dps := p.dps

	if v, ok := dps.acTrx.roots.Load(vk); ok {
		el := v.(*Election)
		dps.acTrx.roots.Delete(el.vote.id)
		dps.acTrx.updatePerfTime(hash, time.Now().UnixNano(), true)

		if el.status.winner.GetHash().String() != hash.String() {
			dps.logger.Infof("hash:%s ...is loser", el.status.winner.GetHash().String())
			el.status.loser = append(el.status.loser, el.status.winner)
		}

		el.status.winner = blk
		el.confirmed = true

		t := el.tally()
		for _, value := range t {
			if value.block.GetHash().String() != hash.String() {
				el.status.loser = append(el.status.loser, value.block)
			}
		}

		dps.acTrx.rollBack(el.status.loser)
		dps.acTrx.addWinner2Ledger(blk)
		p.blocksAcked <- hash
		dps.dispatchAckedBlock(blk, hash, false)
		dps.eb.Publish(string(common.EventConfirmedBlock), blk)
	} else {
		dps.acTrx.addWinner2Ledger(blk)
		p.blocksAcked <- hash
		dps.dispatchAckedBlock(blk, hash, false)
		dps.eb.Publish(string(common.EventConfirmedBlock), blk)
	}
}

func (p *Processor) processFork(newBlock *types.StateBlock) {
	confirmedBlock := p.findAnotherForkedBlock(newBlock)
	isRep := false
	dps := p.dps
	dps.logger.Errorf("fork:%s--%s", newBlock.GetHash(), confirmedBlock.GetHash())

	if dps.acTrx.addToRoots(confirmedBlock) {
		localRepAccount.Range(func(key, value interface{}) bool {
			address := key.(types.Address)
			isRep = true

			weight := dps.ledger.Weight(address)
			if weight.Compare(minVoteWeight) == types.BalanceCompSmaller {
				return true
			}

			va, err := dps.voteGenerateWithSeq(confirmedBlock, address, value.(*types.Account))
			if err != nil {
				return true
			}

			dps.acTrx.vote(va)
			dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, va)
			return true
		})

		if isRep == false {
			dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmReq, confirmedBlock)
		}
	}
}

func (p *Processor) findAnotherForkedBlock(block *types.StateBlock) *types.StateBlock {
	hash := block.Parent()
	dps := p.dps

	forkedHash, err := dps.ledger.GetChild(hash, block.Address)
	if err != nil {
		dps.logger.Error(err)
		return block
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
	result, _ := dps.lv.BlockCheck(bs.Block)
	p.processResult(result, bs)

	if p.isResultValid(result) {
		if bs.BlockFrom == types.Synchronized {
			p.confirmBlock(bs.Block)
			return
		}

		v, e := p.voteCache.Get(bs.Block.GetHash())
		if e == nil {
			vc := v.(*sync.Map)
			vc.Range(func(key, value interface{}) bool {
				dps.acTrx.vote(value.(*protos.ConfirmAckBlock))
				return true
			})

			p.voteCache.Remove(bs.Block.GetHash())
		}

		localRepAccount.Range(func(key, value interface{}) bool {
			address := key.(types.Address)

			va, err := dps.voteGenerate(bs.Block, address, value.(*types.Account))
			if err != nil {
				return true
			}

			dps.acTrx.vote(va)
			dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, va)

			return true
		})
	}
}

func (p *Processor) enqueueUnchecked(result process.ProcessResult, bs *consensus.BlockSource) {
	p.enqueueUncheckedToDb(result, bs)
}

func (p *Processor) enqueueUncheckedToDb(result process.ProcessResult, bs *consensus.BlockSource) {
	blk := bs.Block
	dps := p.dps

	if result == process.GapPrevious {
		err := dps.ledger.AddUncheckedBlock(blk.Previous, blk, types.UncheckedKindPrevious, bs.BlockFrom)
		if err != nil && err != ledger.ErrUncheckedBlockExists {
			dps.logger.Errorf("add unchecked block to ledger err %s", err)
		}
	} else {
		err := dps.ledger.AddUncheckedBlock(blk.Link, blk, types.UncheckedKindLink, bs.BlockFrom)
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

func (p *Processor) dequeueUnchecked(hash types.Hash) {
	p.dequeueUncheckedFromDb(hash)
}

func (p *Processor) dequeueUncheckedFromDb(hash types.Hash) {
	dps := p.dps

	blkLink, bf, _ := dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink)
	if blkLink != nil {
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

	blkPre, bf, _ := dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious)
	if blkPre != nil {
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

			v, e := p.voteCache.Get(bs.Block.GetHash())
			if e == nil {
				vc := v.(*sync.Map)
				vc.Range(func(key, value interface{}) bool {
					dps.acTrx.vote(value.(*protos.ConfirmAckBlock))
					return true
				})

				p.voteCache.Remove(bs.Block.GetHash())
			}

			localRepAccount.Range(func(key, value interface{}) bool {
				address := key.(types.Address)

				va, err := dps.voteGenerate(bs.Block, address, value.(*types.Account))
				if err != nil {
					return true
				}

				dps.logger.Debugf("rep [%s] vote for block [%s]", address, bs.Block.GetHash())
				dps.acTrx.vote(va)
				dps.eb.Publish(string(common.EventBroadcast), p2p.ConfirmAck, va)

				return true
			})
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

func (p *Processor) rollbackUncheckedFromMem(hash types.Hash) {
	if !p.uncheckedCache.Has(hash) {
		return
	}

	m, err := p.uncheckedCache.Get(hash)
	if err != nil {
		p.dps.logger.Errorf("dequeue unchecked err [%s] for hash [%s]", err, hash)
		return
	}

	cm := m.(*sync.Map)
	cm.Range(func(key, value interface{}) bool {
		bs := value.(*consensus.BlockSource)
		p.rollbackUncheckedFromMem(bs.Block.GetHash())
		return true
	})

	r := p.uncheckedCache.Remove(hash)
	if !r {
		p.dps.logger.Error("remove cache for unchecked fail")
	}

	p.voteCache.Remove(hash)

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
	if result == process.GapPrevious || result == process.GapSource {
		return true
	} else {
		return false
	}
}
