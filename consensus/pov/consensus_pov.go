package pov

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/trie"
)

const (
	PovConsensusModeFake = iota
	PovConsensusModePow
)

type PovConsensusChainReader interface {
	GetHeaderByHash(hash types.Hash) *types.PovHeader
	RelativeAncestor(header *types.PovHeader, distance uint64) *types.PovHeader
	GetStateTrie(stateHash *types.Hash) *trie.Trie
	GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState
}

type ConsensusPov interface {
	Init() error
	Start() error
	Stop() error

	PrepareHeader(header *types.PovHeader) error
	FinalizeHeader(header *types.PovHeader) error
	VerifyHeader(header *types.PovHeader) error
	SealHeader(header *types.PovHeader, quitCh chan struct{}, resultCh chan<- *types.PovHeader) error
}

func NewPovConsensus(mode int, chainR PovConsensusChainReader) ConsensusPov {
	if mode == PovConsensusModeFake {
		return NewConsensusFake(chainR)
	} else if mode == PovConsensusModePow {
		return NewConsensusPow(chainR)
	}

	return nil
}
