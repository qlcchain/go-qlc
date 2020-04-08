/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
type ContractAddressList struct {
	AddressList []*types.Address `msg:"a" json:"addressList"`
}

func newContractAddressList(address *types.Address) *ContractAddressList {
	return &ContractAddressList{AddressList: []*types.Address{address}}
}

func (z *ContractAddressList) Append(address *types.Address) bool {
	ret := false
	for _, a := range z.AddressList {
		if a == address {
			ret = true
			break
		}
	}

	if !ret {
		z.AddressList = append(z.AddressList, address)
		return true
	}
	return false
}

func (z *ContractAddressList) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractAddressList) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *ContractAddressList) String() string {
	return util.ToIndentString(z.AddressList)
}
