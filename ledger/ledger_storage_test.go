package ledger

import (
	"encoding/json"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func addtokens(l *Ledger, t *testing.T) []*types.TokenInfo {
	ts, err := l.ListTokens()
	if err != nil {
		t.Fatal(err)
	}
	txn := l.db.NewTransaction(true)
	for _, token := range ts {
		key := getKeyOfHash(token.TokenId, idPrefixToken)
		val, _ := json.Marshal(token)

		if err := txn.Set(key, val); err != nil {
			t.Fatal(err)
		}
	}
	if err := txn.Commit(nil); err != nil {
		t.Fatal(err)
	}
	txn.Discard()
	return ts
}

func TestLedger_ListTokens(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ts := addtokens(l, t)
	tl, err := l.ListTokens()
	if len(tl) != len(ts) {
		t.Fatal()
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range ts {
		t1, err := l.GetTokenById(token.TokenId)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(t1)
		t2, err := l.GetTokenByName(token.TokenName)
		if err != nil {
			t.Fatal(err)
		}
		if t1.TokenName != t2.TokenName || t1.TokenId != t2.TokenId || t1.Owner != t2.Owner {
			t.Fatal("get token err")
		}
	}
}
