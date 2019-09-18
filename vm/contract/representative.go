package contract

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/trie"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type RepReward struct{}

func (m *RepReward) GetMaxRewardInfo(ctx *vmstore.VMContext, account types.Address) (*cabi.RepRewardInfo, error) {
	rewardInfos, err := cabi.GetRepRewardInfosByAccount(ctx, account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, errors.New("failed to get storage for miner")
	}

	maxInfo := cabi.CalcMaxRepRewardInfo(rewardInfos)
	if maxInfo == nil {
		maxInfo = new(cabi.RepRewardInfo)
	}

	return maxInfo, nil
}

func (m *RepReward) GetAllRewardInfos(ctx *vmstore.VMContext, account types.Address) ([]*cabi.RepRewardInfo, error) {
	rewardInfos, err := cabi.GetRepRewardInfosByAccount(ctx, account)
	if err != nil && err != vmstore.ErrStorageNotFound {
		return nil, errors.New("failed to get storage for miner")
	}

	return rewardInfos, nil
}

func (m *RepReward) GetNodeRewardHeight(ctx *vmstore.VMContext) (uint64, error) {
	latestBlock, err := ctx.GetLatestPovBlock()
	if err != nil || latestBlock == nil {
		return 0, errors.New("failed to get latest block")
	}

	nodeHeight := latestBlock.GetHeight()
	if nodeHeight < common.RepRewardHeightGapToLatest {
		return 0, nil
	}
	nodeHeight = nodeHeight - common.RepRewardHeightGapToLatest

	nodeHeight = cabi.RepRoundPovHeight(nodeHeight, common.DPosOnlinePeriod)
	return nodeHeight, nil
}

func (m *RepReward) GetAvailRewardInfo(ctx *vmstore.VMContext, account types.Address, nodeHeight uint64, maxInfo *cabi.RepRewardInfo) (*cabi.RepRewardInfo, error) {
	availInfo := new(cabi.RepRewardInfo)

	startHeight := uint64(0)
	if maxInfo != nil {
		startHeight = maxInfo.EndHeight + 1
	}

	endHeight := cabi.RepCalcRewardEndHeight(startHeight, nodeHeight)
	if endHeight < startHeight {
		return availInfo, nil
	}

	availBlocks, availReward, err := m.calcRewardBlocksByHeight(ctx, account, startHeight, endHeight)
	if err != nil {
		return nil, err
	}

	availInfo.StartHeight = startHeight
	availInfo.EndHeight = endHeight
	availInfo.RewardBlocks = availBlocks
	availInfo.RewardAmount = availReward
	return availInfo, nil
}

func (m *RepReward) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (m *RepReward) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) (err error) {
	param := new(cabi.RepRewardParam)
	err = cabi.RepABI.UnpackMethod(param, cabi.MethodNameRepReward, block.Data)
	if err != nil {
		return err
	}

	if _, err := param.Verify(); err != nil {
		return err
	}

	if param.Account != block.Address {
		return errors.New("account is not representative")
	}

	if block.Token != common.ChainToken() {
		return errors.New("token is not chain token")
	}

	// check account exist
	am, _ := ctx.GetAccountMeta(param.Account)
	if am == nil {
		return errors.New("rep account not exist")
	}

	nodeRewardHeight, err := m.GetNodeRewardHeight(ctx)
	if err != nil {
		return err
	}

	if param.EndHeight > nodeRewardHeight {
		return fmt.Errorf("end height %d greater than node height %d", param.EndHeight, nodeRewardHeight)
	}

	// check same start & end height exist in old reward infos
	err = m.checkParamExistInOldRewardInfos(ctx, param)
	if err != nil {
		return err
	}

	calcRewardBlocks, calcRewardAmount, err := m.calcRewardBlocksByHeight(ctx, param.Account, param.StartHeight, param.EndHeight)
	if err != nil {
		return err
	}
	if calcRewardBlocks != param.RewardBlocks {
		return fmt.Errorf("calc blocks %d not equal param blocks %d", calcRewardBlocks, param.RewardBlocks)
	}
	if calcRewardAmount.Compare(param.RewardAmount) != types.BalanceCompEqual {
		return fmt.Errorf("calc reward %d not equal param reward %v", calcRewardAmount, param.RewardAmount)
	}

	block.Data, err = cabi.RepABI.PackMethod(cabi.MethodNameRepReward, param.Account, param.Beneficial, param.StartHeight, param.EndHeight, param.RewardBlocks)
	if err != nil {
		return err
	}

	return nil
}

