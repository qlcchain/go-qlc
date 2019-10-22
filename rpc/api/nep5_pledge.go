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

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type NEP5PledgeApi struct {
	logger   *zap.SugaredLogger
	ledger   *ledger.Ledger
	pledge   *contract.Nep5Pledge
	withdraw *contract.WithdrawNep5Pledge
}

func NewNEP5PledgeApi(ledger *ledger.Ledger) *NEP5PledgeApi {
	return &NEP5PledgeApi{ledger: ledger, logger: log.NewLogger("api_nep5_pledge"), pledge: &contract.Nep5Pledge{},
		withdraw: &contract.WithdrawNep5Pledge{}}
}

type PledgeParam struct {
	Beneficial    types.Address
	PledgeAddress types.Address
	Amount        types.Balance
	PType         string
	NEP5TxId      string
}

func (p *NEP5PledgeApi) GetPledgeData(param *PledgeParam) ([]byte, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	var t uint8
	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		t = uint8(0)
	case "vote":
		t = uint8(1)
		//TODO: support soon
	//case "storage":
	//	t=uint8(2)
	//case "oracle":
	//	t=uint8(3)
	default:
		return nil, fmt.Errorf("unsupport pledge type %s", param.PType)
	}

	return cabi.NEP5PledgeABI.PackMethod(cabi.MethodNEP5Pledge, param.Beneficial, param.PledgeAddress, t, param.NEP5TxId)
}

func (p *NEP5PledgeApi) GetPledgeBlock(param *PledgeParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	if param.PledgeAddress.IsZero() || param.Beneficial.IsZero() || len(param.PType) == 0 || len(param.NEP5TxId) == 0 {
		return nil, errors.New("invalid param")
	}

	am, err := p.ledger.GetAccountMeta(param.PledgeAddress)
	if am == nil {
		return nil, fmt.Errorf("invalid user account:%s, %s", param.PledgeAddress.String(), err)
	}

	tm := am.Token(common.ChainToken())
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

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.PledgeAddress,
		Balance:        tm.Balance.Sub(param.Amount),
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Previous:       tm.Header,
		Link:           types.Hash(types.NEP5PledgeAddress),
		Representative: tm.Representative,
		Data:           data,
		Timestamp:      common.TimeNow().Unix(),
	}

	err = p.pledge.DoSend(vmstore.NewVMContext(p.ledger), send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (p *NEP5PledgeApi) GetPledgeRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if input == nil {
		return nil, ErrParameterNil
	}
	reward := &types.StateBlock{}
	blocks, err := p.pledge.DoReceive(vmstore.NewVMContext(p.ledger), reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		reward.Timestamp = common.TimeNow().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate pledge reward block")
}

type WithdrawPledgeParam struct {
	Beneficial types.Address `json:"beneficial"`
	Amount     types.Balance `json:"amount"`
	PType      string        `json:"pType"`
	NEP5TxId   string        `json:"nep5TxId"`
}

func (p *NEP5PledgeApi) GetWithdrawPledgeData(param *WithdrawPledgeParam) ([]byte, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	var t uint8
	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		t = uint8(0)
	case "vote":
		t = uint8(1)
		//TODO: support soon
	//case "storage":
	//	t=uint8(2)
	//case "oracle":
	//	t=uint8(3)
	default:
		return nil, fmt.Errorf("unsupport pledge type %s", param.PType)
	}

	return cabi.NEP5PledgeABI.PackMethod(cabi.MethodWithdrawNEP5Pledge, param.Beneficial, param.Amount.Int, t, param.NEP5TxId)
}

