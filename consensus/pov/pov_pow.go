package pov

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

type ConsensusPow struct {
	chainR        PovConsensusChainReader
	logger        *zap.SugaredLogger
	mineWorkerNum int
}

func NewConsensusPow(chainR PovConsensusChainReader) *ConsensusPow {
	consPow := &ConsensusPow{chainR: chainR}
	consPow.logger = log.NewLogger("pov_cs_pow")
	return consPow
}

func (c *ConsensusPow) Init() error {
	/*
		cpuNum := runtime.NumCPU()
		if cpuNum >= 4 {
			c.mineWorkerNum = cpuNum / 2
		} else {
			c.mineWorkerNum = 1
		}
	*/
	c.mineWorkerNum = 1
	return nil
}

func (c *ConsensusPow) Start() error {
	return nil
}

func (c *ConsensusPow) Stop() error {
	return nil
}

func (c *ConsensusPow) PrepareHeader(header *types.PovHeader) error {
	prevHeader := c.chainR.GetHeaderByHash(header.GetPrevious())
	if prevHeader == nil {
		return fmt.Errorf("failed to get previous header %s", header.GetPrevious())
	}

	target, err := c.calcNextRequiredTarget(prevHeader, header)
	if err != nil {
		return err
	}

	header.BasHdr.Bits = target
	return nil
}

func (c *ConsensusPow) FinalizeHeader(header *types.PovHeader) error {
	return nil
}

func (c *ConsensusPow) VerifyHeader(header *types.PovHeader) error {
	var err error

	err = c.verifyTarget(header)
	if err != nil {
		return err
	}

	err = c.verifyProducer(header)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsensusPow) verifyProducer(header *types.PovHeader) error {
	if header.GetHeight() < common.PovMinerVerifyHeightStart {
		return nil
	}

	prevHeader := c.chainR.GetHeaderByHash(header.GetPrevious())
	if prevHeader == nil {
		return errors.New("failed to get previous header")
	}

	prevStateHash := prevHeader.GetStateHash()
	prevTrie := c.chainR.GetStateTrie(&prevStateHash)
	if prevTrie == nil {
		return errors.New("failed to get previous state tire")
	}

	rsKey := types.PovCreateRepStateKey(prevHeader.GetMinerAddr())
	rsVal := prevTrie.GetValue(rsKey)
	if len(rsVal) <= 0 {
		return errors.New("failed to get rep state value")
	}

	rs := types.NewPovRepState()
	err := rs.Deserialize(rsVal)
	if err != nil {
		return errors.New("failed to deserialize rep state value")
	}

	if rs.Vote.Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
		return errors.New("pledge amount not enough")
	}

	return nil
}

func (c *ConsensusPow) verifyTarget(header *types.PovHeader) error {
	prevHeader := c.chainR.GetHeaderByHash(header.GetPrevious())
	if prevHeader == nil {
		return errors.New("failed to get previous header")
	}

	expectedTarget, err := c.calcNextRequiredTarget(prevHeader, header)
	if err != nil {
		return err
	}
	if expectedTarget != header.GetBits() {
		return fmt.Errorf("target 0x%x not equal next required target 0x%x", header.GetBits(), expectedTarget)
	}

	powHash := header.ComputePowHash()
	powInt := powHash.ToBigInt()
	powBits := types.BigToCompact(powInt)

	targetIntAlgo := header.GetAlgoTargetInt()

	if powInt.Cmp(targetIntAlgo) > 0 {
		algoBits := types.BigToCompact(targetIntAlgo)
		return fmt.Errorf("pow hash 0x%x greater than target 0x%x", powBits, algoBits)
	}

	return nil
}

func (c *ConsensusPow) SealHeader(header *types.PovHeader, quitCh chan struct{}, resultCh chan<- *types.PovHeader) error {
	var wgMine sync.WaitGroup

	abortCh := make(chan struct{})
	localCh := make(chan *types.PovHeader)

	for id := 1; id <= c.mineWorkerNum; id++ {
		wgMine.Add(1)
		go func(id int, gap int) {
			defer wgMine.Done()
			c.mineWorker(id, gap, header, abortCh, localCh)
		}(id, c.mineWorkerNum)
	}

	go func() {
		select {
		case <-quitCh:
			close(abortCh)
		case result := <-localCh:
			select {
			case resultCh <- result:
			default:
				// there's no hash
				c.logger.Warnf("failed to send sealing result to miner")
			}
			close(abortCh)
		}
		wgMine.Wait()
	}()

	return nil
}

