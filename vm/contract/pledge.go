/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/vm/abi"
	"math/big"
	"strings"
)

const (
	jsonPledge = `
	[
		{"type":"function","name":"Pledge", "inputs":[{"name":"beneficial","type":"address"}]},
		{"type":"function","name":"WithdrawPledge","inputs":[{"name":"beneficial","type":"address"},{"name":"amount","type":"uint256"}]},
		{"type":"variable","name":"pledgeInfo","inputs":[{"name":"amount","type":"uint256"},{"name":"withdrawTime","type":"uint64"}]},
		{"type":"variable","name":"pledgeBeneficial","inputs":[{"name":"amount","type":"uint256"}]}
	]`

	MethodNamePledge             = "Pledge"
	MethodNameWithdrawPledge     = "WithdrawPledge"
	VariableNamePledgeInfo       = "pledgeInfo"
	VariableNamePledgeBeneficial = "pledgeBeneficial"
)

var (
	ABIPledge, _ = abi.JSONToABIContract(strings.NewReader(jsonPledge))
)

type VariablePledgeBeneficial struct {
	Amount *big.Int
}

type ParamCancelPledge struct {
	Beneficial types.Address
	Amount     *big.Int
}

type PledgeInfo struct {
	Amount         *big.Int
	WithdrawTime   uint64
	BeneficialAddr types.Address
}

type Pledge struct {
}

func (p *Pledge) GetFee(ledger *ledger.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*Pledge) DoSend(ledger *ledger.Ledger, block *types.StateBlock) error {
	panic("implement me")
}

func (*Pledge) DoReceive(ledger *ledger.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (*Pledge) GetRefundData() []byte {
	panic("implement me")
}

func (*Pledge) GetQuota() uint64 {
	return 0
}

type WithdrawPledge struct {
}

func (*WithdrawPledge) GetFee(ledger *ledger.Ledger, block *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (*WithdrawPledge) DoSend(ledger *ledger.Ledger, block *types.StateBlock) error {
	panic("implement me")
}

func (*WithdrawPledge) DoReceive(ledger *ledger.Ledger, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	panic("implement me")
}

func (*WithdrawPledge) GetRefundData() []byte {
	panic("implement me")
}

func (*WithdrawPledge) GetQuota() uint64 {
	return 0
}
