package chaincontract

import (
	"sync"

	"github.com/qlcchain/go-qlc/common/vmcontract"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/vm/contract"
)

var lock = sync.RWMutex{}

func InitChainContract() {
	lock.Lock()
	defer lock.Unlock()
	vmcontract.RegisterContracts(contractaddress.MintageAddress, contract.MintageContract)
	vmcontract.RegisterContracts(contractaddress.NEP5PledgeAddress, contract.Nep5Contract)
	vmcontract.RegisterContracts(contractaddress.RewardsAddress, contract.RewardsContract)
	vmcontract.RegisterContracts(contractaddress.BlackHoleAddress, contract.BlackHoleContract)
	vmcontract.RegisterContracts(contractaddress.MinerAddress, contract.MinerContract)
	vmcontract.RegisterContracts(contractaddress.RepAddress, contract.RepContract)
	vmcontract.RegisterContracts(contractaddress.SettlementAddress, contract.SettlementContract)
	vmcontract.RegisterContracts(contractaddress.PubKeyDistributionAddress, contract.PKDContract)
	vmcontract.RegisterContracts(contractaddress.PermissionAddress, contract.PermissionContract)
	vmcontract.RegisterContracts(contractaddress.PrivacyDemoKVAddress, contract.PdkvContract)
}