func (c *ConsensusPow) mineWorker(id int, gap int, header *types.PovHeader, abortCh chan struct{}, localCh chan *types.PovHeader) {
	copyHdr := header.Copy()
	targetIntAlgo := copyHdr.GetAlgoTargetInt()

	tryCnt := 0
	for nonce := uint32(gap); nonce < common.PovMaxNonce; nonce += uint32(gap) {
		tryCnt++
		if tryCnt >= 100 {
			tryCnt = 0
			select {
			case <-abortCh:
				c.logger.Debugf("mine worker %d abort search nonce", id)
				localCh <- nil
				return
			default:
				//Non-blocking select to fall through
			}
		}

		copyHdr.BasHdr.Nonce = nonce

		powHash := copyHdr.ComputePowHash()
		powInt := powHash.ToBigInt()
		if powInt.Cmp(targetIntAlgo) <= 0 {
			c.logger.Debugf("mine worker %d found nonce %d", id, nonce)
			localCh <- copyHdr
			return
		}
	}

	c.logger.Debugf("mine worker %d exhaust nonce", id)
	localCh <- nil
}

func (c *ConsensusPow) calcNextRequiredTarget(lastHeader *types.PovHeader, curHeader *types.PovHeader) (uint32, error) {
	return c.calcNextRequiredTargetByQLC(lastHeader, curHeader)
	//return c.calcNextRequiredTargetByDGW(lastHeader, curHeader)
	//return c.calcNextRequiredTargetByAlgo(lastHeader, curHeader)
}

func (c *ConsensusPow) calcNextRequiredTargetByQLC(lastHeader *types.PovHeader, curHeader *types.PovHeader) (uint32, error) {
	if (lastHeader.GetHeight()+1)%uint64(common.PovChainTargetCycle) != 0 {
		nextTargetInt := lastHeader.GetNormTargetInt()
		nextTargetIntAlgo := new(big.Int).Mul(nextTargetInt, big.NewInt(int64(curHeader.GetAlgoEfficiency())))
		nextTargetIntBitsAlgo := types.BigToCompact(nextTargetIntAlgo)
		return nextTargetIntBitsAlgo, nil
	}

	// nextTarget = prevTarget * (lastBlock.Timestamp - firstBlock.Timestamp) / (blockInterval * targetCycle)

	distance := uint64(common.PovChainTargetCycle - 1)
	firstHeader := c.chainR.RelativeAncestor(lastHeader, distance)
	if firstHeader == nil {
		c.logger.Errorf("failed to get relative ancestor at height %d distance %d", lastHeader.GetHeight(), distance)
		return 0, ErrPovUnknownAncestor
	}

	targetTimeSpan := uint32(common.PovChainRetargetTimespan)
	minRetargetTimespan := uint32(common.PovChainMinRetargetTimespan)
	maxRetargetTimespan := uint32(common.PovChainMaxRetargetTimespan)

	actualTimespan := uint32(1)
	if lastHeader.GetTimestamp() > firstHeader.GetTimestamp() {
		actualTimespan = lastHeader.GetTimestamp() - firstHeader.GetTimestamp()
	}
	if actualTimespan < minRetargetTimespan {
		actualTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		actualTimespan = maxRetargetTimespan
	}

	// convert to normalized target by algo efficiency
	oldTargetInt := lastHeader.GetNormTargetInt()

	// nextTargetInt = oldTargetInt * actualTimespan / targetTimeSpan
	nextTargetInt := new(big.Int).Mul(oldTargetInt, big.NewInt(int64(actualTimespan)))
	nextTargetInt = new(big.Int).Div(nextTargetInt, big.NewInt(int64(targetTimeSpan)))

	// convert to algo target
	nextTargetIntAlgo := new(big.Int).Mul(nextTargetInt, big.NewInt(int64(curHeader.GetAlgoEfficiency())))

	// at least pow limit
	if nextTargetIntAlgo.Cmp(common.PovPowLimitInt) > 0 {
		nextTargetIntAlgo = common.PovPowLimitInt
	}

	nextTargetBitsAlgo := types.BigToCompact(nextTargetIntAlgo)

	c.logger.Infof("Difficulty target at block height %d", lastHeader.GetHeight()+1)
	c.logger.Infof("Old target %d (%s)", oldTargetInt.BitLen(), oldTargetInt.Text(16))
	c.logger.Infof("New target %d (%s)", nextTargetInt.BitLen(), nextTargetInt.Text(16))
	c.logger.Infof("Actual timespan %v, target timespan %v",
		time.Duration(actualTimespan)*time.Second, time.Duration(targetTimeSpan)*time.Second)

	return nextTargetBitsAlgo, nil
}

