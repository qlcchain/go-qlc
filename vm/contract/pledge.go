/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"
	"math/big"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	minPledgeTime = time.Duration(24 * 30 * 3) // minWithdrawTime 3 months)
)

type Pledge struct {
}

func (p *Pledge) GetFee(ledger *ledger.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*Pledge) DoSend(ledger *ledger.Ledger, block *types.StateBlock) error {
	// check pledge chain coin
	// - address is normal user address
	// - big than min pledge amount
	// transfer quota to beneficial address
	return nil
}

func (*Pledge) DoReceive(ledger *ledger.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	beneficialAddr := new(types.Address)
	_ = cabi.ABIPledge.UnpackMethod(beneficialAddr, cabi.MethodNamePledge, input.Data)
	beneficialKey := cabi.GetPledgeBeneficialKey(*beneficialAddr)
	pledgeKey := cabi.GetPledgeKey(input.Address, beneficialKey)
	oldPledgeData, err := ledger.GetStorage(&block.Address, pledgeKey)
	if err != nil {
		return nil, err
	}
	amount := big.NewInt(0)
	if len(oldPledgeData) > 0 {
		oldPledge := new(cabi.PledgeInfo)
		_ = cabi.ABIPledge.UnpackVariable(oldPledge, cabi.VariableNamePledgeInfo, oldPledgeData)
		amount = oldPledge.Amount
	}
	a, _ := ledger.CalculateAmount(input)
	amount.Add(amount, a.Int)

	pledgeTime := time.Now().UTC().Add(time.Hour * minPledgeTime).Unix()
	pledgeInfo, _ := cabi.ABIPledge.PackVariable(cabi.VariableNamePledgeInfo, amount, pledgeTime)
	_ = ledger.SetStorage(pledgeKey, pledgeInfo)

	oldBeneficialData, err := ledger.GetStorage(&block.Address, beneficialKey)
	beneficialAmount := big.NewInt(0)
	if len(oldBeneficialData) > 0 {
		oldBeneficial := new(cabi.VariablePledgeBeneficial)
		_ = cabi.ABIPledge.UnpackVariable(oldBeneficial, cabi.VariableNamePledgeBeneficial, oldBeneficialData)
		beneficialAmount = oldBeneficial.Amount
	}

	beneficialAmount.Add(beneficialAmount, a.Int)
	beneficialData, _ := cabi.ABIPledge.PackVariable(cabi.VariableNamePledgeBeneficial, beneficialAmount)
	err = ledger.SetStorage(beneficialKey, beneficialData)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (*Pledge) GetRefundData() []byte {
	return []byte{1}
}

func (*Pledge) GetQuota() uint64 {
	return 0
}

type WithdrawPledge struct {
}

func (*WithdrawPledge) GetFee(ledger *ledger.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*WithdrawPledge) DoSend(ledger *ledger.Ledger, block *types.StateBlock) (err error) {
	if amount, err := ledger.CalculateAmount(block); block.Type != types.Send || amount.Compare(types.ZeroBalance) != types.BalanceCompEqual || err != nil {
		return errors.New("invalid block ")
	}
	param := new(cabi.ParamCancelPledge)
	if err := cabi.ABIPledge.UnpackMethod(param, cabi.MethodNameWithdrawPledge, block.Data); err != nil {
		return errors.New("invalid input data")
	}

	if block.Data, err = cabi.ABIPledge.PackMethod(cabi.MethodNameWithdrawPledge, param.Beneficial, param.Amount); err != nil {
		return
	}

	return nil
}

func (*WithdrawPledge) DoReceive(ledger *ledger.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (*WithdrawPledge) GetRefundData() []byte {
	return []byte{2}
}

func (*WithdrawPledge) GetQuota() uint64 {
	return 0
}
