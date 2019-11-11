package dpos

import (
	"context"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluele/gcache"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	repTimeout            = 5 * time.Minute
	voteCacheSize         = 102400
	refreshPriInterval    = 1 * time.Minute
	subAckMaxSize         = 102400
	maxStatisticsPeriod   = 3
	confirmedCacheMaxLen  = 102400
	confirmedCacheMaxTime = 10 * time.Minute
	hashNumPerAck         = 1024
	blockNumPerReq        = 128
)

type subMsgKind byte

const (
	subMsgKindSub subMsgKind = iota
	subMsgKindUnsub
)

const (
	ackTypeCommon uint32 = iota
	ackTypeFindRep
)

type subMsg struct {
	index int
	kind  subMsgKind
	hash  types.Hash
}

type sortContract struct {
	hash      types.Hash
	timestamp int64
}

type frontierStatus byte

const (
	frontierInvalid frontierStatus = iota
	frontierWaitingForVote
	frontierConfirmed
	frontierChainConfirmed
	frontierChainFinished
)

type DPoS struct {
	ledger              *ledger.Ledger
	acTrx               *ActiveTrx
	accounts            []*types.Account
	onlineReps          sync.Map
	logger              *zap.SugaredLogger
	cfg                 *config.Config
	eb                  event.EventBus
	handlerIds          map[common.TopicType]string //topic->handler id
	lv                  *process.LedgerVerifier
	cacheBlocks         chan *consensus.BlockSource
	recvBlocks          chan *consensus.BlockSource
	povState            chan common.SyncState
	syncState           chan common.SyncState
	processors          []*Processor
	processorNum        int
	localRepAccount     sync.Map
	povSyncState        common.SyncState
	blockSyncState      common.SyncState
	minVoteWeight       types.Balance
	voteThreshold       types.Balance
	subAck              gcache.Cache
	subMsg              chan *subMsg
	voteCache           gcache.Cache //vote blocks
	ackSeq              uint32
	hash2el             *sync.Map
	batchVote           chan types.Hash
	ctx                 context.Context
	cancel              context.CancelFunc
	frontiersStatus     *sync.Map
	syncStateNotifyWait *sync.WaitGroup
	totalVote           map[types.Address]types.Balance
	online              gcache.Cache
	confirmedBlocks     gcache.Cache
	lastSendHeight      uint64
	curPovHeight        uint64
	checkFinish         chan struct{}
	gapPovCh            chan *consensus.BlockSource
	povChange           chan *types.PovBlock
	lastGapHeight       uint64
	getFrontier         chan types.StateBlockList
	isFindingRep        int32
	ebDone              chan struct{}
	lastProcessSyncTime time.Time
	updateSync          chan struct{}
	syncDone            chan struct{}
}

func NewDPoS(cfgFile string) *DPoS {
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()

	acTrx := newActiveTrx()
	l := ledger.NewLedger(cfgFile)
	processorNum := runtime.NumCPU()

	ctx, cancel := context.WithCancel(context.Background())

	dps := &DPoS{
		ledger:              l,
		acTrx:               acTrx,
		accounts:            cc.Accounts(),
		logger:              log.NewLogger("dpos"),
		cfg:                 cfg,
		eb:                  cc.EventBus(),
		handlerIds:          make(map[common.TopicType]string),
		lv:                  process.NewLedgerVerifier(l),
		cacheBlocks:         make(chan *consensus.BlockSource, common.DPoSMaxCacheBlocks),
		recvBlocks:          make(chan *consensus.BlockSource, common.DPoSMaxBlocks),
		povState:            make(chan common.SyncState, 1),
		syncState:           make(chan common.SyncState, 1),
		processorNum:        processorNum,
		processors:          newProcessors(processorNum),
		subMsg:              make(chan *subMsg, 10240),
		subAck:              gcache.New(subAckMaxSize).Expiration(confirmReqMaxTimes * time.Minute).LRU().Build(),
		hash2el:             new(sync.Map),
		batchVote:           make(chan types.Hash, 102400),
		ctx:                 ctx,
		cancel:              cancel,
		frontiersStatus:     new(sync.Map),
		syncStateNotifyWait: new(sync.WaitGroup),
		online:              gcache.New(maxStatisticsPeriod).LRU().Build(),
		confirmedBlocks:     gcache.New(confirmedCacheMaxLen).Expiration(confirmedCacheMaxTime).LRU().Build(),
		lastSendHeight:      1,
		curPovHeight:        1,
		checkFinish:         make(chan struct{}, 10240),
		gapPovCh:            make(chan *consensus.BlockSource, 10240),
		povChange:           make(chan *types.PovBlock, 10240),
		voteCache:           gcache.New(voteCacheSize).LRU().Build(),
		blockSyncState:      common.SyncNotStart,
		isFindingRep:        0,
		ebDone:              make(chan struct{}, 1),
		updateSync:          make(chan struct{}, 1),
		syncDone:            make(chan struct{}, 1024),
	}

	dps.acTrx.setDposService(dps)
	for _, p := range dps.processors {
		p.setDposService(dps)
	}

	pb, err := dps.ledger.GetLatestPovBlock()
	if err == nil {
		dps.curPovHeight = pb.Header.BasHdr.Height
	}

	return dps
}