func (m *RepReward) DoPending(block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.RepRewardParam)
	err := cabi.RepABI.UnpackMethod(param, cabi.MethodNameRepReward, block.Data)
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

func (m *RepReward) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.RepRewardParam)
	exist := false
	var calcRewardAmount types.Balance
	var calcRewardBlocks uint64

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

	// check same start & end height exist in old reward infos
	err = m.checkParamExistInOldRewardInfos(ctx, param)
	if err != nil {
		exist = true
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = param.Beneficial
	block.Token = common.GasToken()
	block.Link = input.GetHash()

	// pledge fields only for QLC token
	block.Vote = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance
	block.Network = types.ZeroBalance
	block.PoVHeight = input.PoVHeight

	amBnf, _ := ctx.GetAccountMeta(param.Beneficial)

	//the section has been extracted
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
		calcRewardBlocks, calcRewardAmount, err = m.calcRewardBlocksByHeight(ctx, param.Account, param.StartHeight, param.EndHeight)
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

	// save contract data to storage
	newRepData, err := cabi.RepABI.PackVariable(
		cabi.VariableNameRepRewardInfo,
		param.Beneficial,
		param.StartHeight,
		param.EndHeight,
		param.RewardBlocks,
		param.RewardAmount)
	if err != nil {
		return nil, err
	}

	repKey := cabi.GetRepRewardKey(param.Account, param.StartHeight)
	err = ctx.SetStorage(types.RepAddress.Bytes(), repKey, newRepData)
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

func (m *RepReward) checkParamExistInOldRewardInfos(ctx *vmstore.VMContext, param *cabi.RepRewardParam) error {
	oldRewardInfos, err := cabi.GetRepRewardInfosByAccount(ctx, param.Account)
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

func (m *RepReward) calcRewardBlocksByHeight(ctx *vmstore.VMContext, account types.Address, startHeight, endHeight uint64) (uint64, types.Balance, error) {
	rewardBlocks := uint64(0)
	rewardAmount := types.Balance{}
	keyBytes := types.PovCreateRepStateKey(account)

	for curHeight := startHeight + common.DPosOnlinePeriod - 1; curHeight <= endHeight; curHeight += common.DPosOnlinePeriod {
		block, err := ctx.GetPovHeaderByHeight(curHeight)
		if block == nil {
			return rewardBlocks, rewardAmount, err
		}

		stateHash := block.GetStateHash()
		stateTrie := trie.NewTrie(ctx.GetLedger().Store, &stateHash, nil)

		valBytes := stateTrie.GetValue(keyBytes)
		if len(valBytes) <= 0 {
			continue
		}

		rs := new(types.PovRepState)
		err = rs.Deserialize(valBytes)
		if err != nil {
			ctx.GetLogger().Errorf("deserialize old rep state err %s", err)
			continue
		}

		if rs.Status == types.PovStatusOnline && rs.Height >= curHeight+1-common.DPosOnlinePeriod {
			rewardBlocks += common.DPosOnlinePeriod
		}
	}

	rewardAmount.SetUint64(rewardBlocks * uint64(common.PovMinerRewardPerBlock*common.PovMinerRewardRatioRep/100))
	return rewardBlocks, rewardAmount, nil
}

func (m *RepReward) GetRefundData() []byte {
	return []byte{1}
}
