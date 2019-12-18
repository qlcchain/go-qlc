package dpos

import (
	"encoding/json"
	"math/big"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func getTestEl() *Election {
	dir := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	cm := config.NewCfgManager(dir)
	cfg, _ := cm.Load()
	var mintageBlock, genesisBlock types.StateBlock
	for _, v := range cfg.Genesis.GenesisBlocks {
		_ = json.Unmarshal([]byte(v.Genesis), &genesisBlock)
		_ = json.Unmarshal([]byte(v.Mintage), &mintageBlock)
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: mintageBlock,
			GenesisBlock:        genesisBlock,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}

	dps := NewDPoS(cm.ConfigFile)
	blk := mock.StateBlock()
	return newElection(dps, blk)
}

func TestIfValidAndSetInvalid(t *testing.T) {
	el := getTestEl()

	if !el.ifValidAndSetInvalid() {
		t.Fatal()
	}

	if el.ifValidAndSetInvalid() {
		t.Fatal()
	}
}

func TestIfValid(t *testing.T) {
	el := getTestEl()

	if !el.isValid() {
		t.Fatal()
	}

	if !el.ifValidAndSetInvalid() {
		t.Fatal()
	}

	if el.isValid() {
		t.Fatal()
	}
}

func TestGetGenesisBalance(t *testing.T) {
	el := getTestEl()

	b, err := el.getGenesisBalance()
	if err != nil {
		t.Fatal()
	}

	if b.Cmp(big.NewInt(60000000000000000)) != 0 {
		t.Fatal()
	}
}
