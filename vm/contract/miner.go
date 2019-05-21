package contract

import (
	"errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type MinerReward struct{}

func (m *MinerReward) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (m *MinerReward) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) (err error) {
	param := new(cabi.MinerRewardParam)
	err = cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, block.Data)
	if err != nil {
		return err
	}

	am, _ := ctx.GetAccountMeta(param.Beneficial)
	if am == nil {
		return errors.New("beneficial account not exist")
	}

	block.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Beneficial)
	if err != nil {
		return err
	}

	return nil
}

func (m *MinerReward) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, input.Data)
	if err != nil {
		return nil, err
	}

	key := cabi.GetMinerKey(input.Address)
	minerData, err := ctx.GetStorage(types.MinerAddress.Bytes(), key)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, err
	}

	oldMinerInfo := new(cabi.MinerInfo)
	if len(minerData) > 0 {
		err = cabi.MinerABI.UnpackVariable(oldMinerInfo, cabi.VariableNameMiner, minerData)
		if err != nil {
			return nil, errors.New("invalid miner variable data")
		}
	}

	endHeight, rewardAmount, err := m.calcReward(ctx, input.Address, oldMinerInfo)
	if err != nil {
		return nil, errors.New("failed to calculate reward")
	}

	if endHeight <= oldMinerInfo.RewardHeight {
		return nil, nil
	}

	minerDataNew, err := cabi.MinerABI.PackVariable(
		cabi.VariableNameMiner,
		param.Beneficial,
		endHeight)
	if err != nil {
		return nil, err
	}
	err = ctx.SetStorage(types.MinerAddress.Bytes(), key, minerData)
	if err != nil {
		return nil, err
	}

	am, _ := ctx.GetAccountMeta(param.Beneficial)
	tm := am.Token(common.GasToken())

	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()
	block.Vote = am.CoinVote
	block.Oracle = am.CoinOracle
	block.Storage = am.CoinStorage
	block.Network = am.CoinNetwork
	block.Data = minerDataNew

	if tm != nil {
		block.Representative = tm.Representative
		block.Balance = tm.Balance.Add(rewardAmount)
		block.Previous = tm.Header
	} else {
		block.Representative = block.Address
		block.Balance = rewardAmount
		block.Previous = types.ZeroHash
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.Beneficial,
			BlockType: types.ContractReward,
			Amount:    rewardAmount,
			Token:     common.GasToken(),
			Data:      minerData,
		},
	}, nil
}

func (m *MinerReward) calcReward(ctx *vmstore.VMContext, coinbase types.Address, old *cabi.MinerInfo) (uint64, types.Balance, error) {
	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return old.RewardHeight, types.NewBalance(0), err
	}

	if latestBlock.GetHeight() <= cabi.RewardHeightLimit {
		return old.RewardHeight, types.NewBalance(0), nil
	}

	startHeight := old.RewardHeight + 1
	endHeight := latestBlock.GetHeight() - cabi.RewardHeightLimit
	if endHeight < startHeight {
		return old.RewardHeight, types.NewBalance(0), nil
	}

	rewardAmount := types.NewBalance(0)
	for curHeight := startHeight; curHeight <= endHeight; curHeight++ {
		block, err := ctx.GetPovBlockByHeight(startHeight)
		if block == nil {
			return old.RewardHeight, types.NewBalance(0), err
		}

		if coinbase == block.GetCoinbase() {
			rewardAmount.Add(cabi.RewardPerBlockBalance)
		}
	}

	return endHeight, types.NewBalance(0), nil
}

func (m *MinerReward) GetRefundData() []byte {
	return []byte{1}
}
