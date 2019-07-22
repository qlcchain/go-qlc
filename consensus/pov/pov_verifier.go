package pov

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
	"time"
)

type PovVerifier struct {
	store  ledger.Store
	chain  PovVerifierChainReader
	cs     ConsensusPov
	logger *zap.SugaredLogger
}

type PovVerifyStat struct {
	Result    process.ProcessResult
	ErrMsg    string
	TxResults map[types.Hash]process.ProcessResult

	CurHeader     *types.PovHeader
	PrevHeader    *types.PovHeader
	PrevStateTrie *trie.Trie
	StateTrie     *trie.Trie
	TxBlocks      map[types.Hash]*types.StateBlock
}

func NewPovVerifyStat() *PovVerifyStat {
	pvs := new(PovVerifyStat)
	pvs.TxResults = make(map[types.Hash]process.ProcessResult)
	pvs.TxBlocks = make(map[types.Hash]*types.StateBlock)
	return pvs
}

func (pvs *PovVerifyStat) setResult(result process.ProcessResult, err error) {
	pvs.Result = result
	if err != nil {
		pvs.ErrMsg = err.Error()
	}
}

func (pvs *PovVerifyStat) getCurHeader(pv *PovVerifier, block *types.PovBlock) *types.PovHeader {
	if pvs.CurHeader == nil {
		pvs.CurHeader = block.GetHeader()
	}

	return pvs.CurHeader
}

func (pvs *PovVerifyStat) getPrevHeader(pv *PovVerifier, prevHash types.Hash) *types.PovHeader {
	if pvs.PrevHeader == nil {
		pvs.PrevHeader = pv.chain.GetHeaderByHash(prevHash)
	}

	return pvs.PrevHeader
}

func (pvs *PovVerifyStat) getPrevStateTrie(pv *PovVerifier, prevHash types.Hash) *trie.Trie {
	if pvs.PrevStateTrie == nil {
		prevHeader := pvs.getPrevHeader(pv, prevHash)
		if prevHeader != nil {
			prevStateHash := prevHeader.GetStateHash()
			pvs.PrevStateTrie = pv.chain.GetStateTrie(&prevStateHash)
		}
	}

	return pvs.PrevStateTrie
}

type PovVerifierChainReader interface {
	GetHeaderByHash(hash types.Hash) *types.PovHeader
	CalcPastMedianTime(prevHeader *types.PovHeader) int64
	GenStateTrie(prevStateHash types.Hash, txs []*types.PovTransaction) (*trie.Trie, error)
	GetStateTrie(stateHash *types.Hash) *trie.Trie
	GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState
}

func NewPovVerifier(store ledger.Store, chain PovVerifierChainReader, cs ConsensusPov) *PovVerifier {
	return &PovVerifier{store: store, chain: chain, cs: cs, logger: log.NewLogger("pov_verifier")}
}

func (pv *PovVerifier) Process(block types.Block) (process.ProcessResult, error) {
	return process.Other, nil
}

func (pv *PovVerifier) BlockCheck(block types.Block) (process.ProcessResult, error) {
	return process.Other, nil
}

func (pv *PovVerifier) VerifyNet(block *types.PovBlock) *PovVerifyStat {
	stat := NewPovVerifyStat()

	result, err := pv.verifyDataIntegrity(block, stat)
	if err != nil {
		stat.Result = result
		stat.ErrMsg = err.Error()
		return stat
	}

	result, err = pv.verifyTimestamp(block, stat)
	if err != nil || result != process.Progress {
		stat.setResult(result, err)
		return stat
	}

	stat.Result = process.Progress
	return stat
}

func (pv *PovVerifier) VerifyFull(block *types.PovBlock) *PovVerifyStat {
	stat := NewPovVerifyStat()

	result, err := pv.verifyDataIntegrity(block, stat)
	if err != nil || result != process.Progress {
		stat.setResult(result, err)
		return stat
	}

	result, err = pv.verifyTimestamp(block, stat)
	if err != nil || result != process.Progress {
		stat.setResult(result, err)
		return stat
	}

	result, err = pv.verifyReferred(block, stat)
	if err != nil || result != process.Progress {
		stat.setResult(result, err)
		return stat
	}

	result, err = pv.verifyConsensus(block, stat)
	if err != nil || result != process.Progress {
		stat.setResult(result, err)
		return stat
	}

	result, err = pv.verifyTransactions(block, stat)
	if err != nil || result != process.Progress {
		stat.setResult(result, err)
		return stat
	}

	result, err = pv.verifyState(block, stat)
	if err != nil || result != process.Progress {
		stat.setResult(result, err)
		return stat
	}

	stat.Result = process.Progress
	return stat
}

func (pv *PovVerifier) verifyDataIntegrity(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	if common.PovChainGenesisBlockHeight == block.GetHeight() {
		if !common.IsGenesisPovBlock(block) {
			return process.BadHash, fmt.Errorf("bad genesis block hash %s", block.Hash)
		}
	}

	computedHash := block.ComputeHash()
	if block.Hash.IsZero() || computedHash != block.Hash {
		return process.BadHash, fmt.Errorf("bad hash, %s != %s", computedHash, block.Hash)
	}

	if block.Coinbase.IsZero() {
		return process.BadSignature, errors.New("coinbase is zero")
	}

	if block.Signature.IsZero() {
		return process.BadSignature, errors.New("signature is zero")
	}

	isVerified := block.Coinbase.Verify(block.GetHash().Bytes(), block.GetSignature().Bytes())
	if !isVerified {
		return process.BadSignature, errors.New("bad signature")
	}

	return process.Progress, nil
}

