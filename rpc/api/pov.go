package api

import (
	"encoding/hex"
	"errors"
	"fmt"
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
	AlgoName       string  `json:"algoName"`
	AlgoEfficiency uint    `json:"algoEfficiency"`
	NormBits       uint32  `json:"normBits"`
	NormDifficulty float64 `json:"normDifficulty"`
	AlgoDifficulty float64 `json:"algoDifficulty"`
}

type PovApiBatchHeader struct {
	Count   int             `json:"count"`
	Headers []*PovApiHeader `json:"headers"`
}

type PovApiBlock struct {
	*types.PovBlock
	AlgoName       string  `json:"algoName"`
	AlgoEfficiency uint    `json:"algoEfficiency"`
	NormBits       uint32  `json:"normBits"`
	NormDifficulty float64 `json:"normDifficulty"`
	AlgoDifficulty float64 `json:"algoDifficulty"`
}

type PovApiState struct {
	AccountState *types.PovAccountState `json:"accountState"`
	RepState     *types.PovRepState     `json:"repState"`
}

type PovApiDumpState struct {
	StateHash types.Hash                               `json:"stateHash"`
	Accounts  map[types.Address]*types.PovAccountState `json:"accounts"`
	Reps      map[types.Address]*types.PovRepState     `json:"reps"`
}

type PovApiRepState struct {
	StateHash types.Hash                           `json:"stateHash"`
	Reps      map[types.Address]*types.PovRepState `json:"reps"`
}

type PovApiTxLookup struct {
	TxHash   types.Hash         `json:"txHash"`
	TxLookup *types.PovTxLookup `json:"txLookup"`
}

type PovLedgerStats struct {
	PovBlockCount    uint64 `json:"povBlockCount"`
	PovBestCount     uint64 `json:"povBestCount"`
	PovTxLookupCount uint64 `json:"povTxLookupCount"`
	PovStateTxCount  uint64 `json:"povStateTxCount"`
	StateBlockCount  uint64 `json:"stateBlockCount"`
}

type PovApiTD struct {
	Header *types.PovHeader `json:"header"`
	TD     *types.PovTD     `json:"td"`
}

type PovMinerStatItem struct {
	MainBlockNum     uint32        `json:"mainBlockNum"`
	MainRewardAmount types.Balance `json:"mainRewardAmount"`
	FirstBlockTime   time.Time     `json:"firstBlockTime"`
	LastBlockTime    time.Time     `json:"lastBlockTime"`
	FirstBlockHeight uint64        `json:"firstBlockHeight"`
	LastBlockHeight  uint64        `json:"lastBlockHeight"`
	IsHourOnline     bool          `json:"isHourOnline"`
	IsDayOnline      bool          `json:"isDayOnline"`
}

