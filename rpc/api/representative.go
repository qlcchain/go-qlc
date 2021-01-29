package api

import (
	"errors"
	"fmt"
	"math/big"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type RepApi struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	ledger ledger.Store
	reward *contract.RepReward
}

type RepRewardParam struct {
	Account      types.Address `json:"account"`
	Beneficial   types.Address `json:"beneficial"`
	StartHeight  uint64        `json:"startHeight"`
	EndHeight    uint64        `json:"endHeight"`
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardAmount *big.Int      `json:"rewardAmount"`
}

type RepAvailRewardInfo struct {
	LastEndHeight     uint64        `json:"lastEndHeight"`
	LatestBlockHeight uint64        `json:"latestBlockHeight"`
	NodeRewardHeight  uint64        `json:"nodeRewardHeight"`
	AvailStartHeight  uint64        `json:"availStartHeight"`
	AvailEndHeight    uint64        `json:"availEndHeight"`
	AvailRewardBlocks uint64        `json:"availRewardBlocks"`
	AvailRewardAmount types.Balance `json:"availRewardAmount"`
	NeedCallReward    bool          `json:"needCallReward"`
}

type RepHistoryRewardInfo struct {
	LastEndHeight  uint64        `json:"lastEndHeight"`
	RewardBlocks   uint64        `json:"rewardBlocks"`
	RewardAmount   types.Balance `json:"rewardAmount"`
	LastRewardTime int64         `json:"lastRewardTime"`
}

func NewRepApi(cfg *config.Config, ledger ledger.Store) *RepApi {
	return &RepApi{
		cfg:    cfg,
		logger: log.NewLogger("api_representative"),
		ledger: ledger,
		reward: &contract.RepReward{},
	}
}

func (r *RepApi) GetRewardData(param *RepRewardParam) ([]byte, error) {
	return cabi.RepABI.PackMethod(cabi.MethodNameRepReward, param.Account, param.Beneficial, param.StartHeight, param.EndHeight, param.RewardBlocks, param.RewardAmount)
}

func (r *RepApi) UnpackRewardData(data []byte) (*RepRewardParam, error) {
	abiParam := new(cabi.RepRewardParam)
	err := cabi.RepABI.UnpackMethod(abiParam, cabi.MethodNameRepReward, data)
	if err != nil {
		return nil, err
	}
	apiParam := new(RepRewardParam)
	apiParam.Account = abiParam.Account
	apiParam.Beneficial = abiParam.Beneficial
	apiParam.StartHeight = abiParam.StartHeight
	apiParam.EndHeight = abiParam.EndHeight
	apiParam.RewardBlocks = abiParam.RewardBlocks
	apiParam.RewardAmount = abiParam.RewardAmount
	return apiParam, nil
}

func (r *RepApi) GetAvailRewardInfo(account types.Address) (*RepAvailRewardInfo, error) {
	if !r.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	rsp := new(RepAvailRewardInfo)

	latestPovHeader, err := r.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}
	rsp.LatestBlockHeight = latestPovHeader.GetHeight()

	vmContext := vmstore.NewVMContext(r.ledger, &contractaddress.RepAddress)
	lastRewardHeight, err := r.reward.GetLastRewardHeight(vmContext, account)
	if err != nil {
		return nil, err
	}

	rsp.LastEndHeight = lastRewardHeight
	rsp.NodeRewardHeight, err = r.reward.GetNodeRewardHeight(vmContext)
	if err != nil {
		return nil, err
	}

	lastHeight := uint64(0)
	if lastRewardHeight == 0 {
		lastHeight = common.PovMinerRewardHeightStart - 1
	} else {
		lastHeight = lastRewardHeight
	}

	availInfo := new(cabi.RepRewardInfo)
	availInfo.RewardAmount = big.NewInt(0)
	for {
		info, err := r.reward.GetAvailRewardInfo(vmContext, account, rsp.NodeRewardHeight, lastHeight)
		if err != nil {
			break
		}

		if info.RewardAmount.Int64() > 0 {
			availInfo = info
			break
		} else {
			lastHeight += common.PovMinerMaxRewardBlocksPerCall
		}
	}

	rsp.AvailStartHeight = availInfo.StartHeight
	rsp.AvailEndHeight = availInfo.EndHeight
	rsp.AvailRewardBlocks = availInfo.RewardBlocks
	rsp.AvailRewardAmount = types.Balance{Int: availInfo.RewardAmount}

	if rsp.AvailStartHeight > lastRewardHeight && rsp.AvailEndHeight <= rsp.NodeRewardHeight &&
		rsp.AvailRewardAmount.Int64() > 0 {
		rsp.NeedCallReward = true
	}

	return rsp, nil
}

