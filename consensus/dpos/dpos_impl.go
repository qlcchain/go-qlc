package dpos

import (
	"context"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/vm/contract"

	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/bluele/gcache"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	repTimeout            = 5 * time.Minute
	voteCacheSize         = 102400
	refreshPriInterval    = 1 * time.Minute
	subAckMaxSize         = 1024000
	maxStatisticsPeriod   = 3
	confirmedCacheMaxLen  = 1024000
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

type processorStat struct {
	Index      int `json:"1-index"`
	BlockQueue int `json:"2-blockQueue"`
	SyncQueue  int `json:"3-syncQueue"`
	AckQueue   int `json:"4-ackQueue"`
	DoneQueue  int `json:"5-doneQueue"`
	ChainQueue int `json:"6-chainQueue"`
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
	subscriber          *event.ActorSubscriber
	lv                  *process.LedgerVerifier
	cacheBlocks         chan *consensus.BlockSource
	recvBlocks          chan *consensus.BlockSource
	povState            chan topic.SyncState
	syncState           chan topic.SyncState
	processors          []*Processor
	processorNum        int
	localRepAccount     sync.Map
	povSyncState        topic.SyncState
	blockSyncState      topic.SyncState
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
	confirmedBlocks     *cache
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
	tps                 [10]uint32
	block2Ledger        chan struct{}
	pf                  *perfInfo
	lockPool            *sync.Map
	feb                 *event.FeedEventBus
	febRpcMsgCh         chan *topic.EventRPCSyncCallMsg
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
		lv:                  process.NewLedgerVerifier(l),
		cacheBlocks:         make(chan *consensus.BlockSource, common.DPoSMaxCacheBlocks),
		recvBlocks:          make(chan *consensus.BlockSource, common.DPoSMaxBlocks),
		povState:            make(chan topic.SyncState, 1),
		syncState:           make(chan topic.SyncState, 1),
		processorNum:        processorNum,
		processors:          newProcessors(processorNum),
		subMsg:              make(chan *subMsg, 10240),
		subAck:              gcache.New(subAckMaxSize).Expiration(waitingVoteMaxTime * time.Second).LRU().Build(),
		hash2el:             new(sync.Map),
		batchVote:           make(chan types.Hash, 102400),
		ctx:                 ctx,
		cancel:              cancel,
		frontiersStatus:     new(sync.Map),
		syncStateNotifyWait: new(sync.WaitGroup),
		online:              gcache.New(maxStatisticsPeriod).LRU().Build(),
		confirmedBlocks:     newCache(confirmedCacheMaxLen, confirmedCacheMaxTime),
		lastSendHeight:      1,
		curPovHeight:        0,
		checkFinish:         make(chan struct{}, 10240),
		gapPovCh:            make(chan *consensus.BlockSource, 10240),
		povChange:           make(chan *types.PovBlock, 10240),
		voteCache:           gcache.New(voteCacheSize).LRU().Build(),
		blockSyncState:      topic.SyncNotStart,
		isFindingRep:        0,
		ebDone:              make(chan struct{}, 1),
		updateSync:          make(chan struct{}, 1),
		block2Ledger:        make(chan struct{}, 409600),
		pf:                  new(perfInfo),
		lockPool:            new(sync.Map),
		feb:                 cc.FeedEventBus(),
		febRpcMsgCh:         make(chan *topic.EventRPCSyncCallMsg, 1),
	}

	dps.pf.status.Store(perfTypeClose)

	dps.acTrx.setDPoSService(dps)
	for _, p := range dps.processors {
		p.setDPoSService(dps)
	}

	dps.lastGapHeight = dps.ledger.GetLastGapPovHeight()
	pb, err := dps.ledger.GetLatestPovBlock()
	if err == nil {
		dps.curPovHeight = pb.Header.BasHdr.Height
	}

	dps.confirmedBlocks.evictedFunc = func(key interface{}, val interface{}) {
		hash := key.(types.Hash)
		err = dps.ledger.CleanBlockVoteHistory(hash)
		if err != nil {
			dps.logger.Error("clean vote history err", err)
		}
	}

	return dps
}

