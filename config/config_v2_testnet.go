// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

func DefaultConfigV2(dir string) (*ConfigV2, error) {
	pk, id, err := identityConfig()
	if err != nil {
		return nil, err
	}
	var cfg ConfigV2
	modules := []string{"qlcclassic", "ledger", "account", "net", "util", "wallet", "mintage", "contract", "sms"}
	cfg = ConfigV2{
		Version:             2,
		DataDir:             dir,
		StorageMax:          "10GB",
		AutoGenerateReceive: false,
		LogLevel:            "error",
		PerformanceEnabled:  false,
		RPC: &RPCConfigV2{
			Enable:           false,
			HTTPEnabled:      true,
			HTTPEndpoint:     "tcp4://0.0.0.0:19735",
			HTTPCors:         []string{"*"},
			HttpVirtualHosts: []string{},
			WSEnabled:        true,
			WSEndpoint:       "tcp4://0.0.0.0:19736",
			IPCEnabled:       true,
			IPCEndpoint:      defaultIPCEndpoint(),
			PublicModules:    modules,
		},
		P2P: &P2PConfigV2{
			BootNodes: []string{
				"/ip4/47.103.40.20/tcp/19734/ipfs/QmdFSukPUMF3t1JxjvTo14SEEb5JV9JBT6PukGRo6A2g4f",
				"/ip4/47.112.112.138/tcp/19734/ipfs/QmW9ocg4fRjckCMQvRNYGyKxQd6GiutAY4HBRxMrGrZRfc",
			},
			Listen:       "/ip4/0.0.0.0/tcp/19734",
			SyncInterval: 120,
			Discovery: &DiscoveryConfigV2{
				DiscoveryInterval: 30,
				Limit:             20,
				MDNSEnabled:       false,
				MDNSInterval:      30,
			},
			ID: &IdentityConfigV2{id, pk},
		},
	}
	return &cfg, nil
}