type PovMinerStats struct {
	MinerCount      int `json:"minerCount"`
	HourOnlineCount int `json:"hourOnlineCount"`
	DayOnlineCount  int `json:"dayOnlineCount"`

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

func (api *PovApi) fillHeader(header *PovApiHeader) {
	header.AlgoEfficiency = header.GetAlgoEfficiency()
	header.AlgoName = header.GetAlgoType().String()
	header.NormBits = header.GetNormBits()
	header.NormDifficulty = types.CalcDifficultyRatio(header.NormBits, common.PovPowLimitBits)
	header.AlgoDifficulty = types.CalcDifficultyRatio(header.BasHdr.Bits, common.PovPowLimitBits)
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
	api.fillHeader(apiHeader)

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
	api.fillHeader(apiHeader)

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
	api.fillHeader(apiHeader)

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
	api.fillHeader(apiHeader)

	return apiHeader, nil
}

func (api *PovApi) BatchGetHeadersByHeight(height uint64, count uint64, asc bool) (*PovApiBatchHeader, error) {
	var dbHeaders []*types.PovHeader
	var err error
	if asc {
		dbHeaders, err = api.ledger.BatchGetPovHeadersByHeightAsc(height, count)
	} else {
		dbHeaders, err = api.ledger.BatchGetPovHeadersByHeightDesc(height, count)
	}
	if err != nil {
		return nil, err
	}

	var apiHeaders []*PovApiHeader
	for _, dbHdr := range dbHeaders {
		apiHdr := &PovApiHeader{
			PovHeader: dbHdr,
		}
		api.fillHeader(apiHdr)
		apiHeaders = append(apiHeaders, apiHdr)
	}

	apiHeader := &PovApiBatchHeader{
		Count:   len(apiHeaders),
		Headers: apiHeaders,
	}

	return apiHeader, nil
}

func (api *PovApi) fillBlock(block *PovApiBlock) {
	block.AlgoEfficiency = block.GetAlgoEfficiency()
	block.AlgoName = block.GetAlgoType().String()
	block.NormBits = block.GetHeader().GetNormBits()
	block.NormDifficulty = types.CalcDifficultyRatio(block.NormBits, common.PovPowLimitBits)
	block.AlgoDifficulty = types.CalcDifficultyRatio(block.Header.BasHdr.Bits, common.PovPowLimitBits)
}

func (api *PovApi) GetBlockByHeight(height uint64, txOffset uint32, txLimit uint32) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock: block,
	}
	api.fillBlock(apiBlock)

	apiBlock.PovBlock.Body.Txs = api.pagingTxs(block.Body.Txs, txOffset, txLimit)

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
	api.fillBlock(apiBlock)

	apiBlock.PovBlock.Body.Txs = api.pagingTxs(block.Body.Txs, txOffset, txLimit)

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
	api.fillBlock(apiBlock)

	apiBlock.PovBlock.Body.Txs = api.pagingTxs(block.Body.Txs, txOffset, txLimit)

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

	apiTxl := &PovApiTxLookup{
		TxHash:   txHash,
		TxLookup: txl,
	}

	return apiTxl, nil
}

func (api *PovApi) GetTransactionByBlockHashAndIndex(blockHash types.Hash, index uint32) (*PovApiTxLookup, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	if index >= block.GetTxNum() {
		return nil, errors.New("tx index not exist")
	}
	tx := block.Body.Txs[index]

	return api.GetTransaction(tx.Hash)
}

func (api *PovApi) GetTransactionByBlockHeightAndIndex(height uint64, index uint32) (*PovApiTxLookup, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if index >= block.GetTxNum() {
		return nil, errors.New("tx index not exist")
	}
	tx := block.Body.Txs[index]

	return api.GetTransaction(tx.Hash)
}

func (api *PovApi) GetAccountState(address types.Address, stateHash types.Hash) (*PovApiState, error) {
	apiState := &PovApiState{}

	db := api.ledger.Store
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	asKey := types.PovCreateAccountStateKey(address)
	asVal := stateTrie.GetValue(asKey)
	if len(asVal) <= 0 {
		return nil, errors.New("account state value not exist")
	}

	as := types.NewPovAccountState()
	err := as.Deserialize(asVal)
	if err != nil {
		return nil, err
	}

	apiState.AccountState = as

	rsKey := types.PovCreateRepStateKey(address)
	rsVal := stateTrie.GetValue(rsKey)
	if len(rsVal) > 0 {
		rs := types.NewPovRepState()
		err = rs.Deserialize(rsVal)
		if err != nil {
			return nil, err
		}
		apiState.RepState = rs
	}

	return apiState, nil
}

func (api *PovApi) GetLatestAccountState(address types.Address) (*PovApiState, error) {
	header, err := api.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.GetStateHash())
}

func (api *PovApi) GetAccountStateByBlockHash(address types.Address, blockHash types.Hash) (*PovApiState, error) {
	header, err := api.ledger.GetPovHeaderByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.GetStateHash())
}

func (api *PovApi) GetAccountStateByBlockHeight(address types.Address, height uint64) (*PovApiState, error) {
	header, err := api.ledger.GetPovHeaderByHeight(height)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.GetStateHash())
}

