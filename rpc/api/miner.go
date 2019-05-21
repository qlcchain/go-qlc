package api

import (
	"errors"
	"fmt"
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
	logger    *zap.SugaredLogger
	ledger    *ledger.Ledger
	vmContext *vmstore.VMContext
	reward    *contract.MinerReward
}

type RewardParam struct {
	Beneficial types.Address `json:"beneficial"`
}

func NewMinerApi(ledger *ledger.Ledger) *MinerApi {
	return &MinerApi{
		logger:    log.NewLogger("api_miner"),
		ledger:    ledger,
		vmContext: vmstore.NewVMContext(ledger),
		reward:    &contract.MinerReward{}}
}

func (m *MinerApi) GetRewardData(param *RewardParam) ([]byte, error) {
	return cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Beneficial)
}

func (m *MinerApi) GetRewardContractSendBlock(param *RewardParam) (*types.StateBlock, error) {
	if param.Beneficial.IsZero() {
		return nil, errors.New("invalid reward param")
	}

	am, err := m.ledger.GetAccountMeta(param.Beneficial)
	if am == nil {
		return nil, fmt.Errorf("invalid beneficial account:%s, %s", param.Beneficial, err)
	}

	data, err := m.GetRewardData(param)
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:      types.ContractSend,
		Token:     common.GasToken(),
		Address:   param.Beneficial,
		Balance:   am.CoinBalance,
		Vote:      am.CoinVote,
		Network:   am.CoinNetwork,
		Oracle:    am.CoinOracle,
		Storage:   am.CoinStorage,
		Link:      types.Hash(types.MinerAddress),
		Data:      data,
		Timestamp: common.TimeNow().UTC().Unix(),
	}

	tm := am.Token(common.GasToken())
	if tm != nil {
		send.Previous = tm.Header
		send.Representative = tm.Representative
	} else {
		send.Previous = types.ZeroHash
		send.Representative = send.Address
	}

	err = m.reward.DoSend(m.vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MinerApi) GetRewardContractRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	reward := &types.StateBlock{}

	blocks, err := m.reward.DoReceive(m.vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		reward.Timestamp = common.TimeNow().UTC().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate contract reward block")
}
