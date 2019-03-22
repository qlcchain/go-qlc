/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import "encoding/json"

type MigrationV1ToV2 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV1ToV2() *MigrationV1ToV2 {
	return &MigrationV1ToV2{startVersion: 1, endVersion: 2}
}

func (m *MigrationV1ToV2) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg1 ConfigV1
	err := json.Unmarshal(data, &cfg1)
	if err != nil {
		return data, version, err
	}

	cfg2, err := DefaultConfigV2(cfg1.DataDir)
	if err != nil {
		return data, version, err
	}

	cfg2.AutoGenerateReceive = cfg1.AutoGenerateReceive
	cfg2.PerformanceEnabled = cfg1.PerformanceTest.Enabled
	// p2p
	cfg2.P2P.BootNodes = cfg1.P2P.BootNodes
	cfg2.P2P.Listen = cfg1.P2P.Listen
	cfg2.P2P.SyncInterval = cfg1.P2P.SyncInterval
	cfg2.P2P.Discovery.Limit = cfg1.Discovery.Limit
	cfg2.P2P.Discovery.DiscoveryInterval = cfg1.Discovery.DiscoveryInterval
	cfg2.P2P.Discovery.MDNSInterval = cfg1.Discovery.MDNS.Interval
	cfg2.P2P.ID.PeerID = cfg1.ID.PeerID
	cfg2.P2P.ID.PrivKey = cfg1.ID.PrivKey

	//rpc
	cfg2.RPC.HTTPEndpoint = cfg1.RPC.HTTPEndpoint
	cfg2.RPC.HTTPCors = cfg1.RPC.HTTPCors
	cfg2.RPC.IPCEndpoint = cfg1.RPC.IPCEndpoint
	cfg2.RPC.WSEndpoint = cfg1.RPC.WSEndpoint

	bytes, err := json.Marshal(cfg2)
	return bytes, m.endVersion, err
}

func (m *MigrationV1ToV2) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV1ToV2) EndVersion() int {
	return m.endVersion
}
