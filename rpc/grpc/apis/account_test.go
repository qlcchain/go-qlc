package apis

import (
	"context"

	"github.com/qlcchain/go-qlc/mock"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	"testing"
)

func newTestAccountApi() *AccountApi {
	return NewAccountApi()
}

func TestAccountApi_Create(t *testing.T) {
	seed := "af45bea06bc957bef3a52c118fc576f4d35e6ea0c022efd68642a05b3340c3db"
	api := newTestAccountApi()
	var i uint32
	i = 0
	m, err := api.Create(context.Background(), &pb.CreateRequest{
		SeedStr: seed,
		Index:   i,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m)
}

func TestAccountApi_ForPublicKey(t *testing.T) {
	publickKey := "1676b313888e55547583c01c6f500ad547edf4ef5d2c9bb06e663d5481d9aad6"
	api := newTestAccountApi()
	addr, err := api.ForPublicKey(context.Background(), toString(publickKey))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr)
	if addr.GetAddress() != "qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf" {
		t.Fatal()
	}
}

func TestAccountApi_PublicKey(t *testing.T) {
	addr := mock.Address()
	api := newTestAccountApi()
	addrStr, err := api.PublicKey(context.Background(), &pbtypes.Address{
		Address: addr.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addrStr)
}

func TestAccountApi_NewSeed(t *testing.T) {
	api := newTestAccountApi()
	s, err := api.NewSeed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s.GetValue())
}

func TestAccountApi_NewAccounts(t *testing.T) {
	api := newTestAccountApi()
	r, err := api.NewAccounts(context.Background(), nil)
	if err != nil || len(r.GetAccounts()) != 10 {
		t.Fatal(err, len(r.GetAccounts()))
	}
}

func TestAccountApi_Validate(t *testing.T) {
	api := newTestAccountApi()
	addr := "qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf"
	b, err := api.Validate(context.Background(), toString(addr))
	if err != nil {
		t.Fatal(err)
	}
	if !b.GetValue() {
		t.Fatal(b)
	}
}
