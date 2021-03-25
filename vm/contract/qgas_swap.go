package contract

import (
	"errors"
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var QGasSwapContract = NewChainContract(
	map[string]Contract{
		abi.MethodQGasPledge: &QGasPledge{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
				},
			},
		},
		abi.MethodQGasWithdraw: &QGasWithdraw{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
				},
			},
		},
	},
	abi.QGasSwapABI,
	abi.JsonQGasSwap,
)

type QGasPledge struct {
	BaseContract
}

func (q *QGasPledge) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.Token != cfg.GasToken() {
		return nil, nil, ErrToken
	}

	// make sure the gas account is activated
	if _, err := ctx.GetTokenMeta(block.GetAddress(), cfg.GasToken()); err != nil {
		return nil, nil, err
	}

	param, err := abi.ParseQGasPledgeParam(block.Data)
	if err != nil {
		return nil, nil, err
	}
	// verify block data
	if _, err := param.Verify(); err != nil {
		return nil, nil, err
	}

	pledgeKey := abi.GetQGasSwapKey(param.FromAddress, block.GetHash())
	if _, err := ctx.GetStorage(nil, pledgeKey); err == nil {
		return nil, nil, errors.New("repeatedly pledge info")
	}
	pledgeInfo := abi.QGasSwapInfo{
		SwapType:    abi.QGasPledge,
		SendHash:    block.GetHash(),
		Amount:      param.Amount,
		FromAddress: param.FromAddress,
		ToAddress:   param.ToAddress,
		LinkHash:    types.ZeroHash,
		Time:        time.Now().Unix(),
	}

	pledgeData, err := pledgeInfo.ToABI()
	if err != nil {
		return nil, nil, fmt.Errorf("error pledge info abi: %s", err)
	}
	err = ctx.SetStorage(nil, pledgeKey, pledgeData)
	if err != nil {
		return nil, nil, ErrSetStorage
	} else {
		return &types.PendingKey{
				Address: param.ToAddress,
				Hash:    block.GetHash(),
			}, &types.PendingInfo{
				Type:   block.Token,
				Source: param.FromAddress,
				Amount: types.Balance{Int: param.Amount},
			}, nil
	}
}

func (q *QGasPledge) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	if block.Link != input.GetHash() {
		return nil, fmt.Errorf("invalid link %s, %s", block.Link, input.GetHash())
	}

	pledgeParam, err := abi.ParseQGasPledgeParam(input.Data)
	if err != nil {
		return nil, err
	} else {
		if pledgeParam.FromAddress != input.Address {
			return nil, ErrAccountInvalid
		}
		if _, err := pledgeParam.Verify(); err != nil {
			return nil, err
		}
	}

	pledgeKey := abi.GetQGasSwapKey(pledgeParam.FromAddress, input.GetHash())
	pledgeData, err := ctx.GetStorage(nil, pledgeKey)
	if err != nil {
		return nil, fmt.Errorf("swap info not found, %s", err)
	}
	pledgeInfo, err := abi.ParseQGasSwapInfo(pledgeData)
	if err != nil {
		return nil, fmt.Errorf("parse SwapInfo error: %s", err)
	}
	if pledgeInfo.RewardHash != types.ZeroHash {
		return nil, fmt.Errorf("repeatedly reward block")
	}

	if block.Address.IsZero() {
		// generate contract reward block
		block.Address = pledgeParam.ToAddress
		am, _ := ctx.GetAccountMeta(block.Address)
		if am != nil {
			tm := am.Token(cfg.GasToken())
			if tm == nil {
				block.Balance = types.NewBalance(0)
				if len(am.Tokens) > 0 {
					block.Representative = am.Tokens[0].Representative
				} else {
					block.Representative = input.Representative
				}
				block.Previous = types.ZeroHash
			} else {
				block.Balance = tm.Balance.Add(types.Balance{Int: pledgeParam.Amount})
				block.Vote = types.ToBalance(am.CoinVote)
				block.Network = types.ToBalance(am.CoinNetwork)
				block.Oracle = types.ToBalance(am.CoinOracle)
				block.Storage = types.ToBalance(am.CoinStorage)

				block.Representative = tm.Representative
				block.Previous = tm.Header
			}
		} else {
			block.Balance = types.Balance{Int: pledgeParam.Amount}
			block.Previous = types.ZeroHash
			block.Representative = input.Representative
		}
	}

	pledgeInfo.RewardHash = block.GetHash()
	data, err := pledgeInfo.ToABI()
	if err != nil {
		return nil, fmt.Errorf("pledge info abi: %s", err)
	}
	err = ctx.SetStorage(nil, pledgeKey, data)
	if err != nil {
		return nil, ErrSetStorage
	} else {
		return []*ContractBlock{
			{
				BlockType: types.ContractReward,
				VMContext: ctx,
				Token:     cfg.GasToken(),
				Block:     block,
				ToAddress: pledgeParam.ToAddress,
				Amount:    types.Balance{Int: pledgeParam.Amount},
				Data:      []byte{},
			},
		}, nil
	}
}

