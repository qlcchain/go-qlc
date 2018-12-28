/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mock

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/log"
	"github.com/shopspring/decimal"
)

func init() {
	err := jsoniter.Unmarshal([]byte(genesisBlockstring), &genesisBlocks)
	if err != nil {
		logger.Error(err)
	}

	err = jsoniter.Unmarshal([]byte(scBlockstring), &smartContractBlocks)
	if err != nil {
		logger.Error(err)
	}

	for i := range genesisBlocks {
		block := smartContractBlocks[i]
		hash := block.GetHash()
		if i == 0 {
			chainTokenType = hash
		}

		if _, ok := tokenCache.LoadOrStore(hash, TokenInfo{
			TokenId: hash, TokenName: tokenNames[i], TokenSymbol: tokenSymbols[i],
			Owner: smartContractBlocks[i].InternalAccount, Decimals: uint8(8), TotalSupply: genesisBlocks[i].Balance,
		}); !ok {
			//logger.Debugf("add token[%s] to cache", hash.String())
		}
	}
}

var (
	chainTokenType      types.Hash
	tokenNames          = []string{"QLC", "QN1", "QN2", "QN3", "QN4", "QN5"}
	tokenSymbols        = []string{"qlc", "qn1", "qn2", "qn3", "qn4", "qn5"}
	logger              = log.NewLogger("Mock")
	genesisBlocks       []types.StateBlock
	smartContractBlocks []types.SmartContractBlock
	tokenCache          = sync.Map{}
	genesisBlockstring  = `[
  {
    "token": "4627e2a0c1d68238bb4f848c59f4e18288c36fb7c959d220c9914728db890de8",
    "balance": "600000000",
    "link": "d5ba6c7bb3f4f6545e08b03d6da1258840e0395080378a890601991a2a9e3163",
    "representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "type": "State",
    "address": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000000e408",
    "signature": "c4f00f9558db30c59be7bfd40abffd16eb6d2e2700a12d15fb73e83e5de7c4439284c57fb24f7127fc3eb9f42f3a1e026469a4da3fa24ddb64b8924c5fb5b804"
  },
  {
    "token": "a5d898709a685f1940b33a3e2819c3079240669025af3876d159b20ed59c1c69",
    "balance": "700000000",
    "link": "6f6b4becca470084032121e7923a668da0bbc3ad1ef0c46513cdcf67080bc9b0",
    "representative": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "type": "State",
    "address": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000c662c2",
    "signature": "667bcb0be84eb4ea9c1ffe35d30f29c16f50445c8acd44006d179d05db26a627f7718d951bf41b2268fd2a4de00fc83a9db9f676b477d9f76974e60444a4cd0e"
  },
  {
    "token": "08177e04124d29af243d30dfe1e52821ac6441f9ca9cc1fc11a2c1a4f748619a",
    "balance": "500000000",
    "link": "e4087493c2572dd52ab4ab67ae91c72dff25ca2efececec8205ff1764d75ddbb",
    "representative": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "type": "State",
    "address": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000ce1618",
    "signature": "0cf6b28965c18ae4b5a9d9c41aaf220bbd15031790389790322c3add3126b57bb5c7008e5b38e911f00b0cfad3e2b371661132f08ca741ecbc4b317f95703903"
  },
  {
    "token": "7ec06dc236636cb15ccca4da8473463dca1b5e09d10f1bb82d962576a107f90c",
    "balance": "900000000",
    "link": "2f51df2104bae5a0feaaaa575f9ae458ac7353dfdf86f393a8d068bb5bcea95d",
    "representative": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "type": "State",
    "address": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000075bd2e",
    "signature": "0ef7bca15ae3a1250f9e2891f042b26689ba737914e093849bf053814473d938acb640bee8c3754ae45852bff39d3e83a7557c69ff3d5243532b73acdf739a0b"
  },
  {
    "token": "54da86f58c3196a9fbd4f05b452c11589eeef5e7f4ba688d97b49df978f6adea",
    "balance": "800000000",
    "link": "eee2111fbc1a25c3afb02760b3fdd499d20ca507dec145e2e7f9c07a1d9b9b30",
    "representative": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "type": "State",
    "address": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000008027a6",
    "signature": "34b7aa18266f321502d1a2c6fc839321ed51b61404ea661b305b085741f3bd1492334bcebd8598106b1b2e3184818ed62c6e971285d8e58bd8cf0787c35d9507"
  },
  {
    "token": "9fc49a7899e29518e335e72fdc97663e8469d318981e3b2d96d22d72dc39ec72",
    "balance": "900000000",
    "link": "7b6f0909957f2d5b25a01836b1527599dfd4453705397e7c89967dee27e15286",
    "representative": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "type": "State",
    "address": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009fc06f",
    "signature": "ad68d0592619695dd70b38fca5961563cd8e2a6842c98291d6349290e10cb4b5ee8783c5b2181ad4b75247d961e7ff73dea808a25e139379847ac0b319d17d09"
  }
]
`
	scBlockstring = `[
  {
    "internalAccount": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "contract": {
      "abi": "TAacHjHVOyrJesALAUU720QZSl0Mx2fynH0wuOZWObmGL5K47quVdNZ/mG5dB33LxHGUiHJvcUj+LqJLLXAZwAZAdfz5hOdLJmH0n13V5FiD",
      "abiLength": 81,
      "abiHash": "c175ebefdca681861e2102fcced29eaa9332d79b10daa2a28551c57b099d0dd1"
    },
    "issuer": "qlc_3wkuticdgwxp36oonh4fm953cbkhz1ufmdraphodkn5aouqomoamt9trbxww",
    "type": "SmartContract",
    "address": "qlc_3n47dh3n3maq3fygctz5e9wmcezk4qsjg16w7yn3zowcn6z3mg95hjnkq4ku",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000133602",
    "signature": "e4ee3679d522b06d20956e2e4377bad41d8c413028958b13b0dbf8b67eb2a091db845e4a2458775cc7e9029012328eada9c29d8d91c25416d84214a35f2faf01"
  },
  {
    "internalAccount": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "contract": {
      "abi": "XBcf6qN2bLrKdYgOSrsRc4hYIp3QqX/LqLOfKTLw0YRqwMh7IurVirLtNKstNWlGT475kLbdrLVjPy46TdD2OMuhvAuss7D6eyONzCfBKNN3YMgnfXxO",
      "abiLength": 87,
      "abiHash": "0f2fec3bf1b07f17551a24a34ae5b5b932aead04bcfde1511e75e58ce9d15db0"
    },
    "issuer": "qlc_3s8tydowm7ce6xc4iwubqsc5qxokefxzu5iuk35xieknqxn7pak7ufkb3h1u",
    "type": "SmartContract",
    "address": "qlc_1rq4hd91x8k9oqei5jt811tyaaho7undj5x3ki76epd5oruek9pwdfm3r8t9",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000001212a60",
    "signature": "4a2fc8ddb36abe5f046bcf47efa794632d89c746c307970195d262ec22f7ee01fc4cc408b37ef6eae050852865a193e12fab99f56e76b31fef5ffdacd7939f08"
  },
  {
    "internalAccount": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "contract": {
      "abi": "vQuBj/ZTFa830VuX2jq1WDtFvzj8DlpJoHsdtAzVPxc5onHAw4yII0yipsuZNDs=",
      "abiLength": 47,
      "abiHash": "07acc3715dafde2efb12fee92d67e131903aed48fa13a710f7f6b6962f460062"
    },
    "issuer": "qlc_3u83toxzi5o3ohxtzycbnkorf9i96d6hkikydnk48ob4zen6fosruz6mqzi7",
    "type": "SmartContract",
    "address": "qlc_1hdpob4up9hjzhurzqirpubg74sfsch11ghytajdjo8h8maunog3szjgw9te",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000245418",
    "signature": "dcce8c6a78634d790eedb557cefeb89c4f97503b096b6cbc67d86c48cafc17d8028b7c89e034400c1e013bcfbb0a5d570e4023eccac9507f3a249f07a8949400"
  },
  {
    "internalAccount": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "contract": {
      "abi": "NrgEDMTkMefaXrlQ761kB7+gddN2nRvgs9VDW6Eyxhj/NotFjEsXEdINk96v+KUtnlkJ0ctfm2Yvfj4=",
      "abiLength": 59,
      "abiHash": "8a5f5c624431adeeb2ab131de4ae15fdbf5a68a74140a98cfd6a01896d7be4ea"
    },
    "issuer": "qlc_16o7g3muhar87nw5onsk7mk7kxqxbqtj354ocp7gnzpxjmt39fm9n37i76pp",
    "type": "SmartContract",
    "address": "qlc_18cowq4jmyqw7hk7kb77er8gfn5i64p6kixep3hweo1mh5ji67reayc65sqe",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000156bbef",
    "signature": "ba2c571cca4f35ed9e6c6091d6bdc55642406af78f27a273f765c0d1d747ab85a95ae41d266941566cc044c09e85fd9c821c7223f0aa9ff6ab10dd02c4a80502"
  },
  {
    "internalAccount": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "contract": {
      "abi": "23zYqMrthgG9s63cXr7pCIsoZZbtXMHLMHQey4GmMtKOzYlnDdLB2NIhBl4qDL8hKupH+gNRKHjDGx3NS1Yf7u13ewx5Wp/uhiZ8qIDRvkwm",
      "abiLength": 81,
      "abiHash": "1d0b2a3567a840a5268a5dd6869cd2b01b37c2da0e23d297b38b7a1799c79684"
    },
    "issuer": "qlc_1hsjckbhikcwuardss1y7wpugcb57mezhqtoecoh3p8dfwqc5y5mot8k8wfo",
    "type": "SmartContract",
    "address": "qlc_19i6etepwcxaejdagwaofqsff6n9nch4hj7go73t3ufk7b8swmrk5c1c1fjm",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000022b7c6",
    "signature": "dc8431761ab9123097cf35e3c5b0b9adfcfb11b291473dee4228c1c7c3984fcf667866550f4fc74922fa08cf3dcc5a6efcb784b7364cfbdfd61c5076321b820e"
  },
  {
    "internalAccount": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "contract": {
      "abi": "sFt2joRwc02+gV1f+agS+P8W",
      "abiLength": 18,
      "abiHash": "01d548def8e2b35588dc5c9e319763df4189dc913e1c1d880612010d84680275"
    },
    "issuer": "qlc_3e9ootmpr5dd6o376y6e4wtf7ieqa1819f778zyqu3o13hp6o7pffjsghaj9",
    "type": "SmartContract",
    "address": "qlc_3bugapghmk6w9o3ayby6cky5tffaipbtcdfxohntgq31z5hy7qqwur7t81t8",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000228426",
    "signature": "13b91a4d0519b7ce07f690d9358190f7909089cef4136fad8dd8231ac3e8a9e32a9a52b3e7eeb4217f409a8e524db3b18428a88afab8b4c0d7d08e8b9fe08f0c"
  }
]
`
	units = map[string]decimal.Decimal{
		"raw":  decimal.New(1, 0),
		"qlc":  decimal.New(1, 2),
		"Kqlc": decimal.New(1, 5),
		"Mqlc": decimal.New(1, 8),
		"Gqlc": decimal.New(1, 11),
	}
)

