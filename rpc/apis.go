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
			Service:   api.NewQlcApi(r.ledger, r.eb),
			Public:    true,
		}
	case "account":
		return API{
			Namespace: "account",
			Version:   "1.0",
			Service:   api.NewAccountApi(),
			Public:    true,
		}
	case "ledger":
		return API{
			Namespace: "ledger",
			Version:   "1.0",
			Service:   api.NewLedgerApi(r.ledger, r.relation, r.eb),
			Public:    true,
		}
	case "net":
		return API{
			Namespace: "net",
			Version:   "1.0",
			Service:   api.NewNetApi(r.ledger),
			Public:    true,
		}
	case "util":
		return API{
			Namespace: "util",
			Version:   "1.0",
			Service:   api.NewUtilApi(r.ledger),
			Public:    true,
		}
	case "wallet":
		return API{
			Namespace: "wallet",
			Version:   "1.0",
			Service:   api.NewWalletApi(r.ledger, r.wallet),
			Public:    true,
		}
	case "contract":
		return API{
			Namespace: "contract",
			Version:   "1.0",
			Service:   api.NewContractApi(r.ledger),
			Public:    true,
		}
	case "mintage":
		return API{
			Namespace: "mintage",
			Version:   "1.0",
			Service:   api.NewMintageApi(r.ledger),
			Public:    true,
		}
	case "pledge":
		return API{
			Namespace: "pledge",
			Version:   "1.0",
			Service:   api.NewNEP5PledgeApi(r.ledger),
			Public:    true,
		}
	case "sms":
		return API{
			Namespace: "sms",
			Version:   "1.0",
			Service:   api.NewSMSApi(r.ledger, r.relation),
			Public:    true,
		}
	case "pov":
		return API{
			Namespace: "pov",
			Version:   "1.0",
			Service:   api.NewPovApi(r.ledger),
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
	return r.GetPublicApis()
}

//Ipc apis
func (r *RPC) GetIpcApis() []API {
	return r.GetPublicApis()
}

//Http apis
func (r *RPC) GetHttpApis() []API {
	return r.GetPublicApis()
}

//WS apis
func (r *RPC) GetWSApis() []API {
	return r.GetPublicApis()
}

func (r *RPC) GetPublicApis() []API {
	apiModules := []string{"ledger", "account", "net", "util", "wallet", "mintage", "contract", "sms", "pledge", "pov"}
	return r.GetApis(apiModules...)
}

func (r *RPC) GetAllApis() []API {
	//return GetApis("ledger", "wallet", "private_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "testapi", "pow", "tx", "debug", "dashboard")
	return r.GetApis("qlc")
}
