/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/log"
)

func init() {
	err := json.Unmarshal([]byte(genesisBlockstring), &genesisBlocks)
	if err != nil {
		logger.Error(err)
	}

	err = json.Unmarshal([]byte(scBlockstring), &smartContractBlocks)
	if err != nil {
		logger.Error(err)
	}

	for i := range genesisBlocks {
		block := smartContractBlocks[i]
		hash := block.GetHash()
		if i == 0 {
			chainTokenType = hash
		}

		// fill unit map
		symbol := tokenSymbols[i]
		for i, d := range decimals {
			s := symbol
			if i > 1 {
				s = strings.ToUpper(symbol)
			}
			units[key[i]+s] = d
		}

		if _, ok := tokenCache.LoadOrStore(hash, TokenInfo{
			TokenId: hash, TokenName: tokenNames[i], TokenSymbol: strings.ToUpper(tokenSymbols[i]),
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
    "token": "9bf0dd78eb52f56cf698990d7d3e4f0827de858f6bdabc7713c869482abfd914",
    "balance": "60000000000000000",
    "link": "d5ba6c7bb3f4f6545e08b03d6da1258840e0395080378a890601991a2a9e3163",
    "representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "type": "State",
    "address": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000000e408",
    "signature": "d33ab1facf29f636c482ca7bfb402181bc7cf6fe0bbe1426b3283d4bc6d35ac06a1976d45fd7e3f0607f91bf2d9b2fd3cfd1d96d24bf167bf19eb13247c1190f"
  },
  {
    "token": "02acb4a0e87c1fdd09b794a317ccaa6eddcb143c5fa0139f6fe145e9d94fcb17",
    "balance": "70000000000000000",
    "link": "6f6b4becca470084032121e7923a668da0bbc3ad1ef0c46513cdcf67080bc9b0",
    "representative": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "type": "State",
    "address": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000c662c2",
    "signature": "b27f45042dae3918c5baf69ddb3631aeb8e1076f96c46721f0f36fd4f760cc28774ee23b9dd3e6dfa18d4db6de410bd91d471a3ae22e17d6998da9727aa6b903"
  },
  {
    "token": "3b32f29885c04d9931319a8a564692880f68d1310a4ed527dd65bfd87896e8e4",
    "balance": "50000000000000000",
    "link": "e4087493c2572dd52ab4ab67ae91c72dff25ca2efececec8205ff1764d75ddbb",
    "representative": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "type": "State",
    "address": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000ce1618",
    "signature": "f3f310bf1957bb7eacb738f3b7fc497b3c6973e2e77a693b9c3b5c591766fbfe32c7401ca4bd9a82f193ba454a096f9583d1189e3bea693f964fe742b533070b"
  },
  {
    "token": "6b50aa569c41d3a4efdfbc2138590fb39c556db4a39deb589c3f8b55530ce074",
    "balance": "90000000000000000",
    "link": "2f51df2104bae5a0feaaaa575f9ae458ac7353dfdf86f393a8d068bb5bcea95d",
    "representative": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "type": "State",
    "address": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000075bd2e",
    "signature": "cfe00b7135d6a116bde9faa5ea07bb43bb05190034ec5790cc4e727c4fed78e621268aee9153ddca9598c005b622e6359d85d4d14e6b51583f7648894bce3703"
  },
  {
    "token": "f8a2f79c397875d86cdf9342fc060449a95843cbcf46d2cfcac34f7489660604",
    "balance": "80000000000000000",
    "link": "eee2111fbc1a25c3afb02760b3fdd499d20ca507dec145e2e7f9c07a1d9b9b30",
    "representative": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "type": "State",
    "address": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000008027a6",
    "signature": "13bc561710ae620190f05b17b03d1ae9d1242bccefa7e569850050fa755fd59a01c7e6252c025c74b360ea72e02280ea53eeeed242c990cb76b404be9ae68b05"
  },
  {
    "token": "b9cd84de89deb22305779c39d9a94461ce7b31bf91bf9d4f631d09c1881d18b2",
    "balance": "90000000000000000",
    "link": "7b6f0909957f2d5b25a01836b1527599dfd4453705397e7c89967dee27e15286",
    "representative": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "type": "State",
    "address": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009fc06f",
    "signature": "cafd215d4c3b6f5c71da503587f03b4ac5d94d3c6d015ae1d0d8a86e765e2d0e7c8ef2acf70355569790d18a8947a7242644d3d189c66f447fe240395d1ed803"
  }
]
`
	scBlockstring = `[
  {
    "internalAccount": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "contract": {
      "abi": "mcvnzY+zF5mVDjsvknvPfFgRToMQAVI4wivQGRZBwerbUIvfrKD6/suZJWiFVOI5sbTa98bpY9+1cUhE2T9yidxSCpvZ4kkBVBMfcL3OJIqG",
      "abiLength": 81,
      "abiHash": "79dab43dcc97205918b297c3aba6259e3ab1ed7d0779dc78eec6f57e5d6307ce"
    },
    "issuer": "qlc_1nawsw4yatupd47p3scd5x5i3s9szbsggxbxmfy56f8jroyu945i5seu1cdd",
    "type": "SmartContract",
    "address": "qlc_3watpnwym9i43kbkt35yfp8xnqo7c9ujp3b6udajza71mspjfzpnpdgoydzn",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000007bb1fe",
    "signature": "0e5a3f6b246c99f20bbe302c4cdfb4639ef3358620a85dc996ba38b0e2ae463d3e4a1184b36abfa63d3d11f56c3242446d381f837f2fb6f7336cf303a5cbca08"
  },
  {
    "internalAccount": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "contract": {
      "abi": "VtblZz9y4XqhlMgp+U4ut9TKy9B6jPXm2/Ko4nM74zP3sWI9DQ/Cco0c+/gE8LRS8UFQQQt0o3KXOHlgQ0tWwy5mwanERfP3cmIpsbqW270VIMivbhmY",
      "abiLength": 87,
      "abiHash": "bdc16324e2450ca21293df43e16ffd658e8c486d633521d595f135e23b68ddf1"
    },
    "issuer": "qlc_188wtseurfgqrtsnph6jpd71jqdyb94oqj8zm8x1ger3qemy15g6168xmkyo",
    "type": "SmartContract",
    "address": "qlc_16npf7a86tucn7ei9x1zkccwzqoethz57ajhgt6kfp66hdnht66qqbjhskjq",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000006831fa",
    "signature": "8205c069cf253089880eafa5bb4c676531139553bd9687e673fc1f8f1bb2b73b4a9b55a5cad342a7d17f840b4238f533cdb3507ee0c8ae9745da5d2e83bb6309"
  },
  {
    "internalAccount": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "contract": {
      "abi": "RGNfcfPoQm06RsUvxiEc0LtoCv8Jouy/3J+qGesDyU5QLUaathpqCMtxnYppkUM=",
      "abiLength": 47,
      "abiHash": "e3b27d19ad41ec75971a0910967a88d368f589e9a85d3e2d1f3b7db1de2426c2"
    },
    "issuer": "qlc_1cchdoqsiqs1z6jsmp318pd647cdkg8gsb51mghoo74csmsccrzz8m381wsh",
    "type": "SmartContract",
    "address": "qlc_3mm5zqhnzug3ymmmbpjdmpbc4rgqzwt3aqbnpc69qhunbn1fo6a5168i8uzk",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000004d1eca",
    "signature": "eebd4cfa391f8f9fd1e1af278d844397c6a61d3df26bf9100c7c6f3c9fd4c458c1580089490b869bc3443220adb166298528bab41930fe157e0c7d6ffc7e6806"
  },
  {
    "internalAccount": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "contract": {
      "abi": "yVbs6CUx73/YKh9LcQryGE2gxOUhso2ewhfV8V+sT6VzJAoUAekcZ3LgfAt7KeZ0dO7LBHuz+u2L+sM=",
      "abiLength": 59,
      "abiHash": "6985746f764e467a959ced7f25139978e8ee57bcfaff3bc14bd51ad47aab29ee"
    },
    "issuer": "qlc_3m176fthnjx4pqrkufty7fot8kcmt9ugkmb9cagpec7b3iwaoxzuuzihh8px",
    "type": "SmartContract",
    "address": "qlc_19adrdpewyhsrd9q7hhjc869uqsajyt66qt5dr64rqbtwiresbq4odkh4yit",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009d763f",
    "signature": "ee24e6a845b877b4326ec102d8f759f11f066952eef2d983c38cefcb43f065ce68c67462060bda3dad0a704c517afe2c43839ecd80d8be99b50b8c6dbf36d00b"
  },
  {
    "internalAccount": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "contract": {
      "abi": "NeeipE1zH/sWQLjPqbzcwSFk+AWivOda/OK5FtY82LRz5U6KONz1EMRUsubM1AKU4HasJuCwWvxHw+oujQ7pwZhaVtJ4hxAc36HJULadXJjy",
      "abiLength": 81,
      "abiHash": "b4fcdc674cc456fa24b87cc98209f3edd514515233ae8e6c575511839d2f92f2"
    },
    "issuer": "qlc_31nrezhfzgrka9ae1k9ymn9fx4dwpjntdz6x3fss7h58n3ro1tepudnfmxyo",
    "type": "SmartContract",
    "address": "qlc_1td67f5da5yy3d4qi9fjhpwzubw6hi4du7i8fmhg964664dsnfdpfmqabnra",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000004d6775",
    "signature": "de745f3169b2f610fbc993fc20cc96ffa76b60a3d8843288f4e69e36140088c2b75970130ce99ed72614a6b994efedf42c8f05a67c2f6257d6f1de03ef3c6204"
  },
  {
    "internalAccount": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "contract": {
      "abi": "0jFRlXPYHl2GcEXM7vbUxGp7",
      "abiLength": 18,
      "abiHash": "89fd3c99a475442d27fb8409aaed67ce6e545f757b1abc8375cd69c0a65f9f32"
    },
    "issuer": "qlc_3ucju4ug747fye8r9zyasyap64imekmidjfg7ga83g1gk17dbfs3m1gafefa",
    "type": "SmartContract",
    "address": "qlc_1y74rcprdspfijga3myfh9868p8xjm14k1f75qndzwe1otkwcftw7ypjq88o",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000084273b",
    "signature": "84c86c2ccd62d542c22aed78d9a1da3bbcccd693ca69c6c01099ce2c6446cbe93c6d167f5618a4d27d7446021d9c5e3671a06d9972bc813a03cdd422806ae401"
  }
]
`
	decimals = []*big.Int{
		bigPow(10, 0),
		bigPow(10, 3),
		bigPow(10, 8),
		bigPow(10, 11),
	}

	key = []string{"", "k", "", "M"}

	units = map[string]*big.Int{"raw": big.NewInt(1)}
	//units = map[string]decimal.Decimal{
	//	"qlc":  decimal.New(1, 2),
	//	"Kqlc": decimal.New(1, 5),
	//	"Mqlc": decimal.New(1, 8),
	//	"Gqlc": decimal.New(1, 11),
	//}
)

type TokenInfo struct {
	TokenId     types.Hash    `json:"tokenId"`
	TokenName   string        `json:"tokenName"`
	TokenSymbol string        `json:"tokenSymbol"`
	TotalSupply types.Balance `json:"totalSupply"`
	Decimals    uint8         `json:"decimals"`
	Owner       types.Address `json:"owner"`
}

func Hash() types.Hash {
	h := types.Hash{}
	_ = random.Bytes(h[:])
	return h
}

func AccountMeta(addr types.Address) *types.AccountMeta {
	var am types.AccountMeta
	am.Address = addr
	am.Tokens = []*types.TokenMeta{}
	for i := 0; i < 5; i++ {
		t := TokenMeta(addr)
		am.Tokens = append(am.Tokens, t)
	}
	return &am
}

func BalanceToRaw(b types.Balance, unit string) (types.Balance, error) {
	if v, ok := units[unit]; ok {
		//v = v.Div(v, units["raw"])
		return types.Balance{Int: new(big.Int).Mul(b.Int, v)}, nil
	}
	return b, fmt.Errorf("invalid unit %s", unit)
}

func RawToBalance(b types.Balance, unit string) (types.Balance, error) {
	if v, ok := units[unit]; ok {
		//v = v.Div(v, units["raw"])
		return types.Balance{Int: new(big.Int).Div(b.Int, v)}, nil
	}
	return b, fmt.Errorf("invalid unit %s", unit)
}

func bigPow(base int64, exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(base), big.NewInt(exp), nil)
}

func TokenMeta(addr types.Address) *types.TokenMeta {
	s1, _ := random.Intn(math.MaxInt32)
	i := new(big.Int).SetInt64(int64(s1))
	t := types.TokenMeta{
		//TokenAccount: Address(),
		Type:           Hash(),
		BelongTo:       addr,
		Balance:        types.Balance{Int: i},
		BlockCount:     1,
		OpenBlock:      Hash(),
		Header:         Hash(),
		Representative: Address(),
		Modified:       time.Now().Unix(),
	}

	return &t
}

func Address() types.Address {
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

func Tokens() ([]*TokenInfo, error) {
	var tis []*TokenInfo
	scs := GetSmartContracts()
	for _, sc := range scs {
		hash := sc.GetHash()
		ti, err := GetTokenById(hash)
		if err != nil {
			return nil, err
		}
		tis = append(tis, &ti)
	}
	return tis, nil
}

func GetTokenByName(tokenName string) (*TokenInfo, error) {
	var info TokenInfo
	tokenCache.Range(func(key, value interface{}) bool {
		i := value.(TokenInfo)
		if i.TokenName == tokenName {
			info = i
		}
		return true
	})
	return &info, nil
}

func StateBlock() types.Block {
	a := Account()
	b, _ := types.NewBlock(types.State)
	sb := b.(*types.StateBlock)
	i, _ := random.Intn(math.MaxInt16)
	sb.Type = types.State
	sb.Balance = types.Balance{Int: big.NewInt(int64(i))}
	sb.Address = a.Address()
	sb.Token = chainTokenType
	sb.Previous = Hash()
	sb.Representative = genesisBlocks[0].Representative
	addr := Address()
	sb.Link = addr.ToHash()
	sb.Signature = a.Sign(sb.GetHash())
	var w types.Work
	worker, _ := types.NewWorker(w, sb.Root())

	sb.Work = worker.NewWork()

	return b
}

func BlockChain() ([]*types.StateBlock, error) {
	dir := filepath.Join(config.QlcTestDataDir(), "blocks.json")
	_, err := os.Stat(dir)
	if err == nil {
		if f, err := os.Open(dir); err == nil {
			defer f.Close()
			if r, err := ioutil.ReadAll(f); err == nil {
				var blocks []*types.StateBlock
				if err = json.Unmarshal(r, &blocks); err == nil {
					return blocks, nil
				}
			}
		}
	}

	var blocks []*types.StateBlock
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
	//token2 := Hash()
	//b8 := createBlock(*ac3, types.ZeroHash, token2, types.Balance{Int: big.NewInt(int64(100000000))}, types.Hash(ac3.Address()), ac1.Address()) //new token
	//blocks = append(blocks, b8)

	r, err := json.Marshal(blocks)
	if err != nil {
		return nil, err
	}

	f, error := os.OpenFile(dir, os.O_RDWR|os.O_CREATE, 0766)
	if error != nil {
		fmt.Println(error)
	}
	defer f.Close()
	f.Write(r)
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
