package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/qlcchain/go-qlc/common/storage"

	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
)

type PovApi struct {
	cfg    *config.Config
	l      ledger.Store
	logger *zap.SugaredLogger
	eb     event.EventBus
	feb    *event.FeedEventBus
	pubsub *PovSubscription
	cc     *chainctx.ChainContext
}

type PovStatus struct {
	PovEnabled   bool   `json:"povEnabled"`
	SyncState    int    `json:"syncState"`
	SyncStateStr string `json:"syncStateStr"`
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
	AccountState  *types.PovAccountState  `json:"accountState"`
	RepState      *types.PovRepState      `json:"repState"`
	ContractState *types.PovContractState `json:"contractState"`
}

type PovApiDumpState struct {
	StateHash types.Hash                                `json:"stateHash"`
	Accounts  map[types.Address]*types.PovAccountState  `json:"accounts"`
	Reps      map[types.Address]*types.PovRepState      `json:"reps"`
	Contracts map[types.Address]*types.PovContractState `json:"contracts"`
}

type PovApiRepState struct {
	StateHash types.Hash                           `json:"stateHash"`
	Reps      map[types.Address]*types.PovRepState `json:"reps"`
}

type PovApiKeyValPair struct {
	Key   types.HexBytes `json:"key"`
	Value types.HexBytes
}

type PovApiContractState struct {
	StateHash types.Hash          `json:"stateHash"`
	CodeHash  types.Hash          `json:"codeHash"`
	KVNum     int                 `json:"kvNum"`
	AllKVs    [][2]types.HexBytes `json:"allKVs"`
}

type PovApiTxLookup struct {
	TxHash   types.Hash         `json:"txHash"`
	TxLookup *types.PovTxLookup `json:"txLookup"`

	CoinbaseTx *types.PovCoinBaseTx `json:"coinbaseTx"`
	AccountTx  *types.StateBlock    `json:"accountTx"`
}

type PovLedgerStats struct {
	PovBlockCount   uint64 `json:"povBlockCount"`
	PovBestCount    uint64 `json:"povBestCount"`
	PovAllTxCount   uint64 `json:"povAllTxCount"`
	PovCbTxCount    uint64 `json:"povCbTxCount"`
	PovStateTxCount uint64 `json:"povStateTxCount"`
	StateBlockCount uint64 `json:"stateBlockCount"`
}

type PovApiTD struct {
	Header *types.PovHeader `json:"header"`
	TD     *types.PovTD     `json:"td"`
}

type PovMinerStatItem struct {
	Account            types.Address `json:"account"`
	MainBlockNum       uint32        `json:"mainBlockNum"`
	MainRewardAmount   types.Balance `json:"mainRewardAmount"`
	StableBlockNum     uint32        `json:"stableBlockNum"`
	StableRewardAmount types.Balance `json:"stableRewardAmount"`
	FirstBlockTime     time.Time     `json:"firstBlockTime"`
	LastBlockTime      time.Time     `json:"lastBlockTime"`
	FirstBlockHeight   uint64        `json:"firstBlockHeight"`
	LastBlockHeight    uint64        `json:"lastBlockHeight"`
	IsHourOnline       bool          `json:"isHourOnline"`
	IsDayOnline        bool          `json:"isDayOnline"`
}

type PovMinerStats struct {
	MinerCount      int `json:"minerCount"`
	HourOnlineCount int `json:"hourOnlineCount"`
	DayOnlineCount  int `json:"dayOnlineCount"`

	MinerStats map[types.Address]*PovMinerStatItem `json:"minerStats"`

	TotalBlockNum     uint32        `json:"totalBlockNum"`
	TotalRewardAmount types.Balance `json:"totalRewardAmount"`
	TotalMinerReward  types.Balance `json:"totalMinerReward"`
	TotalRepReward    types.Balance `json:"totalRepReward"`
	LatestBlockHeight uint64        `json:"latestBlockHeight"`
}

type PovRepStatItem struct {
	MainBlockNum       uint32        `json:"mainBlockNum"`
	MainRewardAmount   types.Balance `json:"mainRewardAmount"`
	MainOnlinePeriod   uint32        `json:"mainOnlinePeriod"`
	StableBlockNum     uint32        `json:"stableBlockNum"`
	StableRewardAmount types.Balance `json:"stableRewardAmount"`
	StableOnlinePeriod uint32        `json:"stableOnlinePeriod"`
	LastOnlineTime     time.Time     `json:"lastOnlineTime"`
	LastOnlineHeight   uint64        `json:"lastOnlineHeight"`
	IsOnline           bool          `json:"isOnline"`
}

