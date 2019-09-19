package api

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/trie"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type RepApi struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	ledger *ledger.Ledger
	reward *contract.RepReward
}

type RepRewardParam struct {
	Account      types.Address `json:"account"`
	Beneficial   types.Address `json:"beneficial"`
	StartHeight  uint64        `json:"startHeight"`
	EndHeight    uint64        `json:"endHeight"`
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardAmount types.Balance `json:"rewardAmount"`
}

type RepAvailRewardInfo struct {
	LastBeneficial   string        `json:"lastBeneficial"`
	LastStartHeight  uint64        `json:"lastStartHeight"`
	LastEndHeight    uint64        `json:"lastEndHeight"`
	LastRewardBlocks uint64        `json:"lastRewardBlocks"`
	LastRewardAmount types.Balance `json:"lastRewardAmount"`

	LatestBlockHeight uint64        `json:"latestBlockHeight"`
	NodeRewardHeight  uint64        `json:"nodeRewardHeight"`
	AvailStartHeight  uint64        `json:"availStartHeight"`
	AvailEndHeight    uint64        `json:"availEndHeight"`
	AvailRewardBlocks uint64        `json:"availRewardBlocks"`
	AvailRewardAmount types.Balance `json:"availRewardAmount"`
	NeedCallReward    bool          `json:"needCallReward"`
}

type RepHistoryRewardInfo struct {
	RewardInfos       []*cabi.RepRewardInfo `json:"rewardInfos"`
	FirstRewardHeight uint64                `json:"firstRewardHeight"`
	LastRewardHeight  uint64                `json:"lastRewardHeight"`
	AllRewardBlocks   uint64                `json:"allRewardBlocks"`
	AllRewardAmount   types.Balance         `json:"allRewardAmount"`
}

func NewRepApi(cfg *config.Config, ledger *ledger.Ledger) *RepApi {
	return &RepApi{
		cfg:    cfg,
		logger: log.NewLogger("api_representative"),
		ledger: ledger,
		reward: &contract.RepReward{},
	}
}

func (m *RepApi) GetRewardData(param *RepRewardParam) ([]byte, error) {
	return cabi.RepABI.PackMethod(cabi.MethodNameRepReward, param.Account, param.Beneficial, param.StartHeight, param.EndHeight, param.RewardBlocks, param.RewardAmount)
}

func (m *RepApi) GetHistoryRewardInfos(account types.Address) (*RepHistoryRewardInfo, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	vmContext := vmstore.NewVMContext(m.ledger)
	rewardInfos, err := m.reward.GetAllRewardInfos(vmContext, account)
	if err != nil {
		return nil, err
	}

	apiRsp := new(RepHistoryRewardInfo)
	apiRsp.RewardInfos = rewardInfos

	if len(rewardInfos) > 0 {
		apiRsp.FirstRewardHeight = rewardInfos[0].StartHeight
		apiRsp.LastRewardHeight = rewardInfos[len(rewardInfos)-1].EndHeight

		for _, rewardInfo := range rewardInfos {
			apiRsp.AllRewardBlocks += rewardInfo.RewardBlocks
			apiRsp.AllRewardAmount = apiRsp.AllRewardAmount.Add(rewardInfo.RewardAmount)
		}
	}

	return apiRsp, nil
}

func (m *RepApi) GetAvailRewardInfo(account types.Address) (*RepAvailRewardInfo, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	rsp := new(RepAvailRewardInfo)

	latestPovHeader, err := m.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}
	rsp.LatestBlockHeight = latestPovHeader.GetHeight()

	vmContext := vmstore.NewVMContext(m.ledger)
	lastRewardInfo, err := m.reward.GetMaxRewardInfo(vmContext, account)
	if err != nil {
		return nil, err
	}

	rsp.LastStartHeight = lastRewardInfo.StartHeight
	rsp.LastEndHeight = lastRewardInfo.EndHeight
	rsp.LastRewardBlocks = lastRewardInfo.RewardBlocks
	rsp.LastRewardAmount = lastRewardInfo.RewardAmount
	if !lastRewardInfo.Beneficial.IsZero() {
		rsp.LastBeneficial = lastRewardInfo.Beneficial.String()
	}

	rsp.NodeRewardHeight, err = m.reward.GetNodeRewardHeight(vmContext)
	if err != nil {
		return nil, err
	}

	availInfo, err := m.reward.GetAvailRewardInfo(vmContext, account, rsp.NodeRewardHeight, lastRewardInfo)
	if err != nil {
		return nil, err
	}
	rsp.AvailStartHeight = availInfo.StartHeight
	rsp.AvailEndHeight = availInfo.EndHeight
	rsp.AvailRewardBlocks = availInfo.RewardBlocks
	rsp.AvailRewardAmount = availInfo.RewardAmount

	if (rsp.LastEndHeight > 0 && rsp.AvailStartHeight > rsp.LastEndHeight) ||
		(rsp.LastEndHeight == 0 && rsp.AvailStartHeight == 0) && rsp.AvailEndHeight <= rsp.NodeRewardHeight {
		rsp.NeedCallReward = true
	}

	return rsp, nil
}

