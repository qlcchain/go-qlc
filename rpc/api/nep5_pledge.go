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
	"math/big"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type NEP5PledgeAPI struct {
	logger   *zap.SugaredLogger
	l        ledger.Store
	pledge   *contract.Nep5Pledge
	withdraw *contract.WithdrawNep5Pledge
	cc       *chainctx.ChainContext
}

func NewNEP5PledgeAPI(cfgFile string, l ledger.Store) *NEP5PledgeAPI {
	api := &NEP5PledgeAPI{
		l:        l,
		pledge:   &contract.Nep5Pledge{},
		withdraw: &contract.WithdrawNep5Pledge{},
		logger:   log.NewLogger("api_nep5_pledge"),
		cc:       chainctx.NewChainContext(cfgFile),
	}
	return api
}

type PledgeParam struct {
	Beneficial    types.Address
	PledgeAddress types.Address
	Amount        types.Balance
	PType         string
	NEP5TxId      string
}

func (p *NEP5PledgeAPI) GetPledgeData(param *PledgeParam) ([]byte, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	t, err := cabi.StringToPledgeType(param.PType)
	if err != nil {
		return nil, err
	}

	p2 := &cabi.PledgeParam{
		Beneficial:    param.Beneficial,
		PledgeAddress: param.PledgeAddress,
		PType:         uint8(t),
		NEP5TxId:      param.NEP5TxId,
	}

	return p2.ToABI()
}

func (p *NEP5PledgeAPI) GetPledgeBlock(param *PledgeParam) (*types.StateBlock, error) {
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if param == nil {
		return nil, ErrParameterNil
	}
	if param.PledgeAddress.IsZero() || param.Beneficial.IsZero() || len(param.PType) == 0 || len(param.NEP5TxId) == 0 {
		return nil, errors.New("invalid param")
	}

	am, err := p.l.GetAccountMeta(param.PledgeAddress)
	if am == nil {
		return nil, fmt.Errorf("invalid user account:%s, %s", param.PledgeAddress.String(), err)
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have any chain token", param.PledgeAddress.String())
	}
	if tm.Balance.Compare(param.Amount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("%s have no enough balance, want pledge %s, but only %s", param.PledgeAddress.String(), param.Amount.String(), tm.Balance.String())
	}

	data, err := p.GetPledgeData(param)
	if err != nil {
		return nil, err
	}
	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.PledgeAddress,
		Balance:        tm.Balance.Sub(param.Amount),
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.NEP5PledgeAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	err = p.pledge.DoSend(vmstore.NewVMContext(p.l, &contractaddress.NEP5PledgeAddress), send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (p *NEP5PledgeAPI) GetPledgeRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if input == nil {
		return nil, ErrParameterNil
	}
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	reward := &types.StateBlock{
		Timestamp: common.TimeNow().Unix(),
	}
	blocks, err := p.pledge.DoReceive(vmstore.NewVMContext(p.l, &contractaddress.NEP5PledgeAddress), reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		povHeader, err := p.l.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		reward.PoVHeight = povHeader.GetHeight()
		h := vmstore.TrieHash(blocks[0].VMContext)
		if h != nil {
			reward.Extra = h
		}

		return reward, nil
	}

	return nil, errors.New("can not generate pledge reward block")
}

func (p *NEP5PledgeAPI) GetPledgeRewardBlockBySendHash(sendHash types.Hash) (*types.StateBlock, error) {
	sendBlock, err := p.l.GetStateBlock(sendHash)
	if err != nil {
		return nil, err
	}

	return p.GetPledgeRewardBlock(sendBlock)
}

type WithdrawPledgeParam struct {
	Beneficial types.Address `json:"beneficial"`
	Amount     types.Balance `json:"amount"`
	PType      string        `json:"pType"`
	NEP5TxId   string        `json:"nep5TxId"`
}

func (p *NEP5PledgeAPI) GetWithdrawPledgeData(param *WithdrawPledgeParam) ([]byte, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	t, err := cabi.StringToPledgeType(param.PType)
	if err != nil {
		return nil, err
	}

	p2 := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int,
		PType:      uint8(t),
		NEP5TxId:   param.NEP5TxId,
	}

	return p2.ToABI()
}

