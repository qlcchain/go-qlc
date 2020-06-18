package apis

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/go-qlc/common/types"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type PovAPI struct {
	pov    *api.PovApi
	pubsub *api.PovSubscription
	logger *zap.SugaredLogger
}

func NewPovAPI(ctx context.Context, cfg *config.Config, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *PovAPI {
	return &PovAPI{
		pov:    api.NewPovApi(ctx, cfg, l, eb, cc),
		pubsub: api.NewPovSubscription(ctx, eb),
		logger: log.NewLogger("grpc_pov"),
	}
}

func (p *PovAPI) GetPovStatus(ctx context.Context, param *empty.Empty) (*pb.PovStatus, error) {
	r, err := p.pov.GetPovStatus()
	if err != nil {
		return nil, err
	}
	return toPovStatus(r), nil
}

func (p *PovAPI) GetHeaderByHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiHeader, error) {
	v := toOriginUInt64(param)
	r, err := p.pov.GetHeaderByHeight(v)
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) GetHeaderByHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiHeader, error) {
	h, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetHeaderByHash(h)
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) GetLatestHeader(ctx context.Context, param *empty.Empty) (*pb.PovApiHeader, error) {
	r, err := p.pov.GetLatestHeader()
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) GetFittestHeader(ctx context.Context, param *pb.UInt64) (*pb.PovApiHeader, error) {
	v := toOriginUInt64(param)
	r, err := p.pov.GetFittestHeader(v)
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) BatchGetHeadersByHeight(ctx context.Context, param *pb.HeadersByHeightRequest) (*pb.PovApiBatchHeader, error) {
	height := param.GetHeight()
	count := param.GetCount()
	asc := param.GetAsc()
	r, err := p.pov.BatchGetHeadersByHeight(height, count, asc)
	if err != nil {
		return nil, err
	}
	return toPovApiBatchHeader(r), nil
}

func (p *PovAPI) GetBlockByHeight(ctx context.Context, param *pb.BlockByHeightRequest) (*pb.PovApiBlock, error) {
	height := param.GetHeight()
	txOffset := param.GetTxOffset()
	txLimit := param.GetTxLimit()
	r, err := p.pov.GetBlockByHeight(height, txOffset, txLimit)
	if err != nil {
		return nil, err
	}
	return toPovApiBlock(r), nil
}

func (p *PovAPI) GetBlockByHash(ctx context.Context, param *pb.BlockByHashRequest) (*pb.PovApiBlock, error) {
	blockHash, err := toOriginHashByValue(param.GetBlockHash())
	if err != nil {
		return nil, err
	}
	txOffset := param.GetTxOffset()
	txLimit := param.GetTxLimit()
	r, err := p.pov.GetBlockByHash(blockHash, txOffset, txLimit)
	if err != nil {
		return nil, err
	}
	return toPovApiBlock(r), nil
}

func (p *PovAPI) GetLatestBlock(ctx context.Context, param *pb.LatestBlockRequest) (*pb.PovApiBlock, error) {
	txOffset := param.GetTxOffset()
	txLimit := param.GetTxLimit()
	r, err := p.pov.GetLatestBlock(txOffset, txLimit)
	if err != nil {
		return nil, err
	}
	return toPovApiBlock(r), nil
}

func (p *PovAPI) GetTransaction(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiTxLookup, error) {
	blockHash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetTransaction(blockHash)
	if err != nil {
		return nil, err
	}
	return toPovApiTxLookup(r), nil
}

func (p *PovAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, param *pb.TransactionByBlockHashRequest) (*pb.PovApiTxLookup, error) {
	blockHash, err := toOriginHashByValue(param.GetBlockHash())
	if err != nil {
		return nil, err
	}
	index := param.GetIndex()
	r, err := p.pov.GetTransactionByBlockHashAndIndex(blockHash, index)
	if err != nil {
		return nil, err
	}
	return toPovApiTxLookup(r), nil
}