type TokenInfo struct {
	TokenId     types.Hash    `json:"tokenId"`
	TokenName   string        `json:"tokenName"`
	TokenSymbol string        `json:"tokenSymbol"`
	TotalSupply types.Balance `json:"totalSupply"`
	Decimals    uint8         `json:"decimals"`
	Owner       types.Address `json:"owner"`
}

func MockHash() types.Hash {
	h := types.Hash{}
	_ = random.Bytes(h[:])
	return h
}

func MockAccountMeta(addr types.Address) *types.AccountMeta {
	var am types.AccountMeta
	am.Address = addr
	am.Tokens = []*types.TokenMeta{}
	for i := 0; i < 5; i++ {
		t := MockTokenMeta(addr)
		am.Tokens = append(am.Tokens, t)
	}
	return &am
}

func MockTokenMeta(addr types.Address) *types.TokenMeta {
	s1, _ := random.Intn(math.MaxInt64)
	i := new(big.Int).SetInt64(int64(s1))
	t := types.TokenMeta{
		//TokenAccount: MockAddress(),
		Type:           MockHash(),
		BelongTo:       addr,
		Balance:        types.Balance{Int: i},
		BlockCount:     1,
		OpenBlock:      MockHash(),
		Header:         MockHash(),
		Representative: MockAddress(),
		Modified:       time.Now().Unix(),
	}

	return &t
}

