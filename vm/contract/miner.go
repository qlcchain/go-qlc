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

func (m *MinerReward) GetMaxRewardInfo(ctx *vmstore.VMContext, coinbase types.Address) (*cabi.MinerRewardInfo, error) {
	rewardInfos, err := cabi.GetMinerRewardInfosByCoinbase(ctx, coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, errors.New("failed to get storage for miner")
	}

	maxInfo := cabi.CalcMaxMinerRewardInfo(rewardInfos)
	if maxInfo == nil {
		maxInfo = new(cabi.MinerRewardInfo)
	}

	return maxInfo, nil
}

func (m *MinerReward) GetAllRewardInfos(ctx *vmstore.VMContext, coinbase types.Address) ([]*cabi.MinerRewardInfo, error) {
	rewardInfos, err := cabi.GetMinerRewardInfosByCoinbase(ctx, coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, errors.New("failed to get storage for miner")
	}

	return rewardInfos, nil
}

func (m *MinerReward) GetNodeRewardHeight(ctx *vmstore.VMContext) (uint64, error) {
	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return 0, errors.New("failed to get latest block")
	}

	nodeHeight := latestBlock.GetHeight()
	if nodeHeight < common.PovMinerRewardHeightStart {
		return 0, nil
	}
	if nodeHeight < common.PovMinerRewardHeightGapToLatest {
		return 0, nil
	}
	nodeHeight = nodeHeight - common.PovMinerRewardHeightGapToLatest

	nodeHeight = cabi.MinerRoundPovHeight(nodeHeight, common.PovMinerRewardHeightRound)
	if nodeHeight < common.PovMinerRewardHeightStart {
		return 0, nil
	}

	return nodeHeight, nil
}

func (m *MinerReward) GetAvailRewardInfo(ctx *vmstore.VMContext, coinbase types.Address) (*cabi.MinerRewardInfo, error) {
	availInfo := new(cabi.MinerRewardInfo)

	nodeHeight, err := m.GetNodeRewardHeight(ctx)
	if err != nil {
		return nil, err
	}
	if nodeHeight < common.PovMinerRewardHeightStart {
		return availInfo, nil
	}

	maxInfo, err := m.GetMaxRewardInfo(ctx, coinbase)
	if err != nil {
		return nil, err
	}

	startHeight := uint64(0)
	if maxInfo != nil {
		startHeight = maxInfo.EndHeight + 1
	}
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}

	endHeight := cabi.MinerCalcRewardEndHeight(startHeight, nodeHeight)
	if endHeight < common.PovMinerRewardHeightStart {
		return availInfo, nil
	}

	if endHeight < startHeight {
		return availInfo, nil
	}

	availBlocks, availReward, err := m.calcRewardBlocksByDayStats(ctx, coinbase, startHeight, endHeight)
	if err != nil {
		return nil, err
	}

	availInfo.StartHeight = startHeight
	availInfo.EndHeight = endHeight
	availInfo.RewardBlocks = availBlocks
	availInfo.RewardAmount = availReward
	return availInfo, nil
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

	if _, err := param.Verify(); err != nil {
		return err
	}

	if param.Coinbase != block.Address {
		return errors.New("account is not coinbase")
	}

	if block.Token != common.ChainToken() {
		return errors.New("token is not chain token")
	}

	// check account exist
	amCb, _ := ctx.GetAccountMeta(param.Coinbase)
	if amCb == nil {
		return errors.New("coinbase account not exist")
	}

	nodeRewardHeight, err := m.GetNodeRewardHeight(ctx)
	if err != nil {
		return err
	}
	if param.EndHeight > nodeRewardHeight {
		return fmt.Errorf("end height %d greater than node height %d", param.EndHeight, nodeRewardHeight)
	}

	calcRewardBlocks, calcRewardAmount, err := m.calcRewardBlocksByDayStats(ctx, param.Coinbase, param.StartHeight, param.EndHeight)
	if err != nil {
		return err
	}
	if calcRewardBlocks != param.RewardBlocks {
		return fmt.Errorf("calc blocks %d not equal param blocks %d", calcRewardBlocks, param.RewardBlocks)
	}
	if calcRewardAmount.Compare(param.RewardAmount) != types.BalanceCompEqual {
		return fmt.Errorf("calc reward %d not equal param reward %v", calcRewardAmount, param.RewardAmount)
	}

	block.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial, param.StartHeight, param.EndHeight, param.RewardBlocks)
	if err != nil {
		return err
	}

	return nil
}

func (m *MinerReward) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if _, err := param.Verify(); err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: param.RewardAmount,
			Type:   common.GasToken(),
		}, nil
}