func (p *NEP5PledgeAPI) GetWithdrawPledgeBlock(param *WithdrawPledgeParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	if param.Beneficial.IsZero() || param.Amount.IsZero() || len(param.PType) == 0 || len(param.NEP5TxId) == 0 {
		return nil, errors.New("invalid param")
	}
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	am, err := p.l.GetAccountMeta(param.Beneficial)
	if am == nil {
		return nil, fmt.Errorf("invalid user account:%s, %s", param.Beneficial.String(), err)
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.Beneficial.String())
	}

	data, err := p.GetWithdrawPledgeData(param)
	if err != nil {
		return nil, err
	}
	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Beneficial,
		Balance:        am.CoinBalance,
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.NEP5PledgeAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		send.Network = types.ToBalance(send.GetNetwork().Sub(param.Amount))
	case "vote":
		send.Vote = types.ToBalance(send.GetVote().Sub(param.Amount))
	case "oracle":
		send.Oracle = types.ToBalance(send.GetOracle().Sub(param.Amount))
		//TODO: support soon
	//case "storage":
	//	send.Storage = send.Storage.Sub(param.Amount)
	default:
		return nil, fmt.Errorf("unsupport pledge type %s", param.PType)
	}

	err = p.withdraw.DoSend(vmstore.NewVMContext(p.l, &contractaddress.NEP5PledgeAddress), send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (p *NEP5PledgeAPI) GetWithdrawRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if input == nil {
		return nil, ErrParameterNil
	}

	reward := &types.StateBlock{
		Timestamp: common.TimeNow().Unix(),
	}

	blocks, err := p.withdraw.DoReceive(vmstore.NewVMContext(p.l, &contractaddress.NEP5PledgeAddress), reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		povHeader, err := p.l.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		reward.PoVHeight = povHeader.GetHeight()

		h := vmstore.TrieHash(blocks[0].VMContext)
		if h != nil {
			reward.Extra = h
		}
		return reward, nil
	}

	return nil, errors.New("can not generate pledge withdraw reward block")
}

func (p *NEP5PledgeAPI) GetWithdrawRewardBlockBySendHash(sendHash types.Hash) (*types.StateBlock, error) {
	sendBlock, err := p.l.GetStateBlock(sendHash)
	if err != nil {
		return nil, err
	}

	return p.GetWithdrawRewardBlock(sendBlock)
}

func (p *NEP5PledgeAPI) ParsePledgeInfo(data []byte) (*cabi.NEP5PledgeInfo, error) {
	return cabi.ParsePledgeInfo(data)
}

type NEP5PledgeInfo struct {
	PType         string
	Amount        *big.Int
	WithdrawTime  string
	Beneficial    types.Address
	PledgeAddress types.Address
	NEP5TxId      string
}

type PledgeInfos struct {
	PledgeInfo   []*NEP5PledgeInfo
	TotalAmounts *big.Int
}

