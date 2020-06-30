package dpos

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluele/gcache"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const ConfirmChainParallelNum = 10

type chainKey struct {
	addr  types.Address
	token types.Hash
}

type chainOrderKey struct {
	chainKey
	order uint64
}

type Processor struct {
	index                int
	dps                  *DPoS
	uncheckedCache       gcache.Cache // gap blocks
	quitCh               chan bool
	blocks               chan *consensus.BlockSource
	blocksAcked          chan types.Hash
	tokenCreate          chan types.Hash
	publishBlock         chan types.Hash
	dodSettleStateChange chan types.Hash
	syncBlock            chan *types.StateBlock
	syncAcked            chan types.Hash
	acks                 chan *voteInfo
	frontiers            chan *types.StateBlock
	syncStateChange      chan topic.SyncState
	syncState            topic.SyncState
	orderedChain         *sync.Map
	chainHeight          *sync.Map
	confirmedChain       map[types.Hash]bool
	confirmParallelNum   int32
	ctx                  context.Context
	cancel               context.CancelFunc
	exited               chan struct{}
}

func newProcessors(num int) []*Processor {
	processors := make([]*Processor, 0)

	for i := 0; i < num; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		p := &Processor{
			index:                i,
			quitCh:               make(chan bool, 1),
			blocks:               make(chan *consensus.BlockSource, common.DPoSMaxBlocks),
			blocksAcked:          make(chan types.Hash, common.DPoSMaxBlocks),
			tokenCreate:          make(chan types.Hash, 1024),
			publishBlock:         make(chan types.Hash, common.DPoSMaxBlocks),
			dodSettleStateChange: make(chan types.Hash, common.DPoSMaxBlocks),
			syncBlock:            make(chan *types.StateBlock, common.DPoSMaxBlocks),
			syncAcked:            make(chan types.Hash, common.DPoSMaxBlocks),
			acks:                 make(chan *voteInfo, common.DPoSMaxBlocks),
			frontiers:            make(chan *types.StateBlock, common.DPoSMaxBlocks),
			syncStateChange:      make(chan topic.SyncState, 1),
			syncState:            topic.SyncNotStart,
			orderedChain:         new(sync.Map),
			chainHeight:          new(sync.Map),
			confirmedChain:       make(map[types.Hash]bool),
			confirmParallelNum:   ConfirmChainParallelNum,
			ctx:                  ctx,
			cancel:               cancel,
			exited:               make(chan struct{}, 1),
		}
		processors = append(processors, p)
	}

	return processors
}

func (p *Processor) setDPoSService(dps *DPoS) {
	p.dps = dps
}

func (p *Processor) start() {
	go p.processMsg()
}

func (p *Processor) stop() {
	p.quitCh <- true
	p.cancel()

	<-p.exited
	for atomic.LoadInt32(&p.confirmParallelNum) != ConfirmChainParallelNum {
		time.Sleep(10 * time.Millisecond)
	}
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
				return
			}
		}
	}

	if checked {
		if err := dps.ledger.AddUnconfirmedSyncBlock(hash, block); err != nil {
			dps.logger.Errorf("add unconfirmed sync block err", err)
		}
		p.processConfirmedSync(hash, block)
	} else {
		dps.logger.Errorf("sync block not confirmed %s", hash)
	}
}

func (p *Processor) drainAck() {
	for {
		select {
		case ack := <-p.acks:
			p.processAck(ack)
		default:
			return
		}
	}
}