func (api *PovApi) DumpBlockState(blockHash types.Hash) (*PovApiDumpState, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	stateHash := block.GetStateHash()
	dump := &PovApiDumpState{
		StateHash: stateHash,
		Accounts:  make(map[types.Address]*types.PovAccountState),
		Reps:      make(map[types.Address]*types.PovRepState),
	}

	db := api.ledger.Store
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	it := stateTrie.NewIterator([]byte{types.TriePrefixPovState})
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		if key[0] != types.TriePrefixPovState {
			continue
		}

		if key[1] == types.PovStatePrefixAcc {
			addr, err := types.PovStateKeyToAddress(key)
			if err != nil {
				return nil, err
			}

			as := types.NewPovAccountState()
			err = as.Deserialize(val)
			if err != nil {
				return nil, err
			}

			dump.Accounts[addr] = as
		}

		if key[1] == types.PovStatePrefixRep {
			addr, err := types.PovStateKeyToAddress(key)
			if err != nil {
				return nil, err
			}

			rs := types.NewPovRepState()
			err = rs.Deserialize(val)
			if err != nil {
				return nil, err
			}

			dump.Reps[addr] = rs
		}
	}

	return dump, nil
}

func (api *PovApi) GetAllRepStatesByStateHash(stateHash types.Hash) (*PovApiRepState, error) {
	apiRsp := new(PovApiRepState)

	apiRsp.StateHash = stateHash
	apiRsp.Reps = make(map[types.Address]*types.PovRepState)

	db := api.ledger.Store
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	repPrefix := types.PovCreateStatePrefix(types.PovStatePrefixRep)
	it := stateTrie.NewIterator(repPrefix)
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		addr, err := types.PovStateKeyToAddress(key)
		if err != nil {
			return nil, err
		}

		rs := types.NewPovRepState()
		err = rs.Deserialize(val)
		if err != nil {
			return nil, err
		}

		apiRsp.Reps[addr] = rs
	}

	return apiRsp, nil
}

func (api *PovApi) GetAllRepStatesByBlockHash(blockHash types.Hash) (*PovApiRepState, error) {
	header, err := api.ledger.GetPovHeaderByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return api.GetAllRepStatesByStateHash(header.GetStateHash())
}

func (api *PovApi) GetAllRepStatesByBlockHeight(blockHeight uint64) (*PovApiRepState, error) {
	header, err := api.ledger.GetPovHeaderByHeight(blockHeight)
	if err != nil {
		return nil, err
	}

	return api.GetAllRepStatesByStateHash(header.GetStateHash())
}

func (api *PovApi) GetLedgerStats() (*PovLedgerStats, error) {
	stats := &PovLedgerStats{}

	var err error
	stats.PovBlockCount, err = api.ledger.CountPovBlocks()
	if err != nil {
		return nil, err
	}

	stats.PovBestCount, err = api.ledger.CountPovBestHashs()
	if err != nil {
		return nil, err
	}

	stats.PovTxLookupCount, err = api.ledger.CountPovTxs()
	if err != nil {
		return nil, err
	}

	if stats.PovTxLookupCount > stats.PovBestCount {
		stats.PovStateTxCount = stats.PovTxLookupCount - stats.PovBestCount
	}

	stats.StateBlockCount, err = api.ledger.CountStateBlocks()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (api *PovApi) GetBlockTDByHash(blockHash types.Hash) (*PovApiTD, error) {
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
		Header: header,
		TD:     td,
	}

	return apiTD, nil
}

