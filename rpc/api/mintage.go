/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type MintageAPI struct {
	logger   *zap.SugaredLogger
	l        ledger.Store
	mintage  *contract.Mintage
	withdraw *contract.WithdrawMintage
	cc       *chainctx.ChainContext
}

func NewMintageApi(cfgFile string, l ledger.Store) *MintageAPI {
	api := &MintageAPI{
		l:        l,
		mintage:  &contract.Mintage{},
		withdraw: &contract.WithdrawMintage{},
		logger:   log.NewLogger("api_mintage"),
		cc:       chainctx.NewChainContext(cfgFile),
	}
	return api
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

func (m *MintageAPI) GetMintageData(param *MintageParams) ([]byte, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	tokenId := mintage.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	return mintage.MintageABI.PackMethod(mintage.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol,
		totalSupply, param.Decimals, param.Beneficial, param.NEP5TxId)
}

func (m *MintageAPI) GetMintageBlock(param *MintageParams) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	if !m.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	tokenId := mintage.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	data, err := mintage.MintageABI.PackMethod(mintage.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol,
		totalSupply, param.Decimals, param.Beneficial, param.NEP5TxId)
	if err != nil {
		return nil, err
	}

	am, err := m.l.GetAccountMeta(param.SelfAddr)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.SelfAddr.String())
	}

	minPledgeAmount := types.Balance{Int: contract.MinPledgeAmount}

	if tm.Balance.Compare(minPledgeAmount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have no enough balance %s, expect %s", param.SelfAddr.String(), tm.Balance.String(), minPledgeAmount.String())
	}
	povHeader, err := m.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.SelfAddr,
		Balance:        tm.Balance.Sub(minPledgeAmount),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.MintageAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(m.l, &contractaddress.MintageAddress)
	err = m.mintage.DoSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MintageAPI) GetRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if input == nil {
		return nil, ErrParameterNil
	}
	if !m.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	reward := &types.StateBlock{}
	vmContext := vmstore.NewVMContext(m.l, &contractaddress.MintageAddress)
	blocks, err := m.mintage.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		povHeader, err := m.l.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		reward.PoVHeight = povHeader.GetHeight()
		reward.Timestamp = common.TimeNow().Unix()
		h := vmstore.TrieHash(blocks[0].VMContext)
		if h == nil {
			return nil, errors.New("trie hash is nil")
		}
		reward.Extra = h
		return reward, nil
	}

	return nil, errors.New("can not generate mintage reward block")
}

func (m *MintageAPI) GetWithdrawMintageData(tokenId types.Hash) ([]byte, error) {
	return mintage.MintageABI.PackMethod(mintage.MethodNameMintageWithdraw, tokenId)
}

func (p *MintageAPI) ParseTokenInfo(data []byte) (*types.TokenInfo, error) {
	return mintage.ParseTokenInfo(data)
}

type WithdrawParams struct {
	SelfAddr types.Address `json:"selfAddr"`
	TokenId  types.Hash    `json:"tokenId"`
}

func (m *MintageAPI) GetWithdrawMintageBlock(param *WithdrawParams) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	if !m.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}
	tm, _ := m.l.GetTokenMeta(param.SelfAddr, config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.SelfAddr.String())
	}
	data, err := mintage.MintageABI.PackMethod(mintage.MethodNameMintageWithdraw, param.TokenId)
	if err != nil {
		return nil, err
	}
	povHeader, err := m.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	send := &types.StateBlock{
		Type:    types.ContractSend,
		Token:   tm.Type,
		Address: param.SelfAddr,
		Balance: tm.Balance,
		//Vote:           types.ZeroBalance,
		//Network:        types.ZeroBalance,
		//Storage:        types.ZeroBalance,
		//Oracle:         types.ZeroBalance,
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.MintageAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}
	vmContext := vmstore.NewVMContext(m.l, &contractaddress.MintageAddress)
	err = m.withdraw.DoSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MintageAPI) GetWithdrawRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if input == nil {
		return nil, ErrParameterNil
	}
	if !m.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}
	reward := &types.StateBlock{}
	vmContext := vmstore.NewVMContext(m.l, &contractaddress.MintageAddress)
	blocks, err := m.withdraw.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		povHeader, err := m.l.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		reward.PoVHeight = povHeader.GetHeight()
		reward.Timestamp = common.TimeNow().Unix()
		h := vmstore.TrieHash(blocks[0].VMContext)
		reward.Extra = h
		return reward, nil
	}

	return nil, errors.New("can not generate withdraw reward block")
}
