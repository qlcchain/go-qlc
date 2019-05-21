package api

import (
	"errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
	"math/big"
	"time"
)

type PovApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

type PovApiBlock struct {
	*types.PovBlock
}

type PovApiState struct {
	*types.PovAccountState
}

type PovApiDumpState struct {
	StateHash types.Hash                               `json:"stateHash"`
	Accounts  map[types.Address]*types.PovAccountState `json:"accounts"`
}

type PovApiTxLookup struct {
	TxHash      types.Hash         `json:"txHash"`
	TxLookup    *types.PovTxLookup `json:"txLookup"`
	Transaction *types.StateBlock  `json:"transaction"`
}

type PovLedgerStats struct {
	PovBlockCount   uint64 `json:"povBlockCount"`
	PovTxCount      uint64 `json:"povTxCount"`
	PovBestCount    uint64 `json:"povBestCount"`
	StateBlockCount uint64 `json:"stateBlockCount"`
}

type PovApiTD struct {
	Header   *types.PovHeader `json:"header"`
	TD       *big.Int         `json:"td"`
	TDHex    string           `json:"tdHex"`
	TDBitLen int              `json:"tdBitLen"`
}

type PovMinerStatItem struct {
	MinTime    time.Time `json:"minTime"`
	MaxTime    time.Time `json:"maxTime"`
	MinHeight  uint64    `json:"minHeight"`
	MaxHeight  uint64    `json:"maxHeight"`
	BlockCount uint64    `json:"blockCount"`
}

type PovMinerStats struct {
	Miners map[types.Address]*PovMinerStatItem `json:"miners"`
}

func NewPovApi(ledger *ledger.Ledger) *PovApi {
	return &PovApi{ledger: ledger, logger: log.NewLogger("rpc/pov")}
}

func (api *PovApi) GetBlockByHeight(height uint64) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	return apiBlock, nil
}

func (api *PovApi) GetBlockByHash(blockHash types.Hash) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	return apiBlock, nil
}

func (api *PovApi) GetLatestBlock() (*PovApiBlock, error) {
	block, err := api.ledger.GetLatestPovBlock()
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	return apiBlock, nil
}

func (api *PovApi) GetTransaction(txHash types.Hash) (*PovApiTxLookup, error) {
	txl, err := api.ledger.GetPovTxLookup(txHash)
	if err != nil {
		return nil, err
	}

	txBlock, err := api.ledger.GetStateBlock(txHash)
	if err != nil {
		return nil, err
	}

	apiTxl := &PovApiTxLookup{
		TxHash:      txHash,
		TxLookup:    txl,
		Transaction: txBlock,
	}

	return apiTxl, nil
}

func (api *PovApi) GetTransactionByBlockHashAndIndex(blockHash types.Hash, index uint32) (*PovApiTxLookup, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	if index >= block.TxNum {
		return nil, errors.New("tx index not exist")
	}
	tx := block.Transactions[index]

	return api.GetTransaction(tx.Hash)
}

func (api *PovApi) GetTransactionByBlockHeightAndIndex(height uint64, index uint32) (*PovApiTxLookup, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if index >= block.TxNum {
		return nil, errors.New("tx index not exist")
	}
	tx := block.Transactions[index]

	return api.GetTransaction(tx.Hash)
}

func (api *PovApi) GetAccountState(address types.Address, stateHash types.Hash) (*PovApiState, error) {
	db := api.ledger.Store
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	asBytes := stateTrie.GetValue(address.Bytes())
	if len(asBytes) <= 0 {
		return nil, errors.New("account value not exist")
	}

	as := new(types.PovAccountState)
	err := as.Deserialize(asBytes)
	if err != nil {
		return nil, err
	}

	apiState := &PovApiState{
		PovAccountState: as,
	}

	return apiState, nil
}

func (api *PovApi) GetAccountStateByBlockHash(address types.Address, blockHash types.Hash) (*PovApiState, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, block.StateHash)
}

func (api *PovApi) GetAccountStateByBlockHeight(address types.Address, height uint64) (*PovApiState, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, block.StateHash)
}

func (api *PovApi) DumpBlockState(blockHash types.Hash) (*PovApiDumpState, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	stateHash := block.StateHash
	dump := &PovApiDumpState{
		StateHash: stateHash,
		Accounts:  make(map[types.Address]*types.PovAccountState),
	}

	db := api.ledger.Store
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	it := stateTrie.NewIterator(nil)
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(key) != types.AddressSize {
			continue
		}

		addr, err := types.BytesToAddress(key)
		if err != nil {
			return nil, err
		}

		as := new(types.PovAccountState)
		err = as.Deserialize(val)
		if err != nil {
			return nil, err
		}

		dump.Accounts[addr] = as
	}

	return dump, nil
}

func (api *PovApi) GetLedgerStats() (*PovLedgerStats, error) {
	stats := &PovLedgerStats{}

	var err error
	stats.PovBlockCount, err = api.ledger.CountPovBlocks()
	if err != nil {
		return nil, err
	}

	stats.PovTxCount, err = api.ledger.CountPovTxs()
	if err != nil {
		return nil, err
	}

	stats.PovBestCount, err = api.ledger.CountPovBestHashs()
	if err != nil {
		return nil, err
	}

	stats.StateBlockCount, err = api.ledger.CountStateBlocks()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (api *PovApi) GetBlockTD(blockHash types.Hash) (*PovApiTD, error) {
	height, err := api.ledger.GetPovHeight(blockHash)
	if err != nil {
		return nil, err
	}
	header, err := api.ledger.GetPovHeader(height, blockHash)
	if err != nil {
		return nil, err
	}

	td, err := api.ledger.GetPovTD(blockHash, height)
	if err != nil {
		return nil, err
	}

	apiTD := &PovApiTD{
		Header:   header,
		TD:       td,
		TDHex:    td.Text(16),
		TDBitLen: td.BitLen(),
	}

	return apiTD, nil
}

func (api *PovApi) GetMinerStats(addrs []types.Address) (*PovMinerStats, error) {
	apiRsp := &PovMinerStats{
		Miners: make(map[types.Address]*PovMinerStatItem),
	}

	checkAddrMap := make(map[types.Address]bool)
	if len(addrs) > 0 {
		for _, addr := range addrs {
			checkAddrMap[addr] = true
		}
	}

	err := api.ledger.GetAllPovBlocks(func(block *types.PovBlock) error {
		if len(checkAddrMap) > 0 && checkAddrMap[block.GetCoinbase()] == false {
			return nil
		}

		item, ok := apiRsp.Miners[block.GetCoinbase()]
		if !ok {
			item = &PovMinerStatItem{}
			item.MinTime = time.Unix(block.GetTimestamp(), 0)
			item.MaxTime = time.Unix(block.GetTimestamp(), 0)
			item.MinHeight = block.GetHeight()
			item.MaxHeight = block.GetHeight()
			item.BlockCount = 1

			apiRsp.Miners[block.GetCoinbase()] = item
		} else {
			if item.MinTime.Unix() > block.GetTimestamp() {
				item.MinTime = time.Unix(block.GetTimestamp(), 0)
			}
			if item.MaxTime.Unix() < block.GetTimestamp() {
				item.MaxTime = time.Unix(block.GetTimestamp(), 0)
			}
			if item.MinHeight > block.GetHeight() {
				item.MinHeight = block.GetHeight()
			}
			if item.MaxHeight < block.GetHeight() {
				item.MaxHeight = block.GetHeight()
			}
			item.BlockCount += 1
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return apiRsp, nil
}
