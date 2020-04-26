package contract

import (
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

var PdkvContract = NewChainContract(
	map[string]Contract{
		abi.MethodNamePrivacyDemoKVSet: &PrivacyDemoKVSet{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
				},
			},
		},
	},
	abi.PrivacyDemoKVABI,
	abi.JsonPrivacyDemoKV,
)
