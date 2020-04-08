package abi

import (
	"github.com/qlcchain/go-qlc/vm/abi"
	"strings"
)

const (
	JsonDoDBilling = `[
		{"type":"function","name":"DoDCreateAccount","inputs":[
			{"name":"accountName","type":"string"},
			{"name":"accountInfo","type":"address"},
			{"name":"accountType","type":"string"}
		]},
		{"type":"function","name":"DoDCoupleAccount","inputs":[
			{"name":"uuid","type":"string"},
			{"name":"accountName","type":"string"}
		]},
		{"type":"function","name":"DoDServiceSet","inputs":[
			{"name":"serviceID","type":"string"},
			{"name":"chargeType","type":"string"},
			{"name":"payType","type":"string"},
			{"name":"price","type":"uint32"}
		]},
		{"type":"function","name":"DoDUserServiceSet","inputs":[
			{"name":"accountName","type":"string"},
			{"name":"serviceID","type":"string"},
			{"name":"serviceType","type":"string"},
			{"name":"discount","type":"uint8"},
			{"name":"startTime","type":"string"},
			{"name":"endTime","type":"string"},
			{"name":"usageLimitation","type":"uint64"},
			{"name":"bandwidthLimitation","type":"uint64"}
		]},
		{"type":"function","name":"DoDUsageUpdate","inputs":[
			{"name":"uuid","type":"string"},
			{"name":"usage","type":"uint64"}
		]}
	]`

	MethodNameDoDCreateAccount  = "DoDCreateAccount"
	MethodNameDoDCoupleAccount  = "DoDCoupleAccount"
	MethodNameDoDServiceSet     = "DoDServiceSet"
	MethodNameDoDUserServiceSet = "DoDUserServiceSet"
	MethodNameDoDUsageUpdate    = "DoDUsageUpdate"
)

var (
	DoDBillingABI, _ = abi.JSONToABIContract(strings.NewReader(JsonDoDBilling))
)
