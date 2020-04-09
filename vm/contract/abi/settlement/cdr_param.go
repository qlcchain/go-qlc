/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Sent
Error
Empty
)
*/
type SendingStatus int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Delivered
Rejected
Unknown
Undelivered
Empty
)
*/
type DLRStatus int

//go:generate msgp
type CDRParam struct {
	Index         uint64        `msg:"i" json:"index" validate:"min=1"`
	SmsDt         int64         `msg:"dt" json:"smsDt" validate:"min=1"`
	Account       string        `msg:"ac" json:"account"`
	Sender        string        `msg:"tx" json:"sender" validate:"nonzero"`
	Customer      string        `msg:"c" json:"customer"`
	Destination   string        `msg:"d" json:"destination" validate:"nonzero"`
	SendingStatus SendingStatus `msg:"s" json:"sendingStatus"`
	DlrStatus     DLRStatus     `msg:"ds" json:"dlrStatus"`
	PreStop       string        `msg:"ps" json:"preStop"`
	NextStop      string        `msg:"ns" json:"nextStop"`
}

func (z *CDRParam) String() string {
	return util.ToIndentString(z)
}

func (z *CDRParam) Status() bool {
	switch z.DlrStatus {
	case DLRStatusDelivered:
		return true
	case DLRStatusUndelivered:
		return false
	case DLRStatusUnknown:
		fallthrough
	case DLRStatusEmpty:
		switch z.SendingStatus {
		case SendingStatusSent:
			return true
		default:
			return false
		}
	}
	return false
}

func (z *CDRParam) Verify() error {
	if errs := validator.Validate(z); errs != nil {
		return errs
	}
	return nil
}

func (z *CDRParam) ToHash() (types.Hash, error) {
	return types.HashBytes(util.BE_Uint64ToBytes(z.Index), []byte(z.Sender), []byte(z.Destination))
}

func (z *CDRParam) GetCustomer() string {
	if z.Customer == "" {
		return z.Sender
	}
	return z.Customer
}

//go:generate msgp
type CDRParamList struct {
	ContractAddress types.Address `msg:"a" json:"contractAddress" validate:"qlcaddress"`
	Params          []*CDRParam   `msg:"p" json:"params" validate:"min=1"`
}

func (z *CDRParamList) ToABI() ([]byte, error) {
	id := SettlementABI.Methods[MethodNameProcessCDR].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *CDRParamList) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *CDRParamList) Verify() error {
	if errs := validator.Validate(z); errs != nil {
		return errs
	}
	return nil
}
