package ledger

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func addAccountMeta(t *testing.T, l *Ledger) *types.AccountMeta {
	ac := mock.Account()
	am := mock.AccountMeta(ac.Address())
	if err := l.AddAccountMeta(am); err != nil {
		t.Fatal()
	}
	return am
}

func TestLedger_AddAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addAccountMeta(t, l)
}

func TestLedger_GetAccountMeta_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address := mock.Address()
	_, err := l.GetAccountMeta(address)
	if err != ErrAccountNotFound {
		t.Fatal(err)
	}
}

func TestLedger_GetAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	am := addAccountMeta(t, l)
	a, err := l.GetAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", a)
	for _, token := range a.Tokens {
		t.Log("token,", token)
	}
}

func TestLedger_HasAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	r, err := l.HasAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has account,", r)
}

func TestLedger_DeleteAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	err := l.DeleteAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddOrUpdateAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	token := mock.TokenMeta(am.Address)
	am.Tokens = append(am.Tokens, token)

	err := l.AddOrUpdateAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_UpdateAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	token := mock.TokenMeta(am.Address)
	am.Tokens = append(am.Tokens, token)

	err := l.AddOrUpdateAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}
}

func addTokenMeta(t *testing.T, l *Ledger) *types.TokenMeta {
	tm := addAccountMeta(t, l)
	token := mock.TokenMeta(tm.Address)
	if err := l.AddTokenMeta(token.BelongTo, token); err != nil {
		t.Fatal(err)
	}
	return token
}

func TestLedger_AddTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addTokenMeta(t, l)
}

func TestLedger_GetTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	token, err := l.GetTokenMeta(token.BelongTo, token.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token,", token)
	_, err = l.GetTokenMetaConfirmed(token.BelongTo, token.Type)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddOrUpdateTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	token2 := mock.TokenMeta(token.BelongTo)
	err := l.AddOrUpdateTokenMeta(token.BelongTo, token2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_UpdateTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	token := addTokenMeta(t, l)
	token2 := mock.TokenMeta(token.BelongTo)
	err := l.AddOrUpdateTokenMeta(token.BelongTo, token2)
	if err != nil {
		t.Fatal(err)
	}
	token2.Header = mock.Hash()
	err = l.UpdateTokenMeta(token.BelongTo, token2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_DelTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	token := addTokenMeta(t, l)
	err := l.DeleteTokenMeta(token.BelongTo, token.Type)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_GetAccountMetas(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	addAccountMeta(t, l)

	err := l.GetAccountMetas(func(am *types.AccountMeta) error {
		t.Log(am)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_CountAccountMetas(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	addAccountMeta(t, l)
	num, err := l.CountAccountMetas()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", num)
}

func TestLedger_HasTokenMeta_False(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	token2 := mock.TokenMeta(token.BelongTo)
	has, err := l.HasTokenMeta(token.BelongTo, token2.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", has)
}

func TestLedger_HasTokenMeta_True(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	r, err := l.HasTokenMeta(token.BelongTo, token.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", r)
}

func addAccountMetaCache(t *testing.T, l *Ledger) *types.AccountMeta {
	ac := mock.Account()
	am := mock.AccountMeta(ac.Address())
	if err := l.AddAccountMetaCache(am); err != nil {
		t.Fatal()
	}
	return am
}

func TestLedger_AddAccountMetaCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addAccountMetaCache(t, l)
}

func TestLedger_GetAccountMetaCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	am := addAccountMetaCache(t, l)
	a, err := l.GetAccountMetaCache(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", a)
	for _, token := range a.Tokens {
		t.Log("token,", token)
	}

	var count int
	err = l.GetAccountMetaCaches(func(am *types.AccountMeta) error {
		count++
		return nil
	})
	if err != nil || count != 1 {
		t.Fatal(err)
	}
}

func TestLedger_HasAccountMetaCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMetaCache(t, l)
	r, err := l.HasAccountMetaCache(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	if r == false {
		t.Fatal("should have accountMeta from block cache")
	}
	t.Log("has account,", r)
}

func TestLedger_DeleteAccountMetaCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMetaCache(t, l)
	err := l.DeleteAccountMetaCache(am.Address)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_UpdateAccountMetaCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMetaCache(t, l)
	token := mock.TokenMeta(am.Address)
	am.Tokens = append(am.Tokens, token)

	err := l.UpdateAccountMetaCache(am)
	if err != nil {
		t.Fatal(err)
	}
	err = l.AddOrUpdateAccountMetaCache(am)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_Latest(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	token := mock.TokenMeta(block.GetAddress())
	token.Header = block.GetHash()
	token.Type = block.GetToken()
	ac := types.AccountMeta{Address: token.BelongTo, Tokens: []*types.TokenMeta{token}}
	if err := l.AddAccountMeta(&ac); err != nil {
		t.Fatal()
	}

	hash := l.Latest(ac.Address, token.Type)

	if hash != block.GetHash() {
		t.Fatal("err")
	}
}

func TestLedger_Account(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	token := mock.TokenMeta(block.GetAddress())
	token.Type = block.GetToken()
	token2 := mock.TokenMeta(block.GetAddress())
	token2.Type = block.GetToken()
	ac := types.AccountMeta{Address: token.BelongTo, Tokens: []*types.TokenMeta{token, token2}}
	if err := l.AddAccountMeta(&ac); err != nil {
		t.Fatal()
	}

	am, err := l.Account(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(am.Tokens))
}

func TestLedger_Token(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	token := mock.TokenMeta(block.GetAddress())
	token.Type = block.GetToken()
	ac := types.AccountMeta{Address: token.BelongTo, Tokens: []*types.TokenMeta{token}}
	if err := l.AddAccountMeta(&ac); err != nil {
		t.Fatal()
	}

	tm, err := l.Token(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*tm)
}

func TestLedger_Balance(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	am := addAccountMeta(t, l)
	balances, err := l.Balance(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range balances {
		t.Log(k, v)
	}
}

func TestLedger_TokenBalance(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm := addTokenMeta(t, l)
	balance, err := l.TokenBalance(tm.BelongTo, tm.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(balance)
}
