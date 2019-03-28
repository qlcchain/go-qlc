/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"errors"
	"github.com/qlcchain/go-qlc/common"
	"math/big"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	minPledgeTime = time.Duration(24 * 30 * 6) // minWithdrawTime 6 months)
)

type VotePledge struct {
}

func (p *VotePledge) GetFee(l *ledger.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

// check pledge chain coin
// - address is normal user address
// - small than min pledge amount
// transfer quota to beneficial address
func (*VotePledge) DoSend(l *ledger.Ledger, block *types.StateBlock) error {
	// check pledge amount
	amount, err := l.CalculateAmount(block)
	if err != nil {
		return err
	}

	// check send account is user account
	b, err := l.HasAccountMeta(block.Address)
	if err != nil {
		return err
	}

	if block.Token != common.ChainToken() || amount.IsZero() || !b {
		return errors.New("invalid block data")
	}

	beneficialAddr := new(types.Address)
	if err := cabi.ABIPledge.UnpackMethod(beneficialAddr, cabi.MethodVotePledge, block.Data); err != nil {
		return errors.New("invalid beneficial address")
	}
	block.Data, err = cabi.ABIPledge.PackMethod(cabi.MethodVotePledge, *beneficialAddr)
	if err != nil {
		return err
	}
	return nil
}

func (*VotePledge) DoReceive(ledger *ledger.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	beneficialAddr := new(types.Address)
	_ = cabi.ABIPledge.UnpackMethod(beneficialAddr, cabi.MethodVotePledge, input.Data)
	beneficialKey := cabi.GetPledgeBeneficialKey(*beneficialAddr)
	pledgeKey := cabi.GetPledgeKey(input.Address, beneficialKey, input.Timestamp)
	address := block.Address
	oldPledgeData, err := ledger.GetStorage(address[:], pledgeKey)
	if err != nil {
		return nil, err
	}
	amount := big.NewInt(0)
	if len(oldPledgeData) > 0 {
		oldPledge := new(cabi.VotePledgeInfo)
		_ = cabi.ABIPledge.UnpackVariable(oldPledge, cabi.VariableVotePledgeInfo, oldPledgeData)
		amount = oldPledge.Amount
	}
	a, _ := ledger.CalculateAmount(input)
	amount.Add(amount, a.Int)

	pledgeTime := time.Now().UTC().Add(time.Hour * minPledgeTime).Unix()
	pledgeInfo, _ := cabi.ABIPledge.PackVariable(cabi.VariableVotePledgeInfo, amount, pledgeTime)
	if err := ledger.SetStorage(nil, pledgeKey, pledgeInfo); err != nil {
		return nil, err
	}

	address = block.Address
	oldBeneficialData, err := ledger.GetStorage(address[:], beneficialKey)
	if err != nil {
		return nil, err
	}
	beneficialAmount := big.NewInt(0)
	if len(oldBeneficialData) > 0 {
		oldBeneficial := new(cabi.VariablePledgeBeneficial)
		_ = cabi.ABIPledge.UnpackVariable(oldBeneficial, cabi.VariableVotePledgeBeneficial, oldBeneficialData)
		beneficialAmount = oldBeneficial.Amount
	}

	beneficialAmount.Add(beneficialAmount, a.Int)
	beneficialData, _ := cabi.ABIPledge.PackVariable(cabi.VariableVotePledgeBeneficial, beneficialAmount)
	if err = ledger.SetStorage(nil, beneficialKey, beneficialData); err != nil {
		return nil, err
	}
	return nil, nil
}

func (*VotePledge) GetRefundData() []byte {
	return []byte{1}
}

type WithdrawVotePledge struct {
}

func (*WithdrawVotePledge) GetFee(ledger *ledger.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*WithdrawVotePledge) DoSend(ledger *ledger.Ledger, block *types.StateBlock) (err error) {
	if amount, err := ledger.CalculateAmount(block); err != nil {
		return err
	} else {
		if block.Type != types.Send || amount.Compare(types.ZeroBalance) != types.BalanceCompEqual {
			return errors.New("invalid block ")
		}
	}

	param := new(cabi.WithdrawPledgeParam)
	if err := cabi.ABIPledge.UnpackMethod(param, cabi.MethodWithdrawVotePledge, block.Data); err != nil {
		return errors.New("invalid input data")
	}

	if block.Data, err = cabi.ABIPledge.PackMethod(cabi.MethodWithdrawVotePledge, param.Beneficial, param.Amount); err != nil {
		return
	}

	return nil
}

func (*WithdrawVotePledge) DoReceive(ledger *ledger.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return nil, nil
}

func (*WithdrawVotePledge) GetRefundData() []byte {
	return []byte{2}
}
