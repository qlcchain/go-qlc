/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"time"

	"gopkg.in/validator.v2"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
type Asset struct {
	Mcc         uint64 `msg:"mcc" json:"mcc"`
	Mnc         uint64 `msg:"mnc" json:"mnc"`
	TotalAmount uint64 `msg:"t" json:"totalAmount" validate:"min=1"`
	SLAs        []*SLA `msg:"s" json:"sla,omitempty"`
}

func (z *Asset) ToAssertID() (types.Hash, error) {
	var result []byte
	result = append(result, util.BE_Uint64ToBytes(z.Mcc)...)
	result = append(result, util.BE_Uint64ToBytes(z.Mnc)...)
	result = append(result, util.BE_Uint64ToBytes(z.TotalAmount)...)
	for _, sla := range z.SLAs {
		bts, err := sla.Serialize()
		if err != nil {
			return types.ZeroHash, err
		}
		result = append(result, bts...)
	}

	return types.HashData(result), nil
}

func (z *Asset) String() string {
	return util.ToString(z)
}

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
DeliveredRate
Latency
)
*/
type SLAType int

type Compensation struct {
	Low  float32 `msg:"l" json:"low"`
	High float32 `msg:"h" json:"high"`
	Rate float32 `msg:"r" json:"rate"`
}

type SLA struct {
	SLAType       SLAType         `msg:"t" json:"type"`
	Priority      uint            `msg:"p" json:"priority"`
	Value         float32         `msg:"v" json:"value"`
	Compensations []*Compensation `msg:"c" json:"compensations,omitempty"`
}

func (z *SLA) Serialize() ([]byte, error) {
	return z.MarshalMsg(nil)
}

func (z *SLA) Deserialize(bts []byte) error {
	_, err := z.UnmarshalMsg(bts)
	return err
}

func (z *SLA) String() string {
	return util.ToString(z)
}

func NewDeliveredRate(rate float32, c []*Compensation) *SLA {
	return &SLA{
		SLAType:       SLATypeDeliveredRate,
		Priority:      1,
		Value:         rate,
		Compensations: c,
	}
}

func NewLatency(latency time.Duration, c []*Compensation) *SLA {
	return &SLA{
		SLAType:       SLATypeLatency,
		Priority:      0,
		Value:         float32(latency),
		Compensations: c,
	}
}

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Deactivated
Activated
)
*/
type AssetStatus int

type AssetParam struct {
	Owner     Contractor  `msg:"o" json:"owner"`
	Previous  types.Hash  `msg:"pre,extension" json:"-" validate:"hash"`
	Assets    []*Asset    `msg:"a" json:"assets" validate:"nonzero"`
	SignDate  int64       `msg:"t1" json:"signDate" validate:"min=1"`
	StartDate int64       `msg:"t3" json:"startDate" validate:"min=1"`
	EndDate   int64       `msg:"t4" json:"endDate" validate:"min=1"`
	Status    AssetStatus `msg:"s" json:"status"`
}

func (z *AssetParam) ToABI() ([]byte, error) {
	id := SettlementABI.Methods[MethodNameRegisterAsset].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *AssetParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *AssetParam) ToAddress() (types.Address, error) {
	if h, err := types.HashBytes(z.Owner.Address[:], z.Previous[:],
		util.BE_Int2Bytes(z.StartDate), util.BE_Int2Bytes(z.EndDate)); err == nil {
		return types.Address(h), nil
	} else {
		return types.ZeroAddress, err
	}

}

func (z *AssetParam) Verify() error {
	if err := validator.Validate(z); err != nil {
		return err
	}

	return nil
}

func (z *AssetParam) String() string {
	return util.ToString(z)
}

func ParseAssertParam(bts []byte) (*AssetParam, error) {
	param := new(AssetParam)
	if err := param.FromABI(bts); err != nil {
		return nil, err
	} else {
		return param, nil
	}
}
