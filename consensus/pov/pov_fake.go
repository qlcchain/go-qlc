package pov

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
)

type ConsensusFake struct {
	chainR PovConsensusChainReader
}

func NewConsensusFake(chainR PovConsensusChainReader) *ConsensusFake {
	consFake := &ConsensusFake{chainR: chainR}
	return consFake
}

func (c *ConsensusFake) Init() error {
	return nil
}

func (c *ConsensusFake) Start() error {
	return nil
}

func (c *ConsensusFake) Stop() error {
	return nil
}

func (c *ConsensusFake) PrepareHeader(header *types.PovHeader) error {
	target, err := c.calcNextRequiredTarget(header)
	if err != nil {
		return err
	}
	header.BasHdr.Bits = target
	return nil
}

func (c *ConsensusFake) FinalizeHeader(header *types.PovHeader) error {
	return nil
}

func (c *ConsensusFake) VerifyHeader(header *types.PovHeader) error {
	if header.BasHdr.Nonce != header.BasHdr.Timestamp {
		return errors.New("bad nonce")
	}
	return nil
}

func (c *ConsensusFake) calcNextRequiredTarget(header *types.PovHeader) (uint32, error) {
	return common.PovGenesisPowBits, nil
}
