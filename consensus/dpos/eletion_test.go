package dpos

import (
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"math/big"
	"path/filepath"
	"testing"
)

var (
	el	*Election
)

func init() {
	dir := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	cm := config.NewCfgManager(dir)

	dps := NewDPoS(cm.ConfigFile)
	blk := mock.StateBlock()
	el = newElection(dps, blk)
}

func TestIfValidAndSetInvalid(t *testing.T) {
	if !el.ifValidAndSetInvalid() {
		t.Fatal()
	}

	if el.ifValidAndSetInvalid() {
		t.Fatal()
	}
}

func TestIfValid(t *testing.T) {
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
	b, err := el.getGenesisBalance()
	if err != nil {
		t.Fatal()
	}

	if b.Cmp(big.NewInt(60000000000000000)) != 0 {
		t.Fatal()
	}
}