func (p *Processor) processMsg() {
	timerConfirm := time.NewTicker(time.Second)

	for {
		select {
		case <-p.quitCh:
			p.exited <- struct{}{}
			return
		case p.syncState = <-p.syncStateChange:
			p.dps.syncStateNotifyWait.Done()

			if p.syncState == topic.Syncing {
				p.orderedChain = new(sync.Map)
				p.chainHeight = new(sync.Map)
				p.confirmedChain = make(map[types.Hash]bool)
			}
		case hash := <-p.blocksAcked:
			p.dequeueUnchecked(hash)
		case hash := <-p.tokenCreate:
			p.dequeueGapToken(hash)
		case hash := <-p.publishBlock:
			p.dequeueGapPublish(hash)
		case hash := <-p.dodSettleStateChange:
			p.dequeueGapDoDSettleState(hash)
		case ack := <-p.acks:
			p.processAck(ack)
			p.drainAck()
		case frontier := <-p.frontiers:
			p.processFrontier(frontier)
		case bs := <-p.blocks:
			p.processMsgDo(bs)
		case block := <-p.syncBlock:
			p.dps.updateLastProcessSyncTime()
			p.syncBlockCheck(block)
		case ack := <-p.syncAcked:
			p.dequeueUncheckedSync(ack)
		case <-timerConfirm.C:
			if p.syncState == topic.SyncDone || p.syncState == topic.Syncing {
				for hash, dealt := range p.confirmedChain {
					if dealt {
						continue
					}

					if atomic.LoadInt32(&p.confirmParallelNum) > 0 {
						// use go-routing to forbid blocking the thread
						p.confirmedChain[hash] = true
						go p.confirmChain(hash)
					} else {
						break
					}
				}
			}
		}
	}
}

func (p *Processor) ackedBlockNotify(hash types.Hash) {
	select {
	case p.blocksAcked <- hash:
	default:
	}
}

func (p *Processor) publishBlockNotify(hash types.Hash) {
	select {
	case p.publishBlock <- hash:
	default:
	}
}

func (p *Processor) tokenCreateNotify(hash types.Hash) {
	select {
	case p.tokenCreate <- hash:
	default:
	}
}

func (p *Processor) dodSettleStateChangeNotify(hash types.Hash) {
	select {
	case p.dodSettleStateChange <- hash:
	default:
	}
}

func (p *Processor) processConfirmedSync(hash types.Hash, block *types.StateBlock) {
	var height uint64

	err := block.CheckPrivateRecvRsp()
	if err != nil {
		p.dps.logger.Errorf("block %s check private err %s", hash, err)
		return
	}

	ck := chainKey{
		addr:  block.Address,
		token: block.Token,
	}

	if v, ok := p.chainHeight.Load(ck); ok {
		height = v.(uint64) + 1
	} else {
		height = 1
	}

	p.chainHeight.Store(ck, height)
	cok := chainOrderKey{
		chainKey: ck,
		order:    height,
	}

	p.orderedChain.Store(cok, hash)
	p.processAckedSync(hash)
}

func (p *Processor) processAckedSync(hash types.Hash) {
	if p.dps.isConfirmedFrontier(hash) {
		if s, ok := p.dps.frontiersStatus.Load(hash); ok && s == frontierConfirmed {
			p.dps.frontiersStatus.Store(hash, frontierChainConfirmed)
		}

		if _, ok := p.confirmedChain[hash]; !ok {
			p.confirmedChain[hash] = false
		}
	}

	p.syncAcked <- hash
}

func (p *Processor) localRepVoteFrontier(block *types.StateBlock) {
	hash := block.GetHash()
	dps := p.dps

	dps.localRepAccount.Range(func(key, value interface{}) bool {
		address := key.(types.Address)

		weight := dps.ledger.Weight(address)
		if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
			return true
		}

		if block.Type == types.Online {
			if !dps.isOnline(block.Address) {
				dps.logger.Debugf("block[%s] is not online", hash)
				return false
			}
		}

		vi := &voteInfo{
			hash:    hash,
			account: address,
		}

		select {
		case p.acks <- vi:
			dps.acTrx.setVoteHash(block)
			dps.logger.Debugf("rep [%s] vote for block[%s] type[%s]", address, hash, block.Type)
		default:
		}

		return true
	})
}

