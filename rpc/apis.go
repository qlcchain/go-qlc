package rpc

import (
	"github.com/qlcchain/go-qlc/rpc/api"
)

func GetApi(apiModule string) API {
	switch apiModule {
	case "qlc":
		return API{
			Namespace: "qlcclassic",
			Version:   "1.0",
			Service:   api.NewQlcApi(),
			Public:    true,
		}
	default:
		return API{}
	}
}

func GetApis(apiModule ...string) []API {
	var apis []API
	for _, m := range apiModule {
		apis = append(apis, GetApi(m))
	}
	logger.Info(apis)
	return apis
}

func GetPublicApis() []API {
	//return GetApis("ledger", "public_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "testapi", "pow", "tx", "debug", "dashboard")
	return GetApis("qlc")

}

func GetAllApis() []API {
	//return GetApis("ledger", "wallet", "private_onroad", "net", "contract", "pledge", "register", "vote", "mintage", "consensusGroup", "testapi", "pow", "tx", "debug", "dashboard")
	return GetApis("qlc")
}
