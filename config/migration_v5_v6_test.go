// +build !testnet

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestMigrationV5ToV6_Migration(t *testing.T) {
	dir := filepath.Join(QlcTestDataDir(), "config")
	defer func() { _ = os.RemoveAll(dir) }()
	cfg5, err := DefaultConfigV5(dir)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(cfg5)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMigrationV5ToV6()
	migration, i, err := m.Migration(bytes, 5)
	if err != nil || i != 6 {
		t.Fatal("migration failed")
	}

	cfg6 := &Config{}
	err = json.Unmarshal(migration, cfg6)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg6.Genesis.GenesisBlocks) != 2 {
		t.Fatal("genesis block lens error")
	}

	loadGenesisAccount(cfg6)

	h1, _ := types.NewHash("7201b4c283b7a32e88ec4c5867198da574de1718eb18c7f95ee8ef733c0b5609")
	genesisBlock := GenesisBlock()
	h2 := genesisBlock.GetHash()
	if h2 != h1 {
		t.Log(util.ToString(genesisBlock))
		t.Fatal("invalid genesis block", h2.String(), h1.String())
	}

	h1, _ = types.NewHash("b9e2ea2e4310c38ed82ff492cb83229b4361d89f9c47ebbd6653ddec8a07ebe1")
	gasBlock := GasBlock()
	h2 = gasBlock.GetHash()
	if h2 != h1 {
		t.Log(util.ToString(gasBlock))
		t.Fatal("invalid gas block", h2.String(), h1.String())
	}
}
