package config

import (
	"encoding/base64"
	"path/filepath"
	"time"

	ic "github.com/libp2p/go-libp2p-crypto"
)

type Config struct {
	Version    int        `json:"version"`
	DataDir    string     `json:"DataDir"`
	StorageMax string     `json:"StorageMax"`
	Mode       string     `json:"mode"` // runtime mode: Test,Normal
	LogConfig  *LogConfig `json:"log"`  //log config

	RPC *RPCConfig `json:"rpc"`
	P2P *P2PConfig `json:"p2p"`

	Discovery *DiscoveryConfig `json:"Discovery"`
	ID        *IdentityConfig  `json:"Identity"`
}

type LogConfig struct {
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

type RPCConfig struct {
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

type P2PConfig struct {
	BootNodes []string `json:"BootNode"`
	Listen    string   `json:"Listen"`
}

type DiscoveryConfig struct {
	MDNS MDNS
}

type MDNS struct {
	Enabled bool
	// Time in seconds between discovery rounds
	Interval int
}

// Identity tracks the configuration of the local node's identity.
type IdentityConfig struct {
	PeerID  string
	PrivKey string `json:",omitempty"`
}

// DecodePrivateKey is a helper to decode the users PrivateKey
func (cfg *Config) DecodePrivateKey() (ic.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(cfg.ID.PrivKey)
	if err != nil {
		return nil, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
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