func (p *Processor) processFrontier(block *types.StateBlock) {
	hash := block.GetHash()
	dps := p.dps

	dps.subAckDo(p.index, hash)

	if !p.dps.acTrx.addToRoots(block) {
		if el := dps.acTrx.getVoteInfo(block); el != nil {
			el.blocks.Store(hash, block)
			el.frontier.Store(hash, nil)
			dps.hash2el.Store(hash, el)
		} else {
			dps.logger.Errorf("get election err[%s]", hash)
		}
	} else {
		if el := dps.acTrx.getVoteInfo(block); el != nil {
			el.frontier.Store(hash, nil)
		}
	}

	result, _ := dps.lv.BlockCheck(block)
	if result == process.Progress {
		p.localRepVoteFrontier(block)
	}
}

func (p *Processor) confirmChain(hash types.Hash) {
	atomic.AddInt32(&p.confirmParallelNum, -1)
	defer atomic.AddInt32(&p.confirmParallelNum, 1)

	dps := p.dps
	dps.logger.Debugf("confirm chain %s", hash)

	if block, err := dps.ledger.GetUnconfirmedSyncBlock(hash); err != nil {
		dps.logger.Errorf("get unconfirmed sync block err %s", err)
		return
	} else {
		ck := chainKey{
			addr:  block.Address,
			token: block.Token,
		}

		vh, ok := p.chainHeight.Load(ck)
		if !ok {
			return
		}
		height := vh.(uint64)

		cok := chainOrderKey{
			chainKey: ck,
			order:    1,
		}

		for {
			if h, ok := p.orderedChain.Load(cok); ok {
				bHash := h.(types.Hash)
				if blk, _ := dps.ledger.GetUnconfirmedSyncBlock(bHash); blk != nil {
					dps.logger.Debugf("confirm chain block %s", bHash)
					bs := &consensus.BlockSource{
						Block:     blk,
						BlockFrom: types.Synchronized,
						Type:      consensus.MsgSync,
					}

					select {
					case <-p.ctx.Done():
						return
					case p.blocks <- bs:
					}

					if err := dps.ledger.DeleteUnconfirmedSyncBlock(bHash); err != nil {
						dps.logger.Errorf("delete unconfirmed sync block err", err)
					}
				}
			} else {
				dps.logger.Infof("get chain block err(%s-%s) order(%d)", block.Address, block.Token, cok.order)
				break
			}

			cok.order++
			if cok.order > height {
				break
			}
		}
	}
}

func (p *Processor) processAck(vi *voteInfo) {
	dps := p.dps
	dps.logger.Infof("processor recv confirmAck block[%s]", vi.hash)

	if val, ok := dps.hash2el.Load(vi.hash); ok {
		el := val.(*Election)
		if _, ok := el.frontier.Load(vi.hash); ok {
			if dps.acTrx.voteFrontier(vi) {
				p.dps.frontiersStatus.Store(vi.hash, frontierConfirmed)
				p.processAckedSync(vi.hash)
				dps.logger.Warnf("frontier %s confirmed", vi.hash)
			}
		} else {
			if has, _ := dps.ledger.HasStateBlockConfirmed(vi.hash); !has {
				dps.acTrx.vote(vi)
			} else {
				dps.heartAndVoteInc(vi.hash, vi.account, onlineKindVote)
			}
		}
	}
}

