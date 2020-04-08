package abi

import "github.com/qlcchain/go-qlc/common/types"

const (
	DoDPayTypePrePaid uint8 = iota
	DoDPayTypePostPaid
	DoDPayTypeInvalid
)

const (
	DoDChargeTypeBandwidth uint8 = iota
	DoDChargeTypeUsage
	DoDChargeTypeInvalid
)

func DoDString2PayType(s string) uint8 {
	switch s {
	case "prepaid":
		return DoDPayTypePrePaid
	case "postpaid":
		return DoDPayTypePostPaid
	default:
		return DoDPayTypeInvalid
	}
}

func DoDString2ChargeType(s string) uint8 {
	switch s {
	case "bandwidth":
		return DoDChargeTypeBandwidth
	case "usage":
		return DoDChargeTypeUsage
	default:
		return DoDChargeTypeInvalid
	}
}

//go:generate msgp
type DoDAccount struct {
	AccountName     string        `msg:"-" json:"accountName"`
	AccountInfo     types.Address `msg:"i,extension" json:"accountInfo"`
	AccountType     string        `msg:"t" json:"accountType"`
	UUID            string        `msg:"uu" json:"uuid"`
	PermServiceID   string        `msg:"ps" json:"permanentServiceID"`
	TempServiceID   string        `msg:"ts" json:"temporaryServiceID"`
	StartTime       uint64        `msg:"st" json:"startTime"`
	EndTime         uint64        `msg:"et" json:"endTime"`
	Limitation      uint64        `msg:"l" json:"limitation"`
	Usage           uint64        `msg:"u" json:"usage"`
	UsageLimitation uint64        `msg:"ul" json:"usageLimitation"`
	LastCharge      uint64        `msg:"lc" json:"lastCharge"`
}

//go:generate msgp
type DoDService struct {
	ServiceID  string `msg:"-" json:"serviceID"`
	ChargeType uint8  `msg:"c" json:"chargeType"`
	PayType    uint8  `msg:"t" json:"payType"`
	Price      uint32 `msg:"p" json:"price"`
}
