package types

const (
	OracleTypeEmail uint32 = iota
	OracleTypeWeChat
	OracleTypeInvalid
)

const RandomCodeLen = 16

const (
	OracleExpirePovHeight  = 10
	OracleVerifyMinAccount = 3
	OracleVerifyMaxAccount = 5
)

var (
	MinVerifierPledgeAmount = NewBalance(3e+14) // 3M
	OracleCost              = NewBalance(1e+7)  // 0.1
	PublishCost             = NewBalance(5e+8)  // 5
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
