/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
type TerminateParam struct {
	ContractAddress types.Address `msg:"a,extension" json:"contractAddress"`
	Request         bool          `msg:"r" json:"request"`
}

func (z *TerminateParam) ToABI() ([]byte, error) {
	id := SettlementABI.Methods[MethodNameTerminateContract].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *TerminateParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *TerminateParam) Verify() error {
	if z.ContractAddress.IsZero() {
		return errors.New("invalid contract address")
	}
	return nil
}

func (z *TerminateParam) String() string {
	return util.ToIndentString(z)
}
