package pov

import (
	"errors"
	"time"

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
	header.Target = target
	return nil
}

func (c *ConsensusFake) FinalizeHeader(header *types.PovHeader) error {
	return nil
}

func (c *ConsensusFake) VerifyHeader(header *types.PovHeader) error {
	if header.GetNonce() != header.GetHeight() {
		return errors.New("bad nonce")
	}
	return nil
}

func (c *ConsensusFake) SealHeader(header *types.PovHeader, cbAccount *types.Account, quitCh chan struct{}, resultCh chan<- *types.PovHeader) error {
	go func() {
		copyHdr := header.Copy()

		select {
		case <-quitCh:
		case <-time.After(time.Second):
			copyHdr.Nonce = copyHdr.GetHeight()
			voteHash := copyHdr.ComputeVoteHash()
			copyHdr.VoteSignature = cbAccount.Sign(voteHash)
			select {
			case resultCh <- copyHdr:
			default:
			}
		}
	}()
	return nil
}

func (c *ConsensusFake) calcNextRequiredTarget(header *types.PovHeader) (types.Signature, error) {
	var targetSig types.Signature
	err := targetSig.FromBigInt(common.PovGenesisTargetInt)
	if err != nil {
		return types.ZeroSignature, err
	}

	return targetSig, nil
}
