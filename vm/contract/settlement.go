/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type CreateContract struct {
}

func (c *CreateContract) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (c *CreateContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	// verify send block data
	param := new(cabi.CreateContractParam)
	err := param.FromABI(input.Data)
	if err != nil {
		return nil, err
	}
	address, err := param.Address()
	if err != nil {
		return nil, err
	}
	if b, err := ctx.GetStorage(types.SettlementAddress[:], address[:]); err == nil && len(b) > 0 {
		if _, err := cabi.ParseContractParam(b); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid send block[%s] data", input.GetHash().String())
	}

	rxMeta, _ := ctx.GetAccountMeta(input.Address)
	// qgas token should be exist
	rxToken := rxMeta.Token(input.Token)
	txHash := input.GetHash()

	block.Type = types.ContractReward
	block.Address = input.Address
	block.Link = txHash
	block.Token = input.Token
	block.Extra = types.ZeroHash
	block.Vote = types.ZeroBalance
	block.Network = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance

	block.Balance = rxToken.Balance
	block.Previous = rxToken.Header
	block.Representative = input.Representative

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: input.Address,
			BlockType: types.ContractReward,
			Amount:    types.ZeroBalance,
			Token:     input.Token,
			Data:      []byte{},
		},
	}, nil
}

func (c *CreateContract) GetRefundData() []byte {
	return []byte{1}
}

func (c *CreateContract) GetDescribe() Describe {
	return Describe{withPending: true, withSignature: false}
}

func (c *CreateContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	// check token is QGAS
	if block.Token != common.GasToken() {
		return nil, nil, fmt.Errorf("invalid token: %s", block.Token.String())
	}

	param := new(cabi.CreateContractParam)
	err := param.FromABI(block.Data)
	if err != nil {
		return nil, nil, err
	}

	if b, err := param.Verify(); err == nil && b {
		// check balance
		amount, err := ctx.CalculateAmount(block)
		if err != nil {
			return nil, nil, err
		}

		if exp, err := param.Balance(); err != nil {
			return nil, nil, err
		} else {
			if amount.Compare(exp) == types.BalanceCompSmaller {
				return nil, nil, fmt.Errorf("not enough balance, exp: %s, act: %s", exp.String(), amount.String())
			}
		}

		// make sure that the same block only process once
		address, err := param.Address()
		if err != nil {
			return nil, nil, err
		}
		storage, err := ctx.GetStorage(types.SettlementAddress[:], address[:])
		if err != nil && err != vmstore.ErrStorageNotFound {
			return nil, nil, err
		}

		if len(storage) > 0 {
			// verify saved data
			param2, err := cabi.ParseContractParam(storage)
			if err != nil {
				return nil, nil, err
			}
			if _, err := param2.Equal(param); err != nil {
				return nil, nil, errors.New("invalid saved create contract data")
			}
		} else {
			if data, err := param.ToContractParam().MarshalMsg(nil); err == nil {
				if err := ctx.SetStorage(types.SettlementAddress[:], address[:], data); err != nil {
					return nil, nil, err
				}
			} else {
				return nil, nil, err
			}
		}

		return &types.PendingKey{
				Address: param.PartyA,
				Hash:    block.GetHash(),
			}, &types.PendingInfo{
				Source: types.Address(block.Link),
				Amount: types.ZeroBalance,
				Type:   block.Token,
			}, nil
	} else {
		return nil, nil, err
	}
}

func (c *CreateContract) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

type SignContract struct {
}

func (s *SignContract) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (s *SignContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	// verify send block data
	param := new(cabi.SignContractParam)
	err := param.FromABI(input.Data)
	if err != nil {
		return nil, err
	}
	if b, err := ctx.GetStorage(types.SettlementAddress[:], param.ContractAddress[:]); err == nil && len(b) > 0 {
		if _, err := cabi.ParseContractParam(b); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid send block[%s] data", input.GetHash().String())
	}

	rxMeta, _ := ctx.GetAccountMeta(input.Address)
	// qgas token should be exist
	rxToken := rxMeta.Token(input.Token)
	txHash := input.GetHash()

	block.Type = types.ContractReward
	block.Address = input.Address
	block.Link = txHash
	block.Token = input.Token
	block.Extra = types.ZeroHash
	block.Vote = types.ZeroBalance
	block.Network = types.ZeroBalance
	block.Oracle = types.ZeroBalance
	block.Storage = types.ZeroBalance

	block.Balance = rxToken.Balance
	block.Previous = rxToken.Header
	block.Representative = input.Representative

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: input.Address,
			BlockType: types.ContractReward,
			Amount:    types.ZeroBalance,
			Token:     input.Token,
			Data:      []byte{},
		},
	}, nil
}

func (s *SignContract) GetRefundData() []byte {
	return []byte{1}
}

func (s *SignContract) GetDescribe() Describe {
	return Describe{withSignature: false, withPending: true}
}

func (s *SignContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.SignContractParam)
	err := param.FromABI(block.Data)
	if err != nil {
		return nil, nil, err
	}

	// verify block data
	if _, err := param.Verify(block.Address); err != nil {
		return nil, nil, err
	}

	if storage, err := ctx.GetStorage(types.SettlementAddress[:], param.ContractAddress[:]); err != nil {
		return nil, nil, err
	} else {
		if len(storage) > 0 {
			if cp, err := cabi.ParseContractParam(storage); err != nil {
				return nil, nil, err
			} else {
				// verify partyB
				if cp.PartyB != block.Address {
					return nil, nil, fmt.Errorf("invalid partyB, exp: %s, act: %s", cp.PartyB.String(), block.Address.String())
				}

				cp.ConfirmDate = param.ConfirmDate
				cp.SignatureB = &param.SignatureB

				if data, err := cp.MarshalMsg(nil); err == nil {
					// save confirm data
					if err := ctx.SetStorage(types.SettlementAddress[:], param.ContractAddress[:], data); err == nil {
						return &types.PendingKey{
								Address: block.Address,
								Hash:    block.GetHash(),
							}, &types.PendingInfo{
								Source: types.Address(block.Link),
								Amount: types.ZeroBalance,
								Type:   block.Token,
							}, nil
					} else {
						return nil, nil, err
					}
				} else {
					return nil, nil, err
				}
			}
		} else {
			return nil, nil, fmt.Errorf("invalid saved contract data of %s", param.ContractAddress.String())
		}
	}
}

func (s *SignContract) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}

type ProcessCDR struct {
}

func (p *ProcessCDR) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (p *ProcessCDR) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (p *ProcessCDR) GetRefundData() []byte {
	return []byte{1}
}

func (p *ProcessCDR) GetDescribe() Describe {
	return Describe{withPending: true, withSignature: false}
}

func (p *ProcessCDR) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	panic("implement me")
}

func (p *ProcessCDR) DoGapPov(ctx *vmstore.VMContext, block *types.StateBlock) (uint64, error) {
	return 0, nil
}