func (api *PovApi) GetBlockTDByHeight(height uint64) (*PovApiTD, error) {
	blockHash, err := api.ledger.GetPovBestHash(height)
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
		Header: header,
		TD:     td,
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
				item.MainRewardAmount = types.ZeroBalance
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
			item.MainRewardAmount = item.MainRewardAmount.Add(minerStat.RewardAmount)
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

		minerAddr := header.GetMinerAddr()
		if len(checkAddrMap) > 0 && checkAddrMap[minerAddr] == false {
			continue
		}

		item, ok := apiRsp.MinerStats[minerAddr]
		if !ok {
			item = &PovMinerStatItem{}
			item.MainRewardAmount = types.ZeroBalance
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
		item.MainRewardAmount = item.MainRewardAmount.Add(header.GetMinerReward())
		item.MainBlockNum += 1
		totalBlockNum += 1
	}

	// fill time
	for _, minerItem := range apiRsp.MinerStats {
		firstBlock, _ := api.ledger.GetPovHeaderByHeight(minerItem.FirstBlockHeight)
		if firstBlock != nil {
			minerItem.FirstBlockTime = time.Unix(int64(firstBlock.GetTimestamp()), 0)
		}
		lastBlock, _ := api.ledger.GetPovHeaderByHeight(minerItem.LastBlockHeight)
		if lastBlock != nil {
			minerItem.LastBlockTime = time.Unix(int64(lastBlock.GetTimestamp()), 0)
		}
	}

	apiRsp.MinerCount = len(apiRsp.MinerStats)

	apiRsp.TotalBlockNum = totalBlockNum
	apiRsp.LatestBlockHeight = latestHeader.GetHeight()

	// miner is online if it generate blocks in last N hours
	for _, item := range apiRsp.MinerStats {
		if item.LastBlockTime.Add(time.Hour).After(tmNow) {
			item.IsHourOnline = true
			apiRsp.HourOnlineCount++
		}
		if item.LastBlockTime.Add(24 * time.Hour).After(tmNow) {
			item.IsDayOnline = true
			apiRsp.DayOnlineCount++
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

type PovApiHashInfo struct {
	ChainHashPS   uint64 `json:"chainHashPS"`
	Sha256dHashPS uint64 `json:"sha256dHashPS"`
	ScryptHashPS  uint64 `json:"scryptHashPS"`
	X11HashPS     uint64 `json:"x11HashPS"`
}

func (api *PovApi) GetHashInfo(height uint64, lookup uint64) (*PovApiHashInfo, error) {
	if lookup > uint64(common.POVChainBlocksPerDay) {
		return nil, fmt.Errorf("lookup must be 0 ~ %d", common.POVChainBlocksPerDay)
	}

	if lookup%uint64(common.POVChainBlocksPerHour) != 0 {
		return nil, fmt.Errorf("lookup must be multiplier of %d", common.POVChainBlocksPerHour)
	}

	latestHdr, err := api.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	lastHdr := latestHdr
	if height > 0 && height < latestHdr.GetHeight() {
		lastHdr, err = api.ledger.GetPovHeaderByHeight(height)
		if err != nil {
			return nil, err
		}
	}

	if lastHdr == nil {
		return nil, errors.New("failed to get last block")
	}

	if lookup <= 0 {
		lookup = 120
	}
	if lookup > lastHdr.GetHeight() {
		lookup = lastHdr.GetHeight() + 1
	}

	firstHdr := lastHdr

	minTime := firstHdr.GetTimestamp()
	maxTime := firstHdr.GetTimestamp()
	for i := uint64(1); i < lookup; i++ {
		firstHdr, err = api.ledger.GetPovHeaderByHeight(firstHdr.GetHeight() - 1)
		if err != nil {
			return nil, err
		}
		if firstHdr.GetTimestamp() < minTime {
			minTime = firstHdr.GetTimestamp()
		}
		if firstHdr.GetTimestamp() > maxTime {
			maxTime = firstHdr.GetTimestamp()
		}
	}

	apiRsp := new(PovApiHashInfo)

	// In case there's a situation where minTime == maxTime, we don't want a divide by zero exception.
	if minTime == maxTime {
		return apiRsp, nil
	}
	timeDiffInt := big.NewInt(int64(maxTime - minTime))

	//api.logger.Debugf("minTime:%d, maxTime:%d, timeDiff:%s", minTime, maxTime, timeDiffInt.String())

	lastTD, err := api.ledger.GetPovTD(lastHdr.GetHash(), lastHdr.GetHeight())
	if err != nil {
		return nil, err
	}

	firstTD, err := api.ledger.GetPovTD(firstHdr.GetHash(), firstHdr.GetHeight())
	if err != nil {
		return nil, err
	}

	// chian
	{
		chainWorkDiffInt := new(big.Int).Sub(lastTD.Chain.ToBigInt(), firstTD.Chain.ToBigInt())
		apiRsp.ChainHashPS = new(big.Int).Div(chainWorkDiffInt, timeDiffInt).Uint64()
	}

	// sha256d
	{
		shaWorkDiffInt := new(big.Int).Sub(lastTD.Sha256d.ToBigInt(), firstTD.Sha256d.ToBigInt())
		apiRsp.Sha256dHashPS = new(big.Int).Div(shaWorkDiffInt, timeDiffInt).Uint64()
	}

	// scrypt
	{
		scrWorkDiffInt := new(big.Int).Sub(lastTD.Scrypt.ToBigInt(), firstTD.Scrypt.ToBigInt())
		apiRsp.ScryptHashPS = new(big.Int).Div(scrWorkDiffInt, timeDiffInt).Uint64()
	}

	// x11
	{
		x11WorkDiffInt := new(big.Int).Sub(lastTD.X11.ToBigInt(), firstTD.X11.ToBigInt())
		apiRsp.X11HashPS = new(big.Int).Div(x11WorkDiffInt, timeDiffInt).Uint64()
	}

	return apiRsp, nil
}

type PovApiGetWork struct {
	WorkHash      types.Hash     `json:"workHash"`
	Version       uint32         `json:"version"`
	Previous      types.Hash     `json:"previous"`
	Bits          uint32         `json:"bits"`
	Height        uint64         `json:"height"`
	MinTime       uint32         `json:"minTime"`
	MerkleBranch  []*types.Hash  `json:"merkleBranch"`
	CoinBaseData1 types.HexBytes `json:"coinbaseData1"`
	CoinBaseData2 types.HexBytes `json:"coinbaseData2"`
}

type PovApiSubmitWork struct {
	WorkHash  types.Hash `json:"workHash"`
	BlockHash types.Hash `json:"blockHash"`

	MerkleRoot    types.Hash     `json:"merkleRoot"`
	Timestamp     uint32         `json:"timestamp"`
	Nonce         uint32         `json:"nonce"`
	CoinbaseExtra types.HexBytes `json:"coinbaseExtra"`
	CoinbaseHash  types.Hash     `json:"coinbaseHash"`

	AuxPow *types.PovAuxHeader `json:"auxPow"`
}

func (api *PovApi) StartMining(minerAddr types.Address, algoName string) error {
	if !api.cfg.PoV.PovEnabled {
		return errors.New("pov service is disabled")
	}

	inArgs := make(map[interface{}]interface{})
	inArgs["minerAddr"] = minerAddr
	inArgs["algoName"] = algoName

	outArgs := make(map[interface{}]interface{})
	api.eb.Publish(common.EventRpcSyncCall, "Miner.StartMining", inArgs, outArgs)

	err, ok := outArgs["err"]
	if !ok {
		return errors.New("api not support")
	}
	if err != nil {
		err := outArgs["err"].(error)
		return err
	}

	return nil
}

func (api *PovApi) StopMining() error {
	if !api.cfg.PoV.PovEnabled {
		return errors.New("pov service is disabled")
	}

	inArgs := make(map[interface{}]interface{})

	outArgs := make(map[interface{}]interface{})
	api.eb.Publish(common.EventRpcSyncCall, "Miner.StopMining", inArgs, outArgs)

	err, ok := outArgs["err"]
	if !ok {
		return errors.New("api not support")
	}
	if err != nil {
		err := outArgs["err"].(error)
		return err
	}

	return nil
}

type PovApiGetMiningInfo struct {
	SyncState        int             `json:"syncState"`
	MinerAddr        string          `json:"minerAddr"`
	AlgoName         string          `json:"algoName"`
	AlgoEfficiency   uint            `json:"algoEfficiency"`
	CpuMining        bool            `json:"cpuMining"`
	CurrentBlockHash types.Hash      `json:"currentBlockHash"`
	CurrentBlockSize uint32          `json:"currentBlockSize"`
	CurrentBlockTx   uint32          `json:"currentBlockTx"`
	PooledTx         uint32          `json:"pooledTx"`
	BlockNum         uint64          `json:"blockNum"`
	Difficulty       float64         `json:"difficulty"`
	HashInfo         *PovApiHashInfo `json:"hashInfo"`
}

func (api *PovApi) GetMiningInfo() (*PovApiGetMiningInfo, error) {
	inArgs := make(map[interface{}]interface{})

	outArgs := make(map[interface{}]interface{})
	api.eb.Publish(common.EventRpcSyncCall, "Miner.GetMiningInfo", inArgs, outArgs)

	err, ok := outArgs["err"]
	if !ok {
		return nil, errors.New("api not support")
	}
	if err != nil {
		err := outArgs["err"].(error)
		return nil, err
	}

	hashInfo, err := api.GetHashInfo(0, 0)
	if err != nil {
		return nil, err.(error)
	}

	latestBlock := outArgs["latestBlock"].(*types.PovBlock)
	minerAddr := outArgs["minerAddr"].(types.Address)
	algoType := outArgs["minerAlgo"].(types.PovAlgoType)

	apiRsp := new(PovApiGetMiningInfo)
	apiRsp.SyncState = outArgs["syncState"].(int)
	if !minerAddr.IsZero() {
		apiRsp.MinerAddr = minerAddr.String()
	}
	if algoType != types.ALGO_UNKNOWN {
		apiRsp.AlgoName = algoType.String()
	}
	apiRsp.AlgoEfficiency = latestBlock.GetAlgoEfficiency()
	apiRsp.CpuMining = outArgs["cpuMining"].(bool)

	apiRsp.BlockNum = latestBlock.GetHeight()
	apiRsp.CurrentBlockHash = latestBlock.GetHash()
	apiRsp.CurrentBlockSize = uint32(latestBlock.Msgsize())
	apiRsp.CurrentBlockTx = latestBlock.GetTxNum()

	apiRsp.PooledTx = outArgs["pooledTx"].(uint32)

	apiRsp.Difficulty = types.CalcDifficultyRatio(latestBlock.Header.GetNormBits(), common.PovPowLimitBits)
	apiRsp.HashInfo = hashInfo

	return apiRsp, nil
}

func (api *PovApi) GetWork(minerAddr types.Address, algoName string) (*PovApiGetWork, error) {
	if !api.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	ss := api.syncState.Load().(common.SyncState)
	if ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}

	inArgs := make(map[interface{}]interface{})
	inArgs["minerAddr"] = minerAddr
	inArgs["algoName"] = algoName
	outArgs := make(map[interface{}]interface{})
	api.eb.Publish(common.EventRpcSyncCall, "Miner.GetWork", inArgs, outArgs)

	err, ok := outArgs["err"]
	if !ok {
		return nil, errors.New("api not support")
	}
	if err != nil {
		err := outArgs["err"].(error)
		return nil, err
	}

	mineBlock := outArgs["mineBlock"].(*types.PovMineBlock)

	apiRsp := new(PovApiGetWork)
	apiRsp.WorkHash = mineBlock.WorkHash

	apiRsp.Version = mineBlock.Header.GetVersion()
	apiRsp.Previous = mineBlock.Header.GetPrevious()
	apiRsp.Bits = mineBlock.Header.GetBits()
	apiRsp.Height = mineBlock.Header.GetHeight()

	apiRsp.MinTime = mineBlock.MinTime
	apiRsp.MerkleBranch = mineBlock.CoinbaseBranch

	apiRsp.CoinBaseData1 = mineBlock.Header.CbTx.GetCoinBaseData1()
	apiRsp.CoinBaseData2 = mineBlock.Header.CbTx.GetCoinBaseData2()

	return apiRsp, nil
}

func (api *PovApi) SubmitWork(work *PovApiSubmitWork) error {
	if !api.cfg.PoV.PovEnabled {
		return errors.New("pov service is disabled")
	}

	ss := api.syncState.Load().(common.SyncState)
	if ss != common.SyncDone {
		return errors.New("pov sync is not finished, please check it")
	}

	mineResult := types.NewPovMineResult()
	mineResult.WorkHash = work.WorkHash
	mineResult.BlockHash = work.BlockHash

	mineResult.MerkleRoot = work.MerkleRoot
	mineResult.Timestamp = work.Timestamp
	mineResult.Nonce = work.Nonce
	mineResult.CoinbaseExtra = work.CoinbaseExtra
	mineResult.CoinbaseHash = work.CoinbaseHash

	mineResult.AuxPow = work.AuxPow

	inArgs := make(map[interface{}]interface{})
	inArgs["mineResult"] = mineResult

	outArgs := make(map[interface{}]interface{})
	api.eb.Publish(common.EventRpcSyncCall, "Miner.SubmitWork", inArgs, outArgs)

	err, ok := outArgs["err"]
	if !ok {
		return errors.New("api not support")
	}
	if err != nil {
		err := outArgs["err"].(error)
		return err
	}

	return nil
}

type PovApiGetLastNHourItem struct {
	Hour uint32

	AllBlockNum    uint32
	AllTxNum       uint32
	AllMinerReward types.Balance
	AllRepReward   types.Balance

	Sha256dBlockNum uint32
	X11BlockNum     uint32
	ScryptBlockNum  uint32

	MaxTxPerBlock uint32
	MinTxPerBlock uint32
	AvgTxPerBlock uint32
}

type PovApiGetLastNHourInfo struct {
	MaxTxPerBlock uint32
	MinTxPerBlock uint32
	AvgTxPerBlock uint32

	MaxTxPerHour uint32
	MinTxPerHour uint32
	AvgTxPerHour uint32

	MaxBlockPerHour uint32
	MinBlockPerHour uint32
	AvgBlockPerHour uint32

	AllBlockNum uint32
	AllTxNum    uint32

	Sha256dBlockNum uint32
	X11BlockNum     uint32
	ScryptBlockNum  uint32

	HourItemList []*PovApiGetLastNHourItem
}

func (api *PovApi) GetLastNHourInfo(beginTime uint32, endTime uint32) (*PovApiGetLastNHourInfo, error) {
	endHourTime := uint32(time.Now().Unix())
	beginHourTime := endHourTime - (3600 * 24)

	if beginTime != 0 || endTime != 0 {
		if beginTime >= endTime {
			return nil, errors.New("endTime must be greater than beginTime")
		}
		paraDiffTime := endTime - beginTime
		if paraDiffTime%(2*3600) != 0 {
			return nil, errors.New("(endTime - beginTime) must be multiplier of 2 hour")
		}
		if paraDiffTime < 4*3600 || paraDiffTime > 24*3600 {
			return nil, errors.New("(endTime - beginTime) must be 4 ~ 24 hour")
		}

		endHourTime = endTime
		beginHourTime = beginTime
	}

	minBeginHourTime := beginHourTime - 3600

	latestHeader, err := api.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	apiRsp := new(PovApiGetLastNHourInfo)

	hourItemMap := make(map[uint32]*PovApiGetLastNHourItem)
	maxDiffHour := uint32(0)

	header := latestHeader
	for {
		if header == nil {
			break
		}
		if header.GetTimestamp() < minBeginHourTime {
			break
		}

		if header.GetTimestamp() >= beginHourTime && header.GetTimestamp() <= endHourTime {
			diffTime := endHourTime - header.GetTimestamp()
			diffHour := diffTime / 3600
			if diffHour > maxDiffHour {
				maxDiffHour = diffHour
			}
			hourItem := hourItemMap[diffHour]
			if hourItem == nil {
				hourItem = new(PovApiGetLastNHourItem)
				hourItem.Hour = diffHour
				hourItem.AllMinerReward = types.NewBalance(0)
				hourItem.AllRepReward = types.NewBalance(0)
				hourItemMap[diffHour] = hourItem
			}

			if hourItem.MaxTxPerBlock == 0 || hourItem.MaxTxPerBlock < header.CbTx.TxNum {
				hourItem.MaxTxPerBlock = header.CbTx.TxNum
			}
			if hourItem.MinTxPerBlock == 0 || hourItem.MinTxPerBlock > header.CbTx.TxNum {
				hourItem.MinTxPerBlock = header.CbTx.TxNum
			}

			hourItem.AllBlockNum += 1
			hourItem.AllTxNum += header.CbTx.TxNum

			minerTxOut := header.CbTx.GetMinerTxOut()
			repTxOut := header.CbTx.GetRepTxOut()
			if minerTxOut != nil {
				hourItem.AllMinerReward = hourItem.AllMinerReward.Add(minerTxOut.Value)
			}
			if repTxOut != nil {
				hourItem.AllRepReward = hourItem.AllRepReward.Add(repTxOut.Value)
			}

			algoType := header.GetAlgoType()
			if algoType == types.ALGO_SHA256D {
				hourItem.Sha256dBlockNum++
			} else if algoType == types.ALGO_X11 {
				hourItem.X11BlockNum++
			} else if algoType == types.ALGO_SCRYPT {
				hourItem.ScryptBlockNum++
			}
		}

		header, err = api.ledger.GetPovHeaderByHeight(header.GetHeight() - 1)
	}

	for hourIdx := uint32(0); hourIdx <= maxDiffHour; hourIdx++ {
		hourItem := hourItemMap[hourIdx]
		if hourItem == nil {
			continue
		}
		hourItem.AvgTxPerBlock = hourItem.AllTxNum / hourItem.AllBlockNum
		apiRsp.HourItemList = append(apiRsp.HourItemList, hourItem)

		apiRsp.AllTxNum += hourItem.AllTxNum
		apiRsp.AllBlockNum += hourItem.AllBlockNum

		if apiRsp.MaxTxPerBlock == 0 || apiRsp.MaxTxPerBlock < hourItem.MaxTxPerBlock {
			apiRsp.MaxTxPerBlock = hourItem.MaxTxPerBlock
		}
		if apiRsp.MinTxPerBlock == 0 || apiRsp.MinTxPerBlock > hourItem.MinTxPerBlock {
			apiRsp.MinTxPerBlock = hourItem.MinTxPerBlock
		}
		if apiRsp.MaxTxPerHour == 0 || apiRsp.MaxTxPerHour < hourItem.AllTxNum {
			apiRsp.MaxTxPerHour = hourItem.AllTxNum
		}
		if apiRsp.MinTxPerHour == 0 || apiRsp.MinTxPerHour > hourItem.AllTxNum {
			apiRsp.MinTxPerHour = hourItem.AllTxNum
		}
		if apiRsp.MaxBlockPerHour == 0 || apiRsp.MaxBlockPerHour < hourItem.AllBlockNum {
			apiRsp.MaxBlockPerHour = hourItem.AllBlockNum
		}
		if apiRsp.MinBlockPerHour == 0 || apiRsp.MinBlockPerHour > hourItem.AllBlockNum {
			apiRsp.MinBlockPerHour = hourItem.AllBlockNum
		}

		apiRsp.Sha256dBlockNum += hourItem.Sha256dBlockNum
		apiRsp.X11BlockNum += hourItem.X11BlockNum
		apiRsp.ScryptBlockNum += hourItem.ScryptBlockNum
	}
	apiRsp.AvgTxPerBlock = apiRsp.AllTxNum / apiRsp.AllBlockNum
	apiRsp.AvgTxPerHour = apiRsp.AllTxNum / (maxDiffHour + 1)
	apiRsp.AvgBlockPerHour = apiRsp.AllBlockNum / (maxDiffHour + 1)

	return apiRsp, nil
}
