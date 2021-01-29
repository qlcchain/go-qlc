package contract

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var MinerContract = NewChainContract(
	map[string]Contract{
		cabi.MethodNameMinerReward: &MinerReward{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					pending:   true,
					work:      true,
				},
			},
		},
	},
	cabi.MinerABI,
	cabi.JsonMiner,
)

type MinerReward struct {
	BaseContract
}

func (m *MinerReward) GetLastRewardHeight(ctx *vmstore.VMContext, coinbase types.Address) (uint64, error) {
	height, err := cabi.GetLastMinerRewardHeightByAccount(ctx, coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return 0, errors.New("failed to get storage for miner")
	}

	return height, nil
}

func (m *MinerReward) GetRewardHistory(ctx *vmstore.VMContext, coinbase types.Address) (*cabi.MinerRewardInfo, error) {
	data, err := ctx.GetStorage(contractaddress.MinerAddress[:], coinbase[:])
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
	availInfo.RewardAmount = big.NewInt(0)

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
	availInfo.RewardAmount = availReward.Int
	return availInfo, nil
}

func (m *MinerReward) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	if _, err := param.Verify(); err != nil {
		return nil, nil, ErrCheckParam
	}

	if param.Coinbase != block.Address {
		return nil, nil, ErrAccountInvalid
	}

	if block.Token != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	// check account exist
	amCb, _ := ctx.GetAccountMeta(param.Coinbase)
	if amCb == nil {
		return nil, nil, ErrAccountNotExist
	}

	nodeRewardHeight, err := m.GetNodeRewardHeight(ctx)
	if err != nil {
		return nil, nil, ErrGetNodeHeight
	}

	if param.EndHeight > nodeRewardHeight {
		return nil, nil, ErrEndHeightInvalid
	}

	// check same start & end height exist in old reward infos
	err = m.checkParamExistInOldRewardInfos(ctx, param)
	if err != nil {
		return nil, nil, ErrClaimRepeat
	}

	calcRewardBlocks, calcRewardAmount, err := m.calcRewardBlocksByDayStats(ctx, param.Coinbase, param.StartHeight, param.EndHeight)
	if err != nil {
		return nil, nil, ErrCalcAmount
	}

	if calcRewardBlocks != param.RewardBlocks || calcRewardAmount.Compare(types.Balance{Int: param.RewardAmount}) != types.BalanceCompEqual {
		return nil, nil, ErrCheckParam
	}

	block.Data, err = cabi.MinerABI.PackMethod(cabi.MethodNameMinerReward, param.Coinbase, param.Beneficial,
		param.StartHeight, param.EndHeight, param.RewardBlocks, param.RewardAmount)
	if err != nil {
		return nil, nil, ErrPackMethod
	}

	oldInfo, err := m.GetRewardHistory(ctx, param.Coinbase)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, nil, ErrGetRewardHistory
	}

	if oldInfo == nil {
		oldInfo = new(cabi.MinerRewardInfo)
		oldInfo.RewardAmount = big.NewInt(0)
	}

	data, _ := cabi.MinerABI.PackVariable(cabi.VariableNameMinerReward, param.EndHeight,
		param.RewardBlocks+oldInfo.RewardBlocks, block.Timestamp,
		new(big.Int).Add(param.RewardAmount, oldInfo.RewardAmount))
	err = ctx.SetStorage(contractaddress.MinerAddress.Bytes(), param.Coinbase[:], data)
	if err != nil {
		return nil, nil, ErrSetStorage
	}

	return &types.PendingKey{
			Address: param.Beneficial,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.Balance{Int: param.RewardAmount},
			Type:   cfg.GasToken(),
		}, nil
}

func (m *MinerReward) SetStorage(ctx *vmstore.VMContext, endHeight uint64, RewardAmount *big.Int, RewardBlocks uint64, block *types.StateBlock) error {
	oldInfo, err := m.GetRewardHistory(ctx, block.Address)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return fmt.Errorf("get storage err %s", err)
	}

	if oldInfo == nil {
		oldInfo = new(cabi.MinerRewardInfo)
		oldInfo.RewardAmount = big.NewInt(0)
	}

	data, _ := cabi.MinerABI.PackVariable(cabi.VariableNameMinerReward, endHeight,
		RewardBlocks+oldInfo.RewardBlocks, block.Timestamp,
		new(big.Int).Add(RewardAmount, oldInfo.RewardAmount))
	err = ctx.SetStorage(contractaddress.MinerAddress.Bytes(), block.Address[:], data)
	if err != nil {
		return errors.New("save contract data err")
	}

	return nil
}

func (m *MinerReward) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.MinerRewardParam)

	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, input.Data)
	if err != nil {
		return nil, ErrUnpackMethod
	}

	if _, err := param.Verify(); err != nil {
		return nil, ErrCheckParam
	}

	if param.Coinbase != input.Address {
		return nil, ErrAccountInvalid
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = cfg.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	//block.Vote = types.ZeroBalance
	//block.Oracle = types.ZeroBalance
	//block.Storage = types.ZeroBalance
	//block.Network = types.ZeroBalance

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)
	if amBnf != nil {
		tmBnf := amBnf.Token(cfg.GasToken())
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
			Token:     cfg.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (m *MinerReward) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	param := new(cabi.MinerRewardParam)
	err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, block.Data)
	if err != nil {
		return common.ContractNoGap, nil, err
	}

	needHeight := param.EndHeight + common.PovMinerRewardHeightGapToLatest

	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return common.ContractRewardGapPov, needHeight, nil
	}

	nodeHeight := latestBlock.GetHeight()
	if nodeHeight < needHeight {
		return common.ContractRewardGapPov, needHeight, nil
	}

	return common.ContractNoGap, nil, err
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

func (m *MinerReward) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	param := new(cabi.MinerRewardParam)
	if err := cabi.MinerABI.UnpackMethod(param, cabi.MethodNameMinerReward, block.GetData()); err == nil {
		return param.Beneficial, nil
	} else {
		return types.ZeroAddress, err
	}
}
