/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/crypto/random"
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
		if common.CheckTestMode("POV") {
			if runtime.GOOS == "darwin" {
				return filepath.Join(home, "Library", "Application Support", cfgDir+"_pov")
			} else if runtime.GOOS == "windows" {
				return filepath.Join(home, "AppData", "Roaming", cfgDir+"_pov")
			} else {
				return filepath.Join(home, nixCfgDir+"_pov")
			}
		}
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Application Support", cfgDir)
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", cfgDir)
		} else {
			return filepath.Join(home, nixCfgDir)
		}
	}
	return ""
}

func defaultIPCEndpoint(dir string) string {
	ipc := filepath.Join(dir, ipcName)
	if runtime.GOOS == "windows" {
		prefix := random.RandomHexString(2)
		return fmt.Sprintf(`\\.\pipe\%s-%s`, prefix, ipcName)
	}
	return ipc
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