func (dps *DPoS) Init() {
	if dps.cfg.PoV.PovEnabled {
		dps.povSyncState = common.SyncNotStart

		if id, err := dps.eb.SubscribeSync(common.EventPovSyncState, dps.onPovSyncState); err != nil {
			dps.logger.Errorf("subscribe pov sync state event err")
		} else {
			dps.handlerIds[common.EventPovSyncState] = id
		}
	} else {
		dps.povSyncState = common.SyncDone
	}

	supply := common.GenesisBlock().Balance
	dps.minVoteWeight, _ = supply.Div(common.DposVoteDivisor)
	dps.voteThreshold, _ = supply.Div(2)

	if id, err := dps.eb.SubscribeSync(common.EventRollback, dps.onRollback); err != nil {
		dps.logger.Errorf("subscribe rollback block event err")
	} else {
		dps.handlerIds[common.EventRollback] = id
	}

	if id, err := dps.eb.SubscribeSync(common.EventPovConnectBestBlock, dps.onPovHeightChange); err != nil {
		dps.logger.Errorf("subscribe pov connect best block event err")
	} else {
		dps.handlerIds[common.EventPovConnectBestBlock] = id
	}

	if id, err := dps.eb.SubscribeSync(common.EventRpcSyncCall, dps.onRpcSyncCall); err != nil {
		dps.logger.Errorf("subscribe rpc sync call event err")
	} else {
		dps.handlerIds[common.EventRpcSyncCall] = id
	}

	if id, err := dps.eb.SubscribeSync(common.EventFrontierConsensus, dps.onGetFrontier); err != nil {
		dps.logger.Errorf("subscribe frontier consensus event err")
	} else {
		dps.handlerIds[common.EventFrontierConsensus] = id
	}

	if id, err := dps.eb.SubscribeSync(common.EventFrontierConfirmed, dps.onFrontierConfirmed); err != nil {
		dps.logger.Errorf("subscribe frontier confirm event err")
	} else {
		dps.handlerIds[common.EventFrontierConfirmed] = id
	}

	if id, err := dps.eb.SubscribeSync(common.EventSyncStateChange, dps.onSyncStateChange); err != nil {
		dps.logger.Errorf("subscribe sync state change event err")
	} else {
		dps.handlerIds[common.EventSyncStateChange] = id
	}

	if len(dps.accounts) != 0 {
		dps.refreshAccount()
	} else {
		dps.eb.Publish(common.EventRepresentativeNode, false)
	}
}

func (dps *DPoS) Start() {
	dps.logger.Info("DPOS service started!")

	go dps.acTrx.start()
	go dps.batchVoteStart()
	go dps.processSubMsg()
	go dps.processBlocks()
	go dps.processSyncDone()
	dps.processorStart()

	if err := dps.blockSyncDone(); err != nil {
		dps.logger.Error("block sync down err", err)
	}

	timerRefreshPri := time.NewTicker(refreshPriInterval)
	timerDequeueGap := time.NewTicker(10 * time.Second)
	timerUpdateSyncTime := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Info("Stopped DPOS.")
			return
		case <-timerRefreshPri.C:
			dps.logger.Info("refresh pri info.")
			go dps.refreshAccount()
		case <-dps.checkFinish:
			dps.checkSyncFinished()
		case bs := <-dps.gapPovCh:
			if c, ok, err := contract.GetChainContract(types.Address(bs.Block.Link), bs.Block.Data); ok && err == nil {
				switch cType := c.(type) {
				case contract.ChainContractV2:
					vmCtx := vmstore.NewVMContext(dps.ledger)
					height, _ := cType.DoGapPov(vmCtx, bs.Block)
					if height > 0 {
						dps.logger.Infof("add gap pov[%s][%d]", bs.Block.GetHash(), height)
						err := dps.ledger.AddGapPovBlock(height, bs.Block, bs.BlockFrom)
						if err != nil {
							dps.logger.Errorf("add gap pov block to ledger err %s", err)
						}
					}
				}
			}
		case pb := <-dps.povChange:
			dps.logger.Infof("pov height changed [%d]->[%d]", dps.curPovHeight, pb.Header.BasHdr.Height)
			dps.curPovHeight = pb.Header.BasHdr.Height

			if dps.povSyncState == common.SyncDone {
				// need calculate heart num, so use the pov height to trigger online
				// isFindingRep is used to forbidding too many goroutings
				if dps.curPovHeight%2 == 0 && atomic.CompareAndSwapInt32(&dps.isFindingRep, 0, 1) {
					go func() {
						err := dps.findOnlineRepresentatives()
						if err != nil {
							dps.logger.Error(err)
						}
						dps.cleanOnlineReps()
						atomic.StoreInt32(&dps.isFindingRep, 0)
					}()
				}

				if dps.curPovHeight-dps.lastSendHeight >= common.DPosOnlinePeriod &&
					dps.curPovHeight%common.DPosOnlinePeriod >= common.DPosOnlineSectionLeft &&
					dps.curPovHeight%common.DPosOnlinePeriod <= common.DPosOnlineSectionRight {
					dps.sendOnline(dps.curPovHeight)
					dps.lastSendHeight = pb.Header.BasHdr.Height
				}
			}
		case <-timerDequeueGap.C:
			for {
				if dps.lastGapHeight <= dps.curPovHeight {
					if dps.dequeueGapPovBlocksFromDb(dps.lastGapHeight) {
						dps.lastGapHeight++
					} else {
						break
					}
				} else {
					break
				}
			}
		case dps.blockSyncState = <-dps.syncState:

			// notify processors
			dps.syncStateNotifyWait.Add(dps.processorNum)
			for _, p := range dps.processors {
				p.syncStateChange <- dps.blockSyncState
			}
			dps.syncStateNotifyWait.Wait()

			switch dps.blockSyncState {
			case common.Syncing:
				dps.updateLastProcessSyncTime()
				dps.acTrx.cleanFrontierVotes()
				dps.frontiersStatus = new(sync.Map)
				dps.totalVote = make(map[types.Address]types.Balance)
				dps.CleanSyncCache()
			case common.SyncDone:
			case common.SyncFinish:
				dps.CleanSyncCache()
				dps.logger.Warn("p2p concluded the sync")
			}

			dps.ebDone <- struct{}{}
		case <-timerUpdateSyncTime.C:
			select {
			case <-dps.updateSync:
				dps.lastProcessSyncTime = time.Now()
			default:
			}

			if dps.blockSyncState == common.Syncing || dps.blockSyncState == common.SyncDone {
				if time.Since(dps.lastProcessSyncTime) >= 2*time.Minute {
					dps.logger.Warnf("timeout concluded the sync. last[%s] now[%s]",
						dps.lastProcessSyncTime, time.Now())
					dps.blockSyncState = common.SyncFinish
					dps.syncFinish()
				}
			}
		}
	}
}

