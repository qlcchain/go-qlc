package api

import (
	"encoding/hex"
	"errors"
	"github.com/qlcchain/go-qlc/config"
	"math/big"
	"sync/atomic"
	"time"

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
	AlgoName       string `json:"algoName"`
	AlgoEfficiency uint   `json:"algoEfficiency"`
}

type PovApiBatchHeader struct {
	Count   int             `json:"count"`
	Headers []*PovApiHeader `json:"headers"`
}

type PovApiBlock struct {
	*types.PovBlock
	AlgoName       string `json:"algoName"`
	AlgoEfficiency uint   `json:"algoEfficiency"`
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
	IsOnline         bool          `json:"isOnline"`
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
		PovHeader:      header,
		AlgoEfficiency: header.GetAlgoEfficiency(),
		AlgoName:       header.GetAlgoType().String(),
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
		PovHeader:      header,
		AlgoEfficiency: header.GetAlgoEfficiency(),
		AlgoName:       header.GetAlgoType().String(),
	}

	return apiHeader, nil
}

func (api *PovApi) GetLatestHeader() (*PovApiHeader, error) {
	header, err := api.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	apiHeader := &PovApiHeader{
		PovHeader:      header,
		AlgoEfficiency: header.GetAlgoEfficiency(),
		AlgoName:       header.GetAlgoType().String(),
	}

	return apiHeader, nil
}

func (api *PovApi) GetFittestHeader(gap uint64) (*PovApiHeader, error) {
	if !api.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

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
		PovHeader:      header,
		AlgoEfficiency: header.GetAlgoEfficiency(),
		AlgoName:       header.GetAlgoType().String(),
	}

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
			PovHeader:      dbHdr,
			AlgoEfficiency: dbHdr.GetAlgoEfficiency(),
			AlgoName:       dbHdr.GetAlgoType().String(),
		}
		apiHeaders = append(apiHeaders, apiHdr)
	}

	apiHeader := &PovApiBatchHeader{
		Count:   len(apiHeaders),
		Headers: apiHeaders,
	}

	return apiHeader, nil
}

func (api *PovApi) GetBlockByHeight(height uint64, txOffset uint32, txLimit uint32) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock:       block,
		AlgoEfficiency: block.GetAlgoEfficiency(),
		AlgoName:       block.GetAlgoType().String(),
	}

	apiBlock.PovBlock.Body.Txs = api.pagingTxs(block.Body.Txs, txOffset, txLimit)

	return apiBlock, nil
}

func (api *PovApi) GetBlockByHash(blockHash types.Hash, txOffset uint32, txLimit uint32) (*PovApiBlock, error) {
	block, err := api.ledger.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock:       block,
		AlgoEfficiency: block.GetAlgoEfficiency(),
		AlgoName:       block.GetAlgoType().String(),
	}

	apiBlock.PovBlock.Body.Txs = api.pagingTxs(block.Body.Txs, txOffset, txLimit)

	return apiBlock, nil
}

func (api *PovApi) GetLatestBlock(txOffset uint32, txLimit uint32) (*PovApiBlock, error) {
	block, err := api.ledger.GetLatestPovBlock()
	if err != nil {
		return nil, err
	}

	apiBlock := &PovApiBlock{
		PovBlock:       block,
		AlgoEfficiency: block.GetAlgoEfficiency(),
		AlgoName:       block.GetAlgoType().String(),
	}

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

	as := new(types.PovAccountState)
	err := as.Deserialize(asVal)
	if err != nil {
		return nil, err
	}

	apiState.AccountState = as

	rsKey := types.PovCreateRepStateKey(address)
	rsVal := stateTrie.GetValue(rsKey)
	if len(rsVal) > 0 {
		rs := new(types.PovRepState)
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

	it := stateTrie.NewIterator(nil)
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		if key[0] == types.PovStatePrefixAcc {
			addr, err := types.BytesToAddress(key[1:])
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

		if key[0] == types.PovStatePrefixRep {
			addr, err := types.BytesToAddress(key[1:])
			if err != nil {
				return nil, err
			}

			rs := new(types.PovRepState)
			err = rs.Deserialize(val)
			if err != nil {
				return nil, err
			}

			dump.Reps[addr] = rs
		}
	}

	return dump, nil
}

func (api *PovApi) GetAllRepStatsByStateHash(stateHash types.Hash) (*PovApiRepState, error) {
	apiRsp := new(PovApiRepState)

	apiRsp.StateHash = stateHash
	apiRsp.Reps = make(map[types.Address]*types.PovRepState)

	db := api.ledger.Store
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	it := stateTrie.NewIterator([]byte{types.PovStatePrefixRep})
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		addr, err := types.BytesToAddress(key[1:])
		if err != nil {
			return nil, err
		}

		rs := new(types.PovRepState)
		err = rs.Deserialize(val)
		if err != nil {
			return nil, err
		}

		apiRsp.Reps[addr] = rs
	}

	return apiRsp, nil
}

