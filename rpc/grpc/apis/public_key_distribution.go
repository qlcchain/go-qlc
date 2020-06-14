package apis

import (
	"context"
	"math/big"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

//PublicKeyDistributionAPI
type PublicKeyDistributionAPI struct {
	pkd    *api.PublicKeyDistributionApi
	logger *zap.SugaredLogger
	cc     *qlcchainctx.ChainContext
}

func NewPublicKeyDistributionAPI(cfgFile string, l ledger.Store) *PublicKeyDistributionAPI {
	return &PublicKeyDistributionAPI{
		pkd:    api.NewPublicKeyDistributionApi(cfgFile, l),
		logger: log.NewLogger("grpc_pkd"),
		cc:     qlcchainctx.NewChainContext(cfgFile),
	}
}

func (p *PublicKeyDistributionAPI) GetVerifierRegisterBlock(ctx context.Context, param *pb.VerifierRegParam) (*pbtypes.StateBlock, error) {
	verPara, err := toOriginVerifierRegParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierRegisterBlock(verPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifierUnregisterBlock(ctx context.Context, param *pb.VerifierUnRegParam) (*pbtypes.StateBlock, error) {
	verPara, err := toOriginVerifierUnRegParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierUnregisterBlock(verPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetAllVerifiers(ctx context.Context, param *empty.Empty) (*pb.VerifierRegParams, error) {
	r, err := p.pkd.GetAllVerifiers()
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifiersByType(ctx context.Context, param *pb.String) (*pb.VerifierRegParams, error) {
	r, err := p.pkd.GetVerifiersByType(toOriginString(param))
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetActiveVerifiers(ctx context.Context, param *pb.String) (*pb.VerifierRegParams, error) {
	r, err := p.pkd.GetActiveVerifiers(toOriginString(param))
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifiersByAccount(ctx context.Context, param *pbtypes.Address) (*pb.VerifierRegParams, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifiersByAccount(addr)
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifierStateByBlockHeight(ctx context.Context, param *pb.VerifierStateByBlockHeightRequest) (*pbtypes.PovVerifierState, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierStateByBlockHeight(param.GetHeight(), addr)
	if err != nil {
		return nil, err
	}
	return toPovVerifierState(r), nil
}

func (p *PublicKeyDistributionAPI) GetAllVerifierStatesByBlockHeight(ctx context.Context, param *pb.UInt64) (*pb.PKDVerifierStateList, error) {
	r, err := p.pkd.GetAllVerifierStatesByBlockHeight(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toPKDVerifierStateList(r), nil
}

func (p *PublicKeyDistributionAPI) GetPublishBlock(ctx context.Context, param *pb.PublishParam) (*pb.PublishRet, error) {
	publishParam, err := toOriginPublishParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetPublishBlock(publishParam)
	if err != nil {
		return nil, err
	}
	return toPublishRet(r), nil
}

func (p *PublicKeyDistributionAPI) GetUnPublishBlock(ctx context.Context, param *pb.UnPublishParam) (*pbtypes.StateBlock, error) {
	publishParam, err := toOriginUnPublishParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetUnPublishBlock(publishParam)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetPubKeyByTypeAndID(ctx context.Context, param *pb.TypeAndIDParam) (*pb.PublishInfoStates, error) {
	r, err := p.pkd.GetPubKeyByTypeAndID(param.GetPType(), param.GetPID())
	if err != nil {
		return nil, err
	}
	return toPublishInfoStates(r), nil
}

func (p *PublicKeyDistributionAPI) GetRecommendPubKey(ctx context.Context, param *pb.TypeAndIDParam) (*pb.PublishInfoState, error) {
	r, err := p.pkd.GetRecommendPubKey(param.GetPType(), param.GetPID())
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return toPublishInfoState(r), nil
}

func (p *PublicKeyDistributionAPI) GetPublishInfosByType(ctx context.Context, param *pb.String) (*pb.PublishInfoStates, error) {
	r, err := p.pkd.GetPublishInfosByType(toOriginString(param))
	if err != nil {
		return nil, err
	}
	return toPublishInfoStates(r), nil
}

func (p *PublicKeyDistributionAPI) GetPublishInfosByAccountAndType(ctx context.Context, param *pb.AccountAndTypeParam) (*pb.PublishInfoStates, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetPublishInfosByAccountAndType(addr, param.GetPType())
	if err != nil {
		return nil, err
	}
	return toPublishInfoStates(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleBlock(ctx context.Context, param *pb.OracleParam) (*pbtypes.StateBlock, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetOracleBlock(&api.OracleParam{
		Account: addr,
		OType:   param.GetType(),
		OID:     param.GetId(),
		KeyType: param.GetKeyType(),
		PubKey:  param.GetPubKey(),
		Code:    param.GetCode(),
		Hash:    param.GetHash(),
	})
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleInfosByType(ctx context.Context, param *pb.String) (*pb.OracleParams, error) {
	r, err := p.pkd.GetOracleInfosByType(toOriginString(param))
	if err != nil {
		return nil, err
	}
	return toOracleParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleInfosByTypeAndID(ctx context.Context, param *pb.TypeAndIDParam) (*pb.OracleParams, error) {
	r, err := p.pkd.GetOracleInfosByTypeAndID(param.GetPType(), param.GetPID())
	if err != nil {
		return nil, err
	}
	return toOracleParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleInfosByAccountAndType(ctx context.Context, param *pb.AccountAndTypeParam) (*pb.OracleParams, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetOracleInfosByAccountAndType(addr, param.GetPType())
	if err != nil {
		return nil, err
	}
	return toOracleParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleInfosByHash(ctx context.Context, param *pbtypes.Hash) (*pb.OracleParams, error) {
	r, err := p.pkd.GetOracleInfosByHash(param.GetHash())
	if err != nil {
		return nil, err
	}
	return toOracleParams(r), nil
}

func (p *PublicKeyDistributionAPI) PackRewardData(ctx context.Context, param *pb.PKDRewardParam) (*pb.Bytes, error) {
	rewardPara, err := toOriginPackRewardData(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.PackRewardData(rewardPara)
	if err != nil {
		return nil, err
	}
	return toBytes(r), nil
}

func (p *PublicKeyDistributionAPI) UnpackRewardData(ctx context.Context, param *pb.Bytes) (*pb.PKDRewardParam, error) {
	r, err := p.pkd.UnpackRewardData(toOriginBytes(param))
	if err != nil {
		return nil, err
	}
	return toPackRewardData(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardSendBlock(ctx context.Context, param *pb.PKDRewardParam) (*pbtypes.StateBlock, error) {
	rewardPara, err := toOriginPackRewardData(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardSendBlock(rewardPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardRecvBlock(ctx context.Context, param *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	blk, err := toOriginStateBlock(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardRecvBlock(blk)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardRecvBlockBySendHash(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	h, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardRecvBlockBySendHash(h)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardHistory(ctx context.Context, param *pbtypes.Address) (*pb.PKDHistoryRewardInfo, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardHistory(addr)
	if err != nil {
		return nil, err
	}
	return &pb.PKDHistoryRewardInfo{
		LastEndHeight:  r.LastEndHeight,
		LastBeneficial: toAddressValue(r.LastBeneficial),
		LastRewardTime: r.LastRewardTime,
		RewardAmount:   toBalanceValue(r.RewardAmount),
	}, nil
}

func (p *PublicKeyDistributionAPI) GetAvailRewardInfo(ctx context.Context, param *pbtypes.Address) (*pb.PKDAvailRewardInfo, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetAvailRewardInfo(addr)
	if err != nil {
		return nil, err
	}
	return &pb.PKDAvailRewardInfo{
		LastEndHeight:     r.LastEndHeight,
		LatestBlockHeight: r.LatestBlockHeight,
		NodeRewardHeight:  r.NodeRewardHeight,
		AvailEndHeight:    r.AvailEndHeight,
		AvailRewardAmount: toBalanceValue(r.AvailRewardAmount),
		NeedCallReward:    r.NeedCallReward,
	}, nil
}

func (p *PublicKeyDistributionAPI) GetVerifierHeartBlock(ctx context.Context, param *pb.VerifierHeartBlockRequest) (*pbtypes.StateBlock, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierHeartBlock(addr, param.GetVTypes())
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func toOriginVerifierRegParam(para *pb.VerifierRegParam) (*api.VerifierRegParam, error) {
	addr, err := toOriginAddressByValue(para.GetAccount())
	if err != nil {
		return nil, err
	}
	return &api.VerifierRegParam{
		Account: addr,
		VType:   para.GetType(),
		VInfo:   para.GetId(),
		VKey:    para.GetKey(),
	}, nil
}

func toOriginVerifierUnRegParam(para *pb.VerifierUnRegParam) (*api.VerifierUnRegParam, error) {
	addr, err := toOriginAddressByValue(para.GetAccount())
	if err != nil {
		return nil, err
	}
	return &api.VerifierUnRegParam{
		Account: addr,
		VType:   para.GetType(),
	}, nil
}

func toVerifierRegParam(para *api.VerifierRegParam) *pb.VerifierRegParam {
	return &pb.VerifierRegParam{
		Account: toAddressValue(para.Account),
		Type:    para.VType,
		Id:      para.VInfo,
		Key:     para.VKey,
	}
}

func toVerifierRegParams(para []*api.VerifierRegParam) *pb.VerifierRegParams {
	vs := make([]*pb.VerifierRegParam, 0)
	for _, v := range para {
		vs = append(vs, toVerifierRegParam(v))
	}
	return &pb.VerifierRegParams{Params: vs}
}

func toPovVerifierState(state *types.PovVerifierState) *pbtypes.PovVerifierState {
	r := &pbtypes.PovVerifierState{
		TotalVerify:  state.TotalVerify,
		TotalReward:  toBalanceValueByBigNum(state.TotalReward),
		ActiveHeight: state.ActiveHeight,
	}
	return r
}

func toPKDVerifierStateList(list *api.PKDVerifierStateList) *pb.PKDVerifierStateList {
	r := &pb.PKDVerifierStateList{
		VerifierNum:  int32(list.VerifierNum),
		AllVerifiers: nil,
	}
	if list.AllVerifiers != nil {
		rMap := make(map[string]*pbtypes.PovVerifierState)
		for k, v := range list.AllVerifiers {
			rMap[toAddressValue(k)] = toPovVerifierState(v)
		}
		r.AllVerifiers = rMap
	}
	return r
}

func toOriginPublishParam(param *pb.PublishParam) (*api.PublishParam, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	fee := toOriginBalanceByValue(param.GetFee())
	vers, err := toOriginAddressesByValues(param.GetVerifiers())
	if err != nil {
		return nil, err
	}
	codes, err := toOriginHashesByValues(param.GetCodes())
	if err != nil {
		return nil, err
	}
	return &api.PublishParam{
		Account:   addr,
		PType:     param.GetType(),
		PID:       param.GetId(),
		PubKey:    param.GetPubKey(),
		KeyType:   param.GetKeyType(),
		Fee:       fee,
		Verifiers: vers,
		Codes:     codes,
		Hash:      "",
	}, nil
}

func toPublishRet(ret *api.PublishRet) *pb.PublishRet {
	r := &pb.PublishRet{}
	if ret.Block != nil {
		r.Block = toStateBlock(ret.Block)
	}
	if ret.Verifiers != nil {
		cMap := make(map[string]*pb.VerifierContent)
		for k, v := range ret.Verifiers {
			ct := &pb.VerifierContent{
				Account: toAddressValue(v.Account),
				PubKey:  v.PubKey,
				Code:    v.Code,
				Hash:    toHashValue(v.Hash),
			}
			cMap[k] = ct
		}
		r.Verifiers = cMap
	}
	return r
}

func toOracleParam(p *api.OracleParam) *pb.OracleParam {
	return &pb.OracleParam{
		Account: toAddressValue(p.Account),
		Type:    p.OType,
		Id:      p.OID,
		KeyType: p.KeyType,
		PubKey:  p.PubKey,
		Code:    p.Code,
		Hash:    p.Hash,
	}
}

func toOracleParams(params []*api.OracleParam) *pb.OracleParams {
	os := make([]*pb.OracleParam, 0)
	for _, p := range params {
		os = append(os, toOracleParam(p))
	}
	return &pb.OracleParams{
		Params: os,
	}
}

func toOriginUnPublishParam(param *pb.UnPublishParam) (*api.UnPublishParam, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	return &api.UnPublishParam{
		Account: addr,
		PType:   param.GetType(),
		PID:     param.GetId(),
		PubKey:  param.GetPubKey(),
		KeyType: param.GetKeyType(),
		Hash:    param.GetHash(),
	}, nil
}

func toPublishInfoStates(states []*api.PublishInfoState) *pb.PublishInfoStates {
	ps := make([]*pb.PublishInfoState, 0)
	for _, p := range states {
		pt := toPublishInfoState(p)
		ps = append(ps, pt)
	}
	return &pb.PublishInfoStates{States: ps}
}

func toPublishInfoState(state *api.PublishInfoState) *pb.PublishInfoState {
	r := &pb.PublishInfoState{
		Account:   toAddressValue(state.Account),
		Type:      state.PType,
		Id:        state.PID,
		PubKey:    state.PubKey,
		KeyType:   state.KeyType,
		Fee:       toBalanceValue(state.Fee),
		Verifiers: toAddressValues(state.Verifiers),
		Codes:     toHashesValues(state.Codes),
		Hash:      state.Hash,
		State:     nil,
	}
	if state.State != nil {
		p := &pbtypes.PovPublishState{
			OracleAccounts: toAddressValues(state.State.OracleAccounts),
			PublishHeight:  state.State.PublishHeight,
			VerifiedHeight: state.State.VerifiedHeight,
			VerifiedStatus: int32(state.State.VerifiedStatus),
			BonusFee:       toBalanceValueByBigNum(state.State.BonusFee),
		}
		r.State = p
	}
	return r
}

func toOriginPackRewardData(param *pb.PKDRewardParam) (*api.PKDRewardParam, error) {
	acc, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	bene, err := toOriginAddressByValue(param.GetBeneficial())
	if err != nil {
		return nil, err
	}
	amount := big.NewInt(param.GetRewardAmount())
	return &api.PKDRewardParam{
		Account:      acc,
		Beneficial:   bene,
		EndHeight:    param.GetEndHeight(),
		RewardAmount: amount,
	}, nil
}

func toPackRewardData(param *api.PKDRewardParam) *pb.PKDRewardParam {
	return &pb.PKDRewardParam{
		Account:      toAddressValue(param.Account),
		Beneficial:   toAddressValue(param.Beneficial),
		EndHeight:    param.EndHeight,
		RewardAmount: toBalanceValueByBigInt(param.RewardAmount),
	}
}