func (dps *DPoS) Stop() {
	dps.logger.Info("DPOS service stopped!")

	//do this first
	for k, v := range dps.handlerIds {
		_ = dps.eb.Unsubscribe(k, v)
	}

	dps.cancel()
	dps.processorStop()
	dps.acTrx.stop()
}

func (dps *DPoS) processSyncDone() {
	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Info("Stopped processSyncDone.")
			return
		case <-dps.syncDone:
			if err := dps.blockSyncDone(); err != nil {
				dps.logger.Error("block sync down err", err)
			}

			// notify processors
			dps.syncStateNotifyWait.Add(dps.processorNum)
			for _, p := range dps.processors {
				p.syncStateChange <- common.SyncFinish
			}
			dps.syncStateNotifyWait.Wait()

			dps.eb.Publish(common.EventConsensusSyncFinished)
			dps.logger.Warn("sync finished")
		}
	}
}

func (dps *DPoS) processBlocks() {
	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Info("Stop processBlocks.")
			return
		case bs := <-dps.recvBlocks:
			if dps.povSyncState == common.SyncDone || bs.BlockFrom == types.Synchronized || bs.Type == consensus.MsgConfirmAck {
				dps.dispatchMsg(bs)
			} else {
				if len(dps.cacheBlocks) < cap(dps.cacheBlocks) {
					dps.logger.Debugf("pov sync state[%s] cache blocks", dps.povSyncState)
					dps.cacheBlocks <- bs
				} else {
					dps.logger.Debugf("pov not ready! cache block too much, drop it!")
				}
			}
		case state := <-dps.povState:
			dps.povSyncState = state
			if state == common.SyncDone {
				close(dps.cacheBlocks)

				err := dps.eb.Unsubscribe(common.EventPovSyncState, dps.handlerIds[common.EventPovSyncState])
				if err != nil {
					dps.logger.Errorf("unsubscribe pov sync state err %s", err)
				}

				for bs := range dps.cacheBlocks {
					dps.logger.Infof("process cache block[%s]", bs.Block.GetHash())
					dps.dispatchMsg(bs)
				}
			}
		}
	}
}

func (dps *DPoS) processorStart() {
	for i := 0; i < dps.processorNum; i++ {
		dps.processors[i].start()
	}
}

func (dps *DPoS) processorStop() {
	for i := 0; i < dps.processorNum; i++ {
		dps.processors[i].stop()
	}
}

func (dps *DPoS) onGetFrontier(blocks types.StateBlockList) {
	var unconfirmed []*types.StateBlock

	for _, block := range blocks {
		if block.Token == common.ChainToken() {
			dps.totalVote[block.Address] = block.Balance.Add(block.Vote).Add(block.Oracle).Add(block.Network).Add(block.Storage)
			dps.logger.Infof("account[%s] vote weight[%s]", block.Address, dps.totalVote[block.Address])
		}
	}

	for _, block := range blocks {
		hash := block.GetHash()

		if dps.isReceivedFrontier(hash) {
			continue
		}

		has, _ := dps.ledger.HasStateBlockConfirmed(hash)
		if !has {
			dps.logger.Infof("get frontier %s need ack", hash)
			unconfirmed = append(unconfirmed, block)
			index := dps.getProcessorIndex(block.Address)
			dps.processors[index].frontiers <- block
		}
	}

	unconfirmedLen := len(unconfirmed)
	sendStart := 0
	sendEnd := 0
	sendLen := unconfirmedLen

	for sendLen > 0 {
		if sendLen >= blockNumPerReq {
			sendEnd = sendStart + blockNumPerReq
			sendLen -= blockNumPerReq
		} else {
			sendEnd = sendStart + sendLen
			sendLen = 0
		}

		dps.eb.Publish(common.EventBroadcast, p2p.ConfirmReq, unconfirmed[sendStart:sendEnd])
		sendStart = sendEnd
	}
}

