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
			Service:   api.NewLedgerApi(r.ctx, r.ledger, r.relation, r.eb, r.cc),
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
			Service:   api.NewUtilApi(r.ledger),
			Public:    true,
		}
	case "wallet":
		return rpc.API{
			Namespace: "wallet",
			Version:   "1.0",
			Service:   api.NewWalletApi(r.ledger, r.wallet),
			Public:    true,
		}
	case "contract":
		return rpc.API{
			Namespace: "contract",
			Version:   "1.0",
			Service:   api.NewContractApi(r.ledger),
			Public:    true,
		}
	case "mintage":
		return rpc.API{
			Namespace: "mintage",
			Version:   "1.0",
			Service:   api.NewMintageApi(r.ledger),
			Public:    true,
		}
	case "pledge":
		return rpc.API{
			Namespace: "pledge",
			Version:   "1.0",
			Service:   api.NewNEP5PledgeAPI(r.cfgFile, r.ledger),
			Public:    true,
		}
	case "sms":
		return rpc.API{
			Namespace: "sms",
			Version:   "1.0",
			Service:   api.NewSMSApi(r.ledger, r.relation),
			Public:    true,
		}
	case "rewards":
		return rpc.API{
			Namespace: "rewards",
			Version:   "1.0",
			Service:   api.NewRewardsApi(r.ledger, r.cc),
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
	case "verifier":
		return rpc.API{
			Namespace: "verifier",
			Version:   "1.0",
			Service:   api.NewVerifierApi(r.cfgFile, r.ledger),
			Public:    true,
		}
	case "oracle":
		return rpc.API{
			Namespace: "oracle",
			Version:   "1.0",
			Service:   api.NewOracleApi(r.cfgFile, r.ledger),
			Public:    true,
		}
	case "publisher":
		return rpc.API{
			Namespace: "publisher",
			Version:   "1.0",
			Service:   api.NewPublisherApi(r.cfgFile, r.ledger),
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
	apiModules := []string{"ledger", "account", "net", "util", "wallet", "mintage", "contract", "sms", "pledge",
		"rewards", "pov", "miner", "config", "debug", "destroy", "metrics", "rep", "chain", "verifier", "oracle",
		"publisher"}
	return r.GetApis(apiModules...)
}
