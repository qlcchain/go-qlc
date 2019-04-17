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

type ConfigV1 struct {
	Version             int        `json:"version"`
	DataDir             string     `json:"DataDir"`
	StorageMax          string     `json:"StorageMax"`
	Mode                string     `json:"mode"` // runtime mode: Test,Normal
	AutoGenerateReceive bool       `json:"AutoGenerateReceive"`
	LogConfig           *logConfig `json:"log"` //log config

	RPC *RPCConfigV1 `json:"rpc"`
	P2P *P2PConfigV1 `json:"p2p"`

	Discovery *DiscoveryConfigV1 `json:"Discovery"`
	ID        *IdentityConfigV1  `json:"Identity"`

	PerformanceTest *PerformanceTestConfigV1
}

type logConfig struct {
	Level            string   `json:"level"` // log level: info,warn,debug
	OutputPaths      []string `json:"outputPaths"`
	ErrorOutputPaths []string `json:"errorOutputPaths"`
	Encoding         string   `json:"encoding"`
	EncoderConfig    struct {
		MessageKey   string `json:"messageKey"`
		LevelKey     string `json:"levelKey"`
		LevelEncoder string `json:"levelEncoder"`
	} `json:"encoderConfig"`
}

type RPCConfigV1 struct {
	Enable bool `json:"enable"`
	//Listen string `json:"Listen"`
	HTTPEndpoint     string   `json:"hTTPEndpoint"`
	HTTPEnabled      bool     `json:"hTTPEnabled"`
	HTTPCors         []string `json:"hTTPCors"`
	HttpVirtualHosts []string `json:"httpVirtualHosts"`

	WSEnabled   bool   `json:"wSEnabled"`
	WSEndpoint  string `json:"wSEndpoint"`
	IPCEndpoint string `json:"iPCEndpoint"`

	IPCEnabled bool `json:"iPCEnabled"`
}

type P2PConfigV1 struct {
	BootNodes []string `json:"BootNode"`
	Listen    string   `json:"Listen"`
	//Time in seconds between sync block interval
	SyncInterval int `json:"SyncInterval"`
}

type DiscoveryConfigV1 struct {
	// Time in seconds between remote discovery rounds
	DiscoveryInterval int
	//The maximum number of discovered nodes at a time
	Limit int
	MDNS  MDNSV1
}

type MDNSV1 struct {
	Enabled bool
	// Time in seconds between local discovery rounds
	Interval int
}

// Identity tracks the configuration of the local node's identity.
type IdentityConfigV1 struct {
	PeerID  string
	PrivKey string `json:",omitempty"`
}

type PerformanceTestConfigV1 struct {
	Enabled bool
}

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
