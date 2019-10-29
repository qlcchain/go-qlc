package contract

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type RepReward struct {
	WithSignAndPending
}

func (r *RepReward) GetLastRewardHeight(ctx *vmstore.VMContext, account types.Address) (uint64, error) {
	height, err := cabi.GetLastRepRewardHeightByAccount(ctx, account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return 0, errors.New("failed to get storage for repReward")
	}

	return height, nil
}

func (r *RepReward) GetRewardHistory(ctx *vmstore.VMContext, account types.Address) (*cabi.RepRewardInfo, error) {
	data, err := ctx.GetStorage(types.RepAddress[:], account[:])
	if err == nil {
		info := new(cabi.RepRewardInfo)
		er := cabi.RepABI.UnpackVariable(info, cabi.VariableNameRepReward, data)
		if er != nil {
			return nil, er
		}
		return info, nil
	}

	return nil, err
}

func (r *RepReward) GetNodeRewardHeight(ctx *vmstore.VMContext) (uint64, error) {
	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return 0, errors.New("failed to get latest block")
	}

	nodeHeight := latestBlock.GetHeight()
	if nodeHeight < common.PovMinerRewardHeightGapToLatest {
		return 0, nil
	}
	nodeHeight = nodeHeight - common.PovMinerRewardHeightGapToLatest

	nodeHeight = cabi.RepRoundPovHeight(nodeHeight, common.PovMinerRewardHeightRound)
	return nodeHeight, nil
}

func (r *RepReward) GetAvailRewardInfo(ctx *vmstore.VMContext, account types.Address, nodeHeight uint64, lastRewardHeight uint64) (*cabi.RepRewardInfo, error) {
	availInfo := new(cabi.RepRewardInfo)
	availInfo.RewardAmount = big.NewInt(0)

	startHeight := lastRewardHeight + 1
	endHeight := cabi.RepCalcRewardEndHeight(startHeight, nodeHeight)
	if endHeight < startHeight {
		return nil, fmt.Errorf("calc endheight err [%d-%d]", startHeight, endHeight)
	}

	availBlocks, availReward, err := r.calcRewardBlocksByDayStats(ctx, account, startHeight, endHeight)
	if err != nil {
		return nil, err
	}

	availInfo.StartHeight = startHeight
	availInfo.EndHeight = endHeight
	availInfo.RewardBlocks = availBlocks
	availInfo.RewardAmount = availReward.Int
	return availInfo, nil
}

func (r *RepReward) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.NewBalance(0), nil
}

func (r *RepReward) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.RepRewardParam)
	err := cabi.RepABI.UnpackMethod(param, cabi.MethodNameRepReward, block.Data)
	if err != nil {
		return nil, nil, err
	}

	if _, err := param.Verify(); err != nil {
		return nil, nil, err
	}

	if param.Account != block.Address {
		return nil, nil, errors.New("account is not representative")
	}

	if block.Token != common.ChainToken() {
		return nil, nil, errors.New("token is not chain token")
	}

	// check account exist
	am, _ := ctx.GetAccountMeta(param.Account)
	if am == nil {
		return nil, nil, errors.New("rep account not exist")
	}

	nodeRewardHeight, err := r.GetNodeRewardHeight(ctx)
	if err != nil {
		return nil, nil, err
	}

	if param.EndHeight > nodeRewardHeight {
		return nil, nil, fmt.Errorf("end height %d greater than node height %d", param.EndHeight, nodeRewardHeight)
	}

	// check same start & end height exist in old reward infos
	err = r.checkParamExistInOldRewardInfos(ctx, param)
	if err != nil {
		return nil, nil, errors.New("section exist")
	}

	calcRewardBlocks, calcRewardAmount, err := r.calcRewardBlocksByDayStats(ctx, param.Account, param.StartHeight, param.EndHeight)
	if err != nil {
		return nil, nil, err
	}

	if calcRewardBlocks != param.RewardBlocks {
		return nil, nil, fmt.Errorf("calc blocks %d not equal param blocks %d", calcRewardBlocks, param.RewardBlocks)
	}
	if calcRewardAmount.Compare(types.Balance{Int: param.RewardAmount}) != types.BalanceCompEqual {
		return nil, nil, fmt.Errorf("calc reward %d not equal param reward %v", calcRewardAmount, param.RewardAmount)
	}

	block.Data, err = cabi.RepABI.PackMethod(cabi.MethodNameRepReward, param.Account, param.Beneficial,
		param.StartHeight, param.EndHeight, param.RewardBlocks, param.RewardAmount)
	if err != nil {
		return nil, nil, err
	}

	oldInfo, err := r.GetRewardHistory(ctx, param.Account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, nil, fmt.Errorf("get storage err %s", err)
	}

	if oldInfo == nil {
		oldInfo = new(cabi.RepRewardInfo)
		oldInfo.RewardAmount = big.NewInt(0)
	}

	data, _ := cabi.RepABI.PackVariable(cabi.VariableNameRepReward, param.EndHeight,
		param.RewardBlocks+oldInfo.RewardBlocks, block.Timestamp,
		param.RewardAmount.Add(param.RewardAmount, oldInfo.RewardAmount))
	err = ctx.SetStorage(types.RepAddress.Bytes(), param.Account[:], data)
	if err != nil {
		return nil, nil, errors.New("save contract data err")
	}

	return &types.PendingKey{
			Address: param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.Balance{Int: param.RewardAmount},
			Type:   common.GasToken(),
		}, nil
}

