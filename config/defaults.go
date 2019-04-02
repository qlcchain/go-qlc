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

	ic "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	goqlc "github.com/qlcchain/go-qlc"
)

const (
	QlcConfigFile = "qlc.json"
	configVersion = 2
	cfgDir        = "GQlcchain"
	nixCfgDir     = ".gqlcchain"
	suffix        = "_test"
)

// identityConfig initializes a new identity.
func identityConfig() (string, string, error) {
	sk, pk, err := ic.GenerateKeyPair(ic.RSA, 2048)
	if err != nil {
		return "", "", err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := sk.Bytes()
	if err != nil {
		return "", "", err
	}
	privKey := base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return "", "", err
	}
	peerID := id.Pretty()
	return privKey, peerID, nil
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