func (dps *DPoS) onPovSyncState(state common.SyncState) {
	dps.povState <- state
	dps.logger.Infof("pov sync state to [%s]", state)
}

func (dps *DPoS) batchVoteStart() {
	timerVote := time.NewTicker(time.Second)

	for {
		select {
		case <-dps.ctx.Done():
			return
		case <-timerVote.C:
			dps.localRepAccount.Range(func(key, value interface{}) bool {
				address := key.(types.Address)
				dps.batchVoteDo(address, value.(*types.Account), ackTypeCommon)
				return true
			})
		}
	}
}

func (dps *DPoS) cacheAck(vi *voteInfo) {
	if dps.voteCache.Has(vi.hash) {
		v, err := dps.voteCache.Get(vi.hash)
		if err != nil {
			dps.logger.Error("get vote cache err")
			return
		}

		vc := v.(*sync.Map)
		vc.Store(vi.account, vi)
	} else {
		vc := new(sync.Map)
		vc.Store(vi.account, vi)
		err := dps.voteCache.Set(vi.hash, vc)
		if err != nil {
			dps.logger.Error("set vote cache err")
			return
		}
	}
}

func (dps *DPoS) pubAck(index int, ack *voteInfo) {
	dps.processors[index].acks <- ack
}

func (dps *DPoS) processSubMsg() {
	for {
		select {
		case <-dps.ctx.Done():
			return
		case msg := <-dps.subMsg:
			switch msg.kind {
			case subMsgKindSub:
				_ = dps.subAck.Set(msg.hash, msg.index)

				vote, err := dps.voteCache.Get(msg.hash)
				if err == nil {
					vc := vote.(*sync.Map)
					vc.Range(func(key, value interface{}) bool {
						dps.pubAck(msg.index, value.(*voteInfo))
						return true
					})
					dps.voteCache.Remove(msg.hash)
				}
			case subMsgKindUnsub:
				dps.subAck.Remove(msg.hash)
				dps.voteCache.Remove(msg.hash)
			}
		}
	}
}

func (dps *DPoS) getProcessorIndex(address types.Address) int {
	return int(address[len(address)-1]) % dps.processorNum
}

func (dps *DPoS) dispatchMsg(bs *consensus.BlockSource) {
	if bs.Type == consensus.MsgConfirmAck {
		ack := bs.Para.(*protos.ConfirmAckBlock)
		dps.saveOnlineRep(ack.Account)

		if dps.getAckType(ack.Sequence) != ackTypeFindRep {
			for _, h := range ack.Hash {
				dps.logger.Infof("dps recv confirmAck block[%s]", h)

				if has, _ := dps.ledger.HasStateBlockConfirmed(h); has {
					dps.heartAndVoteInc(h, ack.Account, onlineKindVote)
					continue
				}

				vi := &voteInfo{
					hash:    h,
					account: ack.Account,
				}

				val, err := dps.subAck.Get(vi.hash)
				if err == nil {
					dps.pubAck(val.(int), vi)
				} else {
					dps.cacheAck(vi)
				}
			}
		} else {
			dps.heartAndVoteInc(ack.Hash[0], ack.Account, onlineKindHeart)
		}
	} else {
		index := dps.getProcessorIndex(bs.Block.Address)

		if bs.Type == consensus.MsgSync {
			dps.processors[index].syncBlock <- bs.Block
		} else {
			dps.processors[index].blocks <- bs
		}
	}
}

func (dps *DPoS) subAckDo(index int, hash types.Hash) {
	msg := &subMsg{
		index: index,
		kind:  subMsgKindSub,
		hash:  hash,
	}
	dps.subMsg <- msg
}

func (dps *DPoS) unsubAckDo(hash types.Hash) {
	//msg := &subMsg{
	//	kind: subMsgKindUnsub,
	//	hash: hash,
	//}
	//dps.subMsg <- msg
}