func (api *PovApi) GetAllRepStatsByBlockHash(blockHash types.Hash) (*PovApiRepState, error) {
	header, err := api.ledger.GetPovHeaderByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return api.GetAllRepStatsByStateHash(header.GetStateHash())
}

func (api *PovApi) GetAllRepStatsByBlockHeight(blockHeight uint64) (*PovApiRepState, error) {
	header, err := api.ledger.GetPovHeaderByHeight(blockHeight)
	if err != nil {
		return nil, err
	}

	return api.GetAllRepStatsByStateHash(header.GetStateHash())
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

type PovApiHashStat struct {
	ChainHashPS   float64 `json:"chainHashPS"`
	Sha256dHashPS float64 `json:"sha256dHashPS"`
	ScryptHashPS  float64 `json:"scryptHashPS"`
	X11HashPS     float64 `json:"x11HashPS"`
}

func (api *PovApi) GetHashStats(height uint64, lookup uint64) (*PovApiHashStat, error) {
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

	if lastHdr == nil || lastHdr.GetHeight() == 0 {
		return nil, errors.New("failed to get block")
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

	apiRsp := new(PovApiHashStat)

	// In case there's a situation where minTime == maxTime, we don't want a divide by zero exception.
	if minTime == maxTime {
		return apiRsp, nil
	}
	timeDiffInt := big.NewInt(int64(maxTime - minTime))
	timeDiffFlt := new(big.Float).SetInt(timeDiffInt)

	api.logger.Debugf("minTime:%d, maxTime:%d, timeDiff:%s", minTime, maxTime, timeDiffInt.String())

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
		chainWorkDiffFlt := new(big.Float).SetInt(chainWorkDiffInt)
		apiRsp.ChainHashPS, _ = new(big.Float).Quo(chainWorkDiffFlt, timeDiffFlt).Float64()
	}

	// sha256d
	{
		shaWorkDiffInt := new(big.Int).Sub(lastTD.Sha256d.ToBigInt(), firstTD.Sha256d.ToBigInt())
		shaWorkDiffFlt := new(big.Float).SetInt(shaWorkDiffInt)
		apiRsp.Sha256dHashPS, _ = new(big.Float).Quo(shaWorkDiffFlt, timeDiffFlt).Float64()
	}

	// scrypt
	{
		scrWorkDiffInt := new(big.Int).Sub(lastTD.Scrypt.ToBigInt(), firstTD.Scrypt.ToBigInt())
		scrWorkDiffFlt := new(big.Float).SetInt(scrWorkDiffInt)
		apiRsp.ScryptHashPS, _ = new(big.Float).Quo(scrWorkDiffFlt, timeDiffFlt).Float64()
	}

	// x11
	{
		x11WorkDiffInt := new(big.Int).Sub(lastTD.X11.ToBigInt(), firstTD.X11.ToBigInt())
		x11WorkDiffFlt := new(big.Float).SetInt(x11WorkDiffInt)
		apiRsp.X11HashPS, _ = new(big.Float).Quo(x11WorkDiffFlt, timeDiffFlt).Float64()
	}

	return apiRsp, nil
}

type PovApiGetWork struct {
	WorkID string `json:"workID"`

	Version  uint32     `json:"version"`
	Previous types.Hash `json:"previous"`
	Bits     uint32     `json:"bits"`
	Height   uint64     `json:"height"`

	MinTime      uint32       `json:"minTime"`
	MerkleBranch []types.Hash `json:"merkleBranch"`
	CoinBaseData []byte       `json:"coinbaseData"`
}

type PovApiSubmitWork struct {
	WorkID    string     `json:"workID"`
	BlockHash types.Hash `json:"blockHash"`

	CoinBaseHash  types.Hash      `json:"coinbaseHash"`
	CoinBaseSig   types.Signature `json:"coinbaseSig"`
	CoinBaseExtra []byte          `json:"coinbaseExtra"`

	Timestamp uint32 `json:"timestamp"`
	Nonce     uint32 `json:"nonce"`
}

func (api *PovApi) GetWork(minerAddr types.Address, algoName string) (*PovApiGetWork, error) {
	if !api.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	ss := api.syncState.Load().(common.SyncState)
	if ss != common.Syncdone {
		return nil, errors.New("pov sync is not finished, please check it")
	}

	newBlock := types.NewPovBlock()

	apiRsp := new(PovApiGetWork)

	apiRsp.Version = newBlock.GetVersion()
	apiRsp.Previous = newBlock.GetPrevious()
	apiRsp.Bits = newBlock.GetBits()
	apiRsp.Height = newBlock.GetHeight()

	apiRsp.MinTime = newBlock.GetTimestamp()
	apiRsp.CoinBaseData = newBlock.Header.CbTx.GetCoinBaseData()

	return apiRsp, nil
}

func (api *PovApi) SubmitWork(work *PovApiSubmitWork) error {
	if !api.cfg.PoV.PovEnabled {
		return errors.New("pov service is disabled")
	}

	ss := api.syncState.Load().(common.SyncState)
	if ss != common.Syncdone {
		return errors.New("pov sync is not finished, please check it")
	}

	return nil
}
