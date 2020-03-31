/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import "fmt"

//go:generate msgp
type ContractValue struct {
	BlockHash Hash  `msg:"p,extension" json:"previous"`
	Root      *Hash `msg:"r,extension" json:"trieRoot,omitempty"`
}

func (z ContractValue) Serialize() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractValue) Deserialize(text []byte) error {
	_, err := z.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (z *ContractValue) String() string {
	return fmt.Sprintf("h:%s, r:%v", z.BlockHash, z.Root)
}
