/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"go.uber.org/zap"
)

type MintageApi struct {
	logger    *zap.SugaredLogger
	ledger    *ledger.Ledger
	vmContext *vmstore.VMContext
	mintage   *contract.Mintage
	withdraw  *contract.WithdrawMintage
}

func NewMintageApi(ledger *ledger.Ledger) *MintageApi {
	return &MintageApi{ledger: ledger, vmContext: vmstore.NewVMContext(ledger),
		logger: log.NewLogger("api_mintage"), mintage: &contract.Mintage{},
		withdraw: &contract.WithdrawMintage{}}
}

type MintageParams struct {
	SelfAddr    types.Address `json:"selfAddr"`
	PrevHash    types.Hash    `json:"prevHash"`
	TokenName   string        `json:"tokenName"`
	TokenSymbol string        `json:"tokenSymbol"`
	TotalSupply string        `json:"totalSupply"`
	Decimals    uint8         `json:"decimals"`
	Beneficial  types.Address `json:"beneficial"`
	NEP5TxId    string        `json:"nep5TxId"`
}

func (m *MintageApi) GetMintageData(param *MintageParams) ([]byte, error) {
	tokenId := cabi.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	return cabi.MintageABI.PackMethod(cabi.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol, totalSupply, param.Decimals)
}

func (m *MintageApi) GetMintageBlock(param *MintageParams) (*types.StateBlock, error) {
	tokenId := cabi.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	data, err := cabi.MintageABI.PackMethod(cabi.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol, totalSupply, param.Decimals, param.Beneficial, param.NEP5TxId)
	if err != nil {
		return nil, err
	}

	am, err := m.ledger.GetAccountMeta(param.SelfAddr)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.SelfAddr.String())
	}

	minPledgeAmount := types.Balance{Int: contract.MinPledgeAmount}

	if tm.Balance.Compare(minPledgeAmount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("not enough balance %s, expect %s", tm.Balance, minPledgeAmount)
	}
	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.SelfAddr,
		Balance:        tm.Balance.Sub(minPledgeAmount),
		Previous:       tm.Header,
		Link:           types.Hash(types.MintageAddress),
		Representative: tm.Representative,
		Data:           data,
		Timestamp:      time.Now().UTC().Unix(),
	}

	err = m.mintage.DoSend(m.vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MintageApi) GetRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	reward := &types.StateBlock{}

	blocks, err := m.mintage.DoReceive(m.vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		reward.Timestamp = time.Now().UTC().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		if h == nil {
			return nil, errors.New("trie hash is nil")
		}
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate mintage reward block")
}

func (m *MintageApi) GetWithdrawMintageData(tokenId types.Hash) ([]byte, error) {
	return cabi.MintageABI.PackMethod(cabi.MethodNameMintageWithdraw, tokenId)
}

type WithdrawParams struct {
	SelfAddr types.Address `json:"selfAddr"`
	TokenId  types.Hash    `json:"tokenId"`
}

func (m *MintageApi) GetWithdrawMintageBlock(param *WithdrawParams) (*types.StateBlock, error) {
	tm, _ := m.ledger.GetTokenMeta(param.SelfAddr, common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.SelfAddr.String())
	}
	data, err := cabi.MintageABI.PackMethod(cabi.MethodNameMintageWithdraw, param.TokenId)
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.SelfAddr,
		Balance:        tm.Balance,
		Vote:           types.ZeroBalance,
		Network:        types.ZeroBalance,
		Storage:        types.ZeroBalance,
		Oracle:         types.ZeroBalance,
		Previous:       tm.Header,
		Link:           types.Hash(types.MintageAddress),
		Representative: tm.Representative,
		Data:           data,
		Timestamp:      time.Now().UTC().Unix(),
	}

	err = m.withdraw.DoSend(m.vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MintageApi) GetWithdrawRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	reward := &types.StateBlock{}

	blocks, err := m.withdraw.DoReceive(m.vmContext, reward, input)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		reward.Timestamp = time.Now().UTC().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate withdraw reward block")
}
