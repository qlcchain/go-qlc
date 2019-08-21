package pov

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
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

	target, err := c.calcNextRequiredTarget(prevHeader)
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

	asBytes := prevTrie.GetValue(prevHeader.GetCoinBase().Bytes())
	if len(asBytes) <= 0 {
		return errors.New("failed to get account state value")
	}

	as := new(types.PovAccountState)
	err := as.Deserialize(asBytes)
	if err != nil {
		return errors.New("failed to deserialize account state value")
	}

	if as.RepState == nil {
		return errors.New("account rep state is nil")
	}
	rs := as.RepState

	if rs.Vote.Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
		return errors.New("coinbase pledge amount not enough")
	}

	return nil
}

func (c *ConsensusPow) verifyTarget(header *types.PovHeader) error {
	prevHeader := c.chainR.GetHeaderByHash(header.GetPrevious())
	if prevHeader == nil {
		return errors.New("failed to get previous header")
	}

	expectedTarget, err := c.calcNextRequiredTarget(prevHeader)
	if err != nil {
		return err
	}
	if expectedTarget != header.GetBits() {
		return errors.New("target not equal next required target")
	}

	powHash := header.ComputePowHash()
	powInt := powHash.ToBigInt()

	targetInt := header.GetTargetInt()

	if powInt.Cmp(targetInt) > 0 {
		return errors.New("target greater than vote signature")
	}

	return nil
}

func (c *ConsensusPow) SealHeader(header *types.PovHeader, cbAccount *types.Account, quitCh chan struct{}, resultCh chan<- *types.PovHeader) error {
	var wgMine sync.WaitGroup

	abortCh := make(chan struct{})
	localCh := make(chan *types.PovHeader)

	for id := 1; id <= c.mineWorkerNum; id++ {
		wgMine.Add(1)
		go func(id int, gap int) {
			defer wgMine.Done()
			c.mineWorker(id, gap, header, cbAccount, abortCh, localCh)
		}(id, int(c.mineWorkerNum))
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

func (c *ConsensusPow) mineWorker(id int, gap int, header *types.PovHeader, cbAccount *types.Account, abortCh chan struct{}, localCh chan *types.PovHeader) {
	copyHdr := header.Copy()
	targetInt := copyHdr.GetTargetInt()

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
		if powInt.Cmp(targetInt) <= 0 {
			c.logger.Debugf("mine worker %d found nonce %d", id, nonce)
			localCh <- copyHdr
			return
		}
	}

	c.logger.Debugf("mine worker %d exhaust nonce", id)
	localCh <- nil
}

func (c *ConsensusPow) calcNextRequiredTarget(prevHeader *types.PovHeader) (uint32, error) {
	if (prevHeader.GetHeight()+1)%uint64(common.PovChainTargetCycle) != 0 {
		return prevHeader.GetBits(), nil
	}

	// nextTarget = prevTarget * (lastBlock.Timestamp - firstBlock.Timestamp) / (blockInterval * targetCycle)

	distance := uint64(common.PovChainTargetCycle - 1)
	firstHeader := c.chainR.RelativeAncestor(prevHeader, distance)
	if firstHeader == nil {
		c.logger.Errorf("failed to get relative ancestor at height %d distance %d", prevHeader.GetHeight(), distance)
		return 0, ErrPovUnknownAncestor
	}

	targetTimeSpan := uint32(common.PovChainRetargetTimespan)
	minRetargetTimespan := uint32(common.PovChainMinRetargetTimespan)
	maxRetargetTimespan := uint32(common.PovChainMaxRetargetTimespan)

	actualTimespan := prevHeader.GetTimestamp() - firstHeader.GetTimestamp()
	if actualTimespan < minRetargetTimespan {
		actualTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		actualTimespan = maxRetargetTimespan
	}

	oldTargetInt := prevHeader.GetTargetInt()
	nextTargetInt := new(big.Int).Set(oldTargetInt)
	nextTargetInt.Mul(oldTargetInt, big.NewInt(int64(actualTimespan)))
	nextTargetInt.Div(nextTargetInt, big.NewInt(int64(targetTimeSpan)))

	nextTargetBits := types.BigToCompact(nextTargetInt)

	c.logger.Infof("Difficulty retarget at block height %d", prevHeader.GetHeight()+1)
	c.logger.Infof("Old target %d (%s)", oldTargetInt.BitLen(), oldTargetInt.Text(16))
	c.logger.Infof("New target %d (%s)", nextTargetInt.BitLen(), nextTargetInt.Text(16))
	c.logger.Infof("Actual timespan %v, target timespan %v",
		time.Duration(actualTimespan)*time.Second, time.Duration(targetTimeSpan)*time.Second)

	return nextTargetBits, nil
}
