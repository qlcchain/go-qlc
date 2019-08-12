package dpos

import (
	"errors"
	"github.com/bluele/gcache"
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
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	repTimeout            = 5 * time.Minute
	voteCacheSize         = 102400
	refreshPriInterval    = 1 * time.Minute
	findOnlineRepInterval = 2 * time.Minute
	povBlockNumDay        = 2880
	subAckMaxSize         = 102400
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

type DPoS struct {
	ledger          *ledger.Ledger
	acTrx           *ActiveTrx
	accounts        []*types.Account
	onlineReps      sync.Map
	logger          *zap.SugaredLogger
	cfg             *config.Config
	eb              event.EventBus
	lv              *process.LedgerVerifier
	cacheBlocks     chan *consensus.BlockSource
	povReady        chan bool
	processors      []*Processor
	processorNum    int
	quitCh          chan bool
	localRepAccount sync.Map
	povSyncState    atomic.Value
	minVoteWeight   types.Balance
	voteThreshold   types.Balance
	subAck          gcache.Cache
	subMsg          chan *subMsg
	subMsgQuit      chan bool
	voteCache       gcache.Cache //vote blocks
	ackSeq          uint32
	hash2el         *sync.Map
}

func NewDPoS(cfg *config.Config, accounts []*types.Account, eb event.EventBus) *DPoS {
	acTrx := newActiveTrx()
	l := ledger.NewLedger(cfg.LedgerDir())
	processorNum := runtime.NumCPU()

	dps := &DPoS{
		ledger:       l,
		acTrx:        acTrx,
		accounts:     accounts,
		logger:       log.NewLogger("dpos"),
		cfg:          cfg,
		eb:           eb,
		lv:           process.NewLedgerVerifier(l),
		cacheBlocks:  make(chan *consensus.BlockSource, common.DPoSMaxCacheBlocks),
		povReady:     make(chan bool, 1),
		processorNum: processorNum,
		processors:   newProcessors(processorNum),
		quitCh:       make(chan bool, 1),
		subMsg:       make(chan *subMsg, 10240),
		subMsgQuit:   make(chan bool, 1),
		subAck:       gcache.New(subAckMaxSize).Expiration(confirmReqMaxTimes * time.Minute).LRU().Build(),
		hash2el:      new(sync.Map),
	}

	if common.DPoSVoteCacheEn {
		dps.voteCache = gcache.New(voteCacheSize).LRU().Build()
	}

	dps.acTrx.setDposService(dps)
	for _, p := range dps.processors {
		p.setDposService(dps)
	}

	return dps
}

func (dps *DPoS) Init() {
	if dps.cfg.PoV.PovEnabled {
		dps.povSyncState.Store(common.SyncNotStart)

		err := dps.eb.SubscribeSync(common.EventPovSyncState, dps.onPovSyncState)
		if err != nil {
			dps.logger.Errorf("subscribe pov sync state event err")
		}
	} else {
		dps.povSyncState.Store(common.Syncdone)
	}

	supply := common.GenesisBlock().Balance
	dps.minVoteWeight, _ = supply.Div(common.DposVoteDivisor)
	dps.voteThreshold, _ = supply.Div(2)

	err := dps.eb.SubscribeSync(common.EventRollbackUnchecked, dps.onRollbackUnchecked)
	if err != nil {
		dps.logger.Errorf("subscribe rollback unchecked block event err")
	}

	if len(dps.accounts) != 0 {
		dps.refreshAccount()
	}
}

func (dps *DPoS) Start() {
	dps.logger.Info("DPOS service started!")

	go dps.acTrx.start()
	go dps.processSubMsg()
	dps.processorStart()

	timerFindOnlineRep := time.NewTicker(findOnlineRepInterval)
	timerRefreshPri := time.NewTicker(refreshPriInterval)
	timerDebug := time.NewTicker(time.Minute)

	for {
		select {
		case <-dps.quitCh:
			dps.logger.Info("Stopped DPOS.")
			return
		case <-timerRefreshPri.C:
			dps.logger.Info("refresh pri info.")
			go dps.refreshAccount()
		case <-timerFindOnlineRep.C:
			dps.logger.Info("begin Find Online Representatives.")
			go func() {
				err := dps.findOnlineRepresentatives()
				if err != nil {
					dps.logger.Error(err)
				}
				dps.cleanOnlineReps()
			}()
		case <-dps.povReady:
			err := dps.eb.Unsubscribe(common.EventPovSyncState, dps.onPovSyncState)
			if err != nil {
				dps.logger.Errorf("unsubscribe pov sync state err %s", err)
			}

			for bs := range dps.cacheBlocks {
				dps.logger.Infof("process cache block[%s]", bs.Block.GetHash())
				dps.dispatchMsg(bs)
			}
		case <-timerDebug.C:
			num := 0
			dps.hash2el.Range(func(key, value interface{}) bool {
				num++
				return true
			})
			dps.logger.Debugf("hash2el len:%d", num)
		}
	}
}

func (dps *DPoS) Stop() {
	dps.logger.Info("DPOS service stopped!")
	dps.quitCh <- true
	dps.subMsgQuit <- true
	dps.processorStop()
	dps.acTrx.stop()
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

func (dps *DPoS) getPovSyncState() common.SyncState {
	state := dps.povSyncState.Load()
	return state.(common.SyncState)
}

func (dps *DPoS) onPovSyncState(state common.SyncState) {
	dps.povSyncState.Store(state)
	dps.logger.Infof("pov sync state to [%s]", state)

	if dps.getPovSyncState() == common.Syncdone {
		dps.povReady <- true
		close(dps.cacheBlocks)
	}
}

func (dps *DPoS) cacheAck(ack *protos.ConfirmAckBlock) {
	if dps.voteCache.Has(ack.Hash) {
		v, err := dps.voteCache.Get(ack.Hash)
		if err != nil {
			dps.logger.Error("get vote cache err")
			return
		}

		vc := v.(*sync.Map)
		vc.Store(ack.Account, ack)
	} else {
		vc := new(sync.Map)
		vc.Store(ack.Account, ack)
		err := dps.voteCache.Set(ack.Hash, vc)
		if err != nil {
			dps.logger.Error("set vote cache err")
			return
		}
	}
}

func (dps *DPoS) pubAck(index int, ack *protos.ConfirmAckBlock) {
	dps.processors[index].acks <- ack
}

func (dps *DPoS) processSubMsg() {
	for {
		select {
		case <-dps.subMsgQuit:
			return
		case msg := <-dps.subMsg:
			switch msg.kind {
			case subMsgKindSub:
				_ = dps.subAck.Set(msg.hash, msg.index)

				if common.DPoSVoteCacheEn {
					vote, err := dps.voteCache.Get(msg.hash)
					if err == nil {
						vc := vote.(*sync.Map)
						vc.Range(func(key, value interface{}) bool {
							dps.pubAck(msg.index, value.(*protos.ConfirmAckBlock))
							return true
						})
						dps.voteCache.Remove(msg.hash)
					}
				}
			case subMsgKindUnsub:
				dps.subAck.Remove(msg.hash)

				if common.DPoSVoteCacheEn {
					dps.voteCache.Remove(msg.hash)
				}
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
		dps.eb.Publish(common.EventSendMsgToPeers, p2p.ConfirmAck, ack, bs.MsgFrom)
		dps.logger.Infof("dps recv confirmAck block[%s]", ack.Hash)

		if dps.getAckType(ack.Sequence) != ackTypeFindRep {
			val, err := dps.subAck.Get(ack.Hash)
			if err == nil {
				dps.pubAck(val.(int), ack)
			} else {
				dps.cacheAck(ack)
			}
		}
	} else {
		index := dps.getProcessorIndex(bs.Block.Address)
		dps.processors[index].blocks <- bs
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
	msg := &subMsg{
		kind: subMsgKindUnsub,
		hash: hash,
	}
	dps.subMsg <- msg
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
	case types.ContractReward: //deal gap tokenInfo
		input, err := dps.ledger.GetStateBlock(blk.GetLink())
		if err != nil {
			dps.logger.Errorf("get block link error [%s]", hash)
			return
		}

		if types.Address(input.GetLink()) == types.MintageAddress {
			param := new(cabi.ParamMintage)
			if err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameMintage, input.GetData()); err == nil {
				index := dps.getProcessorIndex(input.Address)
				if localIndex != index {
					dps.processors[index].blocksAcked <- param.TokenId
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
	if dps.getPovSyncState() == common.Syncdone || bs.BlockFrom == types.Synchronized {
		dps.dispatchMsg(bs)
	} else {
		if len(dps.cacheBlocks) < cap(dps.cacheBlocks) {
			dps.cacheBlocks <- bs
		} else {
			dps.logger.Errorf("pov not ready! cache block too much, drop it!")
		}
	}
}

func (dps *DPoS) localRepVote(block *types.StateBlock) {
	dps.localRepAccount.Range(func(key, value interface{}) bool {
		address := key.(types.Address)

		va, err := dps.voteGenerateWithSeq(block, address, value.(*types.Account), ackTypeCommon)
		if err != nil {
			return true
		}

		dps.logger.Debugf("rep [%s] vote for block[%s]", address, block.GetHash())
		dps.acTrx.setVoteHash(block)
		dps.acTrx.vote(va)
		dps.eb.Publish(common.EventBroadcast, p2p.ConfirmAck, va)
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
	if dps.cfg.PoV.PovEnabled {
		//povHeader, err := dps.ledger.GetLatestPovHeader()
		//if povHeader == nil {
		//	dps.logger.Errorf("get pov header err %s", err)
		//	return nil, errors.New("get pov header err")
		//}
		//
		//if block.PoVHeight > povHeader.Height+povBlockNumDay || block.PoVHeight+povBlockNumDay < povHeader.Height {
		//	dps.logger.Errorf("pov height invalid height:%d cur:%d", block.PoVHeight, povHeader.Height)
		//	return nil, errors.New("pov height invalid")
		//}
	}

	weight := dps.ledger.Weight(account)
	if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
		return nil, errors.New("too small weight")
	}

	hash := block.GetHash()
	va := &protos.ConfirmAckBlock{
		Sequence:  0,
		Hash:      hash,
		Account:   account,
		Signature: acc.Sign(hash),
	}
	return va, nil
}

func (dps *DPoS) voteGenerateWithSeq(block *types.StateBlock, account types.Address, acc *types.Account, kind uint32) (*protos.ConfirmAckBlock, error) {
	if dps.cfg.PoV.PovEnabled {
		//povHeader, err := dps.ledger.GetLatestPovHeader()
		//if povHeader == nil {
		//	dps.logger.Errorf("get pov header err %s", err)
		//	return nil, errors.New("get pov header err")
		//}
		//
		//if block.PoVHeight > povHeader.Height+povBlockNumDay || block.PoVHeight+povBlockNumDay < povHeader.Height {
		//	dps.logger.Errorf("pov height invalid height:%d cur:%d", block.PoVHeight, povHeader.Height)
		//	return nil, errors.New("pov height invalid")
		//}
	}

	weight := dps.ledger.Weight(account)
	if weight.Compare(dps.minVoteWeight) == types.BalanceCompSmaller {
		return nil, errors.New("too small weight")
	}

	hash := block.GetHash()
	va := &protos.ConfirmAckBlock{
		Sequence:  dps.getSeq(kind),
		Hash:      hash,
		Account:   account,
		Signature: acc.Sign(hash),
	}
	return va, nil
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
	if count > 1 {
		dps.logger.Error("it is very dangerous to run two or more representatives on one node")
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

		va, err := dps.voteGenerateWithSeq(blk, address, value.(*types.Account), ackTypeFindRep)
		if err != nil {
			return true
		}
		dps.eb.Publish(common.EventBroadcast, p2p.ConfirmAck, va)

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

func (dps *DPoS) onRollbackUnchecked(hash types.Hash) {
	dps.rollbackUncheckedFromDb(hash)
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

	//gap token
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
