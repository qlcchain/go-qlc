package rpc

import (
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/rpc/api"
)

func (r *RPC) getApi(apiModule string) rpc.API {
	switch apiModule {
	case "account":
		return rpc.API{
			Namespace: "account",
			Version:   "1.0",
			Service:   api.NewAccountApi(),
			Public:    true,
		}
	case "ledger":
		return rpc.API{
			Namespace: "ledger",
			Version:   "1.0",
			Service:   api.NewLedgerApi(r.ctx, r.ledger, r.eb, r.cc),
			Public:    true,
		}
	case "net":
		return rpc.API{
			Namespace: "net",
			Version:   "1.0",
			Service:   api.NewNetApi(r.ledger, r.eb, r.cc),
			Public:    true,
		}
	case "util":
		return rpc.API{
			Namespace: "util",
			Version:   "1.0",
			Service:   api.NewUtilAPI(r.ledger),
			Public:    true,
		}
	case "contract":
		return rpc.API{
			Namespace: "contract",
			Version:   "1.0",
			Service:   api.NewContractApi(),
			Public:    true,
		}
	case "mintage":
		return rpc.API{
			Namespace: "mintage",
			Version:   "1.0",
			Service:   api.NewMintageApi(r.cfgFile, r.ledger),
			Public:    true,
		}
	case "pledge":
		return rpc.API{
			Namespace: "pledge",
			Version:   "1.0",
			Service:   api.NewNEP5PledgeAPI(r.cfgFile, r.ledger),
			Public:    true,
		}
	case "rewards":
		return rpc.API{
			Namespace: "rewards",
			Version:   "1.0",
			Service:   api.NewRewardsAPI(r.ledger, r.cc),
			Public:    true,
		}
	case "pov":
		return rpc.API{
			Namespace: "pov",
			Version:   "1.0",
			Service:   api.NewPovApi(r.ctx, r.config, r.ledger, r.eb, r.cc),
			Public:    true,
		}
	case "miner":
		return rpc.API{
			Namespace: "miner",
			Version:   "1.0",
			Service:   api.NewMinerApi(r.config, r.ledger),
			Public:    true,
		}
	case "config":
		return rpc.API{
			Namespace: "config",
			Version:   "1.0",
			Service:   api.NewConfigApi(r.cfgFile),
			Public:    true,
		}
	case "rep":
		return rpc.API{
			Namespace: "rep",
			Version:   "1.0",
			Service:   api.NewRepApi(r.config, r.ledger),
			Public:    true,
		}
	case "debug":
		return rpc.API{
			Namespace: "debug",
			Version:   "1.0",
			Service:   api.NewDebugApi(r.cfgFile, r.eb),
			Public:    true,
		}
	case "destroy":
		return rpc.API{
			Namespace: "destroy",
			Version:   "1.0",
			Service:   api.NewBlackHoleApi(r.ledger, r.cc),
			Public:    true,
		}
	case "metrics":
		return rpc.API{
			Namespace: "metrics",
			Version:   "1.0",
			Service:   api.NewMetricsApi(),
			Public:    true,
		}
	case "chain":
		return rpc.API{
			Namespace: "chain",
			Version:   "1.0",
			Service:   api.NewChainApi(r.ledger),
			Public:    true,
		}
	case "settlement":
		return rpc.API{
			Namespace: "settlement",
			Version:   "1.0",
			Service:   api.NewSettlement(r.ledger, r.cc),
			Public:    true,
		}
	case "dpki":
		return rpc.API{
			Namespace: "dpki",
			Version:   "1.0",
			Service:   api.NewPublicKeyDistributionApi(r.cfgFile, r.ledger),
			Public:    true,
		}
	case "permission":
		return rpc.API{
			Namespace: "permission",
			Version:   "1.0",
			Service:   api.NewPermissionApi(r.cfgFile, r.ledger),
			Public:    true,
		}
	case "privacy":
		return rpc.API{
			Namespace: "privacy",
			Version:   "1.0",
			Service:   api.NewPrivacyApi(r.config, r.ledger, r.eb, r.cc),
			Public:    true,
		}
	default:
		return rpc.API{}
	}
}

func (r *RPC) GetApis(apiModule ...string) []rpc.API {
	var apis []rpc.API
	for _, m := range apiModule {
		apis = append(apis, r.getApi(m))
	}
	return apis
}

//In-proc apis
func (r *RPC) GetInProcessApis() []rpc.API {
	return r.GetPublicApis()
}

//Ipc apis
func (r *RPC) GetIpcApis() []rpc.API {
	return r.GetPublicApis()
}

//Http apis
func (r *RPC) GetHttpApis() []rpc.API {
	return r.GetPublicApis()
}

//WS apis
func (r *RPC) GetWSApis() []rpc.API {
	return r.GetPublicApis()
}

func (r *RPC) GetPublicApis() []rpc.API {
	apiModules := []string{"ledger", "account", "net", "util", "mintage", "contract", "pledge",
		"rewards", "pov", "miner", "config", "debug", "destroy", "metrics", "rep", "chain", "dpki", "settlement",
		"permission", "privacy"}
	return r.GetApis(apiModules...)
}
