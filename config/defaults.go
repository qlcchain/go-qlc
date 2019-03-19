/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	ic "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/qlcchain/go-qlc"
)

const (
	QlcConfigFile = "qlc.json"
	configVersion = 1
	cfgDir        = "GQlcchain"
	nixCfgDir     = ".gqlcchain"
	suffix        = "_test"
)

func DefaultConfig(dir string) (*Config, error) {
	identity, err := identityConfig()
	if err != nil {
		return nil, err
	}

	var logCfg LogConfig
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

	if goqlc.MAINNET {
		return &Config{
			Version:             configVersion,
			DataDir:             dir,
			Mode:                "Normal",
			StorageMax:          "10GB",
			AutoGenerateReceive: false,
			LogConfig:           &logCfg,
			RPC: &RPCConfig{
				Enable:           true,
				HTTPEnabled:      true,
				HTTPEndpoint:     "tcp4://0.0.0.0:9735",
				HTTPCors:         []string{"*"},
				HttpVirtualHosts: []string{},
				WSEnabled:        true,
				WSEndpoint:       "tcp4://0.0.0.0:9736",
				IPCEnabled:       true,
				IPCEndpoint:      defaultIPCEndpoint(),
			},
			P2P: &P2PConfig{
				BootNodes: []string{
					"/ip4/47.244.138.61/tcp/9734/ipfs/QmdFSukPUMF3t1JxjvTo14SEEb5JV9JBT6PukGRo6A2g4f",
					"/ip4/47.75.145.146/tcp/9734/ipfs/QmW9ocg4fRjckCMQvRNYGyKxQd6GiutAY4HBRxMrGrZRfc",
				},
				Listen:       "/ip4/0.0.0.0/tcp/9734",
				SyncInterval: 120,
			},
			Discovery: &DiscoveryConfig{
				DiscoveryInterval: 30,
				Limit:             20,
				MDNS: MDNS{
					Enabled:  true,
					Interval: 30,
				},
			},
			ID: identity,
			PerformanceTest: &PerformanceTestConfig{
				Enabled: false,
			},
		}, nil
	}

	return &Config{
		Version:             configVersion,
		DataDir:             dir,
		Mode:                "Normal",
		StorageMax:          "10GB",
		AutoGenerateReceive: false,
		LogConfig:           &logCfg,
		RPC: &RPCConfig{
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
		P2P: &P2PConfig{
			BootNodes: []string{
				"/ip4/47.244.138.61/tcp/19734/ipfs/QmdFSukPUMF3t1JxjvTo14SEEb5JV9JBT6PukGRo6A2g4f",
				"/ip4/47.75.145.146/tcp/19734/ipfs/QmW9ocg4fRjckCMQvRNYGyKxQd6GiutAY4HBRxMrGrZRfc",
			},
			Listen:       "/ip4/0.0.0.0/tcp/19734",
			SyncInterval: 120,
		},
		Discovery: &DiscoveryConfig{
			DiscoveryInterval: 30,
			Limit:             20,
			MDNS: MDNS{
				Enabled:  true,
				Interval: 30,
			},
		},
		ID: identity,
		PerformanceTest: &PerformanceTestConfig{
			Enabled: false,
		},
	}, nil
}

// identityConfig initializes a new identity.
func identityConfig() (*IdentityConfig, error) {
	ident := IdentityConfig{}

	sk, pk, err := ic.GenerateKeyPair(ic.RSA, 2048)
	if err != nil {
		return &ident, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := sk.Bytes()
	if err != nil {
		return &ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return &ident, err
	}
	ident.PeerID = id.Pretty()
	return &ident, nil
}

// DefaultDataDir is the default data directory to use for the databases and other persistence requirements.
func DefaultDataDir() string {
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			var d string
			if goqlc.MAINNET {
				d = cfgDir
			} else {
				d = cfgDir + suffix
			}
			return filepath.Join(home, "Library", "Application Support", d)
		} else if runtime.GOOS == "windows" {
			var d string
			if goqlc.MAINNET {
				d = cfgDir
			} else {
				d = cfgDir + suffix
			}
			return filepath.Join(home, "AppData", "Roaming", d)
		} else {
			var d string
			if goqlc.MAINNET {
				d = nixCfgDir
			} else {
				d = nixCfgDir + suffix
			}
			return filepath.Join(home, d)
		}
	}
	return ""
}

func defaultIPCEndpoint() string {
	if goqlc.MAINNET {
		dir := filepath.Join(DefaultDataDir(), "gqlc.ipc")
		if runtime.GOOS == "windows" {
			return `\\.\pipe\gqlc.ipc`
		}
		return dir
	} else {
		dir := filepath.Join(DefaultDataDir(), "gqlc_test.ipc")
		if runtime.GOOS == "windows" {
			return `\\.\pipe\gqlc-test.ipc`
		}
		return dir
	}
}

func DefaultConfigFile() string {
	return filepath.Join(DefaultDataDir(), QlcConfigFile)
}

func QlcTestDataDir() string {
	return filepath.Join(DefaultDataDir(), "test")
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
