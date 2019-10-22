package api

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/config"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
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
	RewardAmount types.Balance `json:"rewardAmount"`
}

type MinerAvailRewardInfo struct {
	LastEndHeight     uint64        `json:"lastEndHeight"`
	LatestBlockHeight uint64        `json:"latestBlockHeight"`
	NodeRewardHeight  uint64        `json:"nodeRewardHeight"`
	AvailStartHeight  uint64        `json:"availStartHeight"`
	AvailEndHeight    uint64        `json:"availEndHeight"`
	AvailRewardBlocks uint64        `json:"availRewardBlocks"`
	AvailRewardAmount types.Balance `json:"availRewardAmount"`
	NeedCallReward    bool          `json:"needCallReward"`
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
		reward: &contract.MinerReward{},
	}
}

func (m *MinerApi) GetRewardData(param *RewardParam) ([]byte, error) {
	return cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial, param.StartHeight, param.EndHeight, param.RewardBlocks, param.RewardAmount)
}

func (m *MinerApi) UnpackRewardData(data []byte) (*RewardParam, error) {
	abiParam := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(abiParam, cabi.MethodNameMinerReward, data)
	if err != nil {
		return nil, err
	}
	apiParam := new(RewardParam)
	apiParam.Coinbase = abiParam.Coinbase
	apiParam.Beneficial = abiParam.Beneficial
	apiParam.StartHeight = abiParam.StartHeight
	apiParam.EndHeight = abiParam.EndHeight
	apiParam.RewardBlocks = abiParam.RewardBlocks
	apiParam.RewardAmount = abiParam.RewardAmount
	return apiParam, nil
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
	lastRewardHeight, err := m.reward.GetLastRewardHeight(vmContext, coinbase)
	if err != nil {
		return nil, err
	}

	rsp.LastEndHeight = lastRewardHeight
	rsp.NodeRewardHeight, err = m.reward.GetNodeRewardHeight(vmContext)
	if err != nil {
		return nil, err
	}

	lastHeight := uint64(0)
	if lastRewardHeight == 0 {
		lastHeight = common.PovMinerRewardHeightStart - 1
	} else {
		lastHeight = lastRewardHeight
	}

	availInfo := new(cabi.MinerRewardInfo)
	availInfo.RewardAmount = types.NewBalance(0)
	for {
		info, err := m.reward.GetAvailRewardInfo(vmContext, coinbase, rsp.NodeRewardHeight, lastHeight)
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
	rsp.AvailRewardAmount = availInfo.RewardAmount

	if rsp.AvailStartHeight > lastRewardHeight && rsp.AvailEndHeight <= rsp.NodeRewardHeight &&
		rsp.AvailRewardAmount.Int64() > 0 {
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
		Timestamp: common.TimeNow().Unix(),

		PoVHeight: latestPovHeader.GetHeight(),
	}

	vmContext := vmstore.NewVMContext(m.ledger)
	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
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