func (pv *PovVerifier) verifyTimestamp(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	if block.Timestamp <= 0 {
		return process.InvalidTime, errors.New("timestamp is zero")
	}

	if block.GetTimestamp() > (time.Now().Unix() + int64(common.PovMaxAllowedFutureTimeSec)) {
		return process.InvalidTime, fmt.Errorf("timestamp %d too far from future", block.GetTimestamp())
	}

	return process.Progress, nil
}

func (pv *PovVerifier) verifyReferred(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	prevHeader := stat.getPrevHeader(pv, block.GetPrevious())
	if prevHeader == nil {
		return process.GapPrevious, nil
	}

	if block.GetHeight() != prevHeader.GetHeight()+1 {
		return process.InvalidHeight, fmt.Errorf("height %d not continue with previous %d", block.GetHeight(), prevHeader.GetHeight())
	}

	medianTime := pv.chain.CalcPastMedianTime(prevHeader)

	if block.GetTimestamp() < medianTime {
		return process.InvalidTime, fmt.Errorf("timestamp %d not greater than median time %d", block.GetTimestamp(), medianTime)
	}

	return process.Progress, nil
}

func (pv *PovVerifier) verifyTransactions(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	if block.TxNum != uint32(len(block.Transactions)) {
		return process.InvalidTxNum, nil
	}

	if len(block.Transactions) <= 0 {
		if !block.MerkleRoot.IsZero() {
			return process.BadMerkleRoot, fmt.Errorf("bad merkle root not zero when txs empty")
		}
		return process.Progress, nil
	}

	var txHashList []*types.Hash
	for _, tx := range block.Transactions {
		txHash := tx.Hash
		txHashList = append(txHashList, &txHash)
	}
	merkleRoot := merkle.CalcMerkleTreeRootHash(txHashList)
	if merkleRoot.IsZero() {
		return process.BadMerkleRoot, fmt.Errorf("bad merkle root is zero when txs exist")
	}
	if merkleRoot != block.MerkleRoot {
		return process.BadMerkleRoot, fmt.Errorf("bad merkle root not equals %s != %s", merkleRoot, block.MerkleRoot)
	}
	for _, tx := range block.Transactions {
		txBlock, _ := pv.store.GetStateBlock(tx.Hash)
		if txBlock == nil {
			stat.TxResults[tx.Hash] = process.GapTransaction
		} else {
			tx.Block = txBlock
			stat.TxBlocks[tx.Hash] = txBlock
		}
	}

	if len(stat.TxResults) > 0 {
		return process.GapTransaction, fmt.Errorf("total %d txs in pending", len(stat.TxResults))
	}

	prevTrie := stat.getPrevStateTrie(pv, block.GetPrevious())
	if prevTrie == nil {
		return process.BadStateHash, errors.New("failed to get prev state tire")
	}
	addrTokenPrevHashes := make(map[types.AddressToken]types.Hash)
	for txIdx := 0; txIdx < len(block.Transactions); txIdx++ {
		tx := block.Transactions[txIdx]
		isCA := types.IsContractAddress(tx.Block.GetAddress())
		addrToken := types.AddressToken{Address: tx.Block.GetAddress(), Token: tx.Block.GetToken()}

		prevHashWant, ok := addrTokenPrevHashes[addrToken]
		if !ok {
			// contract address's blocks are all independent, no previous
			if isCA {
				prevHashWant = types.ZeroHash
			} else {
				as := pv.chain.GetAccountState(prevTrie, tx.Block.GetAddress())
				if as != nil {
					ts := as.GetTokenState(tx.Block.GetToken())
					if ts != nil {
						prevHashWant = ts.Hash
					} else {
						prevHashWant = types.ZeroHash
					}
				} else {
					prevHashWant = types.ZeroHash
				}
			}
		}

		//pv.logger.Debugf("address %s token %s block %s", tx.Block.GetAddress(), tx.Block.GetToken(), tx.GetHash())
		//pv.logger.Debugf("prevHashWant %s txPrevHash %s", prevHashWant, tx.Block.GetPrevious())

		if prevHashWant != tx.Block.GetPrevious() {
			return process.InvalidTxOrder, errors.New("tx is not in order")
		}

		// contract address's blocks are all independent, no previous
		if !isCA {
			addrTokenPrevHashes[addrToken] = tx.Block.GetHash()
		}
	}

	return process.Progress, nil
}

func (pv *PovVerifier) verifyState(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	prevHeader := stat.getPrevHeader(pv, block.GetPrevious())
	if prevHeader == nil {
		return process.GapPrevious, fmt.Errorf("prev block %s pending", block.GetPrevious())
	}

	stateTrie, err := pv.chain.GenStateTrie(prevHeader.StateHash, block.Transactions)
	if err != nil {
		return process.BadStateHash, err
	}
	stateHash := types.Hash{}
	if stateTrie != nil {
		stateHash = *stateTrie.Hash()
	}
	if stateHash != block.StateHash {
		return process.BadStateHash, fmt.Errorf("state hash is not equals %s != %s", stateHash, block.StateHash)
	}
	stat.StateTrie = stateTrie

	return process.Progress, nil
}

func (pv *PovVerifier) verifyConsensus(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	header := stat.getCurHeader(pv, block)

	err := pv.cs.VerifyHeader(header)
	if err != nil {
		return process.BadConsensus, err
	}

	return process.Progress, nil
}
