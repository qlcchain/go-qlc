package contract

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type MinerReward struct{}

func (m *MinerReward) GetRewardInfo(ctx *vmstore.VMContext, coinbase types.Address) (*cabi.MinerInfo, error) {
	oldMinerInfo, err := cabi.GetMinerInfoByCoinbase(ctx, coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, errors.New("failed to get storage for miner")
	}
	if oldMinerInfo == nil {
		oldMinerInfo = new(cabi.MinerInfo)
	}

	return oldMinerInfo, nil
}

func (m *MinerReward) GetNodeRewardHeight(ctx *vmstore.VMContext) (uint64, error) {
	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return 0, errors.New("failed to get latest block")
	}

	if latestBlock.GetHeight() < cabi.RewardHeightGapToLatest {
		return 0, nil
	}
	endHeight := latestBlock.GetHeight() - cabi.RewardHeightGapToLatest
	endHeight = cabi.MinerRoundPovHeightByDay(endHeight)

	return endHeight, nil
}

func (m *MinerReward) GetAvailRewardBlocks(ctx *vmstore.VMContext, coinbase types.Address) (uint64, uint64, error) {
	oldMinerInfo, _ := cabi.GetMinerInfoByCoinbase(ctx, coinbase)

	startHeight := uint64(0)
	if oldMinerInfo != nil {
		startHeight = oldMinerInfo.RewardHeight + 1
	}
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}

	nodeHeight, err := m.GetNodeRewardHeight(ctx)
	if err != nil {
		return 0, 0, err
	}
	availHeight := cabi.MinerCalcRewardEndHeight(startHeight, nodeHeight)

	availBlocks, err := m.calcRewardBlocks(ctx, coinbase, startHeight, availHeight)

	return availHeight, availBlocks, err
}

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

	if block.Token != common.ChainToken() {
		return errors.New("token is not chain token")
	}

	amCb, _ := ctx.GetAccountMeta(param.Coinbase)
	if amCb == nil {
		return errors.New("coinbase account not exist")
	}

	// only miner with enough pledge can call this reward contract
	if amCb.GetVote().Compare(common.PovMinerPledgeAmountMin) == types.BalanceCompSmaller {
		return errors.New("coinbase account has not enough pledge amount")
	}

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)
	if amBnf == nil {
		return errors.New("beneficial account not exist")
	}

	// check reward height
	err = cabi.MinerCheckRewardHeight(param.RewardHeight)
	if err != nil {
		return err
	}

	nodeRewardHeight, err := m.GetNodeRewardHeight(ctx)
	if err != nil {
		return err
	}
	if param.RewardHeight > nodeRewardHeight {
		return fmt.Errorf("reward height %d should lesser than node height %d", param.RewardHeight, nodeRewardHeight)
	}

	oldMinerInfo, err := cabi.GetMinerInfoByCoinbase(ctx, param.Coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return errors.New("failed to get storage for miner")
	}
	if oldMinerInfo == nil {
		oldMinerInfo = new(cabi.MinerInfo)
	}
	if param.RewardHeight <= oldMinerInfo.RewardHeight {
		return fmt.Errorf("reward height %d should greath than last height %d", param.RewardHeight, oldMinerInfo.RewardHeight)
	}

	startHeight := oldMinerInfo.RewardHeight + 1
	maxEndHeight := cabi.MinerCalcRewardEndHeight(startHeight, param.RewardHeight)
	if maxEndHeight != param.RewardHeight {
		return fmt.Errorf("reward height %d should lesser than max height %d", param.RewardHeight, maxEndHeight)
	}

	block.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial, param.RewardHeight)
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

	oldMinerInfo, err := cabi.GetMinerInfoByCoinbase(ctx, param.Coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, errors.New("failed to get storage for miner")
	}
	if oldMinerInfo == nil {
		oldMinerInfo = new(cabi.MinerInfo)
	}

	rewardBlocks, err := m.calcRewardBlocks(ctx, input.Address, oldMinerInfo.RewardHeight+1, param.RewardHeight)
	if err != nil {
		return nil, errors.New("failed to calculate reward blocks")
	}

	rewardAmount := cabi.RewardPerBlockBalance.Mul(int64(rewardBlocks))

	newMinerData, err := cabi.MinerABI.PackVariable(
		cabi.VariableNameMiner,
		param.Beneficial,
		param.RewardHeight,
		oldMinerInfo.RewardBlocks+rewardBlocks)
	if err != nil {
		return nil, err
	}
	minerKey := cabi.GetMinerKey(param.Coinbase)
	err = ctx.SetStorage(types.MinerAddress.Bytes(), minerKey, newMinerData)
	if err != nil {
		return nil, err
	}

	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()
	block.Data = newMinerData

	// pledge fields only for QLC token
	block.Vote = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance
	block.Network = types.ZeroBalance

	tmBnf := amBnf.Token(common.GasToken())
	if tmBnf != nil {
		block.Balance = tmBnf.Balance.Add(rewardAmount)
		block.Representative = tmBnf.Representative
		block.Previous = tmBnf.Header
	} else {
		block.Balance = rewardAmount
		if len(amBnf.Tokens) > 0 {
			block.Representative = amBnf.Tokens[0].Representative
		} else {
			block.Representative = input.Representative
		}
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

func (m *MinerReward) calcRewardBlocks(ctx *vmstore.VMContext, coinbase types.Address, startHeight, endHeight uint64) (uint64, error) {
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}

	rewardBlocks := uint64(0)
	for curHeight := startHeight; curHeight <= endHeight; curHeight++ {
		block, err := ctx.GetPovBlockByHeight(curHeight)
		if block == nil {
			return 0, err
		}

		if coinbase == block.GetCoinbase() {
			rewardBlocks++
		}
	}

	return rewardBlocks, nil
}

func (m *MinerReward) GetRefundData() []byte {
	return []byte{1}
}
