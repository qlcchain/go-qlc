/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

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