func MockAddress() types.Address {
	address, _, _ := types.GenerateAddress()

	return address
}

func GetTokenById(tokenId types.Hash) (TokenInfo, error) {
	if v, ok := tokenCache.Load(tokenId); ok {
		return v.(TokenInfo), nil
	}

	return TokenInfo{}, fmt.Errorf("can not find token info by id(%s)", tokenId.String())
}

func GetChainTokenType() types.Hash {
	return chainTokenType
}

func GetGenesis() []types.Block {
	size := len(genesisBlocks)
	b := make([]types.Block, size)
	for i := 0; i < size; i++ {
		b[i] = &genesisBlocks[i]
	}
	return b
}

func GetSmartContracts() []types.Block {
	size := len(smartContractBlocks)
	b := make([]types.Block, size)
	for i := 0; i < size; i++ {
		b[i] = &smartContractBlocks[i]
	}
	return b
}

func Account() *types.Account {
	seed, _ := types.NewSeed()
	_, priv, _ := types.KeypairFromSeed(seed.String(), 0)
	return types.NewAccount(priv)
}

func StateBlock() types.Block {
	a := Account()
	b, _ := types.NewBlock(types.State)
	sb := b.(*types.StateBlock)
	i, _ := random.Intn(math.MaxUint32)
	sb.Type = types.State
	sb.Balance = types.Balance{Int: big.NewInt(int64(i))}
	sb.Address = MockAddress()
	sb.Token = chainTokenType
	sb.Previous = MockHash()
	sb.Representative = genesisBlocks[0].Representative
	addr := MockAddress()
	sb.Link = addr.ToHash()
	sb.Signature = a.Sign(sb.GetHash())
	var w types.Work
	worker, _ := types.NewWorker(w, sb.Root())

	sb.Work = worker.NewWork()

	return b
}