func (p *NEP5PledgeApi) GetWithdrawPledgeBlock(param *WithdrawPledgeParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	if param.Beneficial.IsZero() || param.Amount.IsZero() || len(param.PType) == 0 || len(param.NEP5TxId) == 0 {
		return nil, errors.New("invalid param")
	}

	am, err := p.ledger.GetAccountMeta(param.Beneficial)
	if am == nil {
		return nil, fmt.Errorf("invalid user account:%s, %s", param.Beneficial.String(), err)
	}

	tm := am.Token(common.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.Beneficial.String())
	}

	data, err := p.GetWithdrawPledgeData(param)
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Beneficial,
		Balance:        am.CoinBalance,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Previous:       tm.Header,
		Link:           types.Hash(types.NEP5PledgeAddress),
		Representative: tm.Representative,
		Data:           data,
		Timestamp:      common.TimeNow().Unix(),
	}

	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		send.Network = send.Network.Sub(param.Amount)
	case "vote":
		send.Vote = send.Vote.Sub(param.Amount)
		//TODO: support soon
	//case "storage":
	//	send.Storage = send.Storage.Sub(param.Amount)
	//case "oracle":
	//	send.Oracle = send.Oracle.Sub(param.Amount)
	default:
		return nil, fmt.Errorf("unsupport pledge type %s", param.PType)
	}

	err = p.withdraw.DoSend(vmstore.NewVMContext(p.ledger), send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (p *NEP5PledgeApi) GetWithdrawRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if input == nil {
		return nil, ErrParameterNil
	}
	reward := &types.StateBlock{}

	blocks, err := p.withdraw.DoReceive(vmstore.NewVMContext(p.ledger), reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		reward.Timestamp = common.TimeNow().Unix()
		h := blocks[0].VMContext.Cache.Trie().Hash()
		reward.Extra = *h
		return reward, nil
	}

	return nil, errors.New("can not generate pledge withdraw reward block")
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
func (p *NEP5PledgeApi) GetPledgeInfosByPledgeAddress(addr types.Address) *PledgeInfos {
	infos, am := cabi.GetPledgeInfos(vmstore.NewVMContext(p.ledger), addr)
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
func (p *NEP5PledgeApi) GetPledgeBeneficialTotalAmount(addr types.Address) (*big.Int, error) {
	am, err := cabi.GetPledgeBeneficialTotalAmount(vmstore.NewVMContext(p.ledger), addr)
	return am, err
}

//get pledge info by beneficial,pType ,return PledgeInfos
func (p *NEP5PledgeApi) GetBeneficialPledgeInfosByAddress(beneficial types.Address) *PledgeInfos {
	infos, am := cabi.GetBeneficialInfos(vmstore.NewVMContext(p.ledger), beneficial)
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
func (p *NEP5PledgeApi) GetBeneficialPledgeInfos(beneficial types.Address, pType string) (*PledgeInfos, error) {
	var pt cabi.PledgeType
	switch strings.ToLower(pType) {
	case "network", "confidant":
		pt = cabi.Network
	case "vote":
		pt = cabi.Vote
	default:
		return nil, fmt.Errorf("unsupport type: %s", pType)
	}
	infos, am := cabi.GetBeneficialPledgeInfos(vmstore.NewVMContext(p.ledger), beneficial, pt)
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
func (p *NEP5PledgeApi) GetPledgeBeneficialAmount(beneficial types.Address, pType string) (*big.Int, error) {
	var pt uint8
	switch strings.ToLower(pType) {
	case "network", "confidant":
		pt = uint8(0)
	case "vote":
		pt = uint8(1)
	default:
		return nil, fmt.Errorf("unsupport type: %s", pType)
	}
	am := cabi.GetPledgeBeneficialAmount(vmstore.NewVMContext(p.ledger), beneficial, pt)
	return am, nil
}

//search pledge info by Beneficial,amount pType
func (p *NEP5PledgeApi) GetPledgeInfo(param *WithdrawPledgeParam) ([]*NEP5PledgeInfo, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	var pType uint8
	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		pType = uint8(0)
	case "vote":
		pType = uint8(1)
	default:
		return nil, fmt.Errorf("unsupport type: %s", param.PType)
	}
	pm := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int,
		PType:      pType,
	}
	pr := cabi.SearchBeneficialPledgeInfoIgnoreWithdrawTime(vmstore.NewVMContext(p.ledger), pm)
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
func (p *NEP5PledgeApi) GetPledgeInfoWithNEP5TxId(param *WithdrawPledgeParam) (*NEP5PledgeInfo, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	var pType uint8
	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		pType = uint8(0)
	case "vote":
		pType = uint8(1)
	default:
		return nil, fmt.Errorf("unsupport type: %s", param.PType)
	}
	pm := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int,
		PType:      pType,
		NEP5TxId:   param.NEP5TxId,
	}
	pr := cabi.SearchPledgeInfoWithNEP5TxId(vmstore.NewVMContext(p.ledger), pm)
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
func (p *NEP5PledgeApi) GetPledgeInfoWithTimeExpired(param *WithdrawPledgeParam) ([]*NEP5PledgeInfo, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	var pType uint8
	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		pType = uint8(0)
	case "vote":
		pType = uint8(1)
	default:
		return nil, fmt.Errorf("unsupport type: %s", param.PType)
	}
	pm := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int,
		PType:      pType,
	}
	pr := cabi.SearchBeneficialPledgeInfo(vmstore.NewVMContext(p.ledger), pm)
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
func (p *NEP5PledgeApi) GetAllPledgeInfo() ([]*NEP5PledgeInfo, error) {
	var result []*NEP5PledgeInfo

	infos, err := cabi.SearchAllPledgeInfos(vmstore.NewVMContext(p.ledger))
	if err != nil {
		return nil, err
	}

	for _, val := range infos {
		var t string
		switch val.PType {
		case uint8(0):
			t = "network"
		case uint8(1):
			t = "vote"
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
func (p *NEP5PledgeApi) GetTotalPledgeAmount() (*big.Int, error) {
	return cabi.GetTotalPledgeAmount(vmstore.NewVMContext(p.ledger)), nil
}
