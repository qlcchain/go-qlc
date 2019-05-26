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
	Coinbase   types.Address `json:"coinbase"`
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
	return cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial)
}

func (m *MinerApi) GetRewardSendBlock(param *RewardParam) (*types.StateBlock, error) {
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

	send := &types.StateBlock{
		Type:      types.ContractSend,
		Token:     common.ChainToken(),
		Address:   param.Coinbase,

		Balance:        tmCb.Balance,
		Previous:       tmCb.Header,
		Representative: tmCb.Representative,

		Vote:      amCb.CoinVote,
		Network:   amCb.CoinNetwork,
		Oracle:    amCb.CoinOracle,
		Storage:   amCb.CoinStorage,

		Link:      types.Hash(types.MinerAddress),
		Data:      data,
		Timestamp: common.TimeNow().UTC().Unix(),
	}

	err = m.reward.DoSend(m.vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MinerApi) GetRewardRecvBlock(input *types.StateBlock) (*types.StateBlock, error) {
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

	return nil, errors.New("can not generate reward recv block")
}