type QGasWithdraw struct {
	BaseContract
}

func (q *QGasWithdraw) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.Token != cfg.GasToken() {
		return nil, nil, ErrToken
	}

	// make sure the gas account is activated
	_, err := ctx.GetTokenMeta(block.GetAddress(), cfg.GasToken())
	if err != nil {
		return nil, nil, err
	}

	param, err := abi.ParseQGasWithdrawParam(block.Data)
	if err != nil {
		return nil, nil, err
	}
	// verify block data
	if _, err := param.Verify(); err != nil {
		return nil, nil, err
	}

	pledgeKey := abi.GetQGasSwapKey(param.ToAddress, param.LinkHash)
	if _, err := ctx.GetStorage(nil, pledgeKey); err == nil {
		return nil, nil, errors.New("repeatedly pledge info")
	} else {
		pledgeInfo := abi.QGasSwapInfo{
			Amount:      param.Amount,
			FromAddress: param.FromAddress,
			ToAddress:   param.ToAddress,
			SendHash:    block.GetHash(),
			LinkHash:    param.LinkHash,
			SwapType:    abi.QGasWithdraw,
			Time:        time.Now().Unix(),
		}

		pledgeData, err := pledgeInfo.ToABI()
		if err != nil {
			return nil, nil, fmt.Errorf("error pledge info abi: %s", err)
		}
		err = ctx.SetStorage(nil, pledgeKey, pledgeData)
		if err != nil {
			return nil, nil, ErrSetStorage
		}

		return &types.PendingKey{
				Address: param.ToAddress,
				Hash:    block.GetHash(),
			}, &types.PendingInfo{
				Source: param.FromAddress,
				Amount: types.Balance{Int: param.Amount},
				Type:   block.Token,
			}, nil
	}
}

func (q *QGasWithdraw) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param, err := abi.ParseQGasWithdrawParam(input.Data)
	if err != nil {
		return nil, err
	}
	// verify block data
	if _, err := param.Verify(); err != nil {
		return nil, err
	}

	if param.FromAddress != input.Address {
		return nil, ErrAccountInvalid
	}

	if block.Link != input.GetHash() {
		return nil, fmt.Errorf("invalid link %s, %s", block.Link, input.GetHash())
	}

	pledgeKey := abi.GetQGasSwapKey(param.ToAddress, param.LinkHash)
	pledgeData, err := ctx.GetStorage(nil, pledgeKey)
	if err != nil {
		return nil, fmt.Errorf("swap info not found: %s", err)
	}
	pledgeInfo, err := abi.ParseQGasSwapInfo(pledgeData)
	if err != nil {
		return nil, fmt.Errorf("parse SwapInfo error: %s", err)
	}

	if pledgeInfo.RewardHash != types.ZeroHash {
		return nil, fmt.Errorf("repeatedly reward block")
	}

	if block.Address.IsZero() {
		// generate contract reward block
		block.Address = param.ToAddress
		am, _ := ctx.GetAccountMeta(block.Address)
		if am != nil {
			tm := am.Token(cfg.GasToken())
			if tm != nil {
				block.Balance = tm.Balance.Add(types.Balance{Int: param.Amount})
				block.Vote = types.ToBalance(am.CoinVote)
				block.Network = types.ToBalance(am.CoinNetwork)
				block.Oracle = types.ToBalance(am.CoinOracle)
				block.Storage = types.ToBalance(am.CoinStorage)

				block.Representative = tm.Representative
				block.Previous = tm.Header
			} else {
				block.Balance = types.NewBalance(0)
				if len(am.Tokens) > 0 {
					block.Representative = am.Tokens[0].Representative
				} else {
					block.Representative = input.Representative
				}
				block.Previous = types.ZeroHash
			}
		} else {
			block.Balance = types.Balance{Int: param.Amount}
			block.Representative = input.Representative
			block.Previous = types.ZeroHash
		}
	}

	pledgeInfo.RewardHash = block.GetHash()
	data, err := pledgeInfo.ToABI()
	if err != nil {
		return nil, fmt.Errorf("pledge info abi: %s", err)
	}
	err = ctx.SetStorage(nil, pledgeKey, data)
	if err != nil {
		return nil, ErrSetStorage
	} else {
		return []*ContractBlock{
			{
				VMContext: ctx,
				Block:     block,
				ToAddress: param.ToAddress,
				BlockType: types.ContractReward,
				Amount:    types.Balance{Int: param.Amount},
				Token:     cfg.GasToken(),
				Data:      []byte{},
			},
		}, nil
	}
}
