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
	RewardBlocks uint64        `json:"rewardBlocks"`
	RewardHeight uint64        `json:"rewardHeight"`
}

type MinerRewardInfo struct {
	PledgeVoteAmount types.Balance `json:"pledgeVoteAmount"`

	LastRewardHeight uint64 `json:"lastRewardHeight"`
	LastRewardBlocks uint64 `json:"lastRewardBlocks"`
	LastBeneficial   string `json:"lastBeneficial"`

	LatestBlockHeight uint64 `json:"latestBlockHeight"`
	NodeRewardHeight  uint64 `json:"nodeRewardHeight"`
	AvailRewardHeight uint64 `json:"availRewardHeight"`
	AvailRewardBlocks uint64 `json:"availRewardBlocks"`
	NeedCallReward    bool   `json:"needCallReward"`
}

func NewMinerApi(cfg *config.Config, ledger *ledger.Ledger) *MinerApi {
	return &MinerApi{
		cfg:    cfg,
		logger: log.NewLogger("api_miner"),
		ledger: ledger,
		reward: &contract.MinerReward{}}
}

func (m *MinerApi) GetRewardData(param *RewardParam) ([]byte, error) {
	return cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial, param.RewardBlocks, param.RewardHeight)
}

func (m *MinerApi) GetRewardInfo(coinbase types.Address) (*MinerRewardInfo, error) {
	if !m.cfg.PoV.PovEnabled {
		return nil, errors.New("pov service is disabled")
	}

	rsp := new(MinerRewardInfo)

	latestPovHeader, err := m.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, err
	}
	rsp.LatestBlockHeight = latestPovHeader.GetHeight()

	repCb, err := m.ledger.GetRepresentation(coinbase)
	if err == nil && repCb != nil {
		rsp.PledgeVoteAmount = repCb.Vote
	} else {
		rsp.PledgeVoteAmount = types.ZeroBalance
	}
	hasEnoughVote := true
	if rsp.PledgeVoteAmount.Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
		hasEnoughVote = false
	}

	vmContext := vmstore.NewVMContext(m.ledger)
	oldMinerInfo, err := m.reward.GetRewardInfo(vmContext, coinbase)
	if err != nil {
		return nil, err
	}

	rsp.LastRewardHeight = oldMinerInfo.RewardHeight
	rsp.LastRewardBlocks = oldMinerInfo.RewardBlocks
	if !oldMinerInfo.Beneficial.IsZero() {
		rsp.LastBeneficial = oldMinerInfo.Beneficial.String()
	}

	rsp.NodeRewardHeight, err = m.reward.GetNodeRewardHeight(vmContext)
	if err != nil {
		return nil, err
	}

	if hasEnoughVote {
		rsp.AvailRewardHeight, rsp.AvailRewardBlocks, err = m.reward.GetAvailRewardBlocks(vmContext, coinbase)
		if err != nil {
			return nil, err
		}
	} else {
		rsp.AvailRewardHeight = rsp.LastRewardHeight
	}

	if hasEnoughVote &&
		rsp.AvailRewardHeight > rsp.LastRewardHeight &&
		rsp.AvailRewardHeight <= rsp.NodeRewardHeight {
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