func (p *Processor) processMsgDo(bs *consensus.BlockSource) {
	dps := p.dps
	hash := bs.Block.GetHash()
	var result process.ProcessResult
	var err error

	err = bs.Block.CheckPrivateRecvRsp()
	if err != nil {
		dps.logger.Errorf("block %s check private err %s", hash, err)
		return
	}

	dps.perfBlockProcessCheckPointAdd(hash, checkPointBlockCheck)

	if bs.BlockFrom == types.Synchronized {
		bs.Block.SetFromSync()
		p.dps.updateLastProcessSyncTime()
	}

	result, err = dps.lv.BlockCheck(bs.Block)
	if result == process.Other {
		dps.logger.Infof("block[%s] check err[%s]", hash, err)
		return
	}

	dps.perfBlockProcessCheckPointAdd(hash, checkPointProcessResult)

	p.processResult(result, bs)

	dps.perfBlockProcessCheckPointAdd(hash, checkPointEnd)

	switch bs.Type {
	case consensus.MsgPublishReq:
		dps.logger.Infof("dps recv publishReq block[%s]", hash)
	case consensus.MsgConfirmReq:
		dps.logger.Infof("dps recv confirmReq block[%s]", hash)

		// vote if the result is progress and old
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
		// do nothing
	case consensus.MsgGenerateBlock:
		// cache fork
		if result == process.Progress {
			el := dps.acTrx.getVoteInfo(bs.Block)
			if el != nil && el.status.winner.GetHash() != hash {
				_ = dps.lv.Rollback(hash)
				return
			}
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
			p.confirmBlock(bs.Block)
			dps.logger.Infof("Block %s from sync", hash)
		} else if bs.BlockFrom == types.UnSynchronized {
			dps.logger.Infof("Add block %s to roots", hash)
			// make sure we only vote one of the forked blocks
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
		// dps.processGapSmartContract(blk)
	case process.InvalidData:
		dps.logger.Errorf("InvalidData for block: %s", hash)
	case process.Other:
		dps.logger.Errorf("UnKnow process result for block: %s", hash)
	case process.Fork:
		dps.logger.Errorf("Fork for block: %s", hash)
		p.processFork(bs)
	case process.GapPrevious:
		dps.logger.Infof("block:[%s] Gap previous:[%s]", hash, blk.Previous)
		p.enqueueUnchecked(result, bs)
	case process.GapSource:
		dps.logger.Infof("block:[%s] Gap source:[%s]", hash, blk.Link)
		p.enqueueUnchecked(result, bs)
	case process.GapTokenInfo:
		dps.logger.Infof("block:[%s] Gap tokenInfo", hash)
		p.enqueueUnchecked(result, bs)
	case process.GapPovHeight:
		dps.logger.Infof("block:[%s] Gap pov height", hash)
		p.enqueueUnchecked(result, bs)
	case process.GapPublish:
		dps.logger.Infof("block:[%s] Gap publish", hash)
		p.enqueueUnchecked(result, bs)
	case process.GapDoDSettleState:
		dps.logger.Infof("block:[%s] Gap DoD settle state", hash)
		p.enqueueUnchecked(result, bs)
	}
}

func (p *Processor) confirmBlock(blk *types.StateBlock) {
	hash := blk.GetHash()
	vk := getVoteKey(blk)
	dps := p.dps

	if v, ok := dps.acTrx.roots.Load(vk); ok {
		el := v.(*Election)
		loser := make([]*types.StateBlock, 0)

		if !el.ifValidAndSetInvalid() {
			return
		}

		dps.acTrx.roots.Delete(el.vote.id)

		el.blocks.Range(func(key, value interface{}) bool {
			if key.(types.Hash) != hash {
				loser = append(loser, value.(*types.StateBlock))
			}
			return true
		})

		el.cleanBlockInfo()
		dps.acTrx.rollBack(loser)
		dps.acTrx.addSyncBlock2Ledger(blk)
		p.ackedBlockNotify(hash)
		dps.dispatchAckedBlock(blk, hash, p.index)
		dps.eb.Publish(topic.EventConfirmedBlock, blk)
	} else {
		dps.acTrx.addSyncBlock2Ledger(blk)
		p.ackedBlockNotify(hash)
		dps.dispatchAckedBlock(blk, hash, p.index)
		dps.eb.Publish(topic.EventConfirmedBlock, blk)
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

	if dps.acTrx.addToRoots(confirmedBlock) {
		confirmReqBlocks := make([]*types.StateBlock, 0)
		confirmReqBlocks = append(confirmReqBlocks, confirmedBlock)
		dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.ConfirmReq, Message: confirmReqBlocks})
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
		forkedHash, err = dps.ledger.GetBlockChild(hash)
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

	if has, _ := dps.ledger.HasBlockCache(bs.Block.GetHash()); has {
		dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.PublishReq, Message: bs.Block})
	}

	if bs.BlockFrom == types.Synchronized {
		bs.Block.SetFromSync()
		p.dps.updateLastProcessSyncTime()
	}

	result, _ = dps.lv.BlockCheck(bs.Block)
	p.processResult(result, bs)
}

