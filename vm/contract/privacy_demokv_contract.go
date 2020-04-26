package contract

import (
	"github.com/qlcchain/go-qlc/common/vmcontract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

var PdkvContract = vmcontract.NewChainContract(
	map[string]vmcontract.Contract{
		abi.MethodNamePrivacyDemoKVSet: &PrivacyDemoKVSet{
			BaseContract: BaseContract{
				Describe: vmcontract.Describe{
					SpecVer:   vmcontract.SpecVer2,
					Signature: true,
					Work:      true,
				},
			},
		},
	},
	abi.PrivacyDemoKVABI,
	abi.JsonPrivacyDemoKV,
)