func (m *MinerReward) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.MinerRewardParam)
	exist := false
	calcRewardAmount := types.NewBalance(0)
	var calcRewardBlocks uint64

	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, input.Data)
	if err != nil {
		return nil, err
	}

	if _, err := param.Verify(); err != nil {
		return nil, err
	}

	if param.Coinbase != input.Address {
		return nil, errors.New("input account is not coinbase")
	}

	// check same start & end height exist in old reward infos
	err = m.checkParamExistInOldRewardInfos(ctx, param)
	if err != nil {
		exist = true
	}

	// save contract data to storage
	newMinerData, err := cabi.MinerABI.PackVariable(
		cabi.VariableNameMinerRewardInfo,
		param.Beneficial,
		param.StartHeight,
		param.EndHeight,
		param.RewardBlocks,
		param.RewardAmount)
	if err != nil {
		return nil, err
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()
	block.Data = newMinerData
	block.PoVHeight = input.PoVHeight

	// pledge fields only for QLC token
	block.Vote = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance
	block.Network = types.ZeroBalance

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)

	if exist {
		if amBnf != nil {
			tmBnf := amBnf.Token(common.GasToken())
			if tmBnf != nil {
				block.Balance = tmBnf.Balance
				block.Representative = tmBnf.Representative
				block.Previous = tmBnf.Header
			} else {
				return nil, fmt.Errorf("reward gas token meta does not exist")
			}
		} else {
			return nil, fmt.Errorf("reward account meta does not exist")
		}
	} else {
		calcRewardBlocks, calcRewardAmount, err = m.calcRewardBlocksByDayStats(ctx, param.Coinbase, param.StartHeight, param.EndHeight)
		if err != nil {
			return nil, err
		}
		if calcRewardBlocks != param.RewardBlocks {
			return nil, fmt.Errorf("calc blocks %d not equal param blocks %d", calcRewardBlocks, param.RewardBlocks)
		}
		if calcRewardAmount.Compare(param.RewardAmount) != types.BalanceCompEqual {
			return nil, fmt.Errorf("calc reward %v not equal param reward %v", calcRewardAmount, param.RewardAmount)
		}

		if amBnf != nil {
			tmBnf := amBnf.Token(common.GasToken())
			if tmBnf != nil {
				block.Balance = tmBnf.Balance.Add(calcRewardAmount)
				block.Representative = tmBnf.Representative
				block.Previous = tmBnf.Header
			} else {
				block.Balance = calcRewardAmount
				if len(amBnf.Tokens) > 0 {
					block.Representative = amBnf.Tokens[0].Representative
				} else {
					block.Representative = input.Representative
				}
				block.Previous = types.ZeroHash
			}
		} else {
			block.Balance = calcRewardAmount
			block.Representative = input.Representative
			block.Previous = types.ZeroHash
		}
	}

	minerKey := cabi.GetMinerRewardKey(param.Coinbase, param.StartHeight)
	err = ctx.SetStorage(types.MinerAddress.Bytes(), minerKey, newMinerData)
	if err != nil {
		return nil, err
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.Beneficial,
			BlockType: types.ContractReward,
			Amount:    calcRewardAmount,
			Token:     common.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (m *MinerReward) checkParamExistInOldRewardInfos(ctx *vmstore.VMContext, param *cabi.MinerRewardParam) error {
	oldRewardInfos, err := cabi.GetMinerRewardInfosByCoinbase(ctx, param.Coinbase)
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

func (m *MinerReward) calcRewardBlocksByHeight(ctx *vmstore.VMContext, coinbase types.Address, startHeight, endHeight uint64) (uint64, error) {
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}

	rewardBlocks := uint64(0)
	for curHeight := startHeight; curHeight <= endHeight; curHeight++ {
		block, err := ctx.GetPovHeaderByHeight(curHeight)
		if block == nil {
			return 0, err
		}

		if coinbase == block.GetMinerAddr() {
			rewardBlocks++
		}
	}

	return rewardBlocks, nil
}

func (m *MinerReward) calcRewardBlocksByDayStats(ctx *vmstore.VMContext, coinbase types.Address, startHeight, endHeight uint64) (uint64, types.Balance, error) {
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}

	if startHeight%uint64(common.POVChainBlocksPerDay) != 0 {
		return 0, types.NewBalance(0), errors.New("start height is not integral multiple of blocks per day")
	}
	if (endHeight+1)%uint64(common.POVChainBlocksPerDay) != 0 {
		return 0, types.NewBalance(0), errors.New("end height is not integral multiple of blocks per day")
	}

	cbAddrHex := coinbase.String()
	rewardBlocks := uint64(0)
	startDayIndex := uint32(startHeight / uint64(common.POVChainBlocksPerDay))
	endDayIndex := uint32(endHeight / uint64(common.POVChainBlocksPerDay))

	rewardAmount := types.NewBalance(0)
	for dayIndex := startDayIndex; dayIndex <= endDayIndex; dayIndex++ {
		dayStat, err := ctx.GetPovMinerStat(dayIndex)
		if err != nil {
			return 0, types.NewBalance(0), err
		}

		minerStat := dayStat.MinerStats[cbAddrHex]
		if minerStat == nil {
			continue
		}

		rewardBlocks += uint64(minerStat.BlockNum)
		rewardAmount = rewardAmount.Add(minerStat.RewardAmount)
	}

	return rewardBlocks, rewardAmount, nil
}

func (m *MinerReward) GetRefundData() []byte {
	return []byte{1}
}
