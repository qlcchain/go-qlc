// +build testnet

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

	h1, _ := types.NewHash("5594c690c3618a170a77d2696688f908efec4da2b94363fcb96749516307031d")
	genesisBlock := GenesisBlock()
	h2 := genesisBlock.GetHash()
	if h2 != h1 {
		t.Log(util.ToString(genesisBlock))
		t.Fatal("invalid genesis block", h2.String(), h1.String())
	}

	h1, _ = types.NewHash("424b367da2e0ff991d3086f599ce26547b80ae948b209f1cb7d63e19231ab213")
	gasBlock := GasBlock()
	h2 = gasBlock.GetHash()
	if h2 != h1 {
		t.Log(util.ToString(gasBlock))
		t.Fatal("invalid gas block", h2.String(), h1.String())
	}
}
