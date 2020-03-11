// +build !testnet

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
	h1, _ := types.NewHash("7201b4c283b7a32e88ec4c5867198da574de1718eb18c7f95ee8ef733c0b5609")
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
	addr, _ := types.HexToAddress("qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44")
	if genesisAddress != addr {
		t.Fatal("invalid genesis address", genesisAddress.String(), addr.String())
	}

	h3, _ := types.NewHash("c0d330096ec4ab6ccf5481e06cc54e74b14f534e99e38df486f47d1123cbd1ae")
	genesisMintageBlock := GenesisMintageBlock()
	h4 := genesisMintageBlock.GetHash()
	if h3 != h4 {
		t.Log(util.ToIndentString(genesisMintageBlock))
		t.Fatal("invalid genesis mintage block", h3.String(), h4.String())
	}
	if !IsGenesisBlock(&genesisBlock) {
		t.Fatal("this block should be genesis block: ", h4.String())
	}

	h1, _ = types.NewHash("327531148b1a6302632aa7ad6eb369437d8269a08a55b344bd06b514e4e6ae97")
	b1 := IsGenesisToken(h1)
	if b1 {
		t.Fatal("h1 should not be Genesis Token")
	}
	h2, _ = types.NewHash("45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad")
	b2 := IsGenesisToken(h2)
	if !b2 {
		t.Fatal("h2 should be Genesis Token")
	}

	// test gas info
	h1, _ = types.NewHash("b9e2ea2e4310c38ed82ff492cb83229b4361d89f9c47ebbd6653ddec8a07ebe1")
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
	addr, _ = types.HexToAddress("qlc_3qe19joxq85rnff5wj5ybp6djqtheqqetfgqc3iogxagnjq4rrbmbp1ews7d")
	if gasAddress != addr {
		t.Fatal("invalid genesis address", genesisAddress.String(), addr.String())
	}

	h3, _ = types.NewHash("bdac41b3ff7ac35aee3028d60eabeb9578ea6f7bd148d611133a3b26dfa6a9be")
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
