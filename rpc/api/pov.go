package api

import (
	"errors"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
)

type PovApi struct {
	cfg    *config.Config
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
	eb     event.EventBus

	syncState atomic.Value
}

type PovStatus struct {
	PovEnabled bool `json:"povEnabled"`
	SyncState  int  `json:"syncState"`
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
	MainBlockNum     uint32    `json:"mainBlockNum"`
	FirstBlockTime   time.Time `json:"firstBlockTime"`
	LastBlockTime    time.Time `json:"lastBlockTime"`
	FirstBlockHeight uint64    `json:"firstBlockHeight"`
	LastBlockHeight  uint64    `json:"lastBlockHeight"`
	IsOnline         bool      `json:"isOnline"`
}

type PovMinerStats struct {
	MinerCount  int `json:"minerCount"`
	OnlineCount int `json:"onlineCount"`

	MinerStats map[types.Address]*PovMinerStatItem `json:"minerStats"`

	TotalBlockNum     uint32 `json:"totalBlockNum"`
	LatestBlockHeight uint64 `json:"latestBlockHeight"`
}

func NewPovApi(cfg *config.Config, ledger *ledger.Ledger, eb event.EventBus) *PovApi {
	api := &PovApi{
		cfg:    cfg,
		ledger: ledger,
		eb:     eb,
		logger: log.NewLogger("rpc/pov"),
	}
	api.syncState.Store(common.SyncNotStart)
	_ = eb.SubscribeSync(common.EventPovSyncState, api.OnPovSyncState)
	return api
}

func (api *PovApi) OnPovSyncState(state common.SyncState) {
	api.logger.Infof("receive pov sync state [%s]", state)
	api.syncState.Store(state)
}

func (api *PovApi) GetPovStatus() (*PovStatus, error) {
	apiRsp := new(PovStatus)
	apiRsp.PovEnabled = api.cfg.PoV.PovEnabled
	apiRsp.SyncState = int(api.syncState.Load().(common.SyncState))
	return apiRsp, nil
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
	if !api.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	ss := api.syncState.Load().(common.SyncState)
	if ss != common.SyncDone {
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

func (api *PovApi) GetLatestAccountState(address types.Address) (*PovApiState, error) {
	header, err := api.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.StateHash)
}

func (api *PovApi) GetAccountStateByBlockHash(address types.Address, blockHash types.Hash) (*PovApiState, error) {
	header, err := api.ledger.GetPovHeaderByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.StateHash)
}

func (api *PovApi) GetAccountStateByBlockHeight(address types.Address, height uint64) (*PovApiState, error) {
	header, err := api.ledger.GetPovHeaderByHeight(height)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.StateHash)
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

	tmNow := time.Now()

	checkAddrMap := make(map[types.Address]bool)
	if len(addrs) > 0 {
		for _, addr := range addrs {
			checkAddrMap[addr] = true
		}
	}

	totalBlockNum := uint32(0)

	// scan miner stats per day
	lastDayIndex := uint32(0)
	err := api.ledger.GetAllPovMinerStats(func(stat *types.PovMinerDayStat) error {
		if stat.DayIndex > lastDayIndex {
			lastDayIndex = stat.DayIndex
		}
		for addrHex, minerStat := range stat.MinerStats {
			minerAddr, _ := types.HexToAddress(addrHex)

			if len(checkAddrMap) > 0 && checkAddrMap[minerAddr] == false {
				return nil
			}

			item, ok := apiRsp.MinerStats[minerAddr]
			if !ok {
				item = &PovMinerStatItem{}
				item.FirstBlockHeight = minerStat.FirstHeight
				item.LastBlockHeight = minerStat.LastHeight

				apiRsp.MinerStats[minerAddr] = item
			} else {
				if item.FirstBlockHeight > minerStat.FirstHeight {
					item.FirstBlockHeight = minerStat.FirstHeight
				}
				if item.LastBlockHeight < minerStat.LastHeight {
					item.LastBlockHeight = minerStat.LastHeight
				}
			}
			item.MainBlockNum += minerStat.BlockNum
			totalBlockNum += minerStat.BlockNum
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// scan best block not in miner stats per day
	latestHeader, _ := api.ledger.GetLatestPovHeader()

	notStatHeightStart := uint64(lastDayIndex) * uint64(common.POVChainBlocksPerDay)
	notStatHeightEnd := latestHeader.GetHeight()

	for height := notStatHeightStart; height <= notStatHeightEnd; height++ {
		header, _ := api.ledger.GetPovHeaderByHeight(height)
		if header == nil {
			break
		}

		minerAddr := header.GetCoinbase()
		if len(checkAddrMap) > 0 && checkAddrMap[minerAddr] == false {
			continue
		}

		item, ok := apiRsp.MinerStats[minerAddr]
		if !ok {
			item = &PovMinerStatItem{}
			item.FirstBlockHeight = header.GetHeight()
			item.LastBlockHeight = header.GetHeight()

			apiRsp.MinerStats[minerAddr] = item
		} else {
			if item.FirstBlockHeight > header.GetHeight() {
				item.FirstBlockHeight = header.GetHeight()
			}
			if item.LastBlockHeight < header.GetHeight() {
				item.LastBlockHeight = header.GetHeight()
			}
		}
		item.MainBlockNum += 1
		totalBlockNum += 1
	}

	// fill time
	for _, minerItem := range apiRsp.MinerStats {
		firstBlock, _ := api.ledger.GetPovHeaderByHeight(minerItem.FirstBlockHeight)
		if firstBlock != nil {
			minerItem.FirstBlockTime = time.Unix(firstBlock.GetTimestamp(), 0)
		}
		lastBlock, _ := api.ledger.GetPovHeaderByHeight(minerItem.LastBlockHeight)
		if lastBlock != nil {
			minerItem.LastBlockTime = time.Unix(lastBlock.GetTimestamp(), 0)
		}
	}

	apiRsp.MinerCount = len(apiRsp.MinerStats)

	apiRsp.TotalBlockNum = totalBlockNum
	apiRsp.LatestBlockHeight = latestHeader.GetHeight()

	// miner is online if it generate blocks in last hour
	for _, item := range apiRsp.MinerStats {
		if item.LastBlockTime.Add(time.Hour).After(tmNow) {
			item.IsOnline = true
			apiRsp.OnlineCount++
		}
	}

	return apiRsp, nil
}

func (api *PovApi) GetMinerDayStats(dayIndex uint32) (*types.PovMinerDayStat, error) {
	dayStat, err := api.ledger.GetPovMinerStat(dayIndex)
	if err != nil {
		return nil, err
	}

	return dayStat, nil
}
