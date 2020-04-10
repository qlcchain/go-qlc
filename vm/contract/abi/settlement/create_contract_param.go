/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"errors"
	"fmt"
	"math/big"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
ActiveStage1
Activated
DestroyStage1
Destroyed
Rejected
)
*/
type ContractStatus int

//go:generate msgp
type ContractService struct {
	ServiceId   string  `msg:"id" json:"serviceId" validate:"nonzero"`
	Mcc         uint64  `msg:"mcc" json:"mcc"`
	Mnc         uint64  `msg:"mnc" json:"mnc"`
	TotalAmount uint64  `msg:"t" json:"totalAmount" validate:"min=1"`
	UnitPrice   float64 `msg:"u" json:"unitPrice" validate:"nonzero"`
	Currency    string  `msg:"c" json:"currency" validate:"nonzero"`
}

func (z *ContractService) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractService) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *ContractService) Balance() (types.Balance, error) {
	return types.Balance{Int: big.NewInt(1e8)}, nil
}

//go:generate msgp
type CreateContractParam struct {
	PartyA    Contractor        `msg:"pa" json:"partyA"`
	PartyB    Contractor        `msg:"pb" json:"partyB"`
	Previous  types.Hash        `msg:"pre,extension" json:"-"`
	Services  []ContractService `msg:"s" json:"services"`
	SignDate  int64             `msg:"t1" json:"signDate"`
	StartDate int64             `msg:"t3" json:"startDate"`
	EndDate   int64             `msg:"t4" json:"endDate"`
	//SignatureA *types.Signature  `msg:"sa,extension" json:"signatureA"`
}

func (z *CreateContractParam) Verify() (bool, error) {
	if z.PartyA.Address.IsZero() || len(z.PartyA.Name) == 0 {
		return false, fmt.Errorf("invalid partyA params")
	}

	if z.PartyB.Address.IsZero() || len(z.PartyB.Name) == 0 {
		return false, fmt.Errorf("invalid partyB params")
	}

	if z.Previous.IsZero() {
		return false, errors.New("invalid previous hash")
	}

	if len(z.Services) == 0 {
		return false, errors.New("empty contract services")
	}

	for _, s := range z.Services {
		if err := validator.Validate(s); err != nil {
			return false, err
		}
	}

	if z.SignDate <= 0 {
		return false, fmt.Errorf("invalid sign date %d", z.SignDate)
	}

	if z.StartDate <= 0 {
		return false, fmt.Errorf("invalid start date %d", z.StartDate)
	}

	if z.EndDate <= 0 {
		return false, fmt.Errorf("invalid end date %d", z.EndDate)
	}

	if z.EndDate < z.StartDate {
		return false, fmt.Errorf("invalid end date, should bigger than %d, got: %d", z.StartDate, z.EndDate)
	}

	return true, nil
}

func (z *CreateContractParam) ToContractParam() *ContractParam {
	return &ContractParam{
		CreateContractParam: *z,
		ConfirmDate:         0,
		Status:              ContractStatusActiveStage1,
	}
}

func (z *CreateContractParam) ToABI() ([]byte, error) {
	id := SettlementABI.Methods[MethodNameCreateContract].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *CreateContractParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *CreateContractParam) String() string {
	return util.ToIndentString(z)
}

func ParseContractParam(v []byte) (*ContractParam, error) {
	cp := &ContractParam{}
	if err := cp.FromABI(v); err != nil {
		return nil, err
	} else {
		return cp, nil
	}
}

func (z *CreateContractParam) Address() (types.Address, error) {
	if _, err := z.Verify(); err != nil {
		return types.ZeroAddress, err
	}

	var result []byte
	if data, err := z.PartyA.ToABI(); err != nil {
		return types.ZeroAddress, err
	} else {
		result = append(result, data...)
	}
	if data, err := z.PartyB.ToABI(); err != nil {
		return types.ZeroAddress, err
	} else {
		result = append(result, data...)
	}

	result = append(result, z.Previous[:]...)
	for _, s := range z.Services {
		if data, err := s.ToABI(); err != nil {
			return types.ZeroAddress, err
		} else {
			result = append(result, data...)
		}
	}
	result = append(result, util.BE_Int2Bytes(z.SignDate)...)

	hash := types.HashData(result)

	return types.BytesToAddress(hash[:])
}

func (z *CreateContractParam) Balance() (types.Balance, error) {
	total := types.ZeroBalance
	for _, service := range z.Services {
		b, _ := service.Balance()
		total = total.Add(b)
	}
	return total, nil
}