func (dps *DPoS) Init() {
	// dfile, _ = os.OpenFile("./dfile", os.O_RDWR|os.O_CREATE, 0666)
	supply := common.GenesisBlock().Balance
	dps.minVoteWeight, _ = supply.Div(common.DposVoteDivisor)
	dps.voteThreshold, _ = supply.Div(2)

	subscriber := event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case types.Hash:
			dps.onRollback(msg)
		case *types.PovBlock:
			dps.onPovHeightChange(msg)
		case *types.Tuple:
			dps.onFrontierConfirmed(msg.First.(types.Hash), msg.Second.(*bool))
		}
	}), dps.eb)

	if err := subscriber.Subscribe(topic.EventRollback, topic.EventPovConnectBestBlock); err != nil {
		dps.logger.Errorf("failed to subscribe event %s", err)
	} else {
		dps.subscriber = subscriber
	}

	if dps.cfg.PoV.PovEnabled {
		dps.povSyncState = topic.SyncNotStart

		if err := subscriber.SubscribeOne(topic.EventPovSyncState, event.Spawn(func(c actor.Context) {
			switch msg := c.Message().(type) {
			case topic.SyncState:
				dps.onPovSyncState(msg)
			}
		})); err != nil {
			dps.logger.Errorf("subscribe pov sync state event err")
		}
	} else {
		dps.povSyncState = topic.SyncDone
	}
	if len(dps.accounts) != 0 {
		dps.refreshAccount()
	} else {
		dps.eb.Publish(topic.EventRepresentativeNode, false)
	}
}

func (dps *DPoS) Start() {
	dps.logger.Info("DPoS service started!")

	go dps.acTrx.start()
	go dps.batchVoteStart()
	go dps.processSubMsg()
	go dps.processBlocks()
	go dps.stat()
	dps.processorStart()

	// if err := dps.blockSyncDone(); err != nil {
	// 	dps.logger.Error("block sync down err", err)
	// }

	if err := dps.ledger.CleanAllVoteHistory(); err != nil {
		dps.logger.Error("clean all vote history err")
	}

	timerRefreshPri := time.NewTicker(refreshPriInterval)
	timerDequeueGap := time.NewTicker(10 * time.Second)
	timerUpdateSyncTime := time.NewTicker(5 * time.Second)
	timerGC := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Info("Stopped DPoS.")
			return
		case <-timerRefreshPri.C:
			dps.logger.Info("refresh pri info.")
			go dps.refreshAccount()
		case <-dps.checkFinish:
			dps.checkSyncFinished()
		case pb := <-dps.povChange:
			dps.logger.Infof("pov height changed [%d]->[%d]", dps.curPovHeight, pb.Header.BasHdr.Height)
			dps.curPovHeight = pb.Header.BasHdr.Height

			if dps.povSyncState == topic.SyncDone {
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
			} else {
				// gap pov blocks can not be confirmed, may result in p2p pulling blocks repeatedly
				dps.updateLastProcessSyncTime()
			}
		case <-timerDequeueGap.C:
			for {
				if dps.lastGapHeight <= dps.curPovHeight {
					if dps.dequeueGapPovBlocksFromDb(dps.lastGapHeight) {
						err := dps.ledger.SetLastGapPovHeight(dps.lastGapHeight)
						if err != nil {
							dps.logger.Error(err)
						}

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
			case topic.Syncing:
				dps.updateLastProcessSyncTime()
				dps.totalVote = make(map[types.Address]types.Balance)
				dps.frontiersStatus = new(sync.Map)
			case topic.SyncDone:
			case topic.SyncFinish:
				dps.acTrx.cleanFrontierVotes()
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

			if dps.blockSyncState == topic.Syncing || dps.blockSyncState == topic.SyncDone {
				if time.Since(dps.lastProcessSyncTime) >= 2*time.Minute {
					dps.logger.Warnf("sync timeout. last[%s] now[%s]", dps.lastProcessSyncTime, time.Now())
					dps.syncFinish()
				}
			}
		case <-timerGC.C:
			dps.confirmedBlocks.gc()
		}
	}
}

func (dps *DPoS) Stop() {
	dps.logger.Info("DPoS service stopped!")

	// do this first
	if err := dps.subscriber.UnsubscribeAll(); err != nil {
		dps.logger.Error(err)
	}
	dps.cancel()
	dps.processorStop()
	dps.acTrx.stop()
}

func (dps *DPoS) stat() {
	timerStatInterval := time.NewTicker(15 * time.Second)

	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Info("Stop stat.")
			return
		case <-timerStatInterval.C:
			dps.tps[0] /= 15
			for i := len(dps.tps) - 1; i > 0; i-- {
				dps.tps[i] = dps.tps[i-1]
			}
			dps.tps[0] = 0
		case <-dps.block2Ledger:
			dps.tps[0]++
		}
	}
}