//get pledge info by pledge address ,return pledgeinfos
func (p *NEP5PledgeAPI) GetPledgeInfosByPledgeAddress(addr types.Address) *PledgeInfos {
	infos, am := cabi.GetPledgeInfos(p.l, addr)
	var pledgeInfo []*NEP5PledgeInfo
	for _, v := range infos {
		npi := &NEP5PledgeInfo{
			PType:         cabi.PledgeType(v.PType).String(),
			Amount:        v.Amount,
			WithdrawTime:  time.Unix(v.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    v.Beneficial,
			PledgeAddress: v.PledgeAddress,
			NEP5TxId:      v.NEP5TxId,
		}
		pledgeInfo = append(pledgeInfo, npi)
	}
	pledgeInfos := &PledgeInfos{
		PledgeInfo:   pledgeInfo,
		TotalAmounts: am,
	}
	return pledgeInfos
}

//get pledge total amount by beneficial address ,return total amount
func (p *NEP5PledgeAPI) GetPledgeBeneficialTotalAmount(addr types.Address) (*big.Int, error) {
	am, err := cabi.GetPledgeBeneficialTotalAmount(p.l, addr)
	return am, err
}

//get pledge info by beneficial,pType ,return PledgeInfos
func (p *NEP5PledgeAPI) GetBeneficialPledgeInfosByAddress(beneficial types.Address) *PledgeInfos {
	infos, am := cabi.GetBeneficialInfos(p.l, beneficial)
	var pledgeInfo []*NEP5PledgeInfo
	for _, v := range infos {
		npi := &NEP5PledgeInfo{
			PType:         cabi.PledgeType(v.PType).String(),
			Amount:        v.Amount,
			WithdrawTime:  time.Unix(v.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    v.Beneficial,
			PledgeAddress: v.PledgeAddress,
			NEP5TxId:      v.NEP5TxId,
		}
		pledgeInfo = append(pledgeInfo, npi)
	}
	pledgeInfos := &PledgeInfos{
		PledgeInfo:   pledgeInfo,
		TotalAmounts: am,
	}
	return pledgeInfos
}

//get pledge info by beneficial,pType ,return PledgeInfos
func (p *NEP5PledgeAPI) GetBeneficialPledgeInfos(beneficial types.Address, pType string) (*PledgeInfos, error) {
	pt, err := cabi.StringToPledgeType(pType)
	if err != nil {
		return nil, err
	}

	infos, am := cabi.GetBeneficialPledgeInfos(p.l, beneficial, pt)
	var pledgeInfo []*NEP5PledgeInfo
	for _, v := range infos {
		npi := &NEP5PledgeInfo{
			PType:         cabi.PledgeType(v.PType).String(),
			Amount:        v.Amount,
			WithdrawTime:  time.Unix(v.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    v.Beneficial,
			PledgeAddress: v.PledgeAddress,
			NEP5TxId:      v.NEP5TxId,
		}
		pledgeInfo = append(pledgeInfo, npi)
	}
	pledgeInfos := &PledgeInfos{
		PledgeInfo:   pledgeInfo,
		TotalAmounts: am,
	}
	return pledgeInfos, nil
}

//search pledge info by beneficial address,pType,return total amount
func (p *NEP5PledgeAPI) GetPledgeBeneficialAmount(beneficial types.Address, pType string) (*big.Int, error) {
	pt, err := cabi.StringToPledgeType(pType)
	if err != nil {
		return nil, err
	}

	am := cabi.GetPledgeBeneficialAmount(p.l, beneficial, uint8(pt))
	return am, nil
}

//search pledge info by Beneficial,amount pType
func (p *NEP5PledgeAPI) GetPledgeInfo(param *WithdrawPledgeParam) ([]*NEP5PledgeInfo, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	pType, err := cabi.StringToPledgeType(param.PType)
	if err != nil {
		return nil, err
	}

	pm := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int,
		PType:      uint8(pType),
	}
	pr := cabi.SearchBeneficialPledgeInfoIgnoreWithdrawTime(p.l, pm)
	var pledgeInfo []*NEP5PledgeInfo
	for _, v := range pr {
		npi := &NEP5PledgeInfo{
			PType:         cabi.PledgeType(v.PledgeInfo.PType).String(),
			Amount:        v.PledgeInfo.Amount,
			WithdrawTime:  time.Unix(v.PledgeInfo.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    v.PledgeInfo.Beneficial,
			PledgeAddress: v.PledgeInfo.PledgeAddress,
			NEP5TxId:      v.PledgeInfo.NEP5TxId,
		}
		pledgeInfo = append(pledgeInfo, npi)
	}
	sort.Slice(pledgeInfo, func(i, j int) bool { return pledgeInfo[i].WithdrawTime < pledgeInfo[j].WithdrawTime })
	return pledgeInfo, nil
}

//search pledge info by nep5Txid
func (p *NEP5PledgeAPI) GetPledgeInfoWithNEP5TxId(param *WithdrawPledgeParam) (*NEP5PledgeInfo, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	pType, err := cabi.StringToPledgeType(param.PType)
	if err != nil {
		return nil, err
	}

	pm := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int,
		PType:      uint8(pType),
		NEP5TxId:   param.NEP5TxId,
	}
	pr := cabi.SearchPledgeInfoWithNEP5TxId(p.l, pm)
	if pr != nil {
		pledgeInfo := &NEP5PledgeInfo{
			PType:         cabi.PledgeType(pr.PledgeInfo.PType).String(),
			Amount:        pr.PledgeInfo.Amount,
			WithdrawTime:  time.Unix(pr.PledgeInfo.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    pr.PledgeInfo.Beneficial,
			PledgeAddress: pr.PledgeInfo.PledgeAddress,
			NEP5TxId:      pr.PledgeInfo.NEP5TxId,
		}
		return pledgeInfo, nil
	}
	return nil, errors.New("can not find PledgeInfo")
}

//search pledge info by Beneficial,amount pType
func (p *NEP5PledgeAPI) GetPledgeInfoWithTimeExpired(param *WithdrawPledgeParam) ([]*NEP5PledgeInfo, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	pType, err := cabi.StringToPledgeType(param.PType)
	if err != nil {
		return nil, err
	}

	pm := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int,
		PType:      uint8(pType),
	}
	pr := cabi.SearchBeneficialPledgeInfo(p.l, pm)
	var pledgeInfo []*NEP5PledgeInfo
	for _, v := range pr {
		npi := &NEP5PledgeInfo{
			PType:         cabi.PledgeType(v.PledgeInfo.PType).String(),
			Amount:        v.PledgeInfo.Amount,
			WithdrawTime:  time.Unix(v.PledgeInfo.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    v.PledgeInfo.Beneficial,
			PledgeAddress: v.PledgeInfo.PledgeAddress,
			NEP5TxId:      v.PledgeInfo.NEP5TxId,
		}
		pledgeInfo = append(pledgeInfo, npi)
	}
	return pledgeInfo, nil
}

//search all pledge info
func (p *NEP5PledgeAPI) GetAllPledgeInfo() ([]*NEP5PledgeInfo, error) {
	var result []*NEP5PledgeInfo

	infos, err := cabi.SearchAllPledgeInfos(p.l)
	if err != nil {
		return nil, err
	}

	for _, val := range infos {
		var t string
		switch val.PType {
		case uint8(cabi.Network):
			t = "network"
		case uint8(cabi.Vote):
			t = "vote"
		case uint8(cabi.Oracle):
			t = "oracle"
		}
		p := &NEP5PledgeInfo{
			PType:         t,
			Amount:        val.Amount,
			WithdrawTime:  time.Unix(val.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    val.Beneficial,
			PledgeAddress: val.PledgeAddress,
			NEP5TxId:      val.NEP5TxId,
		}
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].WithdrawTime < result[j].WithdrawTime })
	return result, nil
}

// GetTotalPledgeAmount get all pledge amount
func (p *NEP5PledgeAPI) GetTotalPledgeAmount() (*big.Int, error) {
	return cabi.GetTotalPledgeAmount(p.l), nil
}