func (c *ConsensusPow) calcNextRequiredTargetByDGW(lastHeader *types.PovHeader, curHeader *types.PovHeader) (uint32, error) {
	var nextTargetInt *big.Int

	oldTargetInt := lastHeader.GetNormTargetInt()
	pastBlockCount := common.PovChainTargetCycle
	targetTimeSpan := int64(common.PovChainBlockInterval * pastBlockCount)
	actualTimespan := curHeader.GetTimestamp() - lastHeader.GetTimestamp()

	// Tesetnet ???
	/*
		 else if curHeader.GetTimestamp() > prevHeader.GetTimestamp()+2*60*60 {
			nextTargetInt = common.PovPowLimitInt
			c.logger.Debugf("recent block is more than 2 hours old")
		} else if curHeader.GetTimestamp() > prevHeader.GetTimestamp()+10*60 {
			nextTargetInt = common.PovPowLimitInt
			c.logger.Debugf("recent block is more than 10 minutes old")
		}
	*/

	if lastHeader.GetHeight() < uint64(pastBlockCount) {
		nextTargetInt = common.PovPowLimitInt
	} else {
		// nextTarget = prevAvgTarget * (lastBlock.Timestamp - firstBlock.Timestamp) / (pastBlockCount * blockInterval)

		pastSumTargetInt := big.NewInt(0)
		firstHeader := lastHeader
		for blockCount := 0; blockCount < pastBlockCount; blockCount++ {
			// convert to normalized target by algo efficiency
			scanTargetInt := firstHeader.GetNormTargetInt()

			pastSumTargetInt = new(big.Int).Add(pastSumTargetInt, scanTargetInt)

			firstHeader = c.chainR.GetHeaderByHash(firstHeader.GetPrevious())
			if firstHeader == nil {
				c.logger.Errorf("failed to get previous %s", lastHeader.GetPrevious())
				return 0, ErrPovInvalidPrevious
			}
		}

		pastAvgTargetInt := new(big.Int).Div(pastSumTargetInt, big.NewInt(int64(pastBlockCount)))

		minRetargetTimespan := uint32(targetTimeSpan / 2)
		maxRetargetTimespan := uint32(targetTimeSpan * 2)

		if lastHeader.GetTimestamp() > firstHeader.GetTimestamp() {
			actualTimespan = lastHeader.GetTimestamp() - firstHeader.GetTimestamp()
		} else {
			actualTimespan = 1
		}
		if actualTimespan < minRetargetTimespan {
			actualTimespan = minRetargetTimespan
		} else if actualTimespan > maxRetargetTimespan {
			actualTimespan = maxRetargetTimespan
		}

		// nextTargetInt = pastAvgTargetInt * actualTimespan / targetTimeSpan
		nextTargetInt = new(big.Int).Mul(pastAvgTargetInt, big.NewInt(int64(actualTimespan)))
		nextTargetInt = new(big.Int).Div(nextTargetInt, big.NewInt(targetTimeSpan))
	}

	// convert to algo target
	nextTargetIntAlgo := new(big.Int).Mul(nextTargetInt, big.NewInt(int64(curHeader.GetAlgoEfficiency())))

	// at least pow limit
	if nextTargetIntAlgo.Cmp(common.PovPowLimitInt) > 0 {
		nextTargetIntAlgo = common.PovPowLimitInt
	}
	nextTargetBitsAlgo := types.BigToCompact(nextTargetIntAlgo)

	if (curHeader.GetHeight()+1)%uint64(pastBlockCount) == 0 {
		c.logger.Infof("Difficulty target at block height %d", lastHeader.GetHeight()+1)
		c.logger.Infof("Old target %d (%s)", oldTargetInt.BitLen(), oldTargetInt.Text(16))
		c.logger.Infof("New target %d (%s)", nextTargetInt.BitLen(), nextTargetInt.Text(16))
		c.logger.Infof("actual timespan %s, target timespan %s",
			time.Duration(actualTimespan)*time.Second, time.Duration(targetTimeSpan)*time.Second)
	}

	return nextTargetBitsAlgo, nil
}

