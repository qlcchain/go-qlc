package contract

import (
	"errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
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

	if param.Coinbase != block.Address {
		return errors.New("account is not coinbase")
	}

	if block.Token != common.GasToken() {
		return errors.New("token is not gas token")
	}

	amCb, _ := ctx.GetAccountMeta(param.Coinbase)
	if amCb == nil {
		return errors.New("coinbase account not exist")
	}

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)
	if amBnf == nil {
		return errors.New("beneficial account not exist")
	}

	block.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial)
	if err != nil {
		return err
	}

	return nil
}

func (m *MinerReward) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, errors.New("not implemented")
}

func (m *MinerReward) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, input.Data)
	if err != nil {
		return nil, err
	}

	if param.Coinbase != input.Address {
		return nil, errors.New("input account is not coinbase")
	}

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)
	if amBnf == nil {
		return nil, errors.New("beneficial account not exist")
	}

	key := cabi.GetMinerKey(input.Address)
	oldMinerData, err := ctx.GetStorage(types.MinerAddress.Bytes(), key)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, err
	}

	oldMinerInfo := new(cabi.MinerInfo)
	if len(oldMinerData) > 0 {
		err = cabi.MinerABI.UnpackVariable(oldMinerInfo, cabi.VariableNameMiner, oldMinerData)
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

	newMinerData, err := cabi.MinerABI.PackVariable(
		cabi.VariableNameMiner,
		param.Beneficial,
		endHeight)
	if err != nil {
		return nil, err
	}
	err = ctx.SetStorage(types.MinerAddress.Bytes(), key, newMinerData)
	if err != nil {
		return nil, err
	}

	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()
	block.Data = newMinerData

	block.Vote = amBnf.CoinVote
	block.Oracle = amBnf.CoinOracle
	block.Storage = amBnf.CoinStorage
	block.Network = amBnf.CoinNetwork

	tmBnf := amBnf.Token(common.GasToken())
	if tmBnf != nil {
		block.Balance = tmBnf.Balance.Add(rewardAmount)
		block.Representative = tmBnf.Representative
		block.Previous = tmBnf.Header
	} else {
		block.Balance = rewardAmount
		block.Representative = types.ZeroAddress
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
			Data:      newMinerData,
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
	if old.RewardHeight < common.PovMinerVerifyHeightStart {
		startHeight = common.PovMinerVerifyHeightStart
	}

	endHeight := latestBlock.GetHeight() - cabi.RewardHeightLimit
	if endHeight < startHeight {
		return old.RewardHeight, types.NewBalance(0), nil
	}

	rewardAmountInt := big.NewInt(0)
	rewardCount := 0
	for curHeight := startHeight; curHeight <= endHeight; curHeight++ {
		block, err := ctx.GetPovBlockByHeight(curHeight)
		if block == nil {
			return old.RewardHeight, types.NewBalance(0), err
		}

		if coinbase == block.GetCoinbase() {
			rewardCount++
			rewardAmountInt.Add(rewardAmountInt, cabi.RewardPerBlockInt)
		}
	}

	rewardAmount := types.Balance{Int: rewardAmountInt}

	return endHeight, rewardAmount, nil
}

func (m *MinerReward) GetRefundData() []byte {
	return []byte{1}
}
