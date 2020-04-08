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
type Contractor struct {
	Address types.Address `msg:"a,extension" json:"address" validate:"qlcaddress"`
	Name    string        `msg:"n" json:"name" validate:"nonzero"`
}

func (z *Contractor) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *Contractor) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *Contractor) String() string {
	return util.ToString(z)
}