func (dps *DPoS) dispatchAckedBlock(blk *types.StateBlock, hash types.Hash, localIndex int) {
	if localIndex == -1 {
		localIndex = dps.getProcessorIndex(blk.Address)
		dps.processors[localIndex].blocksAcked <- hash
	}

	switch blk.Type {
	case types.Send:
		index := dps.getProcessorIndex(types.Address(blk.Link))
		if localIndex != index {
			dps.processors[index].blocksAcked <- hash
		}
	case types.ContractSend: //beneficial maybe another account
		dstAddr := types.ZeroAddress

		switch types.Address(blk.GetLink()) {
		case types.MintageAddress:
			data := blk.GetData()
			if method, err := cabi.MintageABI.MethodById(data[0:4]); err == nil {
				if method.Name == cabi.MethodNameMintage {
					param := new(cabi.ParamMintage)
					if err = method.Inputs.Unpack(param, data[4:]); err == nil {
						dstAddr = param.Beneficial
					}
				} else if method.Name == cabi.MethodNameMintageWithdraw {
					tokenId := new(types.Hash)
					if err = method.Inputs.Unpack(tokenId, data[4:]); err == nil {
						ctx := vmstore.NewVMContext(dps.ledger)
						tokenInfoData, err := ctx.GetStorage(types.MintageAddress[:], tokenId[:])
						if err != nil {
							return
						}

						tokenInfo := new(types.TokenInfo)
						err = cabi.MintageABI.UnpackVariable(tokenInfo, cabi.VariableNameToken, tokenInfoData)
						if err == nil {
							dstAddr = tokenInfo.PledgeAddress
						}
					}
				}
			}
		case types.NEP5PledgeAddress:
			data := blk.GetData()
			if method, err := cabi.NEP5PledgeABI.MethodById(data[0:4]); err == nil {
				if method.Name == cabi.MethodNEP5Pledge {
					param := new(cabi.PledgeParam)
					if err = method.Inputs.Unpack(param, data[4:]); err == nil {
						dstAddr = param.Beneficial
					}
				} else if method.Name == cabi.MethodWithdrawNEP5Pledge {
					param := new(cabi.WithdrawPledgeParam)
					if err = method.Inputs.Unpack(param, data[4:]); err == nil {
						pledgeResult := cabi.SearchBeneficialPledgeInfoByTxId(vmstore.NewVMContext(dps.ledger), param)
						if pledgeResult != nil {
							dstAddr = pledgeResult.PledgeInfo.PledgeAddress
						}
					}
				}
			}
		case types.MinerAddress:
			param := new(cabi.MinerRewardParam)
			if err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, blk.GetData()); err == nil {
				dstAddr = param.Beneficial
			}
		case types.RepAddress:
			param := new(cabi.RepRewardParam)
			if err := cabi.RepABI.UnpackMethod(param, cabi.MethodNameRepReward, blk.GetData()); err == nil {
				dstAddr = param.Beneficial
			}
		case types.RewardsAddress:
			param := new(cabi.RewardsParam)
			data := blk.GetData()
			if method, err := cabi.RewardsABI.MethodById(data[0:4]); err == nil {
				if err = method.Inputs.Unpack(param, data[4:]); err == nil {
					dstAddr = param.Beneficial
				}
			}
		default:
			for _, p := range dps.processors {
				if localIndex != p.index {
					p.blocksAcked <- hash
				}
			}
			return
		}

		if dstAddr != types.ZeroAddress {
			index := dps.getProcessorIndex(dstAddr)
			if localIndex != index {
				dps.processors[index].blocksAcked <- hash
			}
		}
		//case types.ContractReward: //deal gap tokenInfo
		//	input, err := dps.ledger.GetStateBlock(blk.GetLink())
		//	if err != nil {
		//		dps.logger.Errorf("get block link error [%s]", hash)
		//		return
		//	}
		//
		//	if types.Address(input.GetLink()) == types.MintageAddress {
		//		param := new(cabi.ParamMintage)
		//		if err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameMintage, input.GetData()); err == nil {
		//			index := dps.getProcessorIndex(input.Address)
		//			if localIndex != index {
		//				dps.processors[index].blocksAcked <- param.TokenId
		//			}
		//		}
		//	}
	}
}

func (dps *DPoS) deleteBlockCache(block *types.StateBlock) {
	hash := block.GetHash()
	if exist, _ := dps.ledger.HasBlockCache(hash); exist {
		err := dps.ledger.DeleteBlockCache(hash)
		if err != nil {
			dps.logger.Error(err)
		} else {
			dps.logger.Debugf("delete block cache [%s] error", hash.String())
		}
		if exit, _ := dps.ledger.HasAccountMetaCache(block.Address); exit {
			err := dps.ledger.DeleteAccountMetaCache(block.Address)
			if err != nil {
				dps.logger.Error(err)
			} else {
				dps.logger.Debugf("delete block account meta cache [%s] error", hash.String())
			}
		}
	}
}

func (dps *DPoS) ProcessMsg(bs *consensus.BlockSource) {
	dps.recvBlocks <- bs
}

func (dps *DPoS) localRepVote(block *types.StateBlock) {
	hash := block.GetHash()

	dps.localRepAccount.Range(func(key, value interface{}) bool {
		address := key.(types.Address)

		weight := dps.ledger.Weight(address)
		if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
			return true
		}

		if block.Type == types.Online {
			has, _ := dps.ledger.HasStateBlockConfirmed(hash)

			if !has && !dps.isOnline(block.Address) {
				dps.logger.Debugf("block[%s] is not online", hash)
				return false
			}
		}

		dps.logger.Debugf("rep [%s] vote for block[%s] type[%s]", address, hash, block.Type)

		vi := &voteInfo{
			hash:    hash,
			account: address,
		}
		dps.acTrx.vote(vi)
		dps.acTrx.setVoteHash(block)
		dps.batchVote <- hash
		return true
	})
}

func (dps *DPoS) hasLocalValidRep() bool {
	has := false
	dps.localRepAccount.Range(func(key, value interface{}) bool {
		address := key.(types.Address)
		weight := dps.ledger.Weight(address)
		if weight.Compare(dps.minVoteWeight) != types.BalanceCompSmaller {
			has = true
		}
		return true
	})
	return has
}

