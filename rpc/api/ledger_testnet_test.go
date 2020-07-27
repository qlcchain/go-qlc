// +build testnet

package api

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestLedgerAPI_GenesisInfo(t *testing.T) {
	teardownTestCase, _, ledgerApi := setupDefaultLedgerAPI(t)
	defer teardownTestCase(t)
	genesisAddress, _ := types.HexToAddress("qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4")
	gasAddress, _ := types.HexToAddress("qlc_3t1mwnf8u4oyn7wc7wuptnsfz83wsbrubs8hdhgkty56xrrez4x7fcttk5f3")
	addr1 := ledgerApi.GenesisAddress()
	if addr1 != genesisAddress {
		t.Fatalf("get genesis address error,should be %s,but get %s", genesisAddress.String(), addr1.String())
	}
	addr2 := ledgerApi.GasAddress()
	if addr2 != gasAddress {
		t.Fatalf("get genesis address error,should be %s,but get %s", gasAddress.String(), addr2.String())
	}
	var chainToken, gasToken types.Hash
	_ = chainToken.Of("a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582")
	_ = gasToken.Of("89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4")
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
