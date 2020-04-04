package chaincontract

import (
	"sync"

	"github.com/qlcchain/go-qlc/common/vmcontract"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var lock = sync.RWMutex{}

func InitChainContract() {
	lock.Lock()
	defer lock.Unlock()
	mintage := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNameMintage: &contract.Mintage{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer1,
						Signature: true,
						Work:      true,
					},
				},
			},
			cabi.MethodNameMintageWithdraw: &contract.WithdrawMintage{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer1,
						Signature: true,
						Work:      true,
					},
				},
			},
		},
		cabi.MintageABI,
		cabi.JsonMintage,
	)
	vmcontract.RegisterContracts(contractaddress.MintageAddress, mintage)

	nep5 := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNEP5Pledge: &contract.Nep5Pledge{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer1,
						Signature: true,
						Work:      true,
					},
				},
			},
			cabi.MethodWithdrawNEP5Pledge: &contract.WithdrawNep5Pledge{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer1,
						Signature: true,
						Work:      true,
					},
				},
			},
		},
		cabi.NEP5PledgeABI,
		cabi.JsonNEP5Pledge,
	)
	vmcontract.RegisterContracts(contractaddress.NEP5PledgeAddress, nep5)

	rewards := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNameAirdropRewards: &contract.AirdropRewards{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer: vmcontract.SpecVer1,
						Pending: true,
					},
				},
			},
			cabi.MethodNameConfidantRewards: &contract.ConfidantRewards{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer: vmcontract.SpecVer1,
						Pending: true,
					},
				},
			},
		},
		cabi.RewardsABI,
		cabi.JsonRewards,
	)
	vmcontract.RegisterContracts(contractaddress.RewardsAddress, rewards)

	blackHole := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNameDestroy: &contract.BlackHole{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer: vmcontract.SpecVer2,
						Pending: true,
					},
				},
			},
		},
		cabi.BlackHoleABI,
		cabi.JsonDestroy,
	)
	vmcontract.RegisterContracts(contractaddress.BlackHoleAddress, blackHole)

	miner := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNameMinerReward: &contract.MinerReward{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Pending:   true,
						Work:      true,
					},
				},
			},
		},
		cabi.MinerABI,
		cabi.JsonMiner,
	)
	vmcontract.RegisterContracts(contractaddress.MinerAddress, miner)

	rep := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNameRepReward: &contract.RepReward{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Pending:   true,
						Work:      true,
					},
				},
			},
		},
		cabi.RepABI,
		cabi.JsonRep,
	)
	vmcontract.RegisterContracts(contractaddress.RepAddress, rep)

	settlement := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNameCreateContract:    &contract.CreateContract{},
			cabi.MethodNameSignContract:      &contract.SignContract{},
			cabi.MethodNameProcessCDR:        &contract.ProcessCDR{},
			cabi.MethodNameAddPreStop:        &contract.AddPreStop{},
			cabi.MethodNameUpdatePreStop:     &contract.UpdatePreStop{},
			cabi.MethodNameRemovePreStop:     &contract.RemovePreStop{},
			cabi.MethodNameAddNextStop:       &contract.AddNextStop{},
			cabi.MethodNameUpdateNextStop:    &contract.UpdateNextStop{},
			cabi.MethodNameRemoveNextStop:    &contract.RemoveNextStop{},
			cabi.MethodNameTerminateContract: &contract.TerminateContract{},
			cabi.MethodNameRegisterAsset:     &contract.RegisterAsset{},
		},
		cabi.SettlementABI,
		cabi.JsonSettlement,
	)
	vmcontract.RegisterContracts(contractaddress.SettlementAddress, settlement)

	pk := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNamePKDVerifierRegister: &contract.VerifierRegister{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Work:      true,
					},
				},
			},
			cabi.MethodNamePKDVerifierUnregister: &contract.VerifierUnregister{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Work:      true,
					},
				},
			},
			cabi.MethodNamePKDPublish: &contract.Publish{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						PovState:  true,
					},
				},
			},
			cabi.MethodNamePKDUnPublish: &contract.UnPublish{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
					},
				},
			},
			cabi.MethodNamePKDOracle: &contract.Oracle{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						PovState:  true,
					},
				},
			},
			cabi.MethodNamePKDReward: &contract.PKDReward{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Pending:   true,
						Work:      true,
					},
				},
			},
			cabi.MethodNamePKDVerifierHeart: &contract.VerifierHeart{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						PovState:  true,
					},
				},
			},
		},
		cabi.PublicKeyDistributionABI,
		cabi.JsonPublicKeyDistribution,
	)
	vmcontract.RegisterContracts(contractaddress.PubKeyDistributionAddress, pk)

	pa := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNamePermissionAdminHandOver: &contract.AdminHandOver{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Work:      true,
						PovState:  true,
					},
				},
			},
			cabi.MethodNamePermissionNodeUpdate: &contract.NodeUpdate{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Work:      true,
						PovState:  true,
					},
				},
			},
		},
		cabi.PermissionABI,
		cabi.JsonPermission,
	)
	vmcontract.RegisterContracts(contractaddress.PermissionAddress, pa)

	pdkv := vmcontract.NewChainContract(
		map[string]vmcontract.Contract{
			cabi.MethodNamePrivacyDemoKVSet: &contract.PrivacyDemoKVSet{
				BaseContract: contract.BaseContract{
					Describe: vmcontract.Describe{
						SpecVer:   vmcontract.SpecVer2,
						Signature: true,
						Work:      true,
					},
				},
			},
		},
		cabi.PrivacyDemoKVABI,
		cabi.JsonPrivacyDemoKV,
	)
	vmcontract.RegisterContracts(contractaddress.PrivacyDemoKVAddress, pdkv)
}