func (dps *DPoS) voteGenerate(block *types.StateBlock, account types.Address, acc *types.Account) (*protos.ConfirmAckBlock, error) {
	//if dps.cfg.PoV.PovEnabled {
	//	povHeader, err := dps.ledger.GetLatestPovHeader()
	//	if povHeader == nil {
	//		dps.logger.Errorf("get pov header err %s", err)
	//		return nil, errors.New("get pov header err")
	//	}
	//
	//	if block.PoVHeight > povHeader.Height+povBlockNumDay || block.PoVHeight+povBlockNumDay < povHeader.Height {
	//		dps.logger.Errorf("pov height invalid height:%d cur:%d", block.PoVHeight, povHeader.Height)
	//		return nil, errors.New("pov height invalid")
	//	}
	//}
	//
	//hash := block.GetHash()
	//va := &protos.ConfirmAckBlock{
	//	Sequence:  0,
	//	Hash:      hash,
	//	Account:   account,
	//	Signature: acc.Sign(hash),
	//}
	//return va, nil
	return nil, nil
}

func (dps *DPoS) voteGenerateWithSeq(block *types.StateBlock, account types.Address, acc *types.Account, kind uint32) (*protos.ConfirmAckBlock, error) {
	hashes := make([]types.Hash, 0)
	hash := block.GetHash()
	hashes = append(hashes, hash)

	hashBytes := make([]byte, 0)
	for _, h := range hashes {
		hashBytes = append(hashBytes, h[:]...)
	}
	signHash, _ := types.HashBytes(hashBytes)

	va := &protos.ConfirmAckBlock{
		Sequence:  dps.getSeq(kind),
		Hash:      hashes,
		Account:   account,
		Signature: acc.Sign(signHash),
	}
	return va, nil
}

func (dps *DPoS) batchVoteDo(account types.Address, acc *types.Account, kind uint32) {
	timerWait := time.NewTimer(10 * time.Millisecond)
	hashes := make([]types.Hash, 0)
	hashBytes := make([]byte, 0)
	total := 0

out:
	for {
		timerWait.Reset(10 * time.Millisecond)

		select {
		case <-timerWait.C:
			timerWait.Stop()
			break out
		case h := <-dps.batchVote:
			hashes = append(hashes, h)
			hashBytes = append(hashBytes, h[:]...)
			total++

			if total == hashNumPerAck {
				hash, _ := types.HashBytes(hashBytes)
				va := &protos.ConfirmAckBlock{
					Sequence:  dps.getSeq(kind),
					Hash:      hashes,
					Account:   account,
					Signature: acc.Sign(hash),
				}

				dps.eb.Publish(common.EventBroadcast, p2p.ConfirmAck, va)

				total = 0
				hashes = make([]types.Hash, 0)
				hashBytes = hashBytes[0:0]
			}
		}
	}

	if len(hashes) == 0 {
		return
	}

	hash, _ := types.HashBytes(hashBytes)
	va := &protos.ConfirmAckBlock{
		Sequence:  dps.getSeq(kind),
		Hash:      hashes,
		Account:   account,
		Signature: acc.Sign(hash),
	}

	dps.eb.Publish(common.EventBroadcast, p2p.ConfirmAck, va)
}

func (dps *DPoS) refreshAccount() {
	var b bool
	var addr types.Address

	for _, v := range dps.accounts {
		addr = v.Address()
		b = dps.isRepresentation(addr)
		if b {
			dps.localRepAccount.Store(addr, v)
			dps.saveOnlineRep(addr)
		}
	}

	var count uint32
	dps.localRepAccount.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	dps.logger.Infof("there is %d local reps", count)
	if count > 0 {
		dps.eb.Publish(common.EventRepresentativeNode, true)

		if count > 1 {
			dps.logger.Error("it is very dangerous to run two or more representatives on one node")
		}
	} else {
		dps.eb.Publish(common.EventRepresentativeNode, false)
	}
}

func (dps *DPoS) isRepresentation(address types.Address) bool {
	if _, err := dps.ledger.GetRepresentation(address); err != nil {
		return false
	}
	return true
}

func (dps *DPoS) saveOnlineRep(addr types.Address) {
	now := time.Now().Add(repTimeout).Unix()
	dps.onlineReps.Store(addr, now)
}

func (dps *DPoS) getOnlineRepresentatives() []types.Address {
	var repAddresses []types.Address

	dps.onlineReps.Range(func(key, value interface{}) bool {
		addr := key.(types.Address)
		repAddresses = append(repAddresses, addr)
		return true
	})

	return repAddresses
}

func (dps *DPoS) findOnlineRepresentatives() error {
	blk, err := dps.ledger.GetRandomStateBlock()
	if err != nil {
		return err
	}

	dps.localRepAccount.Range(func(key, value interface{}) bool {
		address := key.(types.Address)

		if dps.isRepresentation(address) {
			va, err := dps.voteGenerateWithSeq(blk, address, value.(*types.Account), ackTypeFindRep)
			if err != nil {
				return true
			}
			dps.eb.Publish(common.EventBroadcast, p2p.ConfirmAck, va)
			dps.heartAndVoteInc(va.Hash[0], va.Account, onlineKindHeart)
		}

		return true
	})

	return nil
}