type PovRepStats struct {
	RepCount          uint32                            `json:"repCount"`
	RepStats          map[types.Address]*PovRepStatItem `json:"repStats"`
	TotalBlockNum     uint32                            `json:"totalBlockNum"`
	TotalPeriod       uint32                            `json:"totalPeriod"`
	TotalRewardAmount types.Balance                     `json:"totalRewardAmount"`
	LatestBlockHeight uint64                            `json:"latestBlockHeight"`
}

func NewPovApi(ctx context.Context, cfg *config.Config, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *PovApi {
	api := &PovApi{
		cfg:    cfg,
		l:      l,
		eb:     eb,
		feb:    cc.FeedEventBus(),
		pubsub: NewPovSubscription(ctx, eb),
		logger: log.NewLogger("rpc/pov"),
		cc:     cc,
	}
	return api
}

func (api *PovApi) GetPovStatus() (*PovStatus, error) {
	apiRsp := new(PovStatus)
	apiRsp.PovEnabled = api.cfg.PoV.PovEnabled
	ss := api.cc.PoVState()
	apiRsp.SyncState = int(ss)
	apiRsp.SyncStateStr = ss.String()
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
	blockHash, err := api.l.GetPovBestHash(height)
	if err != nil {
		return nil, err
	}

	header, err := api.l.GetPovHeader(height, blockHash)
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
	height, err := api.l.GetPovHeight(blockHash)
	if err != nil {
		return nil, err
	}

	header, err := api.l.GetPovHeader(height, blockHash)
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
	header, err := api.l.GetLatestPovHeader()
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

	if !api.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	var header *types.PovHeader

	latestHeader, err := api.l.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	if gap > 0 && latestHeader.GetHeight() > gap {
		height := latestHeader.GetHeight() - gap
		header, err = api.l.GetPovHeaderByHeight(height)
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
	if count <= 0 || count > uint64(common.POVChainBlocksPerDay) {
		return nil, fmt.Errorf("count should be 1 ~ %d", common.POVChainBlocksPerDay)
	}

	var dbHeaders []*types.PovHeader
	var err error
	if asc {
		dbHeaders, err = api.l.BatchGetPovHeadersByHeightAsc(height, count)
	} else {
		dbHeaders, err = api.l.BatchGetPovHeadersByHeightDesc(height, count)
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
	block, err := api.l.GetPovBlockByHeight(height)
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
	block, err := api.l.GetPovBlockByHash(blockHash)
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
	block, err := api.l.GetLatestPovBlock()
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
	txl, err := api.l.GetPovTxLookup(txHash)
	if err != nil {
		return nil, err
	}

	apiTxl := &PovApiTxLookup{
		TxHash:   txHash,
		TxLookup: txl,
	}

	if txl.TxIndex == 0 {
		header, _ := api.l.GetPovHeaderByHash(txl.BlockHash)
		if header != nil {
			apiTxl.CoinbaseTx = header.CbTx
		}
	} else {
		apiTxl.AccountTx, _ = api.l.GetStateBlockConfirmed(txHash)
	}

	return apiTxl, nil
}

func (api *PovApi) GetTransactionByBlockHashAndIndex(blockHash types.Hash, index uint32) (*PovApiTxLookup, error) {
	block, err := api.l.GetPovBlockByHash(blockHash)
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
	block, err := api.l.GetPovBlockByHeight(height)
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
	stateExist := false

	gsdb := statedb.NewPovGlobalStateDB(api.l.DBStore(), stateHash)

	as, _ := gsdb.GetAccountState(address)
	if as != nil {
		stateExist = true
		apiState.AccountState = as
	}

	rs, _ := gsdb.GetRepState(address)
	if rs != nil {
		stateExist = true
		apiState.RepState = rs
	}

	cs, _ := gsdb.GetContractState(address)
	if cs != nil {
		stateExist = true
		apiState.ContractState = cs
	}

	if !stateExist {
		return nil, errors.New("account state value not exist")
	}

	return apiState, nil
}

func (api *PovApi) GetLatestAccountState(address types.Address) (*PovApiState, error) {
	header, err := api.l.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.GetStateHash())
}

func (api *PovApi) GetAccountStateByBlockHash(address types.Address, blockHash types.Hash) (*PovApiState, error) {
	header, err := api.l.GetPovHeaderByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.GetStateHash())
}

func (api *PovApi) GetAccountStateByBlockHeight(address types.Address, height uint64) (*PovApiState, error) {
	header, err := api.l.GetPovHeaderByHeight(height)
	if err != nil {
		return nil, err
	}

	return api.GetAccountState(address, header.GetStateHash())
}

func (api *PovApi) DumpBlockState(blockHash types.Hash) (*PovApiDumpState, error) {
	block, err := api.l.GetPovBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	stateHash := block.GetStateHash()
	dump := &PovApiDumpState{
		StateHash: stateHash,
		Accounts:  make(map[types.Address]*types.PovAccountState),
		Reps:      make(map[types.Address]*types.PovRepState),
	}

	db := api.l.DBStore()
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	it := stateTrie.NewIterator([]byte{storage.KeyPrefixTriePovState})
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		if key[0] != storage.KeyPrefixTriePovState {
			continue
		}

		if key[1] == statedb.PovGlobalStatePrefixAcc {
			addr, err := statedb.PovStateKeyToAddress(key)
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

		if key[1] == statedb.PovGlobalStatePrefixRep {
			addr, err := statedb.PovStateKeyToAddress(key)
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

		if key[1] == statedb.PovGlobalStatePrefixCS {
			addr, err := statedb.PovStateKeyToAddress(key)
			if err != nil {
				return nil, err
			}

			cs := types.NewPovContractState()
			err = cs.Deserialize(val)
			if err != nil {
				return nil, err
			}

			dump.Contracts[addr] = cs
		}
	}

	return dump, nil
}

func (api *PovApi) DumpContractState(stateHash types.Hash, address types.Address) (*PovApiContractState, error) {
	gsdb := statedb.NewPovGlobalStateDB(api.l.DBStore(), stateHash)
	csdb, err := gsdb.LookupContractStateDB(address)
	if err != nil {
		return nil, err
	}

	apiCs := new(PovApiContractState)
	apiCs.StateHash = csdb.CS.StateHash
	apiCs.CodeHash = csdb.CS.CodeHash

	csTrie := csdb.GetCurTrie()
	it := csTrie.NewIterator(nil)
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		apiCs.AllKVs = append(apiCs.AllKVs, [2]types.HexBytes{key, val})
	}
	apiCs.KVNum = len(apiCs.AllKVs)

	return apiCs, nil
}

func (api *PovApi) GetAllRepStatesByStateHash(stateHash types.Hash) (*PovApiRepState, error) {
	apiRsp := new(PovApiRepState)

	apiRsp.StateHash = stateHash
	apiRsp.Reps = make(map[types.Address]*types.PovRepState)

	db := api.l.DBStore()
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	repPrefix := statedb.PovCreateGlobalStateKey(statedb.PovGlobalStatePrefixRep, nil)
	it := stateTrie.NewIterator(repPrefix)
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		addr, err := statedb.PovStateKeyToAddress(key)
		if err != nil {
			return nil, err
		}

		rs := types.NewPovRepState()
		err = rs.Deserialize(val)
		if err != nil {
			return nil, err
		}

		if rs.Total.Compare(types.ZeroBalance) != types.BalanceCompBigger {
			continue
		}

		apiRsp.Reps[addr] = rs
	}

	return apiRsp, nil
}

