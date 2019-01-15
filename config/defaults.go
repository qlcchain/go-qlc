/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"encoding/base64"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/json-iterator/go"
	ic "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

const (
	QlcConfigFile = "qlc.json"
	configVersion = 1
)

var defaultBootstrapAddresses = []string{
	"/ip4/47.244.138.61/tcp/9734/ipfs/QmeSBhQe5kYtKqEQfSt7K3NR36Z3rBxF9JBBBWTBrDVap3",
}

func DefaultConfig() (*Config, error) {
	identity, err := identityConfig()
	if err != nil {
		return nil, err
	}

	var logCfg LogConfig
	_ = jsoniter.Unmarshal([]byte(`{
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

	cfg := &Config{
		Version:    configVersion,
		DataDir:    DefaultDataDir(),
		Mode:       "Normal",
		StorageMax: "10GB",
		LogConfig:  &logCfg,
		RPC: &RPCConfig{
			Enable: true,
			//Listen:       "/ip4/0.0.0.0/tcp/29735",
			HTTPEnabled:  true,
			HTTPEndpoint: "0.0.0.0:9735",
			WSEnabled:    true,
			WSEndpoint:   "0.0.0.0:9736",
			IPCEnabled:   true,
			IPCEndpoint:  defaultIPCEndpoint(),
		},
		P2P: &P2PConfig{
			BootNodes: defaultBootstrapAddresses,
			Listen:    "/ip4/0.0.0.0/tcp/9734",
		},
		Discovery: &DiscoveryConfig{
			MDNS: MDNS{
				Enabled:  true,
				Interval: 30,
			},
		},
		ID: identity,
	}
	return cfg, nil
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
			return filepath.Join(home, "Library", "Application Support", "GQlcchain")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "GQlcchain")
		} else {
			return filepath.Join(home, ".gqlcchain")
		}
	}
	return ""
}

func defaultIPCEndpoint() string {
	dir := filepath.Join(DefaultDataDir(), "gqlc.ipc")
	if runtime.GOOS == "windows" {
		//if strings.HasPrefix(dir, `\\.\pipe\`) {
		//	return dir
		//}
		return `\\.\pipe\` + dir
	}
	return dir
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
