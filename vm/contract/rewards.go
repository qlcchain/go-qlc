/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type AirdropRewords struct {
}

func (ar *AirdropRewords) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (ar *AirdropRewords) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameAirdropRewards, block.Data)
	if err != nil {
		return err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return err
	}

	return nil
}

func (ar *AirdropRewords) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameAirdropRewards, block.Data)
	if err != nil {
		return nil, err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return nil, err
	}

	//verify is QGAS
	amount, _ := ctx.CalculateAmount(input)
	if amount.Sign() > 0 && input.Token != common.GasToken() {
		txHash := input.GetHash()
		txAddress := input.Address
		txMeta, err := ctx.GetAccountMeta(txAddress)
		if err != nil {
			return nil, err
		}
		txToken := txMeta.Token(common.GasToken())
		rxAddress := types.Address(input.Link)

		rxMeta, _ := ctx.GetAccountMeta(rxAddress)

		block.Type = types.ContractReward
		block.Address = rxAddress
		block.Link = txHash
		block.Token = common.GasToken()
		block.Extra = types.ZeroHash
		block.Timestamp = common.TimeNow().UTC().Unix()

		// already have account
		if rxMeta != nil && len(rxMeta.Tokens) > 0 {
			block.Vote = rxMeta.CoinVote
			block.Oracle = rxMeta.CoinOracle
			block.Network = rxMeta.CoinNetwork
			block.Storage = rxMeta.CoinStorage
			if rxToken := rxMeta.Token(common.GasToken()); rxToken != nil {
				//already have token
				block.Balance = rxToken.Balance.Add(amount)
				block.Previous = rxToken.Header
				block.Representative = txToken.Representative
			} else {
				block.Balance = amount
				block.Previous = types.ZeroHash
				//use other token's rep
				block.Representative = rxMeta.Tokens[0].Representative
			}
		} else {
			block.Balance = amount
			block.Vote = types.ZeroBalance
			block.Network = types.ZeroBalance
			block.Oracle = types.ZeroBalance
			block.Storage = types.ZeroBalance
			block.Previous = types.ZeroHash
			block.Representative = input.Representative
		}

		info := &cabi.RewardsInfo{
			Type:    uint8(cabi.Rewards),
			From:    &input.Address,
			To:      &rxAddress,
			SHeader: txToken.Header[:],
			RHeader: block.Previous[:],
			Amount:  amount.Int,
		}

		key := cabi.GetRewardsKey(param.Id, param.SHeader, param.RHeader)

		if data, err := ctx.GetStorage(types.RewardsAddress[:], key); err != nil && err != vmstore.ErrStorageNotFound {
			return nil, err
		} else {
			//already exist
			if len(data) > 0 {
				if rewardsInfo, err := cabi.ParseRewardsInfo(data); err == nil {
					if !bytes.EqualFold(rewardsInfo.SHeader, info.SHeader) || !bytes.EqualFold(rewardsInfo.RHeader, info.RHeader) ||
						rewardsInfo.Amount.Cmp(info.Amount) != 0 || rewardsInfo.Type != info.Type ||
						rewardsInfo.From != info.From || rewardsInfo.To != info.To {
						return nil, errors.New("invalid saved rewards data")
					}
				} else {
					return nil, err
				}
			} else {
				if data, err := cabi.RewardsABI.PackVariable(cabi.VariableNameRewards, info); err == nil {
					if err := ctx.SetStorage(types.RewardsAddress[:], key, data); err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			}
		}

		return []*ContractBlock{
			{
				VMContext: ctx,
				Block:     block,
				ToAddress: rxAddress,
				BlockType: types.ContractReward,
				Amount:    amount,
				Token:     input.Token,
				Data:      []byte{},
			},
		}, nil

	} else {
		return nil, fmt.Errorf("invalid token hash %s", input.Token.String())
	}
}

func (*AirdropRewords) GetRefundData() []byte {
	return []byte{1}
}

type ConfidantRewards struct {
}

