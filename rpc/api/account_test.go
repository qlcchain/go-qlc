package api

import (
	"testing"

	"github.com/qlcchain/go-qlc/mock"
)

func TestAccountApi_Create(t *testing.T) {
	seed := "af45bea06bc957bef3a52c118fc576f4d35e6ea0c022efd68642a05b3340c3db"
	api := NewAccountApi()
	var i uint32
	i = 0
	m, err := api.Create(seed, &i)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m)
}

func TestAccountApi_ForPublicKey(t *testing.T) {
	publickKey := "1676b313888e55547583c01c6f500ad547edf4ef5d2c9bb06e663d5481d9aad6"
	api := NewAccountApi()
	addr, err := api.ForPublicKey(publickKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr)
	if addr.String() != "qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf" {
		t.Fatal()
	}
}

func TestAccountApi_PublicKey(t *testing.T) {
	addr := mock.Address()
	api := NewAccountApi()
	addrStr := api.PublicKey(addr)
	t.Log(addrStr)
}

func TestAccountApi_NewSeed(t *testing.T) {
	api := NewAccountApi()
	s, err := api.NewSeed()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
}

func TestAccountApi_NewAccounts(t *testing.T) {
	api := NewAccountApi()
	r, err := api.NewAccounts(nil)
	if err != nil || len(r) != 10 {
		t.Fatal(err, len(r))
	}
}

func TestAccountApi_Validate(t *testing.T) {
	api := NewAccountApi()
	b := api.Validate("qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf")
	if !b {
		t.Fatal(b)
	}
}
