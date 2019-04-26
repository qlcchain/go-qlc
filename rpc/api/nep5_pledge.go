/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type NEP5PledgeApi struct {
	logger    *zap.SugaredLogger
	ledger    *ledger.Ledger
	vmContext *vmstore.VMContext
	pledge    *contract.Nep5Pledge
	withdraw  *contract.WithdrawNep5Pledge
}

func NewNEP5PledgeApi(ledger *ledger.Ledger) *NEP5PledgeApi {
	return &NEP5PledgeApi{ledger: ledger, vmContext: vmstore.NewVMContext(ledger),
		logger: log.NewLogger("api_nep5_pledge"), pledge: &contract.Nep5Pledge{},
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
	if param.PledgeAddress.IsZero() || param.Beneficial.IsZero() || len(param.PType) == 0 || len(param.NEP5TxId) == 0 {
		return nil, errors.New("invalid param")
	}

	am, err := p.ledger.GetAccountMeta(param.PledgeAddress)
	if am == nil {
		return nil, fmt.Errorf("invalid user account:%s, %s", param.PledgeAddress.String(), err)
	}

	tm := am.Token(common.ChainToken())
	if tm == nil || tm.Balance.IsZero() {
		return nil, fmt.Errorf("%s do not hava any chain token", param.PledgeAddress.String())
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
		Timestamp:      common.TimeNow().UTC().Unix(),
	}

	err = p.pledge.DoSend(p.vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (p *NEP5PledgeApi) GetPledgeRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	reward := &types.StateBlock{}

	blocks, err := p.pledge.DoReceive(p.vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		reward.Timestamp = common.TimeNow().UTC().Unix()
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
}

func (p *NEP5PledgeApi) GetWithdrawPledgeData(param *WithdrawPledgeParam) ([]byte, error) {
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

	return cabi.NEP5PledgeABI.PackMethod(cabi.MethodWithdrawNEP5Pledge, param.Beneficial, param.Amount.Int, t)
}

func (p *NEP5PledgeApi) GetWithdrawPledgeBlock(param *WithdrawPledgeParam) (*types.StateBlock, error) {
	if param.Beneficial.IsZero() || param.Amount.IsZero() || len(param.PType) == 0 {
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
		Timestamp:      common.TimeNow().UTC().Unix(),
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

	err = p.withdraw.DoSend(p.vmContext, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (p *NEP5PledgeApi) GetWithdrawRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	reward := &types.StateBlock{}

	blocks, err := p.withdraw.DoReceive(p.vmContext, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		reward.Timestamp = common.TimeNow().UTC().Unix()
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

func (p *NEP5PledgeApi) SearchAllPledgeInfo() ([]*NEP5PledgeInfo, error) {
	var result []*NEP5PledgeInfo

	infos, err := cabi.SearchAllPledgeInfos(p.vmContext)
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
	return result, nil
}

func (p *NEP5PledgeApi) SearchPledgeInfo(param *WithdrawPledgeParam) ([]*NEP5PledgeInfo, error) {
	var result []*NEP5PledgeInfo
	var t uint8
	switch strings.ToLower(param.PType) {
	case "network", "confidant":
		t = uint8(0)
	case "vote":
		t = uint8(1)
	default:
		return nil, fmt.Errorf("unsupport type: %s", param.PType)
	}

	pledgeParam := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     param.Amount.Int, PType: t,
	}
	info := cabi.SearchBeneficialPledgeInfo(p.vmContext, pledgeParam)

	for _, val := range info {
		pledgeInfo := val.PledgeInfo
		p := &NEP5PledgeInfo{
			PType:         param.PType,
			Amount:        pledgeInfo.Amount,
			WithdrawTime:  time.Unix(pledgeInfo.WithdrawTime, 0).Format(time.RFC3339),
			Beneficial:    pledgeInfo.Beneficial,
			PledgeAddress: pledgeInfo.PledgeAddress,
			NEP5TxId:      pledgeInfo.NEP5TxId,
		}
		result = append(result, p)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].WithdrawTime < result[j].WithdrawTime })
	return result, nil
}
