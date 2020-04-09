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

	"github.com/qlcchain/go-qlc/common/types"
)

//go:generate msgp
type SignContractParam struct {
	ContractAddress types.Address `msg:"a,extension" json:"contractAddress"`
	ConfirmDate     int64         `msg:"cd" json:"confirmDate"`
}

func (z *SignContractParam) Verify() (bool, error) {
	if z.ContractAddress.IsZero() {
		return false, fmt.Errorf("invalid contract address %s", z.ContractAddress.String())
	}
	if z.ConfirmDate == 0 {
		return false, errors.New("invalid contract confirm date")
	}
	return true, nil
}

func (z *SignContractParam) ToABI() ([]byte, error) {
	id := SettlementABI.Methods[MethodNameSignContract].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *SignContractParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}
