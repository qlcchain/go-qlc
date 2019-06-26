package api

import (
	"errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
	"math/big"
	"sync/atomic"
	"time"
)

type PovApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
	eb     event.EventBus

	syncState atomic.Value
}

type PovApiHeader struct {
	*types.PovHeader
}

type PovApiBatchHeader struct {
	Count   int                `json:"count"`
	Headers []*types.PovHeader `json:"headers"`
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
	BestCount  uint64    `json:"bestCount"`
}

type PovMinerStats struct {
	MinerCount int                                 `json:"minerCount"`
	MinerStats map[types.Address]*PovMinerStatItem `json:"minerStats"`
}

func NewPovApi(ledger *ledger.Ledger, eb event.EventBus) *PovApi {
	api := &PovApi{
		ledger: ledger,
		eb:     eb,
		logger: log.NewLogger("rpc/pov"),
	}
	api.syncState.Store(common.SyncNotStart)
	_ = eb.SubscribeSync(string(common.EventPovSyncState), api.OnPovSyncState)
	return api
}

func (api *PovApi) OnPovSyncState(state common.SyncState) {
	api.logger.Infof("receive pov sync state [%s]", state)
	api.syncState.Store(state)
}

func (api *PovApi) GetHeaderByHeight(height uint64) (*PovApiHeader, error) {
	blockHash, err := api.ledger.GetPovBestHash(height)
	if err != nil {
		return nil, err
	}

	header, err := api.ledger.GetPovHeader(height, blockHash)
	if err != nil {
		return nil, err
	}

	apiHeader := &PovApiHeader{
		PovHeader: header,
	}

	return apiHeader, nil
}

func (api *PovApi) GetHeaderByHash(blockHash types.Hash) (*PovApiHeader, error) {
	height, err := api.ledger.GetPovHeight(blockHash)
	if err != nil {
		return nil, err
	}

	header, err := api.ledger.GetPovHeader(height, blockHash)
	if err != nil {
		return nil, err
	}

	apiHeader := &PovApiHeader{
		PovHeader: header,
	}

	return apiHeader, nil
}

func (api *PovApi) GetLatestHeader() (*PovApiHeader, error) {
	header, err := api.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	apiHeader := &PovApiHeader{
		PovHeader: header,
	}

	return apiHeader, nil
}

func (api *PovApi) GetFittestHeader(gap uint64) (*PovApiHeader, error) {
	ss := api.syncState.Load().(common.SyncState)
	if ss != common.Syncdone {
		return nil, errors.New("pov sync is not finished, please check it")
	}

	var header *types.PovHeader

	latestHeader, err := api.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	if gap > 0 && latestHeader.GetHeight() > gap {
		height := latestHeader.GetHeight() - gap
		header, err = api.ledger.GetPovHeaderByHeight(height)
		if err != nil {
			return nil, err
		}
	} else {
		header = latestHeader
	}

	apiHeader := &PovApiHeader{
		PovHeader: header,
	}

	return apiHeader, nil
}

func (api *PovApi) BatchGetHeadersByHeight(height uint64, count uint64, asc bool) (*PovApiBatchHeader, error) {
	var headers []*types.PovHeader
	var err error
	if asc {
		headers, err = api.ledger.BatchGetPovHeadersByHeightAsc(height, count)
	} else {
		headers, err = api.ledger.BatchGetPovHeadersByHeightDesc(height, count)
	}
	if err != nil {
		return nil, err
	}

	apiHeader := &PovApiBatchHeader{
		Count:   len(headers),
		Headers: headers,
	}

	return apiHeader, nil
}

func (api *PovApi) GetBlockByHeight(height uint64, txOffset uint32, txLimit uint32) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	apiBlock.PovBlock.Transactions = api.pagingTxs(block.Transactions, txOffset, txLimit)

	return apiBlock, nil
}

func (api *PovApi) GetBlockByHash(blockHash types.Hash, txOffset uint32, txLimit uint32) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	apiBlock.PovBlock.Transactions = api.pagingTxs(block.Transactions, txOffset, txLimit)

	return apiBlock, nil
}

func (api *PovApi) GetLatestBlock(txOffset uint32, txLimit uint32) (*PovApiBlock, error) {
	block, err := api.ledger.GetLatestPovBlock()
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}

	apiBlock.PovBlock.Transactions = api.pagingTxs(block.Transactions, txOffset, txLimit)

	return apiBlock, nil
}

func (api *PovApi) pagingTxs(txs []*types.PovTransaction, txOffset uint32, txLimit uint32) []*types.PovTransaction {
	txNum := uint32(len(txs))

	if txOffset >= txNum {
		txOffset = txNum
	}
	if txLimit == 0 {
		txLimit = 100
	}
	if (txOffset + txLimit) > txNum {
		txLimit = txNum - txOffset
	}

	return txs[txOffset : txOffset+txLimit]
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
		MinerStats: make(map[types.Address]*PovMinerStatItem),
	}

	checkAddrMap := make(map[types.Address]bool)
	if len(addrs) > 0 {
		for _, addr := range addrs {
			checkAddrMap[addr] = true
		}
	}

	bestBlockHashMap := make(map[types.Hash]struct{})
	err := api.ledger.GetAllPovBestHashes(func(height uint64, hash types.Hash) error {
		bestBlockHashMap[hash] = struct{}{}
		return nil
	})

	err = api.ledger.GetAllPovBlocks(func(block *types.PovBlock) error {
		if len(checkAddrMap) > 0 && checkAddrMap[block.GetCoinbase()] == false {
			return nil
		}

		item, ok := apiRsp.MinerStats[block.GetCoinbase()]
		if !ok {
			item = &PovMinerStatItem{}
			item.MinTime = time.Unix(block.GetTimestamp(), 0)
			item.MaxTime = time.Unix(block.GetTimestamp(), 0)
			item.MinHeight = block.GetHeight()
			item.MaxHeight = block.GetHeight()
			item.BlockCount = 1

			apiRsp.MinerStats[block.GetCoinbase()] = item
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
		if _, ok := bestBlockHashMap[block.GetHash()]; ok {
			item.BestCount += 1
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	apiRsp.MinerCount = len(apiRsp.MinerStats)

	return apiRsp, nil
}
