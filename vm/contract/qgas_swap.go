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

	pledgeKey := abi.GetQGasSwapKey(param.PledgeAddress, block.GetHash())
	if _, err := ctx.GetStorage(nil, pledgeKey); err == nil {
		return nil, nil, errors.New("repeatedly pledge info")
	} else {
		pledgeInfo := abi.QGasSwapInfo{
			Amount:      param.Amount,
			FromAddress: param.PledgeAddress,
			ToAddress:   param.ToAddress,
			SendHash:    block.GetHash(),
			LinkHash:    types.ZeroHash,
			SwapType:    abi.QGasPledge,
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
				Source: param.PledgeAddress,
				Amount: types.Balance{Int: param.Amount},
				Type:   block.Token,
			}, nil
	}
}

func (q *QGasPledge) DoReceive(ctx *vmstore.VMContext, block, input *types.StateBlock) ([]*ContractBlock, error) {
	param, err := abi.ParseQGasPledgeParam(input.Data)
	if err != nil {
		return nil, err
	}
	// verify block data
	if _, err := param.Verify(); err != nil {
		return nil, err
	}

	if param.PledgeAddress != input.Address {
		return nil, ErrAccountInvalid
	}

	if block.Link != input.GetHash() {
		return nil, fmt.Errorf("invalid link %s, %s", block.Link, input.GetHash())
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

	pledgeKey := abi.GetQGasSwapKey(param.WithdrawAddress, param.LinkHash)
	if _, err := ctx.GetStorage(nil, pledgeKey); err == nil {
		return nil, nil, errors.New("repeatedly pledge info")
	} else {
		pledgeInfo := abi.QGasSwapInfo{
			Amount:      param.Amount,
			FromAddress: param.FromAddress,
			ToAddress:   param.WithdrawAddress,
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
				Address: param.WithdrawAddress,
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

	if block.Address.IsZero() {
		// generate contract reward block
		block.Address = param.WithdrawAddress
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

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: param.WithdrawAddress,
			BlockType: types.ContractReward,
			Amount:    types.Balance{Int: param.Amount},
			Token:     cfg.GasToken(),
			Data:      []byte{},
		},
	}, nil
}
