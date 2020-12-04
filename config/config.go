package config

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"time"

	ic "github.com/libp2p/go-libp2p-core/crypto"
)

type Config ConfigV10

func DefaultConfig(dir string) (*Config, error) {
	v10, err := DefaultConfigV10(dir)
	if err != nil {
		return &Config{}, err
	}
	cfg := Config(*v10)

	return &cfg, nil
}

func (c Config) Clone() (*Config, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	result := new(Config)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DecodePrivateKey is a helper to decode the users PrivateKey
func (c *Config) DecodePrivateKey() (ic.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(c.P2P.ID.PrivKey)
	if err != nil {
		return nil, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO:(security)
	return ic.UnmarshalPrivateKey(pkb)
}

func (c *Config) LogDir() string {
	return filepath.Join(c.DataDir, "log", time.Now().Format("2006-01-02T15-04"))
}

func (c *Config) LedgerDir() string {
	return filepath.Join(c.DataDir, "ledger")
}

func (c *Config) WalletDir() string {
	return filepath.Join(c.DataDir, "wallet")
}

func (c *Config) SqliteDir() string {
	return filepath.Join(c.LedgerDir(), relationDir)
}
