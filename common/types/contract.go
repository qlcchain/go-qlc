package types

import "math/big"

const (
	OracleTypeEmail uint32 = iota
	OracleTypeWeChat
	OracleTypeInvalid
)

const RandomCodeLen = 16

var (
	MinVerifierPledgeAmount = big.NewInt(3e+14) // 3M
	OracleCost              = big.NewInt(1e+7)  // 0.1
	PublishCost             = big.NewInt(5e+8)  // 5
)

func OracleStringToType(os string) uint32 {
	switch os {
	case "email":
		return OracleTypeEmail
	case "weChat":
		return OracleTypeWeChat
	default:
		return OracleTypeInvalid
	}
}

func OracleTypeToString(ot uint32) string {
	switch ot {
	case OracleTypeEmail:
		return "email"
	case OracleTypeWeChat:
		return "weChat"
	default:
		return "invalid"
	}
}
