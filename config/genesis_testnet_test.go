// +build testnet

package config

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
)

func TestGenesisInfo(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	manager := NewCfgManager(configDir)
	_, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}

	// test genesis info
	h1, _ := types.NewHash("5594c690c3618a170a77d2696688f908efec4da2b94363fcb96749516307031d")
	genesisBlock := GenesisBlock()
	h2 := genesisBlock.GetHash()
	if h2 != h1 {
		t.Log(util.ToString(genesisBlock))
		t.Fatal("invalid genesis block", h2.String(), h1.String())
	}
	if !IsGenesisBlock(&genesisBlock) {
		t.Fatal("this block should be genesis block: ", h2.String())
	}

	genesisAddress := GenesisAddress()
	addr, _ := types.HexToAddress("qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4")
	if genesisAddress != addr {
		t.Fatal("invalid genesis address", genesisAddress.String(), addr.String())
	}

	h3, _ := types.NewHash("8b54787c668dddd4f22ad64a8b0d241810871b9a52a989eb97670f345ad5dc90")
	genesisMintageBlock := GenesisMintageBlock()
	h4 := genesisMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(genesisMintageBlock))
		t.Fatal("invalid genesis mintage block", h3.String(), h4.String())
	}
	if !IsGenesisBlock(&genesisBlock) {
		t.Fatal("this block should be genesis block: ", h4.String())
	}

	h1, _ = types.NewHash("a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8580")
	b1 := IsGenesisToken(h1)
	if b1 {
		t.Fatal("h1 should not be Genesis Token")
	}
	h2, _ = types.NewHash("89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4")
	b2 := IsGenesisToken(h2)
	if !b2 {
		t.Fatal("h2 should be Genesis Token")
	}

	// test gas info
	h1, _ = types.NewHash("424b367da2e0ff991d3086f599ce26547b80ae948b209f1cb7d63e19231ab213")
	gasBlock := GasBlock()
	h2 = gasBlock.GetHash()
	if h2 != h1 {
		t.Log(util.ToString(gasBlock))
		t.Fatal("invalid gas block", h2.String(), h1.String())
	}

	if !IsGenesisBlock(&gasBlock) {
		t.Fatal("this block should be genesis block: ", h2.String())
	}

	gasAddress := GasAddress()
	addr, _ = types.HexToAddress("qlc_3t1mwnf8u4oyn7wc7wuptnsfz83wsbrubs8hdhgkty56xrrez4x7fcttk5f3")
	if gasAddress != addr {
		t.Fatal("invalid genesis address", genesisAddress.String(), addr.String())
	}

	h3, _ = types.NewHash("f798089896ffdf45ccce2e039666014b8c666ea0f47f0df4ee7e73b49dac0945")
	gasMintageBlock := GasMintageBlock()
	h4 = gasMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(gasMintageBlock))
		t.Fatal("invalid gas mintage block", h3.String(), h4.String())
	}
	if !IsGenesisBlock(&gasMintageBlock) {
		t.Fatal("this block should be genesis block: ", h4.String())
	}
}
