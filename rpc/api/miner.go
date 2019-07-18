package api

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type MinerApi struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	ledger *ledger.Ledger
	reward *contract.MinerReward
}

type RewardParam struct {
	Coinbase     types.Address `json:"coinbase"`
	Beneficial   types.Address `json:"beneficial"`
	StartHeight  uint64        `json:"startHeight"`
	EndHeight    uint64        `json:"endHeight"`
	RewardBlocks uint64        `json:"rewardBlocks"`
}

type MinerAvailRewardInfo struct {
	LastBeneficial   string `json:"lastBeneficial"`
	LastStartHeight  uint64 `json:"lastStartHeight"`
	LastEndHeight    uint64 `json:"lastEndHeight"`
	LastRewardBlocks uint64 `json:"lastRewardBlocks"`

	LatestBlockHeight uint64 `json:"latestBlockHeight"`
	NodeRewardHeight  uint64 `json:"nodeRewardHeight"`
	AvailStartHeight  uint64 `json:"availStartHeight"`
	AvailEndHeight    uint64 `json:"availEndHeight"`
	AvailRewardBlocks uint64 `json:"availRewardBlocks"`
	NeedCallReward    bool   `json:"needCallReward"`
}

type MinerHistoryRewardInfo struct {
	RewardInfos       []*cabi.MinerRewardInfo `json:"rewardInfos"`
	FirstRewardHeight uint64                  `json:"firstRewardHeight"`
	LastRewardHeight  uint64                  `json:"lastRewardHeight"`
	AllRewardBlocks   uint64                  `json:"allRewardBlocks"`
	AllRewardAmount   types.Balance           `json:"allRewardAmount"`
}

func NewMinerApi(cfg *config.Config, ledger *ledger.Ledger) *MinerApi {
	return &MinerApi{
		cfg:    cfg,
		logger: log.NewLogger("api_miner"),
		ledger: ledger,
		reward: &contract.MinerReward{}}
}

func (m *MinerApi) GetRewardData(param *RewardParam) ([]byte, error) {
	return cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial, param.StartHeight, param.EndHeight, param.RewardBlocks)
}

func (m *MinerApi) GetHistoryRewardInfos(coinbase types.Address) (*MinerHistoryRewardInfo, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	vmContext := vmstore.NewVMContext(m.ledger)
	rewardInfos, err := m.reward.GetAllRewardInfos(vmContext, coinbase)
	if err != nil {
		return nil, err
	}

	apiRsp := new(MinerHistoryRewardInfo)
	apiRsp.RewardInfos = rewardInfos

	if len(rewardInfos) > 0 {
		apiRsp.FirstRewardHeight = rewardInfos[0].StartHeight
		apiRsp.LastRewardHeight = rewardInfos[len(rewardInfos)-1].EndHeight

		for _, rewardInfo := range rewardInfos {
			apiRsp.AllRewardBlocks += rewardInfo.RewardBlocks
		}
		apiRsp.AllRewardAmount = common.PovMinerRewardPerBlockBalance.Mul(int64(apiRsp.AllRewardBlocks))
	}

	return apiRsp, nil
}

func (m *MinerApi) GetAvailRewardInfo(coinbase types.Address) (*MinerAvailRewardInfo, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	rsp := new(MinerAvailRewardInfo)

	latestPovHeader, err := m.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}
	rsp.LatestBlockHeight = latestPovHeader.GetHeight()

	vmContext := vmstore.NewVMContext(m.ledger)
	lastRewardInfo, err := m.reward.GetMaxRewardInfo(vmContext, coinbase)
	if err != nil {
		return nil, err
	}

	rsp.LastStartHeight = lastRewardInfo.StartHeight
	rsp.LastEndHeight = lastRewardInfo.EndHeight
	rsp.LastRewardBlocks = lastRewardInfo.RewardBlocks
	if !lastRewardInfo.Beneficial.IsZero() {
		rsp.LastBeneficial = lastRewardInfo.Beneficial.String()
	}

	rsp.NodeRewardHeight, err = m.reward.GetNodeRewardHeight(vmContext)
	if err != nil {
		return nil, err
	}

	availInfo, err := m.reward.GetAvailRewardInfo(vmContext, coinbase)
	if err != nil {
		return nil, err
	}
	rsp.AvailStartHeight = availInfo.StartHeight
	rsp.AvailEndHeight = availInfo.EndHeight
	rsp.AvailRewardBlocks = availInfo.RewardBlocks

	if rsp.AvailStartHeight > rsp.LastEndHeight && rsp.AvailEndHeight <= rsp.NodeRewardHeight {
		rsp.NeedCallReward = true
	}

	return rsp, nil
}

func (m *MinerApi) GetRewardSendBlock(param *RewardParam) (*types.StateBlock, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	if param.Coinbase.IsZero() {
		return nil, errors.New("invalid reward param coinbase")
	}
	if param.Beneficial.IsZero() {
		return nil, errors.New("invalid reward param beneficial")
	}

	amCb, err := m.ledger.GetAccountMeta(param.Coinbase)
	if amCb == nil {
		return nil, fmt.Errorf("coinbase account not exist, %s", err)
	}

	tmCb := amCb.Token(common.ChainToken())
	if tmCb == nil {
		return nil, fmt.Errorf("coinbase account does not have chain token, %s", err)
	}

	amBnf, err := m.ledger.GetAccountMeta(param.Beneficial)
	if amBnf == nil {
		return nil, fmt.Errorf("invalid beneficial account, %s", err)
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
		Address: param.Coinbase,

		Balance:        tmCb.Balance,
		Previous:       tmCb.Header,
		Representative: tmCb.Representative,

		Vote:    amCb.CoinVote,
		Network: amCb.CoinNetwork,
		Oracle:  amCb.CoinOracle,
		Storage: amCb.CoinStorage,

		Link:      types.Hash(types.MinerAddress),
		Data:      data,
		Timestamp: common.TimeNow().UTC().Unix(),

		PoVHeight: latestPovHeader.GetHeight(),
	}

	vmContext := vmstore.NewVMContext(m.ledger)
	err = m.reward.DoSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MinerApi) GetRewardRecvBlock(input *types.StateBlock) (*types.StateBlock, error) {
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
		reward.Timestamp = common.TimeNow().UTC().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate reward recv block")
}

func (m *MinerApi) GetRewardRecvBlockBySendHash(sendHash types.Hash) (*types.StateBlock, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	input, err := m.ledger.GetStateBlock(sendHash)
	if err != nil {
		return nil, err
	}

	return m.GetRewardRecvBlock(input)
}
