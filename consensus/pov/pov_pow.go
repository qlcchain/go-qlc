package pov

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

type ConsensusPow struct {
	chainR PovConsensusChainReader
	logger *zap.SugaredLogger
}

func NewConsensusPow(chainR PovConsensusChainReader) *ConsensusPow {
	consPow := &ConsensusPow{chainR: chainR}
	consPow.logger = log.NewLogger("pov_cs_pow")
	return consPow
}

func (c *ConsensusPow) Init() error {
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
