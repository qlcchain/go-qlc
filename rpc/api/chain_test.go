package api

import (
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/mock/mocks"
	"testing"
)

func setupMockChainAPI(t *testing.T) (func(t *testing.T), *mocks.Store, *ChainApi) {
	l := new(mocks.Store)
	chainApi := NewChainApi(l)
	return func(t *testing.T) {

	}, l, chainApi
}

func TestChainApi_LedgerSize(t *testing.T) {
	teardownTestCase, l, chainApi := setupMockChainAPI(t)
	defer teardownTestCase(t)

	l.On("Action", storage.Size, 0).Return(map[string]int64{
		"lsm":   100000,
		"vlog":  100000,
		"total": 200000,
	}, nil)
	r, err := chainApi.LedgerSize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestChainApi_Version(t *testing.T) {
	teardownTestCase, _, chainApi := setupMockChainAPI(t)
	defer teardownTestCase(t)
	r, err := chainApi.Version()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
