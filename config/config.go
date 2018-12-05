package config

import (
	"encoding/base64"
	"path"

	"github.com/libp2p/go-libp2p-peer"

	ic "github.com/libp2p/go-libp2p-crypto"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/util"
)

var log = common.NewLogger("config")

const (
	QlcConfigFile        = "qlc.json"
	QlcDataStorePath     = "ledger"
	QlcTestDataStorePath = "test/ledger"
)

var DefaultBootstrapAddresses = []string{
	"/ip4/47.244.138.61/tcp/29735/ipfs/QmaKU4cvFJ7x6A4nSEgcfirvJbn7eJbJgVBnU9QQuk2Kam",
}

type Config struct {
	RPC       *RPCConfig       `json:"rpc"`
	P2P       *P2PConfig       `json:"p2p"`
	Datastore *DatastoreConfig `json:"Datastore"`
	Discovery *DiscoveryConfig `json:"Discovery"`
	ID        *IdentityConfig  `json:"Identity"`
}

func DefaultConfig() (*Config, error) {
	identity, err := identityConfig()
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		RPC:       DefaultRPCConfig(),
		P2P:       DefaultP2PConfig(),
		Datastore: DefaultDatastoreConfig(),
		Discovery: DefaultDiscoveryConfig(),
		ID:        identity,
	}
	return cfg, nil
}

type RPCConfig struct {
	Enable bool   `json:"Enable"`
	Listen string `json:"Listen"`
}

func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		Enable: false,
		Listen: "/ip4/0.0.0.0/tcp/29735",
	}
}

type P2PConfig struct {
	BootNodes []string `json:"BootNode"`
	Listen    string   `json:"Listen"`
}

func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		BootNodes: DefaultBootstrapAddresses,
		Listen:    "/ip4/0.0.0.0/tcp/29734",
	}
}

type DatastoreConfig struct {
	StorageMax        string `json:"StorageMax"`
	DataStorePath     string `json:"DataStorePath"`
	TestDataStorePath string `json:"TestDataStorePath"`
}

func DefaultDatastoreConfig() *DatastoreConfig {
	datapath := path.Join(util.QlcDir(), QlcDataStorePath)
	testdatapath := path.Join(util.QlcDir(), QlcTestDataStorePath)
	return &DatastoreConfig{
		StorageMax:        "10GB",
		DataStorePath:     datapath,
		TestDataStorePath: testdatapath,
	}
}

type DiscoveryConfig struct {
	MDNS MDNS
}
type MDNS struct {
	Enabled bool
	// Time in seconds between discovery rounds
	Interval int
}

func DefaultDiscoveryConfig() *DiscoveryConfig {
	mdns := MDNS{
		Enabled:  true,
		Interval: 30,
	}
	return &DiscoveryConfig{
		MDNS: mdns,
	}
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
