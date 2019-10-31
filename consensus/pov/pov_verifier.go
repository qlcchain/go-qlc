package pov

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
)

type PovVerifier struct {
	store  ledger.Store
	chain  PovVerifierChainReader
	cs     ConsensusPov
	logger *zap.SugaredLogger
}

type PovVerifyStat struct {
	Result process.ProcessResult
	ErrMsg string
	GapTxs map[types.Hash]process.ProcessResult

	CurHeader     *types.PovHeader
	PrevHeader    *types.PovHeader
	PrevStateTrie *trie.Trie
	StateTrie     *trie.Trie
	TxBlocks      map[types.Hash]*types.StateBlock
}

func NewPovVerifyStat() *PovVerifyStat {
	pvs := new(PovVerifyStat)
	pvs.GapTxs = make(map[types.Hash]process.ProcessResult)
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
	CalcPastMedianTime(prevHeader *types.PovHeader) uint32
	GenStateTrie(height uint64, prevStateHash types.Hash, txs []*types.PovTransaction) (*trie.Trie, error)
	GetStateTrie(stateHash *types.Hash) *trie.Trie
	GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState
	CalcBlockReward(header *types.PovHeader) (types.Balance, types.Balance)
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

	result, err = pv.verifyAuxHeader(block, stat)
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
	blkHash := block.GetHash()

	if common.PovChainGenesisBlockHeight == block.GetHeight() {
		if !common.IsGenesisPovBlock(block) {
			return process.BadHash, fmt.Errorf("bad genesis block hash %s", blkHash)
		}
	}

	computedHash := block.ComputeHash()
	if blkHash.IsZero() || computedHash != blkHash {
		return process.BadHash, fmt.Errorf("bad hash, %s != %s", computedHash, blkHash)
	}

	return process.Progress, nil
}

