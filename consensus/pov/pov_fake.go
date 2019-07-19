package pov

import (
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

func (c *ConsensusFake) VerifyHeader(header *types.PovHeader) error {
	return nil
}

func (c *ConsensusFake) SealHeader(header *types.PovHeader, cbAccount *types.Account, quitCh chan struct{}, resultCh chan<- *types.PovHeader) error {
	return nil
}
