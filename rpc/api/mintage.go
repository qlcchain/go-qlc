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
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type MintageApi struct {
	logger    *zap.SugaredLogger
	ledger    *ledger.Ledger
	mintage   *contract.Mintage
	withdraw  *contract.WithdrawMintage
	syncState atomic.Value
}

func NewMintageApi(l *ledger.Ledger, eb event.EventBus) *MintageApi {
	api := &MintageApi{
		ledger:   l,
		mintage:  &contract.Mintage{},
		withdraw: &contract.WithdrawMintage{},
		logger:   log.NewLogger("api_mintage"),
	}
	api.syncState.Store(common.SyncNotStart)
	_, _ = eb.SubscribeSync(common.EventPovSyncState, api.OnPovSyncState)
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

func (m *MintageApi) OnPovSyncState(state common.SyncState) {
	m.logger.Infof("mintage receive pov sync state [%s]", state)
	m.syncState.Store(state)
}

func (m *MintageApi) GetMintageData(param *MintageParams) ([]byte, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	tokenId := cabi.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	return cabi.MintageABI.PackMethod(cabi.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol, totalSupply, param.Decimals)
}

func (m *MintageApi) GetMintageBlock(param *MintageParams) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	if ss := m.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}

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
		return nil, fmt.Errorf("%s have no enough balance %s, expect %s", param.SelfAddr.String(), tm.Balance.String(), minPledgeAmount.String())
	}
	povHeader, err := m.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
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
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(m.ledger)
	err = m.mintage.DoSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MintageApi) GetRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if input == nil {
		return nil, ErrParameterNil
	}
	if ss := m.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}

	reward := &types.StateBlock{}
	vmContext := vmstore.NewVMContext(m.ledger)
	blocks, err := m.mintage.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		povHeader, err := m.ledger.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		reward.PoVHeight = povHeader.GetHeight()
		reward.Timestamp = common.TimeNow().Unix()
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
	if param == nil {
		return nil, ErrParameterNil
	}
	if ss := m.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}
	tm, _ := m.ledger.GetTokenMeta(param.SelfAddr, common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.SelfAddr.String())
	}
	data, err := cabi.MintageABI.PackMethod(cabi.MethodNameMintageWithdraw, param.TokenId)
	if err != nil {
		return nil, err
	}
	povHeader, err := m.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
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
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}
	vmContext := vmstore.NewVMContext(m.ledger)
	err = m.withdraw.DoSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MintageApi) GetWithdrawRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if input == nil {
		return nil, ErrParameterNil
	}
	if ss := m.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}
	reward := &types.StateBlock{}
	vmContext := vmstore.NewVMContext(m.ledger)
	blocks, err := m.withdraw.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		povHeader, err := m.ledger.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		reward.PoVHeight = povHeader.GetHeight()
		reward.Timestamp = common.TimeNow().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate withdraw reward block")
}
