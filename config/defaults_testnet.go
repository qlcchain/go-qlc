// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
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
			d := cfgDir + suffix
			return filepath.Join(home, "Library", "Application Support", d)
		} else if runtime.GOOS == "windows" {
			d := cfgDir + suffix
			return filepath.Join(home, "AppData", "Roaming", d)
		} else {
			d := nixCfgDir + suffix
			return filepath.Join(home, d)
		}
	}
	return ""
}

func defaultIPCEndpoint() string {
	dir := filepath.Join(DefaultDataDir(), "gqlc_test.ipc")
	if runtime.GOOS == "windows" {
		return `\\.\pipe\gqlc-test.ipc`
	}
	return dir
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

func defaultDbConfig() string {
	return ""
}
