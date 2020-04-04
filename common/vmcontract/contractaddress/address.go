package contractaddress

import "github.com/qlcchain/go-qlc/common/types"

var (
	MintageAddress, _    = types.HexToAddress("qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1")
	NEP5PledgeAddress, _ = types.HexToAddress("qlc_3fwi6r1fzjwmiys819pw8jxrcmcottsj4iq56kkgcmzi3b87596jwskwqrr5")
	RewardsAddress, _    = types.HexToAddress("qlc_3oinqggowa7f1rsjfmib476ggz6s4fp8578odjzerzztkrifqkqdz5zjztb3")

	// Builtin contract addresses without private key hold by anyone
	MinerAddress, _              = GenerateBuiltinContractAddress(21)
	BlackHoleAddress, _          = GenerateBuiltinContractAddress(22)
	RepAddress, _                = GenerateBuiltinContractAddress(23)
	PubKeyDistributionAddress, _ = GenerateBuiltinContractAddress(24)
	SettlementAddress, _         = GenerateBuiltinContractAddress(25)
	PermissionAddress, _         = GenerateBuiltinContractAddress(26)
	PrivacyDemoKVAddress, _      = GenerateBuiltinContractAddress(27)

	ChainContractAddressList = []types.Address{NEP5PledgeAddress, MintageAddress, RewardsAddress, MinerAddress,
		BlackHoleAddress, RepAddress, PubKeyDistributionAddress, SettlementAddress, PermissionAddress,
		PrivacyDemoKVAddress,
	}
	RewardContractAddressList = []types.Address{MinerAddress, RepAddress}
)

func GenerateBuiltinContractAddress(suffix byte) (types.Address, error) {
	buf := make([]byte, types.AddressSize)
	buf[types.AddressSize-1] = suffix
	return types.BytesToAddress(buf)
}

func IsChainContractAddress(address types.Address) bool {
	for _, itAddr := range ChainContractAddressList {
		if itAddr == address {
			return true
		}
	}

	return false
}

func IsContractAddress(address types.Address) bool {
	if IsChainContractAddress(address) {
		return true
	}

	return false
}

func IsRewardContractAddress(address types.Address) bool {
	for _, itAddr := range RewardContractAddressList {
		if itAddr == address {
			return true
		}
	}
	return false
}