func (api *PovApi) GetAllRepStatesByBlockHash(blockHash types.Hash) (*PovApiRepState, error) {
	header, err := api.l.GetPovHeaderByHash(blockHash)
	if err != nil {
		return nil, err
	}

	return api.GetAllRepStatesByStateHash(header.GetStateHash())
}

func (api *PovApi) GetAllRepStatesByBlockHeight(blockHeight uint64) (*PovApiRepState, error) {
	header, err := api.l.GetPovHeaderByHeight(blockHeight)
	if err != nil {
		return nil, err
	}

	return api.GetAllRepStatesByStateHash(header.GetStateHash())
}

func (api *PovApi) GetLedgerStats() (*PovLedgerStats, error) {
	stats := &PovLedgerStats{}

	var err error
	stats.PovBlockCount, err = api.l.CountPovBlocks()
	if err != nil {
		return nil, err
	}

	stats.PovBestCount, err = api.l.CountPovBestHashs()
	if err != nil {
		return nil, err
	}

	stats.PovAllTxCount, err = api.l.CountPovTxs()
	if err != nil {
		return nil, err
	}

	stats.PovCbTxCount = stats.PovBestCount
	if stats.PovAllTxCount > stats.PovCbTxCount {
		stats.PovStateTxCount = stats.PovAllTxCount - stats.PovCbTxCount
	}

	stats.StateBlockCount, err = api.l.CountStateBlocks()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (api *PovApi) GetBlockTDByHash(blockHash types.Hash) (*PovApiTD, error) {
	height, err := api.l.GetPovHeight(blockHash)
	if err != nil {
		return nil, err
	}
	header, err := api.l.GetPovHeader(height, blockHash)
	if err != nil {
		return nil, err
	}

	td, err := api.l.GetPovTD(blockHash, height)
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
	blockHash, err := api.l.GetPovBestHash(height)
	if err != nil {
		return nil, err
	}
	header, err := api.l.GetPovHeader(height, blockHash)
	if err != nil {
		return nil, err
	}

	td, err := api.l.GetPovTD(blockHash, height)
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
	totalMinerReward := types.NewBalance(0)
	totalRepReward := types.NewBalance(0)

	// scan miner stats per day
	dbDayCnt := 0
	lastDayIndex := uint32(0)
	err := api.l.GetAllPovMinerStats(func(stat *types.PovMinerDayStat) error {
		dbDayCnt++
		if stat.DayIndex > lastDayIndex {
			lastDayIndex = stat.DayIndex
		}
		for addrHex, minerStat := range stat.MinerStats {
			if minerStat.BlockNum == 0 {
				continue
			}
			minerAddr, _ := types.HexToAddress(addrHex)

			if len(checkAddrMap) > 0 && !checkAddrMap[minerAddr] {
				continue
			}

			totalMinerReward = totalMinerReward.Add(minerStat.RewardAmount)
			totalRepReward = totalRepReward.Add(minerStat.RepReward)

			// just exist rep stats in this item
			if minerStat.BlockNum == 0 {
				continue
			}

			item, ok := apiRsp.MinerStats[minerAddr]
			if !ok {
				item = &PovMinerStatItem{}
				item.Account = minerAddr
				item.MainRewardAmount = types.ZeroBalance
				item.StableRewardAmount = types.ZeroBalance
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
			item.StableRewardAmount = item.MainRewardAmount
			item.MainBlockNum += minerStat.BlockNum
			item.StableBlockNum = item.MainBlockNum
			totalBlockNum += minerStat.BlockNum
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// scan best block not in miner stats per day
	latestHeader, _ := api.l.GetLatestPovHeader()

	notStatHeightStart := uint64(0)
	if dbDayCnt > 0 {
		notStatHeightStart = uint64(lastDayIndex+1) * uint64(common.POVChainBlocksPerDay)
	}

	notStatHeightEnd := latestHeader.GetHeight()

	for height := notStatHeightStart; height <= notStatHeightEnd; height++ {
		header, _ := api.l.GetPovHeaderByHeight(height)
		if header == nil {
			break
		}

		minerAddr := header.GetMinerAddr()
		if len(checkAddrMap) > 0 && !checkAddrMap[minerAddr] {
			continue
		}

		item, ok := apiRsp.MinerStats[minerAddr]
		if !ok {
			item = &PovMinerStatItem{}
			item.Account = minerAddr
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
		totalMinerReward = totalMinerReward.Add(header.GetMinerReward())
		totalRepReward = totalRepReward.Add(header.GetRepReward())
	}

	//exclude genesis block miner
	gBlk := common.GenesisPovBlock()
	delete(apiRsp.MinerStats, gBlk.GetMinerAddr())

	// fill time
	for _, minerItem := range apiRsp.MinerStats {
		firstBlock, _ := api.l.GetPovHeaderByHeight(minerItem.FirstBlockHeight)
		if firstBlock != nil {
			minerItem.FirstBlockTime = time.Unix(int64(firstBlock.GetTimestamp()), 0)
		}
		lastBlock, _ := api.l.GetPovHeaderByHeight(minerItem.LastBlockHeight)
		if lastBlock != nil {
			minerItem.LastBlockTime = time.Unix(int64(lastBlock.GetTimestamp()), 0)
		}
	}

	apiRsp.MinerCount = len(apiRsp.MinerStats)

	apiRsp.TotalBlockNum = totalBlockNum
	apiRsp.LatestBlockHeight = latestHeader.GetHeight()

	apiRsp.TotalMinerReward = totalMinerReward
	apiRsp.TotalRepReward = totalRepReward

	apiRsp.TotalRewardAmount = types.NewBalance(0)
	apiRsp.TotalRewardAmount = apiRsp.TotalRewardAmount.Add(totalMinerReward)
	apiRsp.TotalRewardAmount = apiRsp.TotalRewardAmount.Add(totalRepReward)

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

func (api *PovApi) GetRepStats(addrs []types.Address) (*PovRepStats, error) {
	checkAddrMap := make(map[types.Address]bool)
	if len(addrs) > 0 {
		for _, addr := range addrs {
			checkAddrMap[addr] = true
		}
	}

	rspMap := &PovRepStats{
		RepCount:          0,
		RepStats:          make(map[types.Address]*PovRepStatItem),
		TotalBlockNum:     0,
		TotalPeriod:       0,
		TotalRewardAmount: types.NewBalance(0),
		LatestBlockHeight: 0,
	}

	// scan rep stats per day
	dbDayCnt := 0
	lastDayIndex := uint32(0)
	err := api.l.GetAllPovMinerStats(func(stat *types.PovMinerDayStat) error {
		dbDayCnt++
		if stat.DayIndex > lastDayIndex {
			lastDayIndex = stat.DayIndex
		}

		for addrHex, minerStat := range stat.MinerStats {
			rspMap.TotalBlockNum += minerStat.BlockNum

			if minerStat.RepBlockNum == 0 {
				continue
			}

			repAddr, _ := types.HexToAddress(addrHex)
			if len(checkAddrMap) > 0 && !checkAddrMap[repAddr] {
				continue
			}

			repStat, ok := rspMap.RepStats[repAddr]
			if !ok {
				rspMap.RepStats[repAddr] = new(PovRepStatItem)
				rspMap.RepStats[repAddr].MainRewardAmount = types.NewBalance(0)
				rspMap.RepStats[repAddr].StableRewardAmount = types.NewBalance(0)
				repStat = rspMap.RepStats[repAddr]
			}

			repStat.MainRewardAmount = repStat.MainRewardAmount.Add(minerStat.RepReward)
			repStat.StableRewardAmount = repStat.MainRewardAmount
			repStat.MainBlockNum += minerStat.RepBlockNum
			repStat.StableBlockNum = repStat.MainBlockNum
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// scan best block not in miner stats per day
	latestHeader, _ := api.l.GetLatestPovHeader()
	rspMap.LatestBlockHeight = latestHeader.GetHeight()

	notStatHeightStart := uint64(0)
	if dbDayCnt > 0 {
		notStatHeightStart = uint64(lastDayIndex+1)*uint64(common.POVChainBlocksPerDay) + common.DPosOnlinePeriod - 1
	}
	notStatHeightEnd := latestHeader.GetHeight()

	var height uint64
	for height = notStatHeightStart; height <= notStatHeightEnd; height += common.DPosOnlinePeriod {
		header, _ := api.l.GetPovHeaderByHeight(height)
		if header == nil {
			break
		}

		// calc repReward
		repStates := api.GetAllOnlineRepStates(header)
		api.logger.Debugf("get online rep states %d at block height %d", len(repStates), height)

		// calc total weight of all reps
		repWeightTotal := types.NewBalance(0)
		for _, rep := range repStates {
			repWeightTotal = repWeightTotal.Add(rep.CalcTotal())
		}
		repWeightTotalBig := big.NewInt(repWeightTotal.Int64())

		// calc total reward of all blocks in period
		var i uint64
		for i = 0; i < common.DPosOnlinePeriod; i++ {
			rspMap.TotalBlockNum++

			povHead, err := api.l.GetPovHeaderByHeight(height - i)
			if err != nil {
				api.logger.Warnf("failed to get pov header %d, err %s", height-i, err)
				break
			}
			repRewardAllBig := povHead.GetRepReward()

			// divide reward to each rep
			for _, rep := range repStates {
				if len(checkAddrMap) > 0 && !checkAddrMap[rep.Account] {
					continue
				}

				// repReward = totalReward / repWeight * totalWeight
				repRewardBig := big.NewInt(rep.CalcTotal().Int64())
				repRewardBig = repRewardBig.Mul(repRewardBig, repRewardAllBig.Int)
				amountBig := repRewardBig.Div(repRewardBig, repWeightTotalBig)
				amount := types.NewBalance(amountBig.Int64())

				repStat, ok := rspMap.RepStats[rep.Account]
				if !ok {
					rspMap.RepStats[rep.Account] = new(PovRepStatItem)
					rspMap.RepStats[rep.Account].MainRewardAmount = types.NewBalance(0)
					rspMap.RepStats[rep.Account].StableRewardAmount = types.NewBalance(0)
					repStat = rspMap.RepStats[rep.Account]
				}

				repStat.MainBlockNum++
				repStat.MainRewardAmount = repStat.MainRewardAmount.Add(amount)
			}
		}
	}

	lastHeader, err := api.l.GetPovHeaderByHeight(height - common.DPosOnlinePeriod)
	if err != nil {
		return nil, fmt.Errorf("get pov header[%d] err", height)
	}

	ols, err := api.l.GetOnlineRepresentations()
	if err != nil {
		return nil, fmt.Errorf("get online reps err")
	}

	rspMap.TotalPeriod = rspMap.TotalBlockNum / uint32(common.DPosOnlinePeriod)
	for acc, r := range rspMap.RepStats {
		rspMap.RepCount++
		rspMap.TotalRewardAmount = rspMap.TotalRewardAmount.Add(r.MainRewardAmount)
		r.MainOnlinePeriod = r.MainBlockNum / uint32(common.DPosOnlinePeriod)
		r.StableOnlinePeriod = r.StableBlockNum / uint32(common.DPosOnlinePeriod)

		rs := api.GetRepStatesByHeightAndAccount(lastHeader, acc)
		if rs != nil {
			r.LastOnlineHeight = rs.Height / common.DPosOnlinePeriod * common.DPosOnlinePeriod

			pb, err := api.l.GetPovHeaderByHeight(r.LastOnlineHeight)
			if err != nil {
				return nil, fmt.Errorf("get pov header[%d] err", r.LastOnlineHeight)
			}

			r.LastOnlineTime = time.Unix(int64(pb.GetTimestamp()), 0)
		}

		for _, ac := range ols {
			if ac == acc {
				r.IsOnline = true
				break
			}
		}
	}

	return rspMap, nil
}

func (api *PovApi) GetMinerDayStat(dayIndex uint32) (*types.PovMinerDayStat, error) {
	dayStat, err := api.l.GetPovMinerStat(dayIndex)
	if err != nil {
		return nil, err
	}

	return dayStat, nil
}

func (api *PovApi) GetMinerDayStatByHeight(height uint64) (*types.PovMinerDayStat, error) {
	dayIndex := uint32(height / uint64(common.POVChainBlocksPerDay))

	dayStat, err := api.l.GetPovMinerStat(dayIndex)
	if err != nil {
		return nil, err
	}

	return dayStat, nil
}

func (api *PovApi) GetDiffDayStat(dayIndex uint32) (*types.PovDiffDayStat, error) {
	dayStat, err := api.l.GetPovDiffStat(dayIndex)
	if err != nil {
		return nil, err
	}

	return dayStat, nil
}

func (api *PovApi) GetDiffDayStatByHeight(height uint64) (*types.PovDiffDayStat, error) {
	dayIndex := uint32(height / uint64(common.POVChainBlocksPerDay))

	dayStat, err := api.l.GetPovDiffStat(dayIndex)
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

	latestHdr, err := api.l.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	lastHdr := latestHdr
	if height > 0 && height < latestHdr.GetHeight() {
		lastHdr, err = api.l.GetPovHeaderByHeight(height)
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
		firstHdr, err = api.l.GetPovHeaderByHeight(firstHdr.GetHeight() - 1)
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

	lastTD, err := api.l.GetPovTD(lastHdr.GetHash(), lastHdr.GetHeight())
	if err != nil {
		return nil, err
	}

	firstTD, err := api.l.GetPovTD(firstHdr.GetHash(), firstHdr.GetHeight())
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

type PovApiGetMiningInfo struct {
	SyncState          int               `json:"syncState"`
	SyncStateStr       string            `json:"syncStateStr"`
	CurrentBlockHeight uint64            `json:"currentBlockHeight"`
	CurrentBlockHash   types.Hash        `json:"currentBlockHash"`
	CurrentBlockSize   uint32            `json:"currentBlockSize"`
	CurrentBlockTx     uint32            `json:"currentBlockTx"`
	CurrentBlockAlgo   types.PovAlgoType `json:"currentBlockAlgo"`
	PooledTx           uint32            `json:"pooledTx"`
	Difficulty         float64           `json:"difficulty"`
	HashInfo           *PovApiHashInfo   `json:"hashInfo"`
}

func (api *PovApi) GetMiningInfo() (*PovApiGetMiningInfo, error) {
	inArgs := make(map[interface{}]interface{})

	outArgs := make(map[interface{}]interface{})
	api.feb.RpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Miner.GetMiningInfo", In: inArgs, Out: outArgs})

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

	apiRsp := new(PovApiGetMiningInfo)
	apiRsp.SyncState = outArgs["syncState"].(int)
	apiRsp.SyncStateStr = topic.SyncState(apiRsp.SyncState).String()

	apiRsp.CurrentBlockAlgo = latestBlock.GetAlgoType()
	apiRsp.CurrentBlockHeight = latestBlock.GetHeight()
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
	if !api.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	inArgs := make(map[interface{}]interface{})
	inArgs["minerAddr"] = minerAddr
	inArgs["algoName"] = algoName
	outArgs := make(map[interface{}]interface{})
	api.feb.RpcSyncCallWithTime(&topic.EventRPCSyncCallMsg{Name: "Miner.GetWork", In: inArgs, Out: outArgs}, 300*time.Second)

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

	if !api.cc.IsPoVDone() {
		return chainctx.ErrPoVNotFinish
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
	api.feb.RpcSyncCallWithTime(&topic.EventRPCSyncCallMsg{Name: "Miner.SubmitWork", In: inArgs, Out: outArgs}, 300*time.Second)

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
	AuxBlockNum     uint32

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
	AuxBlockNum     uint32

	HourItemList []*PovApiGetLastNHourItem
}

func (api *PovApi) GetLastNHourInfo(endHeight uint64, timeSpan uint32) (*PovApiGetLastNHourInfo, error) {
	endHourTime := uint32(0)
	beginHourTime := uint32(0)

	var err error
	var latestHeader *types.PovHeader
	if endHeight > 0 {
		latestHeader, err = api.l.GetPovHeaderByHeight(endHeight)
		if err != nil {
			return nil, err
		}
	} else {
		latestHeader, err = api.l.GetLatestPovHeader()
		if err != nil {
			return nil, err
		}
	}
	endHourTime = latestHeader.GetTimestamp()

	if timeSpan <= 0 {
		beginHourTime = endHourTime - (24 * 3600)
	} else if timeSpan >= 3600 {
		if timeSpan%(2*3600) != 0 {
			return nil, errors.New("timeSpan must be multiplier of 7200 seconds")
		}
		if (timeSpan < (2 * 3600)) || (timeSpan > (24 * 3600)) {
			return nil, errors.New("timeSpan must be in range [2 * 3600 ~ 24 * 3600] hour")
		}
		beginHourTime = endHourTime - timeSpan
	} else {
		if timeSpan%2 != 0 {
			return nil, errors.New("timeSpan must be multiplier of 2 hours")
		}
		if (timeSpan < 2) || (timeSpan > 24) {
			return nil, errors.New("timeSpan must be in range [2 ~ 24] hour")
		}
		beginHourTime = endHourTime - (timeSpan * 3600)
	}

	minBeginHourTime := beginHourTime - 3600

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

			if header.AuxHdr != nil {
				hourItem.AuxBlockNum++
			}
		}

		header, err = api.l.GetPovHeaderByHeight(header.GetHeight() - 1)
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
		apiRsp.AuxBlockNum += hourItem.AuxBlockNum
	}
	apiRsp.AvgTxPerBlock = apiRsp.AllTxNum / apiRsp.AllBlockNum
	apiRsp.AvgTxPerHour = apiRsp.AllTxNum / (maxDiffHour + 1)
	apiRsp.AvgBlockPerHour = apiRsp.AllBlockNum / (maxDiffHour + 1)

	return apiRsp, nil
}

func (api *PovApi) GetAllOnlineRepStates(header *types.PovHeader) []*types.PovRepState {
	var allRss []*types.PovRepState
	supply := config.GenesisBlock().Balance
	minVoteWeight, _ := supply.Div(common.DposVoteDivisor)

	stateHash := header.GetStateHash()
	stateTrie := trie.NewTrie(api.l.DBStore(), &stateHash, nil)
	if stateTrie == nil {
		return nil
	}

	repPrefix := statedb.PovCreateGlobalStateKey(statedb.PovGlobalStatePrefixRep, nil)
	it := stateTrie.NewIterator(repPrefix)
	if it == nil {
		return nil
	}

	key, valBytes, ok := it.Next()
	for ; ok; key, valBytes, ok = it.Next() {
		if len(valBytes) == 0 {
			continue
		}

		rs := types.NewPovRepState()
		err := rs.Deserialize(valBytes)
		if err != nil {
			api.logger.Errorf("deserialize old rep state, key %s err %s", hex.EncodeToString(key), err)
			return nil
		}

		if rs.Status != statedb.PovStatusOnline {
			continue
		}

		if rs.Height > header.GetHeight() || rs.Height < (header.GetHeight()+1-common.DPosOnlinePeriod) {
			continue
		}

		if rs.CalcTotal().Compare(minVoteWeight) == types.BalanceCompSmaller {
			continue
		}

		allRss = append(allRss, rs)
	}

	return allRss
}

func (api *PovApi) GetRepStatesByHeightAndAccount(header *types.PovHeader, acc types.Address) *types.PovRepState {
	stateHash := header.GetStateHash()
	stateTrie := trie.NewTrie(api.l.DBStore(), &stateHash, nil)
	if stateTrie == nil {
		return nil
	}

	repPrefix := statedb.PovCreateGlobalStateKey(statedb.PovGlobalStatePrefixRep, nil)
	it := stateTrie.NewIterator(repPrefix)
	if it == nil {
		return nil
	}

	key, valBytes, ok := it.Next()
	for ; ok; key, valBytes, ok = it.Next() {
		if len(valBytes) == 0 {
			continue
		}

		rs := types.NewPovRepState()
		err := rs.Deserialize(valBytes)
		if err != nil {
			api.logger.Errorf("deserialize old rep state, key %s err %s", hex.EncodeToString(key), err)
			return nil
		}

		if rs.Account == acc {
			return rs
		}
	}

	return nil
}

func (api *PovApi) NewBlock(ctx context.Context) (*rpc.Subscription, error) {
	return CreatePovSubscription(ctx, func(notifier *rpc.Notifier, subscription *rpc.Subscription) {
		go func() {
			notifyCh := make(chan struct{})
			api.pubsub.addChan(subscription.ID, notifyCh)
			defer api.pubsub.removeChan(subscription.ID)

			for {
				select {
				case <-notifyCh:
					blocks := api.pubsub.fetchBlocks(subscription.ID)

					for _, block := range blocks {
						header := block.GetHeader()
						apiHdr := &PovApiHeader{PovHeader: header}
						api.fillHeader(apiHdr)
						err := notifier.Notify(subscription.ID, apiHdr)
						if err != nil {
							api.logger.Errorf("notify pov header %d/%s error: %s",
								err, header.GetHeight(), header.GetHash())
							return
						}
					}
				case err := <-subscription.Err():
					api.logger.Infof("subscription exception %s", err)
					return
				}
			}
		}()
	})
}

type PovApiCheckStateRsp struct {
	AccountStates map[types.Address]*types.PovAccountState
	AccountMetas  map[types.Address]*types.AccountMeta
	RepStates     map[types.Address]*types.PovRepState
	RepMetas      map[types.Address]*types.Benefit
}

func (api *PovApi) CheckAllAccountStates() (*PovApiCheckStateRsp, error) {
	apiRsp := new(PovApiCheckStateRsp)
	apiRsp.AccountStates = make(map[types.Address]*types.PovAccountState)
	apiRsp.AccountMetas = make(map[types.Address]*types.AccountMeta)
	apiRsp.RepStates = make(map[types.Address]*types.PovRepState)
	apiRsp.RepMetas = make(map[types.Address]*types.Benefit)

	latestHdr, err := api.l.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}
	stateHash := latestHdr.GetStateHash()

	db := api.l.DBStore()
	stateTrie := trie.NewTrie(db, &stateHash, nil)

	it := stateTrie.NewIterator([]byte{storage.KeyPrefixTriePovState})
	for key, val, ok := it.Next(); ok; key, val, ok = it.Next() {
		if len(val) <= 0 {
			api.logger.Debugf("key %s got empty value", hex.EncodeToString(key))
			continue
		}

		if key[0] != storage.KeyPrefixTriePovState {
			continue
		}

		if key[1] == statedb.PovGlobalStatePrefixAcc {
			addr, err := statedb.PovStateKeyToAddress(key)
			if err != nil {
				return nil, err
			}

			as := types.NewPovAccountState()
			err = as.Deserialize(val)
			if err != nil {
				return nil, err
			}

			chkOk := true
			am, _ := api.l.GetAccountMeta(addr)
			if am != nil {
				if am.CoinBalance.Compare(as.Balance) != types.BalanceCompEqual {
					chkOk = false
				}
				if am.CoinVote.Compare(as.Vote) != types.BalanceCompEqual {
					chkOk = false
				}
				if am.CoinOracle.Compare(as.Oracle) != types.BalanceCompEqual {
					chkOk = false
				}
				if am.CoinNetwork.Compare(as.Network) != types.BalanceCompEqual {
					chkOk = false
				}
				if am.CoinStorage.Compare(as.Storage) != types.BalanceCompEqual {
					chkOk = false
				}

				for _, tm := range am.Tokens {
					ts := as.GetTokenState(tm.Type)
					if ts != nil {
						if ts.Balance.Compare(tm.Balance) != types.BalanceCompEqual {
							chkOk = false
						}
						if ts.Representative != tm.Representative {
							chkOk = false
						}
					}
				}
			} else {
				chkOk = false
			}

			if !chkOk {
				apiRsp.AccountStates[addr] = as
				apiRsp.AccountMetas[addr] = am
			}
		}

		if key[1] == statedb.PovGlobalStatePrefixRep {
			addr, err := statedb.PovStateKeyToAddress(key)
			if err != nil {
				return nil, err
			}

			rs := types.NewPovRepState()
			err = rs.Deserialize(val)
			if err != nil {
				return nil, err
			}

			chkOk := true
			rm, _ := api.l.GetRepresentation(addr)
			if rm != nil {
				if rm.Balance.Compare(rs.Balance) != types.BalanceCompEqual {
					chkOk = false
				}
				if rm.Vote.Compare(rs.Vote) != types.BalanceCompEqual {
					chkOk = false
				}
				if rm.Oracle.Compare(rs.Oracle) != types.BalanceCompEqual {
					chkOk = false
				}
				if rm.Network.Compare(rs.Network) != types.BalanceCompEqual {
					chkOk = false
				}
				if rm.Storage.Compare(rs.Storage) != types.BalanceCompEqual {
					chkOk = false
				}
				if rm.Total.Compare(rs.Total) != types.BalanceCompEqual {
					chkOk = false
				}
			} else {
				chkOk = false
			}

			if !chkOk {
				apiRsp.RepStates[addr] = rs
				apiRsp.RepMetas[addr] = rm
			}
		}
	}

	return apiRsp, nil
}