func (dps *DPoS) cleanOnlineReps() {
	var repAddresses []*types.Address
	now := time.Now().Unix()

	dps.onlineReps.Range(func(key, value interface{}) bool {
		addr := key.(types.Address)
		v := value.(int64)
		if v < now {
			dps.onlineReps.Delete(addr)
		} else {
			repAddresses = append(repAddresses, &addr)
		}
		return true
	})

	_ = dps.ledger.SetOnlineRepresentations(repAddresses)
}

func (dps *DPoS) calculateAckHash(va *protos.ConfirmAckBlock) (types.Hash, error) {
	data, err := protos.ConfirmAckBlockToProto(va)
	if err != nil {
		return types.ZeroHash, err
	}

	version := dps.cfg.Version
	message := p2p.NewQlcMessage(data, byte(version), p2p.ConfirmAck)
	hash, err := types.HashBytes(message)
	if err != nil {
		return types.ZeroHash, err
	}

	return hash, nil
}

func (dps *DPoS) onRollback(hash types.Hash) {
	dps.rollbackUncheckedFromDb(hash)

	blk, err := dps.ledger.GetStateBlockConfirmed(hash)
	if err == nil && blk.Type == types.Online {
		for _, acc := range dps.accounts {
			if acc.Address() == blk.Address {
				dps.sendOnlineWithAccount(acc, blk.PoVHeight)
				break
			}
		}
	}
}

func (dps *DPoS) rollbackUncheckedFromDb(hash types.Hash) {
	if blkLink, _, _ := dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindLink); blkLink != nil {
		err := dps.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindLink)
		if err != nil {
			dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindLink", err, blkLink.GetHash())
		}
	}

	if blkPrevious, _, _ := dps.ledger.GetUncheckedBlock(hash, types.UncheckedKindPrevious); blkPrevious != nil {
		err := dps.ledger.DeleteUncheckedBlock(hash, types.UncheckedKindPrevious)
		if err != nil {
			dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindPrevious", err, blkPrevious.GetHash())
		}
	}

	// gap token
	blk, err := dps.ledger.GetStateBlock(hash)
	if err != nil {
		dps.logger.Errorf("get block error [%s]", hash)
		return
	}

	if blk.GetType() == types.ContractReward {
		input, err := dps.ledger.GetStateBlock(blk.GetLink())
		if err != nil {
			dps.logger.Errorf("dequeue get block link error [%s]", hash)
			return
		}

		address := types.Address(input.GetLink())
		if address == types.MintageAddress {
			var param = new(cabi.ParamMintage)
			if err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameMintage, input.GetData()); err == nil {
				if blkToken, _, _ := dps.ledger.GetUncheckedBlock(param.TokenId, types.UncheckedKindTokenInfo); blkToken != nil {
					err := dps.ledger.DeleteUncheckedBlock(param.TokenId, types.UncheckedKindTokenInfo)
					if err != nil {
						dps.logger.Errorf("Get err [%s] for hash: [%s] when delete UncheckedKindTokenInfo", err, blkToken.GetHash())
					}
				}
			}
		}
	}
}

func (dps *DPoS) getSeq(kind uint32) uint32 {
	var seq, ackSeq uint32

	seq = atomic.AddUint32(&dps.ackSeq, 1)
	ackSeq = (kind << 28) | ((seq - 1) & 0xFFFFFFF)

	return ackSeq
}

func (dps *DPoS) getAckType(seq uint32) uint32 {
	return seq >> 28
}

func (dps *DPoS) onRpcSyncCall(name string, in interface{}, out interface{}) {
	switch name {
	case "DPoS.Online":
		dps.onGetOnlineInfo(in, out)
	}
}

func (dps *DPoS) isWaitingFrontier(hash types.Hash) (bool, frontierStatus) {
	if status, ok := dps.frontiersStatus.Load(hash); ok {
		return true, status.(frontierStatus)
	}
	return false, frontierInvalid
}

func (dps *DPoS) isConfirmedFrontier(hash types.Hash) bool {
	if status, ok := dps.frontiersStatus.Load(hash); ok && status == frontierConfirmed {
		if has, _ := dps.ledger.HasUnconfirmedSyncBlock(hash); has {
			return true
		}
	}
	return false
}

func (dps *DPoS) isReceivedFrontier(hash types.Hash) bool {
	if _, ok := dps.frontiersStatus.Load(hash); ok {
		return true
	}
	return false
}

func (dps *DPoS) chainFinished(hash types.Hash) {
	if _, ok := dps.frontiersStatus.Load(hash); ok {
		dps.frontiersStatus.Store(hash, frontierChainFinished)
		dps.checkFinish <- struct{}{}
	}
}

