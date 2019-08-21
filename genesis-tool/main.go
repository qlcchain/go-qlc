package main

import (
	"encoding/hex"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/merkle"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/util/hexutil"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/trie"
	"time"
)

type genesisConfig struct {
	netID      string
	addrHex    string
	priKeyHex  string
	targetBits uint32
	time       uint32
}

var testnetCfg = genesisConfig{
	netID:      "testnet",
	addrHex:    "qlc_176f1aj1361y5i4yu8ccyp8xphjcbxmmu4ryh4jecnsncse1eiud7uncz8bj",
	priKeyHex:  "fc41f6fcee50f13eab858588735f811b95b4298f7859e251a96d5a8ed23c8e75148d022200901e1c05ed994af58ddb3e2a4f673d8b1e78a2c55334565806436b",
	targetBits: common.PovGenesisPowBits,
	time:       uint32(1566345600), // 2019/8/21 8:0:0
}

var mainnetCfg = genesisConfig{
	netID:      "mainnet",
	addrHex:    "qlc_1afnoj8acwikgsazz1ocgakss3hck6htgxcm4wafuit7439izg8kzdu6twou",
	priKeyHex:  "ddd5f2aaa7c9fab11bf6bffd734a7339ca59cf80c72dac3a4553b1598f33068921b4ac4c8572127651ff82aa72259c85ea911fa775531710ddc345104f0fb8d2",
	targetBits: common.PovGenesisPowBits,
	time:       uint32(1569024000), // 2019/9/21 8:0:0
}

func makeGenesisPovBlock(cfg *genesisConfig) {
	fmt.Printf("================================================================\n")
	fmt.Printf("genesisConfig: %+v\n", cfg)

	address, err := types.HexToAddress(cfg.addrHex)
	if err != nil {
		panic(err)
	}
	priKey, err := hexutil.Decode("0x" + cfg.priKeyHex)
	if err != nil {
		panic(err)
	}
	account := types.NewAccount(priKey)

	if account.Address() != address {
		fmt.Printf("priKeyHex's address %s != addrHex %s", account.Address(), cfg.addrHex)
		return
	}

	targetInt := types.CompactToBig(cfg.targetBits)
	fmt.Printf("target %d/%s\n", targetInt.BitLen(), targetInt.Text(16))

	targetSig := types.Signature{}
	targetSig.FromBigInt(targetInt)

	stateTrie := trie.NewTrie(nil, nil, nil)
	stateKeys, stateValues := common.GenesisPovStateKVs()
	for i := range stateKeys {
		fmt.Printf("%d key %s value %s\n", i, hex.EncodeToString(stateKeys[i]), hex.EncodeToString(stateValues[i]))
		stateTrie.SetValue(stateKeys[i], stateValues[i])
	}
	stateHash := stateTrie.Hash()

	genesisBlock := &types.PovBlock{}
	genesisBlock.Header.BasHdr.Timestamp = cfg.time
	genesisBlock.Header.BasHdr.Bits = cfg.targetBits

	genesisBlock.Header.CbTx.TxNum = 1
	genesisBlock.Header.CbTx.StateHash = *stateHash
	genesisBlock.Header.CbTx.Reward = common.PovMinerRewardPerBlockBalance
	genesisBlock.Header.CbTx.CoinBase = account.Address()
	genesisBlock.Header.CbTx.Hash = genesisBlock.Header.CbTx.ComputeHash()

	cbtxSig := ed25519.Sign(priKey, genesisBlock.Header.CbTx.Hash[:])
	copy(genesisBlock.Header.CbTx.Signature[:], cbtxSig)

	cbtxHash := &types.PovTransaction{Hash: genesisBlock.Header.CbTx.Hash}
	genesisBlock.Body.Txs = append(genesisBlock.Body.Txs, cbtxHash)

	var txHashes []*types.Hash
	txHashes = append(txHashes, &cbtxHash.Hash)
	genesisBlock.Header.BasHdr.MerkleRoot = merkle.CalcMerkleTreeRootHash(txHashes)

	startTm := time.Now()
	for i := uint32(0); i <= common.PovMaxNonce; i++ {
		genesisBlock.Header.BasHdr.Nonce = uint32(i)

		powHash := genesisBlock.Header.ComputePowHash()
		chkTarget := powHash.ToBigInt()

		if chkTarget.Cmp(targetInt) > 0 {
			continue
		}

		break
	}
	usedTime := time.Since(startTm)

	genesisBlock.Header.BasHdr.Hash = genesisBlock.ComputeHash()

	fmt.Printf("usedTime: %s\n", usedTime)
	fmt.Printf("genesisBlock: %s\n", util.ToString(genesisBlock))
}

func checkTarget() {
	fmt.Printf("GenesisPow: %d/%s/0x%x\n", common.PovGenesisPowInt.BitLen(), common.PovGenesisPowInt.Text(16), common.PovGenesisPowBits)
	fmt.Printf("PowLimit: %d/%s/0x%x\n", common.PovPowLimitInt.BitLen(), common.PovPowLimitInt.Text(16), common.PovPowLimitBits)

	vlsPowBits := uint32(0x1e0ffff0)
	vlsPowInt := types.CompactToBig(vlsPowBits)
	fmt.Printf("vlsPow: %d/%s/0x%x\n", vlsPowInt.BitLen(), vlsPowInt.Text(16), vlsPowBits)
}

func main() {
	checkTarget()

	// go run -tags "testnet" main.go
	//makeGenesisPovBlock(&testnetCfg)

	// go run -tags "mainnet" main.go
	makeGenesisPovBlock(&mainnetCfg)
}