func (p *Processor) enqueueUnchecked(result process.ProcessResult, bs *consensus.BlockSource) {
	// frontier confirmReq may cause problem
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
		err = mintage.MintageABI.UnpackMethod(tokenId, mintage.MethodNameMintageWithdraw, input.GetData())
		if err != nil {
			dps.logger.Errorf("get token info err %s", err)
			return
		}
		err = dps.ledger.AddUncheckedBlock(*tokenId, blk, types.UncheckedKindTokenInfo, bs.BlockFrom)
		if err != nil && err != ledger.ErrUncheckedBlockExists {
			dps.logger.Errorf("add unchecked block to ledger err %s", err)
		}
	case process.GapPublish:
		info := new(cabi.OracleInfo)
		err := cabi.PublicKeyDistributionABI.UnpackMethod(info, cabi.MethodNamePKDOracle, blk.GetData())
		if err != nil {
			dps.logger.Errorf("unpack oracle data err %s", err)
		}

		err = dps.ledger.AddGapPublishBlock(info.Hash, blk, bs.BlockFrom)
		if err != nil && err != ledger.ErrUncheckedBlockExists {
			dps.logger.Errorf("add gap publish block to ledger err %s", err)
		}
	case process.GapPovHeight:
		// check private tx
		if bs.Block.IsPrivate() && !bs.Block.IsRecipient() {
			break
		}

		if c, ok, err := contract.GetChainContract(types.Address(bs.Block.Link), bs.Block.GetPayload()); ok && err == nil {
			d := c.GetDescribe()
			switch d.GetVersion() {
			case contract.SpecVer2:
				vmCtx := vmstore.NewVMContextWithBlock(dps.ledger, bs.Block)
				if vmCtx == nil {
					dps.logger.Errorf("enqueue unchecked: can not get vm context")
					return
				}

				gapResult, gapInfo, err := c.DoGap(vmCtx, bs.Block)
				if err != nil || gapResult != common.ContractRewardGapPov {
					dps.logger.Errorf("check gap %s", err)
				}

				height := gapInfo.(uint64)
				if height > 0 {
					dps.logger.Infof("add gap pov[%s][%d]", bs.Block.GetHash(), height)
					err := dps.ledger.AddGapPovBlock(height, bs.Block, bs.BlockFrom)
					if err != nil {
						dps.logger.Errorf("add gap pov block to ledger err %s", err)
					}

					dps.gapHeight <- height
				}
			}
		}
	case process.GapDoDSettleState:
		// check private tx
		if bs.Block.IsPrivate() && !bs.Block.IsRecipient() {
			break
		}

		if c, ok, err := contract.GetChainContract(types.Address(bs.Block.Link), bs.Block.GetPayload()); ok && err == nil {
			d := c.GetDescribe()
			switch d.GetVersion() {
			case contract.SpecVer2:
				vmCtx := vmstore.NewVMContextWithBlock(dps.ledger, bs.Block)
				if vmCtx == nil {
					dps.logger.Errorf("enqueue unchecked: can not get vm context")
					return
				}

				gapResult, gapInfo, err := c.DoGap(vmCtx, bs.Block)
				if err != nil || gapResult != common.ContractDoDOrderState {
					dps.logger.Errorf("check gap err %s", err)
				}

				internalId := gapInfo.(types.Hash)
				if !internalId.IsZero() {
					dps.logger.Infof("add gap dod settle state[%s][%s]", bs.Block.GetHash(), internalId)
					err := dps.ledger.AddGapDoDSettleStateBlock(internalId, bs.Block, bs.BlockFrom)
					if err != nil && err != ledger.ErrUncheckedBlockExists {
						dps.logger.Errorf("add gap dod settle state block to ledger err %s", err)
					}
				}
			}
		}
	}
}

