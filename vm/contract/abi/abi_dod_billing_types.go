package abi

import "github.com/qlcchain/go-qlc/common/types"

const (
	DoDPaidRulePrePaid uint8 = iota
	DoDPaidRulePostPaid
	DoDPaidRuleInvalid
)

const (
	DoDChargeTypeBandwidth uint8 = iota
	DoDChargeTypeUsage
	DoDChargeTypeInvalid
)

const (
	DoDBuyModeCombo uint8 = iota
	DoDBuyModeCommon
	DoDBuyModeInvalid
)

const (
	DoDDataTypeAccount uint8 = iota
	DoDDataTypeConnection
	DoDDataTypeUsage
)

func DoDString2PaidRule(s string) uint8 {
	switch s {
	case "prepaid":
		return DoDPaidRulePrePaid
	case "postpaid":
		return DoDPaidRulePostPaid
	default:
		return DoDPaidRuleInvalid
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

func DoDString2BuyMode(s string) uint8 {
	switch s {
	case "combo":
		return DoDBuyModeCombo
	case "common":
		return DoDBuyModeCommon
	default:
		return DoDBuyModeInvalid
	}
}

//go:generate msgp
type DoDAccount struct {
	Account     types.Address `msg:"-" json:"account"`
	AccountName string        `msg:"-" json:"accountName"`
	AccountInfo string        `msg:"ai" json:"accountInfo"`
	AccountType string        `msg:"at" json:"accountType"`
	UUID        string        `msg:"uu" json:"uuid"`
	Connections []string      `msg:"cs" json:"connections"`
}

type DoDConnection struct {
	AccountName     string  `msg:"-" json:"accountName"`
	ConnectionID    string  `msg:"-" json:"connectionID"`
	ServiceType     string  `msg:"st" json:"serviceType"`
	ChargeType      uint8   `msg:"ct" json:"chargeType"`
	PaidRule        uint8   `msg:"pr" json:"paidRule"`
	BuyMode         uint8   `msg:"bm" json:"buyMode"`
	Location        string  `msg:"lo" json:"location"`
	StartTime       uint64  `msg:"sti" json:"startTime"`
	EndTime         uint64  `msg:"et" json:"endTime"`
	Price           float64 `msg:"p" json:"price"`
	Unit            string  `msg:"u" json:"unit"`
	Currency        string  `msg:"c" json:"currency"`
	Balance         float64 `msg:"b" json:"balance"`
	TempStartTime   uint64  `msg:"tst" json:"tempStartTime"`
	TempEndTime     uint64  `msg:"tet" json:"tempEndTime"`
	TempPrice       float64 `msg:"tp" json:"tempPrice"`
	TempBandwidth   string  `msg:"tb" json:"tempBandwidth"`
	Bandwidth       string  `msg:"bw" json:"bandwidth"`
	Quota           float64 `msg:"q" json:"quota"`
	UsageLimitation float64 `msg:"ul" json:"usageLimitation"`
	MinBandwidth    string  `msg:"mbw" json:"minBandwidth"`
	ExpireTime      uint64  `msg:"eti" json:"expireTime"`
}

type DoDUsage struct {
	Account      types.Address `msg:"-" json:"account"`
	ConnectionID string        `msg:"-" json:"connectionID"`
	ChargeType   uint8         `msg:"c" json:"chargeType"`
	SourceType   string        `msg:"st" json:"sourceType"`
	SourceDevice string        `msg:"sd" json:"sourceDevice"`
	IP           string        `msg:"i" json:"ip"`
	Usage        float64       `msg:"u" json:"usage"`
	Timestamp    uint64        `msg:"t" json:"timestamp"`
}

type DoDInvoice struct {
	ConnectionID string  `json:"connectionID"`
	StartTime    uint64  `json:"startTime"`
	EndTime      uint64  `json:"endTime"`
	Usage        float64 `json:"usage"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	TotalPrice   float64 `json:"totalPrice"`
}
