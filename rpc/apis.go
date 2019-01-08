package rpc

import (
	"github.com/qlcchain/go-qlc/rpc/api"
)

func (r *RPC) getApi(apiModule string) API {
	switch apiModule {
	case "qlcclassic":
		return API{
			Namespace: "qlcclassic",
			Version:   "1.0",
			Service:   api.NewQlcApi(r.ledger, r.dpos),
			Public:    true,
		}
	default:
		return API{}
	}
}

func (r *RPC) GetApis(apiModule ...string) []API {
	var apis []API
	for _, m := range apiModule {
		apis = append(apis, r.getApi(m))
	}
	return apis
}

//In-proc apis
func (r *RPC) GetInProcessApis() []API {
	//return GetApis("ledger", "wallet", "private_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "testapi", "pow", "tx")
	return r.GetApis("qlcclassic")
}

//Ipc apis
func (r *RPC) GetIpcApis() []API {
	return r.GetApis("qlcclassic")
	//return GetApis("ledger", "wallet", "private_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "testapi", "pow", "tx")
}

//Http apis
func (r *RPC) GetHttpApis() []API {
	apiModules := []string{"qlcclassic"}
	//apiModules := []string{"ledger", "public_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "pow", "tx"}
	//if node.Config().NetID > 1 {
	//	apiModules = append(apiModules, "testapi")
	//}
	return r.GetApis(apiModules...)
}

//WS apis
func (r *RPC) GetWSApis() []API {
	apiModules := []string{"qlcclassic"}
	//apiModules := []string{"ledger", "public_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "pow", "tx"}
	//if node.Config().NetID > 1 {
	//	apiModules = append(apiModules, "testapi")
	//}
	return r.GetApis(apiModules...)
}

func (r *RPC) GetPublicApis() []API {
	//return GetApis("ledger", "public_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "testapi", "pow", "tx", "debug", "dashboard")
	return r.GetApis("qlc")
}

func (r *RPC) GetAllApis() []API {
	//return GetApis("ledger", "wallet", "private_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "testapi", "pow", "tx", "debug", "dashboard")
	return r.GetApis("qlc")
}
