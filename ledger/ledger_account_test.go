package ledger

import (
	"math/big"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func addAccountMeta(t *testing.T, l *Ledger) *types.AccountMeta {
	ac := mock.Account()
	am := mock.AccountMeta(ac.Address())
	if err := l.AddAccountMeta(am, l.cache.GetCache()); err != nil {
		t.Fatal()
	}
	return am
}

func TestLedger_AddAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	a, err := l.GetAccountMetaConfirmed(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	amount := types.Balance{Int: big.NewInt(50)}
	a.CoinVote = amount
	a.Tokens[0].Balance = amount
	b, err := l.GetAccountMetaConfirmed(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	if b.CoinVote.Equal(amount) {
		t.Fatal()
	}
	for _, tm := range b.Tokens {
		if tm.Type == a.Tokens[0].Type && tm.Balance.Equal(amount) {
			t.Fatal()
		}
	}
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

	// get account from confirmed
	am := addAccountMeta(t, l)
	a, err := l.GetAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", a)
	for _, token := range a.Tokens {
		t.Log("token,", token)
	}

	// get account from unConfirmed
	am2 := mock.AccountMeta(mock.Address())
	if err := l.AddAccountMetaCache(am2); err != nil {
		t.Fatal()
	}
	if _, err = l.GetAccountMeta(am2.Address); err != nil {
		t.Fatal(err)
	}

	// get account from confirmed and unConfirmed
	am3 := mock.AccountMeta(mock.Address())
	tm3 := mock.TokenMeta(am3.Address)
	tm3.Type = config.ChainToken()
	am3.Tokens = append(am3.Tokens, tm3)
	if err := l.AddAccountMetaCache(am3); err != nil {
		t.Fatal()
	}
	tm3.BlockCount++
	if err := l.AddAccountMeta(am3, l.cache.GetCache()); err != nil {
		t.Fatal()
	}
	if _, err = l.GetAccountMeta(am3.Address); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_HasAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	am2 := addAccountMeta(t, l)
	if err := l.DeleteAccountMeta(am2.Address, l.cache.GetCache()); err != nil {
		t.Fatal()
	}

	tests := []struct {
		name       string
		args       types.Address
		wantReturn bool
		wantErr    bool
	}{
		{
			name:       "f1",
			args:       am.Address,
			wantReturn: true,
			wantErr:    false,
		},
		{
			name:       "f2",
			args:       mock.Address(),
			wantReturn: false,
			wantErr:    false,
		},
		{
			name:       "f3",
			args:       am2.Address,
			wantReturn: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := l.HasAccountMetaConfirmed(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasAccountMetaConfirmed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if r != tt.wantReturn {
				t.Errorf("HasAccountMetaConfirmed() value = %v, want %v", r, tt.wantReturn)
			}
		})
	}
}

func TestLedger_DeleteAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	err := l.DeleteAccountMeta(am.Address, l.cache.GetCache())
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

	err := l.UpdateAccountMeta(am, l.cache.GetCache())
	if err != nil {
		t.Fatal(err)
	}
}

func addTokenMeta(t *testing.T, l *Ledger) *types.TokenMeta {
	tm := addAccountMeta(t, l)
	token := mock.TokenMeta(tm.Address)
	if err := l.AddTokenMetaConfirmed(token.BelongTo, token, l.cache.GetCache()); err != nil {
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

func TestLedger_DelTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	token := addTokenMeta(t, l)
	err := l.DeleteTokenMetaConfirmed(token.BelongTo, token.Type, l.cache.GetCache())
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
	has, _ := l.HasTokenMeta(token.BelongTo, token2.Type)
	if has {
		t.Fatal()
	}
	t.Log("has token,", has)
}

func TestLedger_HasTokenMeta_True(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	r, _ := l.HasTokenMeta(token.BelongTo, token.Type)
	if !r {
		t.Fatal()
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
	a, err := l.GetAccountMeteCache(am.Address)
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
	r, _ := l.HasAccountMetaCache(am.Address)
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
	a, err := l.GetAccountMeteCache(am.Address)
	if err != nil {
		t.Fatal(err)
	}

	amount1 := types.Balance{Int: big.NewInt(101)}
	am.CoinBalance = amount1
	err = l.UpdateAccountMeteCache(am)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := l.GetAccountMeteCache(am.Address)
	am2, _ := l.GetAccountMeteCache(am.Address)
	if !b.CoinBalance.Equal(amount1) || !am2.CoinBalance.Equal(amount1) {
		t.Fatal("amount not equal")
	}

	token := mock.TokenMeta(am.Address)
	am.Tokens = append(am.Tokens, token)
	err = l.AddOrUpdateAccountMetaCache(am)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := l.GetAccountMeteCache(am.Address)
	if len(a.Tokens)+1 != len(c.Tokens) {
		t.Fatal(len(a.Tokens), len(c.Tokens))
	}
}

func TestLedger_DeleteTokenMetaCache(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMetaCache(t, l)
	if _, err := l.GetTokenMeta(am.Address, am.Tokens[0].Type); err != nil {
		t.Fatal(err)
	}
	err := l.DeleteTokenMetaCache(am.Address, am.Tokens[0].Type)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.GetTokenMeta(am.Address, am.Tokens[0].Type); err == nil {
		t.Fatal(err)
	}
}

func TestLedger_Weight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ac := addRepresentationWeight(t, l)
	r := l.Weight(ac.Address)
	t.Log(r)
}

func TestLedger_CalculateAmount(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	//open
	blk := mock.StateBlockWithoutWork()
	r, err := l.CalculateAmount(blk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