func (c *ConsensusPow) calcNextRequiredTargetByAlgo(lastHeader *types.PovHeader, curHeader *types.PovHeader) (uint32, error) {
	ALGO_ACTIVE_COUNT := uint64(5)
	nPastAlgoFastBlocks := uint64(5)
	nPastAlgoBlocks := nPastAlgoFastBlocks * ALGO_ACTIVE_COUNT

	nPastFastBlocks := nPastAlgoFastBlocks * 2
	nPastBlocks := nPastFastBlocks * ALGO_ACTIVE_COUNT

	// stabilizing block spacing
	nPastBlocks = nPastBlocks * 100

	// make sure we have at least ALGO_ACTIVE_COUNT blocks, otherwise just return powLimit
	if lastHeader.GetHeight() < nPastBlocks {
		if lastHeader.GetHeight() < nPastAlgoBlocks {
			return common.PovPowLimitBits, nil
		} else {
			nPastBlocks = lastHeader.GetHeight()
		}
	}

	pindexLast := lastHeader
	pindex := pindexLast
	pindexFast := pindexLast
	bnPastTargetAvg := big.NewInt(0)
	bnPastTargetAvgFast := big.NewInt(0)

	var pindexAlgo *types.PovHeader
	var pindexAlgoFast *types.PovHeader
	var pindexAlgoLast *types.PovHeader
	bnPastAlgoTargetAvg := big.NewInt(0)
	bnPastAlgoTargetAvgFast := big.NewInt(0)

	// count blocks mined by actual algo for secondary average
	curAlgo := curHeader.BasHdr.Version & uint32(types.ALGO_VERSION_MASK)

	nCountBlocks := uint64(0)
	nCountFastBlocks := uint64(0)
	nCountAlgoBlocks := uint64(0)
	nCountAlgoFastBlocks := uint64(0)

	for nCountBlocks < nPastBlocks && nCountAlgoBlocks < nPastAlgoBlocks {
		// convert to normalized target by algo efficiency
		bnTarget := pindex.GetNormTargetInt()

		// calculate algo average
		if curAlgo == (pindex.BasHdr.Version & uint32(types.ALGO_VERSION_MASK)) {
			nCountAlgoBlocks++

			pindexAlgo = pindex
			if pindexAlgoLast == nil {
				pindexAlgoLast = pindex
			}

			// algo average
			// bnPastAlgoTargetAvg = (bnPastAlgoTargetAvg * (nCountAlgoBlocks - 1) + bnTarget) / nCountAlgoBlocks
			bnPastAlgoTargetAvg = new(big.Int).Mul(bnPastAlgoTargetAvg, big.NewInt(int64(nCountAlgoBlocks-1)))
			bnPastAlgoTargetAvg = new(big.Int).Add(bnPastAlgoTargetAvg, bnTarget)
			bnPastAlgoTargetAvg = new(big.Int).Div(bnPastAlgoTargetAvg, big.NewInt(int64(nCountAlgoBlocks)))

			// fast algo average
			if nCountAlgoBlocks <= nPastAlgoFastBlocks {
				nCountAlgoFastBlocks++
				pindexAlgoFast = pindex
				bnPastAlgoTargetAvgFast = bnPastAlgoTargetAvg
			}
		}

		nCountBlocks++

		// average
		// bnPastTargetAvg = (bnPastTargetAvg * (nCountBlocks - 1) + bnTarget) / nCountBlocks
		bnPastTargetAvg = new(big.Int).Mul(bnPastTargetAvg, big.NewInt(int64(nCountBlocks-1)))
		bnPastTargetAvg = new(big.Int).Add(bnPastTargetAvg, bnTarget)
		bnPastTargetAvg = new(big.Int).Div(bnPastTargetAvg, big.NewInt(int64(nCountBlocks)))
		// fast average
		if nCountBlocks <= nPastFastBlocks {
			nCountFastBlocks++
			pindexFast = pindex
			bnPastTargetAvgFast = bnPastTargetAvg
		}

		// next block
		if nCountBlocks != nPastBlocks {
			pindex = c.chainR.GetHeaderByHash(pindex.GetPrevious())
			if pindex == nil { // should never fail
				c.logger.Errorf("failed to get previous %s", lastHeader.GetPrevious())
				return 0, ErrPovInvalidPrevious
			}
		}
	}

	// instamine protection for blockchain
	if (pindexLast.GetTimestamp() - pindexFast.GetTimestamp()) < uint32(common.PovChainBlockInterval/2) {
		nCountBlocks = nCountFastBlocks
		pindex = pindexFast
		bnPastTargetAvg = bnPastTargetAvgFast
	}

	bnNew := bnPastTargetAvg

	if pindexAlgo != nil && pindexAlgoLast != nil && nCountAlgoBlocks > 1 {
		// instamine protection for algo
		if (pindexLast.GetTimestamp() - pindexAlgoFast.GetTimestamp()) < uint32(uint64(common.PovChainBlockInterval)*ALGO_ACTIVE_COUNT/2) {
			nCountAlgoBlocks = nCountAlgoFastBlocks
			pindexAlgo = pindexAlgoFast
			bnPastAlgoTargetAvg = bnPastAlgoTargetAvgFast
		}

		bnNew = bnPastAlgoTargetAvg

		// pindexLast instead of pindexAlgoLst on purpose
		nActualTimespan := uint64(pindexLast.GetTimestamp() - pindexAlgo.GetTimestamp())
		nTargetTimespan := nCountAlgoBlocks * uint64(common.PovChainBlockInterval) * ALGO_ACTIVE_COUNT

		// higher algo diff faster
		if nActualTimespan < 1 {
			nActualTimespan = 1
		}
		// lower algo diff slower
		if nActualTimespan > nTargetTimespan*2 {
			nActualTimespan = nTargetTimespan * 2
		}

		// Retarget algo, bnNew = bnNew * nActualTimespan / nTargetTimespan
		bnNew = new(big.Int).Mul(bnNew, big.NewInt(int64(nActualTimespan)))
		bnNew = new(big.Int).Div(bnNew, big.NewInt(int64(nTargetTimespan)))
	} else {
		bnNew = common.PovPowLimitInt
	}

	//c.logger.Infof("Algo New target %d (%s)", bnNew.BitLen(), bnNew.Text(16))

	nActualTimespan := uint64(pindexLast.GetTimestamp() - pindex.GetTimestamp())
	nTargetTimespan := nCountBlocks * uint64(common.PovChainBlockInterval)

	// higher diff faster
	if nActualTimespan < 1 {
		nActualTimespan = 1
	}
	// lower diff slower
	if nActualTimespan > nTargetTimespan*2 {
		nActualTimespan = nTargetTimespan * 2
	}

	// Retarget, bnNew = bnNew * nActualTimespan / nTargetTimespan
	bnNew = new(big.Int).Mul(bnNew, big.NewInt(int64(nActualTimespan)))
	bnNew = new(big.Int).Div(bnNew, big.NewInt(int64(nTargetTimespan)))

	//c.logger.Infof("Chain New target %d (%s)", bnNew.BitLen(), bnNew.Text(16))

	// convert to algo target
	bnNewAlgo := new(big.Int).Mul(bnNew, big.NewInt(int64(curHeader.GetAlgoEfficiency())))

	// at least PoW limit
	if bnNewAlgo.Cmp(common.PovPowLimitInt) > 0 {
		bnNewAlgo = common.PovPowLimitInt
	}
	bnNewBitsAlgo := types.BigToCompact(bnNewAlgo)

	if (curHeader.GetHeight()+1)%uint64(10) == 0 {
		bnOld := lastHeader.GetNormTargetInt()
		c.logger.Infof("Difficulty target at block height %d", lastHeader.GetHeight()+1)
		c.logger.Infof("Old target %d (%s)", bnOld.BitLen(), bnOld.Text(16))
		c.logger.Infof("New target %d (%s)", bnNew.BitLen(), bnNew.Text(16))
		c.logger.Infof("actual timespan %s, target timespan %s",
			time.Duration(nActualTimespan)*time.Second, time.Duration(nTargetTimespan)*time.Second)
	}

	return bnNewBitsAlgo, nil
}