func (m *RepApi) checkParamExistInOldRewardInfos(param *RepRewardParam) error {
	ctx := vmstore.NewVMContext(m.ledger)

	oldRewardInfos, err := cabi.GetRepRewardInfosByAccount(ctx, param.Account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return errors.New("failed to get storage for miner")
	}

	for _, oldRewardInfo := range oldRewardInfos {
		if param.StartHeight >= oldRewardInfo.StartHeight && param.StartHeight <= oldRewardInfo.EndHeight {
			return fmt.Errorf("start height %d exist in old reward info %d-%d", param.StartHeight, oldRewardInfo.StartHeight, oldRewardInfo.EndHeight)
		}
		if param.EndHeight >= oldRewardInfo.StartHeight && param.EndHeight <= oldRewardInfo.EndHeight {
			return fmt.Errorf("end height %d exist in old reward info %d-%d", param.StartHeight, oldRewardInfo.StartHeight, oldRewardInfo.EndHeight)
		}
	}

	return nil
}

func (m *RepApi) GetRewardSendBlock(param *RepRewardParam) (*types.StateBlock, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	if param.Account.IsZero() {
		return nil, errors.New("invalid reward param account")
	}

	if param.Beneficial.IsZero() {
		return nil, errors.New("invalid reward param beneficial")
	}

	// check same start & end height exist in old reward infos
	err := m.checkParamExistInOldRewardInfos(param)
	if err != nil {
		return nil, err
	}

	am, err := m.ledger.GetAccountMeta(param.Account)
	if am == nil {
		return nil, fmt.Errorf("rep account not exist, %s", err)
	}

	tm := am.Token(common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("rep account does not have chain token, %s", err)
	}

	data, err := m.GetRewardData(param)
	if err != nil {
		return nil, err
	}

	latestPovHeader, err := m.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:    types.ContractSend,
		Token:   common.ChainToken(),
		Address: param.Account,

		Balance:        tm.Balance,
		Previous:       tm.Header,
		Representative: tm.Representative,

		Vote:    am.CoinVote,
		Network: am.CoinNetwork,
		Oracle:  am.CoinOracle,
		Storage: am.CoinStorage,

		Link:      types.Hash(types.RepAddress),
		Data:      data,
		Timestamp: common.TimeNow().Unix(),

		PoVHeight: latestPovHeader.GetHeight(),
	}

	vmContext := vmstore.NewVMContext(m.ledger)
	err = m.reward.DoSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *RepApi) GetRewardRecvBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	if input.GetType() != types.ContractSend {
		return nil, errors.New("input block type is not contract send")
	}
	if input.GetLink() != types.MinerAddress.ToHash() {
		return nil, errors.New("input address is not contract miner")
	}

	reward := &types.StateBlock{}

	vmContext := vmstore.NewVMContext(m.ledger)
	blocks, err := m.reward.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		reward.Timestamp = common.TimeNow().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate reward recv block")
}

func (m *RepApi) GetRewardRecvBlockBySendHash(sendHash types.Hash) (*types.StateBlock, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	input, err := m.ledger.GetStateBlock(sendHash)
	if err != nil {
		return nil, err
	}

	return m.GetRewardRecvBlock(input)
}

type RepStateParams struct {
	Account types.Address
	Height  uint64
}

func (m *RepApi) GetRepStateWithHeight(params *RepStateParams) (*types.PovRepState, error) {
	ctx := vmstore.NewVMContext(m.ledger)
	block, err := ctx.GetPovHeaderByHeight(params.Height)
	if block == nil {
		return nil, fmt.Errorf("get pov block with height[%d] err", params.Height)
	}

	stateHash := block.GetStateHash()
	stateTrie := trie.NewTrie(ctx.GetLedger().Store, &stateHash, nil)
	keyBytes := types.PovCreateRepStateKey(params.Account)

	valBytes := stateTrie.GetValue(keyBytes)
	if len(valBytes) <= 0 {
		return nil, fmt.Errorf("get trie err")
	}

	rs := new(types.PovRepState)
	err = rs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize old rep state err %s", err)
	}

	return rs, nil
}