func MockBlockChain() ([]types.Block, error) {
	var blocks []types.Block
	ac1 := Account()
	ac2 := Account()
	//ac3 := Account()

	token := GetChainTokenType()

	b0 := createBlock(*ac1, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(10000000))}, types.Hash(ac1.Address()), ac1.Address()) // genesis
	blocks = append(blocks, b0)

	b1 := createBlock(*ac1, b0.GetHash(), token, types.Balance{Int: big.NewInt(int64(4000000))}, types.Hash(ac2.Address()), ac1.Address()) //a1 send
	blocks = append(blocks, b1)

	b2 := createBlock(*ac2, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(6000000))}, b1.GetHash(), ac1.Address()) //a2 open
	blocks = append(blocks, b2)

	b3 := createBlock(*ac2, b2.GetHash(), token, types.Balance{Int: big.NewInt(int64(6000000))}, types.ZeroHash, ac2.Address()) //a2 change
	blocks = append(blocks, b3)

	b4 := createBlock(*ac2, b3.GetHash(), token, types.Balance{Int: big.NewInt(int64(3000000))}, types.Hash(ac1.Address()), ac2.Address()) //a2 send
	blocks = append(blocks, b4)

	b5 := createBlock(*ac1, b1.GetHash(), token, types.Balance{Int: big.NewInt(int64(7000000))}, b4.GetHash(), ac1.Address()) //a1 receive
	blocks = append(blocks, b5)

	//b6 := createBlock(*ac2, b4.GetHash(), token, types.Balance{Int: big.NewInt(int64(2000000))}, types.Hash(ac3.Address()), ac2.Address()) //a2 send
	//blocks = append(blocks, b6)
	//
	//b7 := createBlock(*ac3, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(1000000))}, b6.GetHash(), ac1.Address()) //a3 open
	//blocks = append(blocks, b7)
	//
	//token2 := MockHash()
	//b8 := createBlock(*ac3, types.ZeroHash, token2, types.Balance{Int: big.NewInt(int64(100000000))}, types.Hash(ac3.Address()), ac1.Address()) //new token
	//blocks = append(blocks, b8)

	return blocks, nil
}

func createBlock(ac types.Account, pre types.Hash, token types.Hash, balance types.Balance, link types.Hash, rep types.Address) *types.StateBlock {
	blk := new(types.StateBlock)
	blk.Type = types.State
	blk.Address = ac.Address()
	blk.Previous = pre
	blk.Token = token
	blk.Balance = balance
	blk.Link = link
	blk.Representative = rep
	blk.Signature = ac.Sign(blk.GetHash())
	var w types.Work
	worker, _ := types.NewWorker(w, blk.Root())
	blk.Work = worker.NewWork()
	return blk
}