func (dps *DPoS) RPC(kind uint, in, out interface{}) {
	switch kind {
	case common.RpcDPoSOnlineInfo:
		dps.onGetOnlineInfo(in, out)
	case common.RpcDPoSConsInfo:
		dps.info(in, out)
	case common.RpcDPoSSetConsPerf:
		dps.setPerf(in, out)
	case common.RpcDPoSGetConsPerf:
		dps.getPerf(in, out)
	case common.RpcDPoSProcessFrontier:
		blocks := in.(types.StateBlockList)
		dps.onGetFrontier(blocks)
	case common.RpcDPoSOnSyncStateChange:
		state := in.(topic.SyncState)
		dps.onSyncStateChange(state)
	case common.RpcDPoSFeed:
		go dps.feedBlocks()
	}
}

func (dps *DPoS) statBlockInc() {
	select {
	case dps.block2Ledger <- struct{}{}:
	default:
	}
}

func (dps *DPoS) processBlocks() {
	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Info("Stop processBlocks.")
			return
		case bs := <-dps.recvBlocks:
			if dps.povSyncState == topic.SyncDone || bs.BlockFrom == types.Synchronized || bs.Type == consensus.MsgConfirmAck {
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
			if state == topic.SyncDone {
				close(dps.cacheBlocks)

				err := dps.subscriber.Unsubscribe(topic.EventPovSyncState)
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
			if _, ok := dps.totalVote[block.Representative]; ok {
				dps.totalVote[block.Representative] = dps.totalVote[block.Representative].Add(block.GetBalance()).Add(block.GetVote()).Add(block.GetOracle()).Add(block.GetNetwork()).Add(block.GetStorage())
			} else {
				dps.totalVote[block.Representative] = block.GetBalance().Add(block.GetVote()).Add(block.GetOracle()).Add(block.GetNetwork()).Add(block.GetStorage())
			}
		}
	}

	for addr, vote := range dps.totalVote {
		dps.logger.Infof("account[%s] vote weight[%s]", addr, vote)
	}

	for _, block := range blocks {
		hash := block.GetHash()

		if dps.isReceivedFrontier(hash) {
			continue
		}

		has, _ := dps.ledger.HasStateBlockConfirmed(hash)
		if !has {
			dps.logger.Infof("get frontier [%s-%s] need ack", block.Address, hash)
			dps.frontiersStatus.Store(hash, frontierWaitingForVote)
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

		dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.ConfirmReq, Message: unconfirmed[sendStart:sendEnd]})
		sendStart = sendEnd
	}
}