func (*ConfidantRewards) GetFee(ctx *vmstore.VMContext, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*ConfidantRewards) DoSend(ctx *vmstore.VMContext, block *types.StateBlock) error {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameConfidantRewards, block.Data)
	if err != nil {
		return err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return err
	}

	return nil
}

func (*ConfidantRewards) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(cabi.RewardsParam)
	err := cabi.MintageABI.UnpackMethod(param, cabi.MethodNameConfidantRewards, block.Data)
	if err != nil {
		return nil, err
	}

	if _, err := param.Verify(block.Address); err != nil {
		return nil, err
	}

	//verify is QGAS
	amount, _ := ctx.CalculateAmount(input)
	if amount.Sign() > 0 && input.Token != common.GasToken() {
		txHash := input.GetHash()
		txAddress := input.Address
		txMeta, err := ctx.GetAccountMeta(txAddress)
		if err != nil {
			return nil, err
		}
		txToken := txMeta.Token(common.GasToken())
		rxAddress := types.Address(input.Link)

		rxMeta, _ := ctx.GetAccountMeta(rxAddress)

		block.Type = types.ContractReward
		block.Address = rxAddress
		block.Link = txHash
		block.Token = common.GasToken()
		block.Extra = types.ZeroHash
		block.Timestamp = common.TimeNow().UTC().Unix()

		// already have account
		if rxMeta != nil && len(rxMeta.Tokens) > 0 {
			block.Vote = rxMeta.CoinVote
			block.Oracle = rxMeta.CoinOracle
			block.Network = rxMeta.CoinNetwork
			block.Storage = rxMeta.CoinStorage
			if rxToken := rxMeta.Token(common.GasToken()); rxToken != nil {
				//already have token
				block.Balance = rxToken.Balance.Add(amount)
				block.Previous = rxToken.Header
				block.Representative = txToken.Representative
			} else {
				block.Balance = amount
				block.Previous = types.ZeroHash
				//use other token's rep
				block.Representative = rxMeta.Tokens[0].Representative
			}
		} else {
			block.Balance = amount
			block.Vote = types.ZeroBalance
			block.Network = types.ZeroBalance
			block.Oracle = types.ZeroBalance
			block.Storage = types.ZeroBalance
			block.Previous = types.ZeroHash
			block.Representative = input.Representative
		}

		info := &cabi.RewardsInfo{
			Type:    uint8(cabi.Confidant),
			From:    &input.Address,
			To:      &rxAddress,
			SHeader: txToken.Header[:],
			RHeader: block.Previous[:],
			Amount:  amount.Int,
		}

		key := cabi.GetConfidantKey(rxAddress, param.Id, param.SHeader, param.RHeader)

		if data, err := ctx.GetStorage(types.RewardsAddress[:], key); err != nil && err != vmstore.ErrStorageNotFound {
			return nil, err
		} else {
			//already exist
			if len(data) > 0 {
				if rewardsInfo, err := cabi.ParseRewardsInfo(data); err == nil {
					if !bytes.EqualFold(rewardsInfo.SHeader, info.SHeader) || !bytes.EqualFold(rewardsInfo.RHeader, info.RHeader) ||
						rewardsInfo.Amount.Cmp(info.Amount) != 0 || rewardsInfo.Type != info.Type ||
						rewardsInfo.From != info.From || rewardsInfo.To != info.To {
						return nil, errors.New("invalid saved confidant data")
					}
				} else {
					return nil, err
				}
			} else {
				if data, err := cabi.RewardsABI.PackVariable(cabi.VariableNameRewards, info); err == nil {
					if err := ctx.SetStorage(types.RewardsAddress[:], key, data); err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			}
		}

		return []*ContractBlock{
			{
				VMContext: ctx,
				Block:     block,
				ToAddress: rxAddress,
				BlockType: types.ContractReward,
				Amount:    amount,
				Token:     input.Token,
				Data:      []byte{},
			},
		}, nil

	} else {
		return nil, fmt.Errorf("invalid token hash %s", input.Token.String())
	}
}

func (*ConfidantRewards) GetRefundData() []byte {
	return []byte{2}
}