func (p *PovAPI) GetTransactionByBlockHeightAndIndex(ctx context.Context, param *pb.TransactionByBlockHeightRequest) (*pb.PovApiTxLookup, error) {
	height := param.GetHeight()
	index := param.GetIndex()
	r, err := p.pov.GetTransactionByBlockHeightAndIndex(height, index)
	if err != nil {
		return nil, err
	}
	return toPovApiTxLookup(r), nil
}

func (p *PovAPI) GetAccountState(ctx context.Context, param *pb.AccountStateRequest) (*pb.PovApiState, error) {
	address, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	stateHash, err := toOriginHashByValue(param.GetStateHash())
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAccountState(address, stateHash)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetLatestAccountState(ctx context.Context, param *pbtypes.Address) (*pb.PovApiState, error) {
	address, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetLatestAccountState(address)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetAccountStateByBlockHash(ctx context.Context, param *pb.AccountStateByHashRequest) (*pb.PovApiState, error) {
	address, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	hash, err := toOriginHashByValue(param.GetBlockHash())
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAccountStateByBlockHash(address, hash)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetAccountStateByBlockHeight(ctx context.Context, param *pb.AccountStateByHeightRequest) (*pb.PovApiState, error) {
	address, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	height := param.GetHeight()
	r, err := p.pov.GetAccountStateByBlockHeight(address, height)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetAllRepStatesByStateHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiRepState, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAllRepStatesByStateHash(hash)
	if err != nil {
		return nil, err
	}
	return toPovApiRepState(r), nil
}

func (p *PovAPI) GetAllRepStatesByBlockHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiRepState, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAllRepStatesByBlockHash(hash)
	if err != nil {
		return nil, err
	}
	return toPovApiRepState(r), nil
}

func (p *PovAPI) GetAllRepStatesByBlockHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiRepState, error) {
	v := toOriginUInt64(param)
	r, err := p.pov.GetAllRepStatesByBlockHeight(v)
	if err != nil {
		return nil, err
	}
	return toPovApiRepState(r), nil

}

func (p *PovAPI) GetLedgerStats(ctx context.Context, param *empty.Empty) (*pb.PovLedgerStats, error) {
	r, err := p.pov.GetLedgerStats()
	if err != nil {
		return nil, err
	}
	return toPovLedgerStats(r), nil
}

//func (p *PovAPI) GetBlockTDByHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiTD, error) {
//	hash, err := toOriginHash(param)
//	if err != nil {
//		return nil, err
//	}
//	r, err := p.pov.GetBlockTDByHash(hash)
//	if err != nil {
//		return nil, err
//	}
//	return toPovApiTD(r), nil
//}
//
//func (p *PovAPI) GetBlockTDByHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiTD, error) {
//	v := toOriginUInt64(param)
//	r, err := p.pov.GetBlockTDByHeight(v)
//	if err != nil {
//		return nil, err
//	}
//	return toPovApiTD(r), nil
//}

func (p *PovAPI) GetMinerStats(ctx context.Context, param *pbtypes.Addresses) (*pb.PovMinerStats, error) {
	address, err := toOriginAddresses(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetMinerStats(address)
	if err != nil {
		return nil, err
	}
	return toPovMinerStats(r), nil
}

func (p *PovAPI) GetRepStats(ctx context.Context, param *pbtypes.Addresses) (*pb.PovRepStats, error) {
	address, err := toOriginAddresses(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetRepStats(address)
	if err != nil {
		return nil, err
	}
	return toPovRepStats(r), nil
}

func (p *PovAPI) GetMinerDayStat(ctx context.Context, param *pb.UInt32) (*pbtypes.PovMinerDayStat, error) {
	pa := toOriginUInt32(param)
	r, err := p.pov.GetMinerDayStat(pa)
	if err != nil {
		return nil, err
	}
	return toPovMinerDayStat(r), nil
}

func (p *PovAPI) GetMinerDayStatByHeight(ctx context.Context, param *pb.UInt64) (*pbtypes.PovMinerDayStat, error) {
	pa := toOriginUInt64(param)
	r, err := p.pov.GetMinerDayStatByHeight(pa)
	if err != nil {
		return nil, err
	}
	return toPovMinerDayStat(r), nil
}

func (p *PovAPI) GetDiffDayStat(ctx context.Context, param *pb.UInt32) (*pbtypes.PovDiffDayStat, error) {
	pa := toOriginUInt32(param)
	r, err := p.pov.GetDiffDayStat(pa)
	if err != nil {
		return nil, err
	}
	return toPovDiffDayStat(r), nil
}

func (p *PovAPI) GetDiffDayStatByHeight(ctx context.Context, param *pb.UInt64) (*pbtypes.PovDiffDayStat, error) {
	pa := toOriginUInt64(param)
	r, err := p.pov.GetDiffDayStatByHeight(pa)
	if err != nil {
		return nil, err
	}
	return toPovDiffDayStat(r), nil
}

func (p *PovAPI) GetHashInfo(ctx context.Context, param *pb.HashInfoRequest) (*pb.PovApiHashInfo, error) {
	height := param.GetHeight()
	lookup := param.GetLookup()
	r, err := p.pov.GetHashInfo(height, lookup)
	if err != nil {
		return nil, err
	}
	return toPovApiHashInfo(r), nil
}

func (p *PovAPI) StartMining(ctx context.Context, param *pb.StartMiningRequest) (*empty.Empty, error) {
	minerAddr, err := toOriginAddressByValue(param.GetMinerAddr())
	if err != nil {
		return nil, err
	}
	algoName := param.GetAlgoName()
	err = p.pov.StartMining(minerAddr, algoName)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (p *PovAPI) StopMining(ctx context.Context, param *empty.Empty) (*empty.Empty, error) {
	err := p.pov.StopMining()
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (p *PovAPI) GetMiningInfo(ctx context.Context, param *empty.Empty) (*pb.PovApiGetMiningInfo, error) {
	r, err := p.pov.GetMiningInfo()
	if err != nil {
		return nil, err
	}
	return toPovApiGetMiningInfo(r), nil
}

func (p *PovAPI) GetWork(ctx context.Context, param *pb.WorkRequest) (*pb.PovApiGetWork, error) {
	minerAddr, err := toOriginAddressByValue(param.GetMinerAddr())
	if err != nil {
		return nil, err
	}
	algoName := param.GetAlgoName()
	r, err := p.pov.GetWork(minerAddr, algoName)
	if err != nil {
		return nil, err
	}
	return toPovApiGetWork(r), nil
}

//func (p *PovAPI) SubmitWork(ctx context.Context, param *pb.PovApiSubmitWork) (*empty.Empty, error) {
//	pa, err := toOriginPovApiSubmitWork(param)
//	if err != nil {
//		return nil, err
//	}
//	err = p.pov.SubmitWork(pa)
//	if err != nil {
//		return nil, err
//	}
//	return &empty.Empty{}, nil
//}

func (p *PovAPI) GetLastNHourInfo(ctx context.Context, param *pb.LastNHourInfoRequest) (*pb.PovApiGetLastNHourInfo, error) {
	endHeight := param.GetEndHeight()
	timeSpan := param.GetTimeSpan()
	r, err := p.pov.GetLastNHourInfo(endHeight, timeSpan)
	if err != nil {
		return nil, err
	}
	return toPovApiGetLastNHourInfo(r), nil
}

//func (p *PovAPI) GetAllOnlineRepStates(ctx context.Context, param *pbtypes.PovHeader) (*pbtypes.PovRepStates, error) {
//	panic("implement me")
//}
//
//func (p *PovAPI) GetRepStatesByHeightAndAccount(ctx context.Context, param *pb.RepStatesByHeightRequset) (*pbtypes.PovRepState, error) {
//	panic("implement me")
//}

//func (p *PovAPI) CheckAllAccountStates(ctx context.Context, param *empty.Empty) (*pb.PovApiCheckStateRsp, error) {
//	panic("implement me")
//}

func (p *PovAPI) NewBlock(em *empty.Empty, srv pb.PovAPI_NewBlockServer) error {
	id := getReqId()
	ch := make(chan struct{})
	p.logger.Infof("subscription pov block done, %s", id)
	p.pubsub.AddChan(id, ch)
	defer p.pubsub.RemoveChan(id)

	for {
		select {
		case <-ch:
			blocks := p.pubsub.FetchBlocks(id)

			for _, block := range blocks {
				header := block.GetHeader()
				apiHdr := &api.PovApiHeader{PovHeader: header}
				api.FillHeader(apiHdr)

				if err := srv.Send(toPovApiHeader(apiHdr)); err != nil {
					p.logger.Errorf("notify pov header %d/%s error: %s",
						err, header.GetHeight(), header.GetHash())
					return err
				}

			}
		case <-srv.Context().Done():
			p.logger.Infof("subscription  pov block finished, %s ", id)
			return nil
		}
	}

}

func toPovStatus(s *api.PovStatus) *pb.PovStatus {
	return &pb.PovStatus{
		PovEnabled:   s.PovEnabled,
		SyncState:    int32(s.SyncState),
		SyncStateStr: s.SyncStateStr,
	}
}

func toPovApiHeader(ph *api.PovApiHeader) *pb.PovApiHeader {
	r := &pb.PovApiHeader{
		BasHdr:         toPovBaseHeader(ph.BasHdr),
		AuxHdr:         nil,
		CbTx:           nil,
		AlgoName:       ph.AlgoName,
		AlgoEfficiency: uint32(ph.AlgoEfficiency),
		NormBits:       ph.NormBits,
		NormDifficulty: ph.NormDifficulty,
		AlgoDifficulty: ph.AlgoDifficulty,
	}
	if ph.AuxHdr != nil {
		r.AuxHdr = toPovAuxHeader(ph.AuxHdr)
	}
	if ph.CbTx != nil {
		r.CbTx = toPovCoinBaseTx(ph.CbTx)
	}
	return r
}

func toPovBaseHeader(header types.PovBaseHeader) *pbtypes.PovBaseHeader {
	r := &pbtypes.PovBaseHeader{
		Version:    header.Version,
		Previous:   toHashValue(header.Previous),
		MerkleRoot: toHashValue(header.MerkleRoot),
		Timestamp:  header.Timestamp,
		Bits:       header.Bits,
		Nonce:      header.Nonce,
		Hash:       toHashValue(header.Hash),
		Height:     header.Height,
	}
	return r
}

func toPovAuxHeader(header *types.PovAuxHeader) *pbtypes.PovAuxHeader {
	aux := &pbtypes.PovAuxHeader{
		AuxMerkleBranch:   toHashesValuesByPoint(header.AuxMerkleBranch),
		AuxMerkleIndex:    int32(header.AuxMerkleIndex),
		ParCoinBaseTx:     nil,
		ParCoinBaseMerkle: toHashesValuesByPoint(header.ParCoinBaseMerkle),
		ParMerkleIndex:    int32(header.ParMerkleIndex),
		ParBlockHeader: &pbtypes.PovBtcHeader{
			Version:    header.ParBlockHeader.Version,
			Previous:   toHashValue(header.ParBlockHeader.Previous),
			MerkleRoot: toHashValue(header.ParBlockHeader.MerkleRoot),
			Timestamp:  header.ParBlockHeader.Timestamp,
			Bits:       header.ParBlockHeader.Bits,
			Nonce:      header.ParBlockHeader.Nonce,
		},
		ParentHash: toHashValue(header.ParentHash),
	}
	aux.ParCoinBaseTx = &pbtypes.PovBtcTx{
		Version:  header.ParCoinBaseTx.Version,
		TxIn:     nil,
		TxOut:    nil,
		LockTime: header.ParCoinBaseTx.LockTime,
	}
	return aux
}

func toPovCoinBaseTx(header *types.PovCoinBaseTx) *pbtypes.PovCoinBaseTx {
	cbTx := &pbtypes.PovCoinBaseTx{
		Version:   header.Version,
		TxIns:     nil,
		TxOuts:    nil,
		StateHash: toHashValue(header.StateHash),
		TxNum:     header.TxNum,
		Hash:      toHashValue(header.Hash),
	}
	if header.TxIns != nil {
		cbTx.TxIns = toPovCoinBaseTxIns(header.TxIns)
	}
	if header.TxOuts != nil {
		cbTx.TxOuts = toPovCoinBaseTxOuts(header.TxOuts)
	}
	return cbTx
}

func toPovCoinBaseTxIns(ts []*types.PovCoinBaseTxIn) []*pbtypes.PovCoinBaseTxIn {
	ti := make([]*pbtypes.PovCoinBaseTxIn, 0)
	for _, tp := range ts {
		tt := &pbtypes.PovCoinBaseTxIn{
			PrevTxHash: toHashValue(tp.PrevTxHash),
			PrevTxIdx:  tp.PrevTxIdx,
			Extra:      tp.Extra.String(),
			Sequence:   tp.Sequence,
		}
		ti = append(ti, tt)
	}
	return ti
}

func toPovCoinBaseTxOuts(ts []*types.PovCoinBaseTxOut) []*pbtypes.PovCoinBaseTxOut {
	to := make([]*pbtypes.PovCoinBaseTxOut, 0)
	for _, tp := range ts {
		tt := &pbtypes.PovCoinBaseTxOut{
			Value:   toBalanceValue(tp.Value),
			Address: toAddressValue(tp.Address),
		}
		to = append(to, tt)
	}
	return to
}

func toPovApiBatchHeader(s *api.PovApiBatchHeader) *pb.PovApiBatchHeader {
	r := &pb.PovApiBatchHeader{
		Count:   0,
		Headers: nil,
	}
	if s.Headers != nil {
		headers := make([]*pb.PovApiHeader, 0)
		for _, h := range s.Headers {
			headers = append(headers, toPovApiHeader(h))
		}
		r.Headers = headers
	}
	return r
}

func toPovApiBlock(s *api.PovApiBlock) *pb.PovApiBlock {
	r := &pb.PovApiBlock{
		Header:         nil,
		Body:           nil,
		AlgoName:       s.AlgoName,
		AlgoEfficiency: uint32(s.AlgoEfficiency),
		NormBits:       s.NormBits,
		NormDifficulty: s.NormDifficulty,
		AlgoDifficulty: s.AlgoDifficulty,
	}
	header := &pbtypes.PovHeader{}
	header.BasHdr = toPovBaseHeader(s.Header.BasHdr)
	if s.Header.AuxHdr != nil {
		header.AuxHdr = toPovAuxHeader(s.Header.AuxHdr)
	}
	if s.Header.CbTx != nil {
		header.CbTx = toPovCoinBaseTx(s.Header.CbTx)
	}
	r.Header = header
	r.Body = toPovBody(s.Body)
	return r
}

func toPovBody(ts types.PovBody) *pbtypes.PovBody {
	ps := make([]*pbtypes.PovTransaction, 0)
	for _, tp := range ts.Txs {
		pt := &pbtypes.PovTransaction{
			Hash: toHashValue(tp.Hash),
		}
		ps = append(ps, pt)
	}
	return &pbtypes.PovBody{Txs: ps}
}

func toPovApiTxLookup(s *api.PovApiTxLookup) *pb.PovApiTxLookup {
	r := &pb.PovApiTxLookup{
		TxHash:     toHashValue(s.TxHash),
		TxLookup:   nil,
		CoinbaseTx: nil,
		AccountTx:  nil,
	}
	if s.TxLookup != nil {
		r.TxLookup = &pbtypes.PovTxLookup{
			BlockHash:   toHashValue(s.TxLookup.BlockHash),
			BlockHeight: s.TxLookup.BlockHeight,
			TxIndex:     s.TxLookup.TxIndex,
		}
	}
	if s.CoinbaseTx != nil {
		ct := &pbtypes.PovCoinBaseTx{
			Version:   s.CoinbaseTx.Version,
			TxIns:     nil,
			TxOuts:    nil,
			StateHash: toHashValue(s.CoinbaseTx.StateHash),
			TxNum:     s.CoinbaseTx.TxNum,
			Hash:      toHashValue(s.CoinbaseTx.Hash),
		}
		if s.CoinbaseTx.TxIns != nil {
			ct.TxIns = toPovCoinBaseTxIns(s.CoinbaseTx.TxIns)
		}
		if s.CoinbaseTx.TxOuts != nil {
			ct.TxOuts = toPovCoinBaseTxOuts(s.CoinbaseTx.TxOuts)
		}
		r.CoinbaseTx = ct
	}
	if s.AccountTx != nil {
		r.AccountTx = toStateBlock(s.AccountTx)
	}
	return r
}

func toPovApiState(s *api.PovApiState) *pb.PovApiState {
	r := &pb.PovApiState{
		AccountState:  nil,
		RepState:      nil,
		ContractState: nil,
	}
	if s.AccountState != nil {
		as := &pbtypes.PovAccountState{
			Account:     toAddressValue(s.AccountState.Account),
			Balance:     toBalanceValue(s.AccountState.Balance),
			Vote:        toBalanceValue(s.AccountState.Vote),
			Network:     toBalanceValue(s.AccountState.Network),
			Storage:     toBalanceValue(s.AccountState.Storage),
			Oracle:      toBalanceValue(s.AccountState.Oracle),
			TokenStates: nil,
		}
		if s.AccountState.TokenStates != nil {
			ts := make([]*pbtypes.PovTokenState, 0)
			for _, tp := range s.AccountState.TokenStates {
				tt := &pbtypes.PovTokenState{
					Type:           toHashValue(tp.Type),
					Hash:           toHashValue(tp.Hash),
					Representative: toAddressValue(tp.Representative),
					Balance:        toBalanceValue(tp.Balance),
				}
				ts = append(ts, tt)
			}
			as.TokenStates = ts
		}
		r.AccountState = as
	}
	if s.RepState != nil {
		rs := &pbtypes.PovRepState{
			Account: toAddressValue(s.RepState.Account),
			Balance: toBalanceValue(s.RepState.Balance),
			Vote:    toBalanceValue(s.RepState.Vote),
			Network: toBalanceValue(s.RepState.Network),
			Storage: toBalanceValue(s.RepState.Storage),
			Oracle:  toBalanceValue(s.RepState.Oracle),
			Total:   toBalanceValue(s.RepState.Total),
			Status:  s.RepState.Status,
			Height:  s.RepState.Height,
		}
		r.RepState = rs
	}
	if s.ContractState != nil {
		cs := &pbtypes.PovContractState{
			StateHash: toHashValue(s.ContractState.StateHash),
			CodeHash:  toHashValue(s.ContractState.CodeHash),
		}
		r.ContractState = cs
	}
	return r
}

func toPovApiRepState(s *api.PovApiRepState) *pb.PovApiRepState {
	r := &pb.PovApiRepState{
		StateHash: toHashValue(s.StateHash),
		Reps:      nil,
	}
	so := make(map[string]*pbtypes.PovRepState)
	for k, v := range s.Reps {
		so[toAddressValue(k)] = &pbtypes.PovRepState{
			Account: toAddressValue(v.Account),
			Balance: toBalanceValue(v.Balance),
			Vote:    toBalanceValue(v.Vote),
			Network: toBalanceValue(v.Network),
			Storage: toBalanceValue(v.Storage),
			Oracle:  toBalanceValue(v.Oracle),
			Total:   toBalanceValue(v.Total),
			Status:  v.Status,
			Height:  v.Height,
		}
	}
	r.Reps = so
	return r
}

func toPovLedgerStats(s *api.PovLedgerStats) *pb.PovLedgerStats {
	return &pb.PovLedgerStats{
		PovBlockCount:   s.PovBlockCount,
		PovBestCount:    s.PovBestCount,
		PovAllTxCount:   s.PovAllTxCount,
		PovCbTxCount:    s.PovCbTxCount,
		PovStateTxCount: s.PovStateTxCount,
		StateBlockCount: s.StateBlockCount,
	}
}

//func toPovApiTD(s *api.PovApiTD) *pb.PovApiTD {
//	return &pb.PovApiTD{
//		Header: nil,
//		Td:     nil,
//	}
//}

func toPovMinerStats(s *api.PovMinerStats) *pb.PovMinerStats {
	r := &pb.PovMinerStats{
		MinerCount:        int32(s.MinerCount),
		HourOnlineCount:   int32(s.HourOnlineCount),
		DayOnlineCount:    int32(s.DayOnlineCount),
		MinerStats:        nil,
		TotalBlockNum:     s.TotalBlockNum,
		TotalRewardAmount: toBalanceValue(s.TotalRewardAmount),
		TotalMinerReward:  toBalanceValue(s.TotalMinerReward),
		TotalRepReward:    toBalanceValue(s.TotalRepReward),
		LatestBlockHeight: s.LatestBlockHeight,
	}
	so := make(map[string]*pb.PovMinerStatItem)
	for k, v := range s.MinerStats {
		so[toAddressValue(k)] = &pb.PovMinerStatItem{
			Account:            toAddressValue(v.Account),
			MainBlockNum:       v.MainBlockNum,
			MainRewardAmount:   toBalanceValue(v.MainRewardAmount),
			StableBlockNum:     v.StableBlockNum,
			StableRewardAmount: toBalanceValue(v.StableRewardAmount),
			FirstBlockTime:     v.FirstBlockTime.Unix(),
			LastBlockTime:      v.LastBlockTime.Unix(),
			FirstBlockHeight:   v.FirstBlockHeight,
			LastBlockHeight:    v.LastBlockHeight,
			IsHourOnline:       v.IsHourOnline,
			IsDayOnline:        v.IsDayOnline,
		}
	}
	r.MinerStats = so
	return r
}

func toPovRepStats(s *api.PovRepStats) *pb.PovRepStats {
	return &pb.PovRepStats{
		RepCount:          s.RepCount,
		RepStats:          nil,
		TotalBlockNum:     s.TotalBlockNum,
		TotalPeriod:       s.TotalPeriod,
		TotalRewardAmount: toBalanceValue(s.TotalRewardAmount),
		LatestBlockHeight: s.LatestBlockHeight,
	}
}

func toPovMinerDayStat(s *types.PovMinerDayStat) *pbtypes.PovMinerDayStat {
	r := &pbtypes.PovMinerDayStat{
		DayIndex:   s.DayIndex,
		MinerNum:   s.MinerNum,
		MinerStats: nil,
	}
	rp := make(map[string]*pbtypes.PovMinerStatItem)
	for k, v := range s.MinerStats {
		rp[k] = &pbtypes.PovMinerStatItem{
			FirstHeight:  v.FirstHeight,
			LastHeight:   v.LastHeight,
			BlockNum:     v.BlockNum,
			RewardAmount: toBalanceValue(v.RewardAmount),
			RepBlockNum:  v.RepBlockNum,
			RepReward:    toBalanceValue(v.RepReward),
			IsMiner:      v.IsMiner,
		}
	}
	r.MinerStats = rp
	return r
}

func toPovDiffDayStat(s *types.PovDiffDayStat) *pbtypes.PovDiffDayStat {
	return &pbtypes.PovDiffDayStat{
		DayIndex:     s.DayIndex,
		AvgDiffRatio: s.AvgDiffRatio,
		MaxDiffRatio: s.MaxDiffRatio,
		MinDiffRatio: s.MinDiffRatio,
		MaxBlockTime: s.MaxBlockTime,
		MinBlockTime: s.MinBlockTime,
	}
}

func toPovApiHashInfo(s *api.PovApiHashInfo) *pb.PovApiHashInfo {
	return &pb.PovApiHashInfo{
		ChainHashPS:   s.ChainHashPS,
		Sha256DHashPS: s.Sha256dHashPS,
		ScryptHashPS:  s.ScryptHashPS,
		X11HashPS:     s.X11HashPS,
	}
}

func toPovApiGetMiningInfo(s *api.PovApiGetMiningInfo) *pb.PovApiGetMiningInfo {
	r := &pb.PovApiGetMiningInfo{
		SyncState:          int32(s.SyncState),
		SyncStateStr:       s.SyncStateStr,
		CurrentBlockHeight: s.CurrentBlockHeight,
		CurrentBlockHash:   toHashValue(s.CurrentBlockHash),
		CurrentBlockSize:   s.CurrentBlockSize,
		CurrentBlockTx:     s.CurrentBlockTx,
		CurrentBlockAlgo:   s.CurrentBlockAlgo.String(),
		PooledTx:           s.PooledTx,
		Difficulty:         s.Difficulty,
		HashInfo:           nil,
		MinerAddr:          s.MinerAddr,
		AlgoName:           s.AlgoName,
		AlgoEfficiency:     uint32(s.AlgoEfficiency),
		CpuMining:          s.CpuMining,
	}
	if s.HashInfo != nil {
		r.HashInfo = toPovApiHashInfo(s.HashInfo)
	}
	return r
}

func toPovApiGetWork(s *api.PovApiGetWork) *pb.PovApiGetWork {
	return &pb.PovApiGetWork{
		WorkHash:      toHashValue(s.WorkHash),
		Version:       s.Version,
		Previous:      toHashValue(s.Previous),
		Bits:          s.Bits,
		Height:        s.Height,
		MinTime:       s.MinTime,
		MerkleBranch:  toHashesValuesByPoint(s.MerkleBranch),
		CoinBaseData1: s.CoinBaseData1.String(),
		CoinBaseData2: s.CoinBaseData2.String(),
	}
}

//func toOriginPovApiSubmitWork(param *pb.PovApiSubmitWork) (*api.PovApiSubmitWork, error) {
//	return &api.PovApiSubmitWork{
//		WorkHash:      types.Hash{},
//		BlockHash:     types.Hash{},
//		MerkleRoot:    types.Hash{},
//		Timestamp:     0,
//		Nonce:         0,
//		CoinbaseExtra: nil,
//		CoinbaseHash:  types.Hash{},
//		AuxPow:        nil,
//	}, nil
//}

func toPovApiGetLastNHourInfo(s *api.PovApiGetLastNHourInfo) *pb.PovApiGetLastNHourInfo {
	r := &pb.PovApiGetLastNHourInfo{
		MaxTxPerBlock:   s.MaxTxPerBlock,
		MinTxPerBlock:   s.MinTxPerBlock,
		AvgTxPerBlock:   s.AvgTxPerBlock,
		MaxTxPerHour:    s.MaxTxPerHour,
		MinTxPerHour:    s.MinTxPerHour,
		AvgTxPerHour:    s.AvgTxPerHour,
		MaxBlockPerHour: s.MaxBlockPerHour,
		MinBlockPerHour: s.MinBlockPerHour,
		AvgBlockPerHour: s.AvgBlockPerHour,
		AllBlockNum:     s.AllBlockNum,
		AllTxNum:        s.AllTxNum,
		Sha256DBlockNum: s.Sha256dBlockNum,
		X11BlockNum:     s.X11BlockNum,
		ScryptBlockNum:  s.ScryptBlockNum,
		AuxBlockNum:     s.AuxBlockNum,
		HourItemList:    nil,
	}
	if s.HourItemList != nil {
		hi := make([]*pb.PovApiGetLastNHourItem, 0)
		for _, hp := range s.HourItemList {
			ht := &pb.PovApiGetLastNHourItem{
				Hour:            hp.Hour,
				AllBlockNum:     hp.AllBlockNum,
				AllTxNum:        hp.AllTxNum,
				AllMinerReward:  toBalanceValue(hp.AllMinerReward),
				AllRepReward:    toBalanceValue(hp.AllRepReward),
				Sha256DBlockNum: hp.Sha256dBlockNum,
				X11BlockNum:     hp.X11BlockNum,
				ScryptBlockNum:  hp.ScryptBlockNum,
				AuxBlockNum:     hp.AuxBlockNum,
				MaxTxPerBlock:   hp.MaxTxPerBlock,
				MinTxPerBlock:   hp.MinTxPerBlock,
				AvgTxPerBlock:   hp.AvgTxPerBlock,
			}
			hi = append(hi, ht)
		}
		r.HourItemList = hi
	}
	return r
}
