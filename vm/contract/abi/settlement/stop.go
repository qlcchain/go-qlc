/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"fmt"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common/types"
)

//go:generate msgp
type StopParam struct {
	ContractAddress types.Address `msg:"ca" json:"contractAddress"`
	StopName        string        `msg:"n" json:"stopName" validate:"nonzero"`
}

func (z *StopParam) ToABI(methodName string) ([]byte, error) {
	id := SettlementABI.Methods[methodName].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *StopParam) FromABI(methodName string, data []byte) error {
	if method, err := SettlementABI.MethodById(data[:4]); err == nil && method.Name == methodName {
		_, err := z.UnmarshalMsg(data[4:])
		return err
	} else {
		return fmt.Errorf("could not locate named method: %s", methodName)
	}
}

func (z *StopParam) Verify() error {
	return validator.Validate(z)
}

//go:generate msgp
type UpdateStopParam struct {
	ContractAddress types.Address `msg:"ca" json:"contractAddress"`
	StopName        string        `msg:"n" json:"stopName" validate:"nonzero"`
	New             string        `msg:"n2" json:"newName" validate:"nonzero"`
}

func (z *UpdateStopParam) ToABI(methodName string) ([]byte, error) {
	id := SettlementABI.Methods[methodName].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *UpdateStopParam) FromABI(methodName string, data []byte) error {
	if method, err := SettlementABI.MethodById(data[:4]); err == nil && method.Name == methodName {
		_, err := z.UnmarshalMsg(data[4:])
		return err
	} else {
		return fmt.Errorf("could not locate named method: %s", methodName)
	}
}

func (z *UpdateStopParam) Verify() error {
	return validator.Validate(z)
}