func (r *RepReward) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.RepRewardParam)

	err := cabi.RepABI.UnpackMethod(param, cabi.MethodNameRepReward, input.Data)
	if err != nil {
		return nil, err
	}

	if _, err := param.Verify(); err != nil {
		return nil, err
	}

	if param.Account != input.Address {
		return nil, errors.New("input account is not rep")
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	block.Vote = types.NewBalance(0)
	block.Oracle = types.NewBalance(0)
	block.Storage = types.NewBalance(0)
	block.Network = types.NewBalance(0)

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)
	if amBnf != nil {
		tmBnf := amBnf.Token(common.GasToken())
		if tmBnf != nil {
			block.Balance = tmBnf.Balance.Add(types.Balance{Int: param.RewardAmount})
			block.Representative = tmBnf.Representative
			block.Previous = tmBnf.Header
		} else {
			block.Balance = types.Balance{Int: param.RewardAmount}
			if len(amBnf.Tokens) > 0 {
				block.Representative = amBnf.Tokens[0].Representative
			} else {
				block.Representative = input.Representative
			}
			block.Previous = types.ZeroHash
		}
	} else {
		block.Balance = types.Balance{Int: param.RewardAmount}
		block.Representative = input.Representative
		block.Previous = types.ZeroHash
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.Beneficial,
			BlockType: types.ContractReward,
			Amount:    types.Balance{Int: param.RewardAmount},
			Token:     common.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (r *RepReward) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	param := new(cabi.RepRewardParam)
	err := cabi.RepABI.UnpackMethod(param, cabi.MethodNameRepReward, block.Data)
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

func (r *RepReward) checkParamExistInOldRewardInfos(ctx *vmstore.VMContext, param *cabi.RepRewardParam) error {
	lastRewardHeight, err := cabi.GetLastRepRewardHeightByAccount(ctx, param.Account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return errors.New("failed to get storage for repReward")
	}

	if param.StartHeight <= lastRewardHeight {
		return fmt.Errorf("start height[%d] err, last height[%d]", param.StartHeight, lastRewardHeight)
	}

	return nil
}

func (r *RepReward) calcRewardBlocksByDayStats(ctx *vmstore.VMContext, account types.Address, startHeight, endHeight uint64) (uint64, types.Balance, error) {
	if startHeight < common.PovMinerRewardHeightStart {
		startHeight = common.PovMinerRewardHeightStart
	}

	if startHeight%uint64(common.POVChainBlocksPerDay) != 0 {
		return 0, types.NewBalance(0), errors.New("start height is not integral multiple of blocks per day")
	}
	if (endHeight+1)%uint64(common.POVChainBlocksPerDay) != 0 {
		return 0, types.NewBalance(0), errors.New("end height is not integral multiple of blocks per day")
	}

	repAddrStr := account.String()
	rewardBlocks := uint64(0)
	startDayIndex := uint32(startHeight / uint64(common.POVChainBlocksPerDay))
	endDayIndex := uint32(endHeight / uint64(common.POVChainBlocksPerDay))

	rewardAmount := types.NewBalance(0)
	for dayIndex := startDayIndex; dayIndex <= endDayIndex; dayIndex++ {
		dayStat, err := ctx.GetPovMinerStat(dayIndex)
		if err != nil {
			return 0, types.NewBalance(0), fmt.Errorf("get pov miner state err[%d]", dayIndex)
		}

		minerStat := dayStat.MinerStats[repAddrStr]
		if minerStat == nil {
			continue
		}

		rewardBlocks += uint64(minerStat.RepBlockNum)
		rewardAmount = rewardAmount.Add(minerStat.RepReward)
	}

	return rewardBlocks, rewardAmount, nil
}

func (r *RepReward) GetRefundData() []byte {
	return []byte{1}
}