// func (p *Processor) enqueueUncheckedToMem(hash types.Hash, depHash types.Hash, bs *consensus.BlockSource) {
// 	if !p.uncheckedCache.Has(depHash) {
// 		consensus.GlobalUncheckedBlockNum.Inc()
// 		blocks := new(sync.Map)
// 		blocks.DBStore(hash, bs)
//
// 		err := p.uncheckedCache.Set(depHash, blocks)
// 		if err != nil {
// 			p.dps.logger.Errorf("Gap previous set cache err for block:%s", hash)
// 		}
// 	} else {
// 		c, err := p.uncheckedCache.Get(depHash)
// 		if err != nil {
// 			p.dps.logger.Errorf("Gap previous get cache err for block:%s", hash)
// 		}
//
// 		blocks := c.(*sync.Map)
// 		blocks.DBStore(hash, bs)
// 	}
// }

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

	// check if the block is existed, may be from another thread
	if has, _ := dps.ledger.HasStateBlockConfirmed(hash); !has {
		dps.logger.Debugf("get confirmed block fail %s", hash)
		time.Sleep(time.Millisecond)
		p.ackedBlockNotify(hash)
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

func (p *Processor) dequeueGapToken(hash types.Hash) {
	dps := p.dps

	// check if the block is existed, may be from another thread
	blk, err := dps.ledger.GetStateBlockConfirmed(hash)
	if err != nil || blk == nil {
		dps.logger.Debugf("get confirmed block fail %s", hash)
		time.Sleep(time.Millisecond)
		p.tokenCreateNotify(hash)
		return
	}

	input, err := dps.ledger.GetStateBlockConfirmed(blk.GetLink())
	if err != nil {
		dps.logger.Errorf("get block link error [%s]", hash)
		return
	}

	param := new(mintage.ParamMintage)
	if err := mintage.MintageABI.UnpackMethod(param, mintage.MethodNameMintage, input.GetData()); err != nil {
		return
	}

	tokenId := param.TokenId
	if blkToken, bf, _ := dps.ledger.GetUncheckedBlock(tokenId, types.UncheckedKindTokenInfo); blkToken != nil {
		if dps.getProcessorIndex(blkToken.Address) == p.index {
			dps.logger.Debugf("dequeue gap token info[%s] block[%s]", tokenId, blkToken.GetHash())
			bs := &consensus.BlockSource{
				Block:     blkToken,
				BlockFrom: bf,
			}

			p.processUncheckedBlock(bs)

			err := dps.ledger.DeleteUncheckedBlock(tokenId, types.UncheckedKindTokenInfo)
			if err != nil {
				dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindTokenInfo", err, blkToken.GetHash())
			}
		}
	}
}

func (p *Processor) dequeueGapPublish(hash types.Hash) {
	dps := p.dps

	// check if the block is existed, may be from another thread
	blk, err := dps.ledger.GetStateBlockConfirmed(hash)
	if err != nil || blk == nil {
		dps.logger.Debugf("get confirmed block fail %s", hash)
		time.Sleep(time.Millisecond)
		p.publishBlockNotify(hash)
		return
	}

	// deal gap publish
	depHash := blk.Previous
	if err := dps.ledger.GetGapPublishBlock(depHash, func(block *types.StateBlock, bf types.SynchronizedKind) error {
		if dps.getProcessorIndex(block.Address) == p.index {
			dps.logger.Debugf("dequeue gap publish[%s] block[%s]", hash, block.GetHash())
			bs := &consensus.BlockSource{
				Block:     block,
				BlockFrom: bf,
			}

			p.processUncheckedBlock(bs)

			err := dps.ledger.DeleteGapPublishBlock(depHash, block.GetHash())
			if err != nil {
				dps.logger.Errorf("Get err [%s] for hash: [%s] when delete gapPublishBlock", err, block.GetHash())
			}
		}

		return nil
	}); err != nil {
		dps.logger.Errorf("dequeue gap publish err %s", err)
	}
}

func (p *Processor) dequeueGapDoDSettleState(hash types.Hash) {
	dps := p.dps

	// check if the block is existed, may be from another thread
	blk, err := dps.ledger.GetStateBlockConfirmed(hash)
	if err != nil || blk == nil {
		dps.logger.Debugf("get confirmed block fail %s", hash)
		time.Sleep(time.Millisecond)
		p.dodSettleStateChangeNotify(hash)
		return
	}

	var internalId types.Hash

	if blk.Type == types.ContractSend {
		param := new(cabi.DoDSettleUpdateOrderInfoParam)
		err := param.FromABI(blk.GetPayload())
		if err != nil {
			dps.logger.Errorf("unpack data err %s", err)
			return
		}

		internalId = param.InternalId
	} else {
		input, err := dps.ledger.GetStateBlockConfirmed(blk.GetLink())
		if err != nil {
			dps.logger.Errorf("get block link error [%s]", hash)
			return
		}

		internalId = input.Previous
	}

	if err = dps.ledger.GetGapDoDSettleStateBlock(internalId, func(block *types.StateBlock, sync types.SynchronizedKind) error {
		if dps.getProcessorIndex(block.Address) == p.index {
			dps.logger.Debugf("dequeue gap dod settle state internal id[%s] block[%s]", internalId, block.GetHash())
			bs := &consensus.BlockSource{
				Block:     block,
				BlockFrom: sync,
			}

			p.processUncheckedBlock(bs)

			blkHash := block.GetHash()
			err := dps.ledger.DeleteGapDoDSettleStateBlock(internalId, blkHash)
			if err != nil {
				dps.logger.Errorf("Get err [%s] for hash: [%s] when delete GapDoDSettleStateBlock", err, blkHash)
			}
		}
		return nil
	}); err != nil {
		dps.logger.Errorf("dequeue gap dod settle publish err %s", err)
	}
}

// func (p *Processor) dequeueUncheckedFromMem(hash types.Hash) {
// 	dps := p.dps
// 	dps.logger.Debugf("dequeue gap[%s]", hash.String())
//
// 	if !p.uncheckedCache.Has(hash) {
// 		return
// 	}
//
// 	m, err := p.uncheckedCache.Get(hash)
// 	if err != nil {
// 		dps.logger.Errorf("dequeue unchecked err [%s] for hash [%s]", err, hash)
// 		return
// 	}
//
// 	cm := m.(*sync.Map)
// 	cm.Range(func(key, value interface{}) bool {
// 		bs := value.(*consensus.BlockSource)
// 		dps.logger.Debugf("dequeue gap[%s] block[%s]", hash, bs.Block.GetHash())
//
// 		result, _ := dps.lv.BlockCheck(bs.Block)
// 		p.processResult(result, bs)
//
// 		if p.isResultValid(result) {
// 			if bs.BlockFrom == types.Synchronized {
// 				p.confirmBlock(bs.Block)
// 				return true
// 			}
//
// 			dps.localRepVote(bs.Block)
// 		}
//
// 		return true
// 	})
//
// 	r := p.uncheckedCache.Remove(hash)
// 	if !r {
// 		dps.logger.Error("remove cache for unchecked fail")
// 	}
//
// 	if consensus.GlobalUncheckedBlockNum.Load() > 0 {
// 		consensus.GlobalUncheckedBlockNum.Dec()
// 	}
// }

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

// func (p *Processor) rollbackUncheckedFromMem(hash types.Hash) {
// 	if !p.uncheckedCache.Has(hash) {
// 		return
// 	}
//
// 	_, err := p.uncheckedCache.Get(hash)
// 	if err != nil {
// 		p.dps.logger.Errorf("dequeue unchecked err [%s] for hash [%s]", err, hash)
// 		return
// 	}
//
// 	r := p.uncheckedCache.Remove(hash)
// 	if !r {
// 		p.dps.logger.Error("remove cache for unchecked fail")
// 	}
//
// 	if consensus.GlobalUncheckedBlockNum.Load() > 0 {
// 		consensus.GlobalUncheckedBlockNum.Dec()
// 	}
// }
