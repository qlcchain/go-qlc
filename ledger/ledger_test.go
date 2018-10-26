package ledger

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/ledger/genesis"
)

func TestNewLedger(t *testing.T) {
	store, err := db.NewBadgerStore("./testdatabase")
	defer store.Close()
	if err != nil {
		t.Fatal(err)
	}
	gen, err := genesis.Get()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = NewLedger(store, LedgerOptions{Genesis: *gen}); err != nil {
		fmt.Println(err)
		if err == ErrBadGenesis || err == db.ErrAccountExists {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
	}
}
