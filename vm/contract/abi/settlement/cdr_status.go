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
	"github.com/qlcchain/go-qlc/common/util"
)

var (
	customerFn = func(status *CDRStatus) (string, error) {
		_, customer, _, err := status.ExtractCustomer()
		return customer, err
	}

	accountFn = func(status *CDRStatus) (string, error) {
		_, account, _, err := status.ExtractAccount()
		return account, err
	}
)

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
unknown
stage1
success
failure
missing
duplicate
)
*/
type SettlementStatus int

//go:generate msgp
type SettlementCDR struct {
	CDRParam
	From types.Address `msg:"f,extension" json:"from"`
}

//go:generate msgp
type CDRStatus struct {
	Params map[string][]CDRParam `msg:"p" json:"params"`
	Status SettlementStatus      `msg:"s" json:"status"`
}

func (z *CDRStatus) ToABI() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *CDRStatus) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data)
	return err
}

func (z *CDRStatus) ToHash() (types.Hash, error) {
	if len(z.Params) > 0 {
		for _, v := range z.Params {
			if len(v) > 0 {
				return v[0].ToHash()
			}
		}
	}
	return types.ZeroHash, errors.New("no cdr record")
}

func (z *CDRStatus) State(addr *types.Address, fn func(status *CDRStatus) (string, error)) (sender string, isMatching, state bool, err error) {
	sender, err = fn(z)
	if err != nil {
		return "", false, false, err
	}
	isMatching = len(z.Params) == 2
	if params, ok := z.Params[addr.String()]; ok {
		switch size := len(params); {
		case size == 1:
			return sender, isMatching, params[0].Status(), nil
		case size > 1: //upload multi-times or normalize time error
			return sender, isMatching, false, nil
		default:
			return "", isMatching, false, nil
		}
	} else {
		return "", isMatching, false, fmt.Errorf("can not find data of %s", addr.String())
	}
}

func (z *CDRStatus) IsInCycle(start, end int64) bool {
	if len(z.Params) == 0 {
		return false
	}

	if start != 0 && end != 0 {
		i := 0
		for _, params := range z.Params {
			if len(params) > 0 {
				param := params[0]
				if param.SmsDt >= start && param.SmsDt <= end {
					return true
				}
			}
			i++
		}
		if i == len(z.Params) {
			return false
		}
	}
	return true
}

// ExtractID fetch SMS sender/destination/datetime, if `Customer` is not empty, use it instead of sender
func (z *CDRStatus) ExtractID() (dt int64, sender, destination string, err error) {
	if len(z.Params) > 0 {
		for _, params := range z.Params {
			if len(params) >= 1 {
				dt = params[0].SmsDt
				destination = params[0].Destination
				sender = params[0].GetCustomer()
				for _, param := range params {
					if param.Customer != "" {
						sender = param.Customer
						return
					}
				}
			}
		}
		return
	}

	return 0, "", "", errors.New("can not find any CDR param")
}

func (z *CDRStatus) String() string {
	return util.ToIndentString(z)
}

func (z *CDRStatus) ExtractCustomer() (dt int64, customer, destination string, err error) {
	if len(z.Params) > 0 {
		for _, params := range z.Params {
			if len(params) >= 1 {
				dt = params[0].SmsDt
				destination = params[0].Destination
				for _, param := range params {
					if param.Customer != "" {
						customer = param.Customer
						return
					}
				}
			}
		}
		return
	}

	return 0, "", "", errors.New("can not find any CDR param")
}

func (z *CDRStatus) ExtractAccount() (dt int64, account, destination string, err error) {
	if len(z.Params) > 0 {
		for _, params := range z.Params {
			if len(params) >= 1 {
				dt = params[0].SmsDt
				destination = params[0].Destination
				for _, param := range params {
					if param.Account != "" {
						account = param.Account
						return
					}
				}
			}
		}
		return
	}

	return 0, "", "", errors.New("can not find any CDR param")
}

// DoSettlement process settlement
// @param cdr  cdr data
func (z *CDRStatus) DoSettlement(cdr SettlementCDR) (err error) {
	if z.Params == nil {
		z.Params = make(map[string][]CDRParam, 0)
	}

	from := cdr.From.String()
	if params, ok := z.Params[from]; ok {
		params = append(params, cdr.CDRParam)
		z.Params[from] = params
	} else {
		z.Params[from] = []CDRParam{cdr.CDRParam}
	}

	switch size := len(z.Params); {
	//case size == 0:
	//	z.Status = SettlementStatusUnknown
	//	break
	case size == 1:
		z.Status = SettlementStatusStage1
		break
	case size == 2:
		z.Status = SettlementStatusSuccess
		b := true
		// combine all status
		for _, params := range z.Params {
			//for _, param := range params {
			//	b = b && param.Status()
			//}
			switch l := len(params); {
			case l > 1:
				z.Status = SettlementStatusDuplicate
				return
			case l == 1:
				b = b && params[0].Status()
				break
			}
		}
		if !b {
			z.Status = SettlementStatusFailure
		}
	case size > 2:
		err = fmt.Errorf("invalid params size %d", size)
	}
	return err
}

func ParseCDRStatus(v []byte) (*CDRStatus, error) {
	state := &CDRStatus{}
	if err := state.FromABI(v); err != nil {
		return nil, err
	} else {
		return state, nil
	}
}
