/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

type ConfigV2 struct {
	Version             int    `json:"version"`
	DataDir             string `json:"dataDir"`
	StorageMax          string `json:"storageMax"`
	AutoGenerateReceive bool   `json:"autoGenerateReceive"`
	LogLevel            string `json:"logLevel"` //info,warn,debug
	PerformanceEnabled  bool   `json:"performanceEnabled"`

	RPC *RPCConfigV2 `json:"rpc"`
	P2P *P2PConfigV2 `json:"p2p"`
}

type P2PConfigV2 struct {
	BootNodes          []string `json:"bootNode" mapstructure:"bootNode"`
	IsBootNode         bool     `json:"isBootNode"`
	BootNodeHttpServer string   `json:"bootNodeHttpServer"`
	Listen             string   `json:"listen"`
	WhiteListMode      bool     `json:"whiteListMode"`
	//Time in seconds between sync block interval
	SyncInterval int                `json:"syncInterval"`
	Discovery    *DiscoveryConfigV2 `json:"discovery"`
	ID           *IdentityConfigV2  `json:"identity" mapstructure:"identity"`
}

type RPCConfigV2 struct {
	Enable bool `json:"rpcEnabled" mapstructure:"rpcEnabled"`
	//Listen string `json:"Listen"`
	HTTPEndpoint     string   `json:"httpEndpoint"`
	HTTPEnabled      bool     `json:"httpEnabled"`
	HTTPCors         []string `json:"httpCors"`
	HttpVirtualHosts []string `json:"httpVirtualHosts"`

	WSEnabled  bool   `json:"webSocketEnabled" mapstructure:"webSocketEnabled"`
	WSEndpoint string `json:"webSocketEndpoint" mapstructure:"webSocketEndpoint"`

	IPCEndpoint   string   `json:"ipcEndpoint"`
	IPCEnabled    bool     `json:"ipcEnabled"`
	PublicModules []string `json:"publicModules"`
}

type DiscoveryConfigV2 struct {
	// Time in seconds between remote discovery rounds
	DiscoveryInterval int `json:"discoveryInterval"`
	//The maximum number of discovered nodes at a time
	Limit       int  `json:"limit"`
	MDNSEnabled bool `json:"mDNSEnabled"`
	// Time in seconds between local discovery rounds
	MDNSInterval int `json:"mDNSInterval"`
}

type IdentityConfigV2 struct {
	PeerID  string `json:"peerId"`
	PrivKey string `json:"privateKey,omitempty" mapstructure:"privateKey"`
}
