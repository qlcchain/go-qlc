// +build !testnet

package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"testing"
)

func TestLedgerAPI_GenesisInfo1(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	genesisAddress, _ := types.HexToAddress("qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44")
	gasAddress, _ := types.HexToAddress("qlc_3qe19joxq85rnff5wj5ybp6djqtheqqetfgqc3iogxagnjq4rrbmbp1ews7d")
	addr1 := ledgerApi.GenesisAddress()
	if addr1 != genesisAddress {
		t.Fatalf("get genesis address error,should be %s,but get %s", genesisAddress.String(), addr1.String())
	}
	addr2 := ledgerApi.GasAddress()
	if addr2 != gasAddress {
		t.Fatalf("get genesis address error,should be %s,but get %s", gasAddress.String(), addr2.String())
	}
	var chainToken, gasToken types.Hash
	_ = chainToken.Of("45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad")
	_ = gasToken.Of("ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81")
	token1 := ledgerApi.ChainToken()
	token2 := ledgerApi.GasToken()
	if token1 != chainToken || token2 != gasToken {
		t.Fatal("get chain token or gas token error")
	}
	blk1 := ledgerApi.GenesisMintageBlock()
	blk2 := ledgerApi.GenesisBlock()
	blk3 := ledgerApi.GasBlock()
	if blk1.GetHash() != ledgerApi.GenesisMintageHash() || blk2.GetHash() != ledgerApi.GenesisBlockHash() || blk3.GetHash() != ledgerApi.GasBlockHash() {
		t.Fatal("get genesis block hash error")
	}
	if !ledgerApi.IsGenesisToken(chainToken) {
		t.Fatal("chain token error")
	}
	bs := ledgerApi.AllGenesisBlocks()
	if len(bs) != 4 {
		t.Fatal("get all genesis block error")
	}
	if !ledgerApi.IsGenesisBlock(&bs[0]) {
		t.Fatal("genesis block error")
	}
}
