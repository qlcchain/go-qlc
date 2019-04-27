// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"encoding/json"
)

func DefaultConfigV1(dir string) (*ConfigV1, error) {
	pk, id, err := identityConfig()
	if err != nil {
		return nil, err
	}

	var logCfg logConfig
	_ = json.Unmarshal([]byte(`{
		"level": "info",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoding": "json",
		"encoderConfig": {
			"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
		}
	}`), &logCfg)

	var cfg ConfigV1
	cfg = ConfigV1{
		Version:             1,
		DataDir:             dir,
		Mode:                "Normal",
		StorageMax:          "10GB",
		AutoGenerateReceive: false,
		LogConfig:           &logCfg,
		RPC: &RPCConfigV1{
			Enable:           true,
			HTTPEnabled:      true,
			HTTPEndpoint:     "tcp4://0.0.0.0:19735",
			HTTPCors:         []string{"*"},
			HttpVirtualHosts: []string{},
			WSEnabled:        true,
			WSEndpoint:       "tcp4://0.0.0.0:19736",
			IPCEnabled:       true,
			IPCEndpoint:      defaultIPCEndpoint(),
		},
		P2P: &P2PConfigV1{
			BootNodes: []string{
				"/ip4/47.103.40.20/tcp/19734/ipfs/QmdFSukPUMF3t1JxjvTo14SEEb5JV9JBT6PukGRo6A2g4f",
				"/ip4/47.112.112.138/tcp/19734/ipfs/QmW9ocg4fRjckCMQvRNYGyKxQd6GiutAY4HBRxMrGrZRfc",
			},
			Listen:       "/ip4/0.0.0.0/tcp/19734",
			SyncInterval: 120,
		},
		Discovery: &DiscoveryConfigV1{
			DiscoveryInterval: 30,
			Limit:             20,
			MDNS: MDNSV1{
				Enabled:  true,
				Interval: 30,
			},
		},
		ID: &IdentityConfigV1{PeerID: id, PrivKey: pk},
		PerformanceTest: &PerformanceTestConfigV1{
			Enabled: false,
		},
	}

	return &cfg, nil
}