func (pv *PovVerifier) verifyTimestamp(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	if block.GetTimestamp() == 0 {
		return process.InvalidTime, errors.New("timestamp is zero")
	}

	if block.GetTimestamp() > uint32(time.Now().Unix()+int64(common.PovMaxAllowedFutureTimeSec)) {
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
	allTxs := block.GetAllTxs()
	if block.GetTxNum() != uint32(len(allTxs)) {
		return process.InvalidTxNum, fmt.Errorf("txNum %d != txs %d", block.GetTxNum(), len(allTxs))
	}

	if len(allTxs) < 1 {
		return process.InvalidTxNum, fmt.Errorf("coinbase tx not exist")
	}

	var txHashList []*types.Hash
	for _, tx := range allTxs {
		txHash := tx.Hash
		txHashList = append(txHashList, &txHash)
	}
	merkleRoot := merkle.CalcMerkleTreeRootHash(txHashList)
	if merkleRoot.IsZero() {
		return process.BadMerkleRoot, fmt.Errorf("bad merkle root is zero when txs exist")
	}
	if merkleRoot != block.GetMerkleRoot() {
		return process.BadMerkleRoot, fmt.Errorf("bad merkle root not equals %s != %s", merkleRoot, block.GetMerkleRoot())
	}

	cbTx := block.Header.CbTx
	cbTx.Hash = cbTx.ComputeHash()

	result, err := pv.verifyCoinBaseTx(cbTx, stat)
	if err != nil {
		return result, err
	}

	for txIdx, tx := range allTxs {
		if txIdx == 0 {
			if tx.Hash != cbTx.Hash {
				return process.BadHash, errors.New("invalid coinbase tx hash")
			}
			tx.CbTx = cbTx
		} else {
			txBlock, _ := pv.store.GetStateBlockConfirmed(tx.Hash)
			if txBlock == nil {
				stat.GapTxs[tx.Hash] = process.GapTransaction
			} else {
				tx.Block = txBlock
				stat.TxBlocks[tx.Hash] = txBlock
			}
		}
	}

	if len(stat.GapTxs) > 0 {
		return process.GapTransaction, fmt.Errorf("total %d txs in pending", len(stat.GapTxs))
	}

	prevTrie := stat.getPrevStateTrie(pv, block.GetPrevious())
	if prevTrie == nil {
		return process.BadStateHash, errors.New("failed to get prev state tire")
	}
	addrTokenPrevHashes := make(map[types.AddressToken]types.Hash)

	accTxs := allTxs[1:]

	for txIdx := 0; txIdx < len(accTxs); txIdx++ {
		tx := accTxs[txIdx]
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

func (pv *PovVerifier) verifyCoinBaseTx(cbTx *types.PovCoinBaseTx, stat *PovVerifyStat) (process.ProcessResult, error) {
	minerTxOut := cbTx.GetMinerTxOut()
	if minerTxOut == nil || minerTxOut.Address.IsZero() {
		return process.BadCoinbase, errors.New("miner address is zero")
	}

	repTxOut := cbTx.GetRepTxOut()
	if repTxOut == nil || repTxOut.Address.IsZero() {
		return process.BadCoinbase, errors.New("rep address is zero")
	}

	calcMinerRwd, calcRepRwd := pv.chain.CalcBlockReward(stat.CurHeader)

	minerReward := minerTxOut.Value
	if minerReward.Compare(calcMinerRwd) == types.BalanceCompBigger {
		return process.BadCoinbase, fmt.Errorf("miner got bad reward, %v > %v", minerReward, calcMinerRwd)
	}

	repReward := repTxOut.Value
	if repReward.Compare(calcRepRwd) == types.BalanceCompBigger {
		return process.BadCoinbase, fmt.Errorf("rep got bad reward, %v > %v", repReward, calcRepRwd)
	}

	return process.Progress, nil
}

func (pv *PovVerifier) verifyState(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	prevHeader := stat.getPrevHeader(pv, block.GetPrevious())
	if prevHeader == nil {
		return process.GapPrevious, fmt.Errorf("prev block %s pending", block.GetPrevious())
	}

	stateTrie, err := pv.chain.GenStateTrie(block.GetHeight(), prevHeader.GetStateHash(), block.GetAccountTxs())
	if err != nil {
		return process.BadStateHash, err
	}
	if stateTrie == nil {
		return process.BadStateHash, errors.New("failed to gen state trie")
	}

	stateHash := *stateTrie.Hash()
	if stateHash != block.GetStateHash() {
		return process.BadStateHash, fmt.Errorf("state hash is not equals %s != %s", stateHash, block.GetStateHash())
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

func (pv *PovVerifier) verifyAuxHeader(block *types.PovBlock, stat *PovVerifyStat) (process.ProcessResult, error) {
	if block.Header.AuxHdr == nil {
		return process.Progress, nil
	}

	/*
		Merged mining coinbase

		Insert exactly one of these headers into the scriptSig of the coinbase transaction in the parent block.
		Field Size 	Description 	Data type 	Comments
		4 	magic 	char[4] 	0xfa, 0xbe, 'm', 'm' (only required if over 20 bytes past the start of the script; optional otherwise)
		32 	block_hash 	char[32] 	Hash of the AuxPOW block header / root of the chain merkle branch
		4 	merkle_size 	int32_t 	Number of entries in aux work merkle tree. (Must be a power of 2)
		4 	merkle_nonce 	int32_t 	Nonce used to calculate indexes into aux work merkle tree; you may as well leave this at zero

		That string of 44 bytes being part of the coinbase script means that the miner constructed the AuxPOW Block before creating the coinbase.
	*/

	ap := block.Header.AuxHdr
	calcParMR := merkle.CalcMerkleRootByIndex(ap.ParCoinBaseTx.ComputeHash(), ap.ParCoinBaseMerkle, ap.ParMerkleIndex)
	if calcParMR != ap.ParBlockHeader.MerkleRoot {
		return process.BadAuxHeader, fmt.Errorf("parent merkle root not equal, %s != %s", calcParMR, ap.ParBlockHeader.MerkleRoot)
	}

	// reverse the hashAuxBlock
	hashAuxBlock := block.GetHash().ReverseByte()

	auxRootHash := merkle.CalcMerkleRootByIndex(hashAuxBlock, ap.AuxMerkleBranch, ap.AuxMerkleIndex)

	script := ap.ParCoinBaseTx.TxIn[0].SignatureScript
	scriptStr := hex.EncodeToString(script)

	// reverse the auxRootHash
	auxRootHashReverseStr := hex.EncodeToString(auxRootHash.ReverseByte().Bytes())
	pchMergedMiningHeaderStr := hex.EncodeToString(types.PovAuxPowHeaderMagic)

	headerIndex := strings.Index(scriptStr, pchMergedMiningHeaderStr)
	rootHashIndex := strings.Index(scriptStr, auxRootHashReverseStr)

	if headerIndex == -1 {
		return process.BadAuxHeader, errors.New("coinbase aux magic not exist")
	}
	if rootHashIndex == -1 {
		return process.BadAuxHeader, errors.New("coinbase aux root hash not exist")
	}

	if strings.Index(scriptStr[headerIndex+2:], pchMergedMiningHeaderStr) != -1 {
		return process.BadAuxHeader, errors.New("coinbase aux magic not exist")
	}

	if headerIndex+len(pchMergedMiningHeaderStr) != rootHashIndex {
		return process.BadAuxHeader, errors.New("coinbase aux root hash not exist")
	}

	rootHashIndex += len(auxRootHashReverseStr)
	if len(scriptStr)-rootHashIndex < 8 {
		return process.BadAuxHeader, errors.New("coinbase script size too small")
	}

	size := binary.LittleEndian.Uint32(script[rootHashIndex/2 : rootHashIndex/2+4])
	merkleHeight := len(ap.AuxMerkleBranch)
	if size != uint32(1<<uint32(merkleHeight)) {
		return process.BadAuxHeader, errors.New("aux merkle branch height not equal")
	}

	nonce := binary.LittleEndian.Uint32(script[rootHashIndex/2+4 : rootHashIndex/2+8])
	if ap.AuxMerkleIndex != merkle.CalcAuxPowExpectedIndex(nonce, types.PovAuxPowChainID, merkleHeight) {
		return process.BadAuxHeader, errors.New("aux merkle index not equal")
	}

	return process.Progress, nil
}
