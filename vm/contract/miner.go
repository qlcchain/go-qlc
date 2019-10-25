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

func (m *MinerReward) GetLastRewardHeight(ctx *vmstore.VMContext, coinbase types.Address) (uint64, error) {
	height, err := cabi.GetLastMinerRewardHeightByAccount(ctx, coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return 0, errors.New("failed to get storage for miner")
	}

	return height, nil
}

func (m *MinerReward) GetRewardHistory(ctx *vmstore.VMContext, coinbase types.Address) (*cabi.MinerRewardInfo, error) {
	data, err := ctx.GetStorage(types.MinerAddress[:], coinbase[:])
	if err == nil {
		info := new(cabi.MinerRewardInfo)
		er := cabi.MinerABI.UnpackVariable(info, cabi.VariableNameMinerReward, data)
		if er != nil {
			return nil, er
		}
		return info, nil
	}

	return nil, err
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

func (m *MinerReward) GetAvailRewardInfo(ctx *vmstore.VMContext, coinbase types.Address, nodeHeight uint64, lastRewardHeight uint64) (*cabi.MinerRewardInfo, error) {
	availInfo := new(cabi.MinerRewardInfo)
	availInfo.RewardAmount = types.NewBalance(0)

	startHeight := lastRewardHeight + 1
	endHeight := cabi.MinerCalcRewardEndHeight(startHeight, nodeHeight)
	if endHeight < startHeight {
		return nil, fmt.Errorf("calc endheight err [%d-%d]", startHeight, endHeight)
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

func (m *MinerReward) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if _, err := param.Verify(); err != nil {
		return nil, nil, err
	}

	if param.Coinbase != block.Address {
		return nil, nil, errors.New("account is not coinbase")
	}

	if block.Token != common.ChainToken() {
		return nil, nil, errors.New("token is not chain token")
	}

	// check account exist
	amCb, _ := ctx.GetAccountMeta(param.Coinbase)
	if amCb == nil {
		return nil, nil, errors.New("coinbase account not exist")
	}

	nodeRewardHeight, err := m.GetNodeRewardHeight(ctx)
	if err != nil {
		return nil, nil, err
	}
	if param.EndHeight > nodeRewardHeight {
		return nil, nil, fmt.Errorf("end height %d greater than node height %d", param.EndHeight, nodeRewardHeight)
	}

	// check same start & end height exist in old reward infos
	err = m.checkParamExistInOldRewardInfos(ctx, param)
	if err != nil {
		return nil, nil, errors.New("section exist")
	}

	calcRewardBlocks, calcRewardAmount, err := m.calcRewardBlocksByDayStats(ctx, param.Coinbase, param.StartHeight, param.EndHeight)
	if err != nil {
		return nil, nil, err
	}
	if calcRewardBlocks != param.RewardBlocks {
		return nil, nil, fmt.Errorf("calc blocks %d not equal param blocks %d", calcRewardBlocks, param.RewardBlocks)
	}
	if calcRewardAmount.Compare(param.RewardAmount) != types.BalanceCompEqual {
		return nil, nil, fmt.Errorf("calc reward %d not equal param reward %v", calcRewardAmount, param.RewardAmount)
	}

	block.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial,
		param.StartHeight, param.EndHeight, param.RewardBlocks, param.RewardAmount)
	if err != nil {
		return nil, nil, err
	}

	oldInfo, err := m.GetRewardHistory(ctx, param.Coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, nil, fmt.Errorf("get storage err %s", err)
	}

	if oldInfo == nil {
		oldInfo = new(cabi.MinerRewardInfo)
		oldInfo.RewardAmount = types.NewBalance(0)
	}

	data, _ := cabi.MinerABI.PackVariable(cabi.VariableNameMinerReward, param.EndHeight,
		param.RewardBlocks+oldInfo.RewardBlocks, block.Timestamp, param.RewardAmount.Add(oldInfo.RewardAmount))
	err = ctx.SetStorage(types.MinerAddress.Bytes(), param.Coinbase[:], data)
	if err != nil {
		return nil, nil, errors.New("save contract data err")
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

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	block.Vote = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance
	block.Network = types.ZeroBalance

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)
	if amBnf != nil {
		tmBnf := amBnf.Token(common.GasToken())
		if tmBnf != nil {
			block.Balance = tmBnf.Balance.Add(param.RewardAmount)
			block.Representative = tmBnf.Representative
			block.Previous = tmBnf.Header
		} else {
			block.Balance = param.RewardAmount
			if len(amBnf.Tokens) > 0 {
				block.Representative = amBnf.Tokens[0].Representative
			} else {
				block.Representative = input.Representative
			}
			block.Previous = types.ZeroHash
		}
	} else {
		block.Balance = param.RewardAmount
		block.Representative = input.Representative
		block.Previous = types.ZeroHash
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.Beneficial,
			BlockType: types.ContractReward,
			Amount:    param.RewardAmount,
			Token:     common.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (m *MinerReward) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	param := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, block.Data)
	if err != nil {
		return 0, err
	}

	needHeight := param.EndHeight + common.PovMinerRewardHeightGapToLatest

	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return needHeight, nil
	}

	nodeHeight := latestBlock.GetHeight()
	if nodeHeight < needHeight {
		return needHeight, nil
	}

	return 0, nil
}

func (m *MinerReward) checkParamExistInOldRewardInfos(ctx *vmstore.VMContext, param *cabi.MinerRewardParam) error {
	lastRewardHeight, err := cabi.GetLastMinerRewardHeightByAccount(ctx, param.Coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return errors.New("failed to get storage for Miner")
	}

	if param.StartHeight <= lastRewardHeight {
		return fmt.Errorf("start height[%d] err, last height[%d]", param.StartHeight, lastRewardHeight)
	}

	return nil
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
			return 0, types.NewBalance(0), fmt.Errorf("get pov miner state err[%d]", dayIndex)
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