func (dps *DPoS) blockSyncDone() error {
	dps.logger.Warn("begin: process sync cache blocks")
	dps.eb.Publish(common.EventAddSyncBlocks, &types.StateBlock{}, true)

	scs := make([]*sortContract, 0)
	if err := dps.ledger.GetSyncCacheBlocks(func(block *types.StateBlock) error {
		if b, err := dps.isRelatedOrderBlock(block); err != nil {
			dps.logger.Infof("block[%s] type check error, %s", block.GetHash(), err)
		} else {
			if b {
				sortC := &sortContract{
					hash:      block.GetHash(),
					timestamp: block.Timestamp,
				}
				scs = append(scs, sortC)
			} else {
				index := dps.getProcessorIndex(block.Address)
				dps.processors[index].doneBlock <- block
			}
		}
		return nil
	}); err != nil {
		return err
	}

	dps.logger.Info("sync done, common block process finished, order blocks:  ", len(scs))
	if len(scs) > 0 {
		sort.Slice(scs, func(i, j int) bool {
			return scs[i].timestamp < scs[j].timestamp
		})

		for _, sc := range scs {
			dps.updateLastProcessSyncTime()

			blk, err := dps.ledger.GetSyncCacheBlock(sc.hash)
			if err != nil {
				dps.logger.Errorf("get sync cache block[%s] err[%s]", sc.hash, err)
				continue
			}

			if err := dps.lv.BlockSyncDoneProcess(blk); err != nil {
				dps.logger.Warn("contract block(%s) sync done error: %s", blk.GetHash(), err)
			}
		}
	}

	var allDone bool
	timerCheckProcessor := time.NewTicker(time.Second)
	defer timerCheckProcessor.Stop()

	for range timerCheckProcessor.C {
		allDone = true
		for _, p := range dps.processors {
			if len(p.doneBlock) > 0 {
				allDone = false
				break
			}
		}

		if allDone {
			dps.logger.Warn("end: process sync cache blocks")
			return nil
		}
	}

	return nil
}

func (dps *DPoS) isRelatedOrderBlock(block *types.StateBlock) (bool, error) {
	switch block.GetType() {
	case types.ContractReward:
		sendblk, err := dps.ledger.GetStateBlockConfirmed(block.GetLink())
		if err != nil {
			return false, err
		}
		switch types.Address(sendblk.GetLink()) {
		case types.NEP5PledgeAddress:
			return true, nil
		case types.MintageAddress:
			return true, nil
		}
	}
	return false, nil
}

func (dps *DPoS) syncFinish() {
	dps.logger.Warn("process state blocks finished")

	select {
	case dps.syncDone <- struct{}{}:
	default:
	}
}

func (dps *DPoS) checkSyncFinished() {
	allFinished := true
	dps.frontiersStatus.Range(func(k, v interface{}) bool {
		status := v.(frontierStatus)
		if status != frontierChainFinished {
			allFinished = false
			return false
		}
		return true
	})

	if allFinished {
		checkLedgerTimer := time.NewTicker(1 * time.Second)
		defer checkLedgerTimer.Stop()
		checkMaxTimes := 5
		checkResult := true

	checkOut:
		for {
			select {
			case <-checkLedgerTimer.C:
				if checkMaxTimes <= 0 {
					dps.logger.Errorf("check frontier block confirmed err")
					return
				}
				checkMaxTimes--

				dps.frontiersStatus.Range(func(k, v interface{}) bool {
					hash := k.(types.Hash)
					if has, _ := dps.ledger.HasStateBlockConfirmed(hash); !has {
						checkResult = false
						return false
					}
					return true
				})

				if checkResult {
					break checkOut
				}
			}
		}

		dps.syncFinish()
	}
}

func (dps *DPoS) syncBlockRollback(hash types.Hash) {
	block, err := dps.ledger.GetUnconfirmedSyncBlock(hash)
	if err != nil {
		return
	}

	_ = dps.ledger.DeleteUnconfirmedSyncBlock(hash)
	dps.syncBlockRollback(block.Previous)
}

func (dps *DPoS) onFrontierConfirmed(hash types.Hash, result *bool) {
	if status, ok := dps.frontiersStatus.Load(hash); ok {
		if status == frontierChainConfirmed || status == frontierChainFinished {
			*result = true
		} else {
			*result = false
		}
	} else {
		// filter confirmed blocks
		if has, _ := dps.ledger.HasStateBlockConfirmed(hash); has {
			*result = true
		} else {
			*result = false
		}
	}
}

func (dps *DPoS) CleanSyncCache() {
	dps.ledger.CleanSyncCache()
}

func (dps *DPoS) onSyncStateChange(state common.SyncState) {
	dps.syncState <- state
	<-dps.ebDone
}

func (dps *DPoS) dequeueGapPovBlocksFromDb(height uint64) bool {
	if height < common.PovMinerRewardHeightGapToLatest {
		return true
	}

	if blocks, kind, err := dps.ledger.GetGapPovBlock(height); err != nil {
		return true
	} else {
		if len(blocks) == 0 {
			return true
		}

		dayIndex := uint32((height - common.PovMinerRewardHeightGapToLatest) / uint64(common.POVChainBlocksPerDay))
		if !dps.ledger.HasPovMinerStat(dayIndex) {
			dps.logger.Infof("miner stat [%d] not exist", dayIndex)
			return false
		}

		for i, b := range blocks {
			bs := &consensus.BlockSource{
				Block:     b,
				BlockFrom: kind[i],
			}

			dps.logger.Infof("dequeue gap pov[%s][%s]", b.GetHash(), kind[i])
			index := dps.getProcessorIndex(b.GetAddress())
			dps.processors[index].blocks <- bs
		}

		if err := dps.ledger.DeleteGapPovBlock(height); err != nil {
			dps.logger.Errorf("del gap pov block err", err)
		}

		return true
	}
}

func (dps *DPoS) updateLastProcessSyncTime() {
	select {
	case dps.updateSync <- struct{}{}:
	default:
	}
}