func (r *RepApi) checkParamExistInOldRewardInfos(param *RepRewardParam) error {
	ctx := vmstore.NewVMContext(r.ledger, &contractaddress.RepAddress)

	lastRewardHeight, err := cabi.GetLastRepRewardHeightByAccount(ctx, param.Account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return errors.New("failed to get storage for repReward")
	}

	if param.StartHeight <= lastRewardHeight {
		return fmt.Errorf("start height[%d] err, last height[%d]", param.StartHeight, lastRewardHeight)
	}

	return nil
}

func (r *RepApi) GetRewardSendBlock(param *RepRewardParam) (*types.StateBlock, error) {
	if !r.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	if param.Account.IsZero() {
		return nil, errors.New("invalid reward param account")
	}

	if param.Beneficial.IsZero() {
		return nil, errors.New("invalid reward param beneficial")
	}

	am, err := r.ledger.GetAccountMeta(param.Account)
	if am == nil {
		return nil, fmt.Errorf("rep account not exist, %s", err)
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("rep account does not have chain token, %s", err)
	}

	data, err := r.GetRewardData(param)
	if err != nil {
		return nil, err
	}

	latestPovHeader, err := r.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:    types.ContractSend,
		Token:   config.ChainToken(),
		Address: param.Account,

		Balance:        tm.Balance,
		Previous:       tm.Header,
		Representative: tm.Representative,

		Vote:    types.ToBalance(am.CoinVote),
		Network: types.ToBalance(am.CoinNetwork),
		Oracle:  types.ToBalance(am.CoinOracle),
		Storage: types.ToBalance(am.CoinStorage),

		Link:      types.Hash(contractaddress.RepAddress),
		Data:      data,
		Timestamp: common.TimeNow().Unix(),

		PoVHeight: latestPovHeader.GetHeight(),
	}

	vmContext := vmstore.NewVMContext(r.ledger, &contractaddress.RepAddress)
	_, _, err = r.reward.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = h
	}

	return send, nil
}

func (r *RepApi) GetRewardRecvBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if !r.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	if input.GetType() != types.ContractSend {
		return nil, errors.New("input block type is not contract send")
	}
	if input.GetLink() != contractaddress.RepAddress.ToHash() {
		return nil, errors.New("input address is not contract repReward")
	}

	reward := &types.StateBlock{}

	vmContext := vmstore.NewVMContext(r.ledger, &contractaddress.RepAddress)
	blocks, err := r.reward.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		return reward, nil
	}

	return nil, errors.New("can not generate reward recv block")
}

func (r *RepApi) GetRewardRecvBlockBySendHash(sendHash types.Hash) (*types.StateBlock, error) {
	if !r.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	input, err := r.ledger.GetStateBlockConfirmed(sendHash)
	if err != nil {
		return nil, err
	}

	return r.GetRewardRecvBlock(input)
}

type RepStateParams struct {
	Account types.Address
	Height  uint64
}

func (r *RepApi) GetRepStateWithHeight(params *RepStateParams) (*types.PovRepState, error) {
	block, err := r.ledger.GetPovHeaderByHeight(params.Height)
	if block == nil {
		return nil, fmt.Errorf("get pov block with height[%d] err", params.Height)
	}

	stateHash := block.GetStateHash()
	stateTrie := trie.NewTrie(r.ledger.DBStore(), &stateHash, nil)
	keyBytes := statedb.PovCreateRepStateKey(params.Account)

	valBytes := stateTrie.GetValue(keyBytes)
	if len(valBytes) <= 0 {
		return nil, fmt.Errorf("get trie err")
	}

	rs := types.NewPovRepState()
	err = rs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize old rep state err %s", err)
	}

	return rs, nil
}

func (r *RepApi) GetRewardHistory(account types.Address) (*RepHistoryRewardInfo, error) {
	history := new(RepHistoryRewardInfo)
	vmContext := vmstore.NewVMContext(r.ledger, &contractaddress.RepAddress)
	info, err := r.reward.GetRewardHistory(vmContext, account)
	if err != nil {
		return nil, err
	}

	history.LastEndHeight = info.EndHeight
	history.RewardBlocks = info.RewardBlocks
	history.RewardAmount = types.Balance{Int: info.RewardAmount}
	history.LastRewardTime = info.Timestamp

	return history, nil
}