func (dps *DPoS) onPovSyncState(state topic.SyncState) {
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
				dps.heartAndVoteInc(h, ack.Account, onlineKindVote)

				if has, _ := dps.ledger.HasStateBlockConfirmed(h); has {
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
	// msg := &subMsg{
	// 	kind: subMsgKindUnsub,
	// 	hash: hash,
	// }
	// dps.subMsg <- msg
}

func (dps *DPoS) dispatchAckedBlock(blk *types.StateBlock, hash types.Hash, localIndex int) {
	if localIndex == -1 {
		localIndex = dps.getProcessorIndex(blk.Address)
		dps.processors[localIndex].ackedBlockNotify(hash)
	}

	switch blk.Type {
	case types.Send:
		index := dps.getProcessorIndex(types.Address(blk.Link))
		if localIndex != index {
			dps.processors[index].ackedBlockNotify(hash)
		}
	case types.ContractSend: // beneficial maybe another account
		dstAddr := types.ZeroAddress

		if c, ok, err := contract.GetChainContract(types.Address(blk.GetLink()), blk.GetData()); ok && err == nil {
			ctx := vmstore.NewVMContext(dps.ledger)
			dstAddr, err = c.GetTargetReceiver(ctx, blk)
			if err != nil {
				dps.logger.Error(err)
			}
		}

		if dstAddr != types.ZeroAddress {
			index := dps.getProcessorIndex(dstAddr)
			if localIndex != index {
				dps.processors[index].ackedBlockNotify(hash)
			}
		}

		if types.Address(blk.GetLink()) == types.PubKeyDistributionAddress {
			method, err := cabi.PublicKeyDistributionABI.MethodById(blk.Data)
			if err == nil {
				if method.Name == cabi.MethodNamePKDPublish {
					for _, p := range dps.processors {
						if localIndex != p.index {
							p.publishBlockNotify(hash)
						}
					}
				}
			} else {
				dps.logger.Errorf("get contract method err")
			}
		}
	case types.ContractReward: // deal gap tokenInfo
		if blk.IsOpen() {
			input, err := dps.ledger.GetStateBlockConfirmed(blk.GetLink())
			if err != nil {
				dps.logger.Errorf("get block link error [%s]", hash)
				return
			}

			if types.Address(input.GetLink()) == types.MintageAddress {
				param := new(cabi.ParamMintage)
				if err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameMintage, input.GetData()); err == nil {
					index := dps.getProcessorIndex(input.Address)
					if localIndex != index {
						dps.processors[index].tokenCreateNotify(hash)
					}
				}
			}
		}
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
	dps.localRepAccount.Range(func(key, value interface{}) bool {
		hash := block.GetHash()
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
	hashes := make([]types.Hash, 0)
	hashBytes := make([]byte, 0)
	total := 0

out:
	for {
		select {
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

				dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.ConfirmAck, Message: va})

				total = 0
				hashes = make([]types.Hash, 0)
				hashBytes = hashBytes[0:0]
			}
		default:
			break out
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

	dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.ConfirmAck, Message: va})
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
		dps.eb.Publish(topic.EventRepresentativeNode, true)
	} else {
		dps.eb.Publish(topic.EventRepresentativeNode, false)
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

func (dps *DPoS) findOnlineRepresentatives() error {
	blk := common.GenesisBlock()
	dps.localRepAccount.Range(func(key, value interface{}) bool {
		address := key.(types.Address)

		if dps.isRepresentation(address) {
			va, err := dps.voteGenerateWithSeq(&blk, address, value.(*types.Account), ackTypeFindRep)
			if err != nil {
				return true
			}
			dps.eb.Publish(topic.EventBroadcast, &p2p.EventBroadcastMsg{Type: p2p.ConfirmAck, Message: va})
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

func (dps *DPoS) onRollback(hash types.Hash) {
	// dps.rollbackUncheckedFromDb(hash)

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

func (dps *DPoS) getSeq(kind uint32) uint32 {
	var seq, ackSeq uint32

	seq = atomic.AddUint32(&dps.ackSeq, 1)
	ackSeq = (kind << 28) | ((seq - 1) & 0xFFFFFFF)

	return ackSeq
}

func (dps *DPoS) getAckType(seq uint32) uint32 {
	return seq >> 28
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

// func (dps *DPoS) blockSyncDone() error {
// 	dps.logger.Warn("begin: process sync cache blocks")
// 	dps.eb.Publish(topic.EventAddSyncBlocks, types.NewTuple(&types.StateBlock{}, true))
//
// 	scs := make([]*sortContract, 0)
// 	if err := dps.ledger.GetSyncCacheBlocks(func(block *types.StateBlock) error {
// 		if b, err := dps.isRelatedOrderBlock(block); err != nil {
// 			dps.logger.Infof("block[%s] type check error, %s", block.GetHash(), err)
// 		} else {
// 			if b {
// 				sortC := &sortContract{
// 					hash:      block.GetHash(),
// 					timestamp: block.Timestamp,
// 				}
// 				scs = append(scs, sortC)
// 			} else {
// 				index := dps.getProcessorIndex(block.Address)
// 				dps.processors[index].doneBlock <- block
// 			}
// 		}
// 		return nil
// 	}); err != nil {
// 		return err
// 	}
//
// 	dps.logger.Info("sync done, common block process finished, order blocks:  ", len(scs))
// 	if len(scs) > 0 {
// 		sort.Slice(scs, func(i, j int) bool {
// 			return scs[i].timestamp < scs[j].timestamp
// 		})
//
// 		for _, sc := range scs {
// 			dps.updateLastProcessSyncTime()
//
// 			blk, err := dps.ledger.GetSyncCacheBlock(sc.hash)
// 			if err != nil {
// 				dps.logger.Errorf("get sync cache block[%s] err[%s]", sc.hash, err)
// 				continue
// 			}
//
// 			if err := dps.lv.BlockSyncDoneProcess(blk); err != nil {
// 				dps.logger.Warnf("contract block(%s) sync done error: %s", blk.GetHash(), err)
// 			}
// 		}
// 	}
//
// 	var allDone bool
// 	timerCheckProcessor := time.NewTicker(time.Second)
// 	defer timerCheckProcessor.Stop()
//
// 	for range timerCheckProcessor.C {
// 		allDone = true
// 		for _, p := range dps.processors {
// 			if len(p.doneBlock) > 0 {
// 				allDone = false
// 				break
// 			}
// 		}
//
// 		if allDone {
// 			dps.logger.Warn("end: process sync cache blocks")
// 			return nil
// 		}
// 	}
//
// 	return nil
// }

// func (dps *DPoS) isRelatedOrderBlock(block *types.StateBlock) (bool, error) {
// 	switch block.GetType() {
// 	case types.ContractReward:
// 		sendblk, err := dps.ledger.GetStateBlockConfirmed(block.GetLink())
// 		if err != nil {
// 			return false, err
// 		}
// 		switch types.Address(sendblk.GetLink()) {
// 		case types.NEP5PledgeAddress:
// 			return true, nil
// 		case types.MintageAddress:
// 			return true, nil
// 		case types.BlackHoleAddress:
// 			return true, nil
// 		}
// 	case types.ContractSend:
// 		switch types.Address(block.GetLink()) {
// 		case types.BlackHoleAddress:
// 			return true, nil
// 		case types.MinerAddress, types.RepAddress:
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

func (dps *DPoS) syncFinish() {
	dps.blockSyncState = topic.SyncFinish
	dps.acTrx.cleanFrontierVotes()
	dps.CleanSyncCache()

	// notify processors
	dps.syncStateNotifyWait.Add(dps.processorNum)
	for _, p := range dps.processors {
		p.syncStateChange <- dps.blockSyncState
	}
	dps.syncStateNotifyWait.Wait()

	dps.eb.Publish(topic.EventConsensusSyncFinished, &topic.EventP2PSyncStateMsg{P2pSyncState: topic.SyncFinish})
	dps.logger.Warn("sync finished")
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
		dps.syncFinish()
	}
}

// func (dps *DPoS) syncBlockRollback(hash types.Hash) {
// 	block, err := dps.ledger.GetUnconfirmedSyncBlock(hash)
// 	if err != nil {
// 		return
// 	}
//
// 	_ = dps.ledger.DeleteUnconfirmedSyncBlock(hash)
// 	dps.syncBlockRollback(block.Previous)
// }

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

func (dps *DPoS) onSyncStateChange(state topic.SyncState) {
	dps.syncState <- state
	<-dps.ebDone
}

func (dps *DPoS) dequeueGapPovBlocksFromDb(height uint64) bool {
	if height < common.PovMinerRewardHeightGapToLatest {
		return true
	}

	dayIndex := uint32((height - common.PovMinerRewardHeightGapToLatest) / uint64(common.POVChainBlocksPerDay))
	if !dps.ledger.HasPovMinerStat(dayIndex) {
		dps.logger.Infof("miner stat [%d] not exist", dayIndex)
		return false
	}

	_ = dps.ledger.WalkGapPovBlocksWithHeight(height, func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		bs := &consensus.BlockSource{
			Block:     block,
			BlockFrom: sync,
		}

		hash := block.GetHash()
		index := dps.getProcessorIndex(block.GetAddress())
		dps.processors[index].blocks <- bs
		dps.logger.Infof("dequeue gap pov block[%s] height[%d]", hash, height)

		if err := dps.ledger.DeleteGapPovBlock(height, hash); err != nil {
			dps.logger.Errorf("del gap pov block err", err)
		}

		return nil
	})

	return true
}

func (dps *DPoS) updateLastProcessSyncTime() {
	select {
	case dps.updateSync <- struct{}{}:
	default:
	}
}

func (dps *DPoS) info(in interface{}, out interface{}) {
	outArgs := out.(map[string]interface{})
	rootsNum := 0

	outArgs["err"] = nil
	outArgs["tps"] = dps.tps
	outArgs["povSyncState"] = dps.povSyncState.String()
	outArgs["blockSyncState"] = dps.blockSyncState.String()
	outArgs["blockQueue"] = len(dps.recvBlocks)

	pStats := make([]*processorStat, 0)
	for _, p := range dps.processors {
		ps := &processorStat{
			Index:      p.index,
			BlockQueue: len(p.blocks),
			SyncQueue:  len(p.syncBlock),
			AckQueue:   len(p.acks),
		}

		for _, dealt := range p.confirmedChain {
			if !dealt {
				ps.ChainQueue++
			}
		}

		ps.ChainQueue += ConfirmChainParallelNum - int(p.confirmParallelNum)
		pStats = append(pStats, ps)
	}
	outArgs["processors"] = pStats

	dps.acTrx.roots.Range(func(key, value interface{}) bool {
		rootsNum++
		return true
	})
	outArgs["rootsNum"] = rootsNum
}

func (dps *DPoS) feedBlocks() {
	dps.logger.Warnf("...............feed blocks.............")
	l := ledger.NewLedger("dataFeed/qlc.json")
	blks := make(map[types.AddressToken][]*types.StateBlock)
	count := 0

	for {
		if dps.povSyncState == topic.SyncDone {
			break
		}
		time.Sleep(time.Second)
	}

	dps.logger.Warnf("...................generating blocks.................")

	frontiers, err := l.GetFrontiers()
	if err != nil {
		dps.logger.Error(err)
	}

	for _, f := range frontiers {
		header, err := l.GetStateBlockConfirmed(f.HeaderBlock)
		if err != nil {
			dps.logger.Error(err)
		}

		key := types.AddressToken{
			Address: header.Address,
			Token:   header.Token,
		}

		blocks := make([]*types.StateBlock, 0)
		block := header
		if header.Type == types.Send {
			for {
				blocks = append(blocks, block)
				count++

				block, err = l.GetStateBlockConfirmed(block.Previous)
				if err != nil || block.Type != types.Send {
					break
				}
			}

			blks[key] = blocks
		}
	}

	dps.logger.Warnf("..............total %d blocks, will begin the test in 5 seconds.............", count)
	time.Sleep(5 * time.Second)

	for _, b := range blks {
		for i := len(b) - 1; i >= 0; i-- {
			blk := b[i]
			bs := &consensus.BlockSource{
				Block:     blk,
				BlockFrom: types.UnSynchronized,
				Type:      consensus.MsgGenerateBlock,
			}
			dps.ProcessMsg(bs)
		}
	}
}
