package process

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type PovVerifier struct {
	store ledger.Store
	chain PovVerifierChainReader
	logger *zap.SugaredLogger
}

type PovVerifierChainReader interface {
	GetBlockByHash(hash types.Hash) *types.PovBlock
	CalcNextRequiredTarget(block *types.PovBlock) (types.Signature, error)
}

func NewPovVerifier(store ledger.Store, chain PovVerifierChainReader) *PovVerifier {
	return &PovVerifier{store: store, chain: chain, logger: log.NewLogger("pov_verifier")}
}

func (pv *PovVerifier) Process(block types.Block) (ProcessResult, error) {
	return Progress, nil
}

func (pv *PovVerifier) BlockCheck(block types.Block) (ProcessResult, error) {
	povBlock, ok := block.(*types.PovBlock)
	if !ok {
		return Other, errors.New("invalid block")
	}

	result, err := pv.verifyHeader(povBlock)
	if err != nil {
		return result, err
	}

	result, err = pv.verifyTimestamp(povBlock)
	if err != nil {
		return result, err
	}

	result, err = pv.verifyReferred(povBlock)
	if err != nil {
		return result, err
	}

	result, err = pv.verifyProducer(povBlock)
	if err != nil {
		return result, err
	}

	result, err = pv.verifyTarget(povBlock)
	if err != nil {
		return result, err
	}

	result, err = pv.verifyTransactions(povBlock)
	if err != nil {
		return result, err
	}

	return Progress, nil
}

func (pv *PovVerifier) verifyHeader(block *types.PovBlock) (ProcessResult, error) {
	computedHash := block.ComputeHash()
	if block.Hash.IsZero() || computedHash != block.Hash {
		return BadHash, fmt.Errorf("bad hash, %s != %s", computedHash, block.Hash)
	}

	if len(block.Signature) == 0 || len(block.Coinbase) == 0 {
		return BadSignature, errors.New("signature or coinbase is nil")
	}

	isVerified := block.Coinbase.Verify(block.GetHash().Bytes(), block.GetSignature().Bytes())
	if !isVerified {
		return BadSignature, errors.New("bad signature")
	}

	return Progress, nil
}

func (pv *PovVerifier) verifyTimestamp(block *types.PovBlock) (ProcessResult, error) {
	if block.Timestamp <= 0 {
		return InvalidTime, errors.New("timestamp is 0")
	}

	return Progress, nil
}

func (pv *PovVerifier) verifyReferred(block *types.PovBlock) (ProcessResult, error) {
	prevBlock := pv.chain.GetBlockByHash(block.GetPrevious())
	if prevBlock == nil {
		return GapPrevious, nil
	}

	if prevBlock.Timestamp > block.Timestamp {
		return InvalidTime, errors.New("timestamp must be greater than previous")
	}

	return Progress, nil
}

func (pv *PovVerifier) verifyTransactions(block *types.PovBlock) (ProcessResult, error) {
	if block.TxNum != uint32(len(block.Transactions)) {
		return InvalidTxNum, nil
	}

	txHashs := make([]*types.Hash, len(block.Transactions))
	for _, tx := range block.Transactions {
		txHashs = append(txHashs, &tx.Hash)
	}
	merkleRoot := merkle.CalcMerkleTreeRootHash(txHashs)
	if *merkleRoot != block.MerkleRoot {
		return BadMerkleRoot, fmt.Errorf("bad merkle root, %s != %s", merkleRoot, block.MerkleRoot)
	}

	for _, tx := range block.Transactions {
		txBlock, _ := pv.store.GetStateBlock(tx.Hash)
		if txBlock == nil {
			return GapTransaction, nil
		}
	}

	return Progress, nil
}

func (pv *PovVerifier) verifyTarget(block *types.PovBlock) (ProcessResult, error) {
	prevBlock := pv.chain.GetBlockByHash(block.GetPrevious())
	if prevBlock == nil {
		return GapPrevious, nil
	}

	expectedTarget, err := pv.chain.CalcNextRequiredTarget(prevBlock)
	if err != nil {
		return BadTarget, err
	}
	if expectedTarget != block.Target {
		return BadTarget, err
	}

	voteHash := block.ComputeVoteHash()
	voteSig := block.GetVoteSignature()

	isVerified := block.GetCoinbase().Verify(voteHash.Bytes(), voteSig.Bytes())
	if !isVerified {
		return BadSignature, errors.New("bad vote signature")
	}

	voteSigInt := voteSig.ToBigInt()

	targetSig := block.GetTarget()
	targetInt := targetSig.ToBigInt()

	if voteSigInt.Cmp(targetInt) > 0 {
		return BadTarget, errors.New("bad target")
	}

	return Progress, nil
}

func (pv *PovVerifier) verifyProducer(block *types.PovBlock) (ProcessResult, error) {
	// TODO: check coinbase is valid producer or not
	return Progress, nil
}