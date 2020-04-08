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

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
type Terminator struct {
	Address types.Address `msg:"a,extension" json:"address"` // terminator qlc address
	Request bool          `msg:"r" json:"request"`           // request operate, true or false
}

func (z *Terminator) String() string {
	return util.ToString(z)
}

//go:generate msgp
type ContractParam struct {
	CreateContractParam
	PreStops    []string       `msg:"pre" json:"preStops,omitempty"`
	NextStops   []string       `msg:"nex" json:"nextStops,omitempty"`
	ConfirmDate int64          `msg:"t2" json:"confirmDate"`
	Status      ContractStatus `msg:"s" json:"status"`
	Terminator  *Terminator    `msg:"t" json:"terminator,omitempty"`
}

func (z *ContractParam) IsPreStop(n string) bool {
	if len(z.PreStops) == 0 {
		return false
	}
	for _, stop := range z.PreStops {
		if stop == n {
			return true
		}
	}
	return false
}

func (z *ContractParam) IsNextStop(n string) bool {
	if len(z.NextStops) == 0 {
		return false
	}
	for _, stop := range z.NextStops {
		if stop == n {
			return true
		}
	}
	return false
}

func (z *ContractParam) IsAvailable() bool {
	unix := common.TimeNow().Unix()
	return z.Status == ContractStatusActivated && unix >= z.StartDate && unix <= z.EndDate
}

func (z *ContractParam) IsExpired() bool {
	unix := common.TimeNow().Unix()
	return z.Status == ContractStatusActivated && unix > z.EndDate
}

func (z *ContractParam) IsContractor(addr types.Address) bool {
	return z.PartyA.Address == addr || z.PartyB.Address == addr
}

func (z *ContractParam) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *ContractParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *ContractParam) Equal(cp *CreateContractParam) (bool, error) {
	if cp == nil {
		return false, errors.New("invalid input value")
	}

	a1, err := z.Address()
	if err != nil {
		return false, err
	}

	a2, err := cp.Address()
	if err != nil {
		return false, err
	}

	if a1 == a2 {
		return true, nil
	}
	return false, fmt.Errorf("invalid address, exp: %s,act: %s", a1.String(), a2.String())
}

func (z *ContractParam) DoActive(operator types.Address) error {
	if z.PartyB.Address != operator {
		return fmt.Errorf("invalid partyB, exp: %s, act: %s", z.PartyB.Address.String(), operator.String())
	}

	if z.Status == ContractStatusActiveStage1 {
		z.Status = ContractStatusActivated
		return nil
	} else if z.Status == ContractStatusDestroyed {
		return errors.New("contract has been destroyed")
	} else {
		return fmt.Errorf("invalid contract status, %s", z.Status.String())
	}
}

func (z *ContractParam) DoTerminate(operator *Terminator) error {
	if operator == nil || operator.Address.IsZero() {
		return errors.New("invalid terminal operator")
	}

	if b := z.IsContractor(operator.Address); !b {
		return fmt.Errorf("permission denied, only contractor can terminate it, exp: %s or %s, act: %s",
			z.PartyA.Address.String(), z.PartyB.Address.String(), operator.Address.String())
	}

	if z.Terminator != nil {
		if operator.Address == z.Terminator.Address && operator.Request == z.Terminator.Request {
			return fmt.Errorf("%s already terminated contract", operator.String())
		}

		// confirmed, only allow deal with ContractStatusDestroyStage1
		// - terminator cancel by himself, request is false
		// - confirm by the other one, request is true
		// - reject by the other one, request is false
		if z.Status == ContractStatusDestroyStage1 {
			// cancel himself
			if operator.Address == z.Terminator.Address {
				if !operator.Request {
					z.Status = ContractStatusActivated
					z.Terminator = nil
				}
			} else {
				// confirm by the other one
				if operator.Request {
					z.Status = ContractStatusDestroyed
				} else {
					// reject, back to activated
					z.Status = ContractStatusActivated
					z.Terminator = nil
				}
			}
		} else {
			return fmt.Errorf("invalid contract status, %s", z.Status.String())
		}
	} else {
		if !operator.Request {
			return fmt.Errorf("invalid request(%s) on %s", operator.String(), z.Status.String())
		}
		if z.Status == ContractStatusActiveStage1 {
			// request only allow true
			// first operate, request should always true, allow
			// - partyA close contract by himself
			// - partyB reject partyA's contract
			// - partyA or partyB start to close a signed contract
			if z.PartyA.Address == operator.Address {
				z.Status = ContractStatusDestroyed
			} else {
				z.Status = ContractStatusRejected
			}
		} else if z.Status == ContractStatusActivated {
			z.Terminator = operator
			z.Status = ContractStatusDestroyStage1
		} else {
			return fmt.Errorf("invalid contract status, %s", z.Status.String())
		}
	}

	return nil
}

func (z *ContractParam) String() string {
	return util.ToIndentString(z)
}
