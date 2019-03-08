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
	"github.com/qlcchain/go-qlc/common/util"
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

		if _, ok := tokenCache.LoadOrStore(hash, types.TokenInfo{
			TokenId: hash, TokenName: tokenNames[i], TokenSymbol: strings.ToUpper(tokenSymbols[i]),
			Owner: smartContractBlocks[i].InternalAccount, Decimals: uint8(8), TotalSupply: genesisBlocks[i].Balance.Int,
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
    "token": "cfb64601dee031fc045a2880ea0b8b4823c4f0ce9241d245a012d40910137536",
    "balance": "60000000000000000",
    "link": "d5ba6c7bb3f4f6545e08b03d6da1258840e0395080378a890601991a2a9e3163",
    "representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "type": "Open",
    "address": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"sender": "",
	"receiver": "",
	"data": "",
	"quota": 0,
	"timestamp": 0,
	"message": "0000000000000000000000000000000000000000000000000000000000000000",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000000e408",
    "signature": "64404de19abc9e978e1a0186c3a1a498ba3e9ad367375240e6620f86893095ba69b6e09e27a56a481b7cb1d327010a77d40d86662252ac0020d89acda31e7209"
  },
  {
    "token": "822d045630542776a79c1cc84f4c3a0ae7bec75c6507a265e4fb4050f9745812",
    "balance": "70000000000000000",
    "link": "6f6b4becca470084032121e7923a668da0bbc3ad1ef0c46513cdcf67080bc9b0",
    "representative": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "type": "Open",
    "address": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
	"sender": "",
	"receiver": "",
	"data": "",
	"quota": 0,
	"timestamp": 0,
	"message": "0000000000000000000000000000000000000000000000000000000000000000",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000c662c2",
    "signature": "28ccec0faf5dc737dee1fa4ff603afb1ad3f0f9e7fdbf827be4cd8ae0612beb25f6a2b6ec83da2c296f0ed91a23ca9f004a44afa4934cf9654897b0e8acf2900"
  },
  {
    "token": "d58e691b12dadbb2b4f0918293fbd0bdf10e5f5612e6de7924d1a4a61385a11c",
    "balance": "50000000000000000",
    "link": "e4087493c2572dd52ab4ab67ae91c72dff25ca2efececec8205ff1764d75ddbb",
    "representative": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "type": "Open",
    "address": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
	"sender": "",
	"receiver": "",
	"data": "",
	"quota": 0,
	"timestamp": 0,
	"message": "0000000000000000000000000000000000000000000000000000000000000000",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000ce1618",
    "signature": "cfd16cf088093cf5a7d7cf178d9beff1d59ebf56582aa6fc0cf529c3c376b1cd260f2e483f2ddcb9644188d97034ac3be39cc4a9dd5284659f3d2a846cded90e"
  },
  {
    "token": "3a3ac14c452b69ab9cfcf4338fd8b5f7134435c8fa0c1420cb2aa427b1591960",
    "balance": "90000000000000000",
    "link": "2f51df2104bae5a0feaaaa575f9ae458ac7353dfdf86f393a8d068bb5bcea95d",
    "representative": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "type": "Open",
    "address": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
	"sender": "",
	"receiver": "",
	"data": "",
	"quota": 0,
	"timestamp": 0,
	"message": "0000000000000000000000000000000000000000000000000000000000000000",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000075bd2e",
    "signature": "6ad03d54d498f8294661ed348bda82e42fd58d10b3f2aae8f958c5719049fb5dfd625affb225fb43564f47383d923165f67ce7071ec4174d69da86981876c600"
  },
  {
    "token": "fb5c54408d07cdcff5c237f0f08409bc73c8665d5d2696a98774fab0350e5254",
    "balance": "80000000000000000",
    "link": "eee2111fbc1a25c3afb02760b3fdd499d20ca507dec145e2e7f9c07a1d9b9b30",
    "representative": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "type": "Open",
    "address": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
	"sender": "",
	"receiver": "",
	"data": "",
	"quota": 0,
	"timestamp": 0,
	"message": "0000000000000000000000000000000000000000000000000000000000000000",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000008027a6",
    "signature": "00ba9d4291f9d0d653ea7e22464bc7d5b9161f887ec5391ebdd1e2993f343ea729b9b58ba37a32485b9b150d5a4e5dfb9df8a7af7ce44ffeed59b00173aae009"
  },
  {
    "token": "4dbc7ee15cb22213555ed1b3cf835d0318f0a3f801333cc96c91c47dae1aa40e",
    "balance": "90000000000000000",
    "link": "7b6f0909957f2d5b25a01836b1527599dfd4453705397e7c89967dee27e15286",
    "representative": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "type": "Open",
    "address": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
	"sender": "",
	"receiver": "",
	"data": "",
	"quota": 0,
	"timestamp": 0,
	"message": "0000000000000000000000000000000000000000000000000000000000000000",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009fc06f",
    "signature": "161cf2568a89cc6ca23ba43e8a9b85c341c41b729eaf4e7b5f78a05904253cfc32e24bbd46d9b6fcbe81cd78c4c714225452e8a20e360411981fb5f23689ee0c"
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
    "owner": "qlc_1nawsw4yatupd47p3scd5x5i3s9szbsggxbxmfy56f8jroyu945i5seu1cdd",
	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_3watpnwym9i43kbkt35yfp8xnqo7c9ujp3b6udajza71mspjfzpnpdgoydzn",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000007bb1fe",
    "signature": "d9d71c82eccdca0324e102c089b28c1430b0ae61f2af809e6134b289d5186b16cbcb6fcd4bfc1424fd34aa40e9bdd05069bc56d05fecf833470d80d047048a05"
  },
  {
    "internalAccount": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "contract": {
      "abi": "VtblZz9y4XqhlMgp+U4ut9TKy9B6jPXm2/Ko4nM74zP3sWI9DQ/Cco0c+/gE8LRS8UFQQQt0o3KXOHlgQ0tWwy5mwanERfP3cmIpsbqW270VIMivbhmY",
      "abiLength": 87,
      "abiHash": "bdc16324e2450ca21293df43e16ffd658e8c486d633521d595f135e23b68ddf1"
    },
    "owner": "qlc_188wtseurfgqrtsnph6jpd71jqdyb94oqj8zm8x1ger3qemy15g6168xmkyo",
 	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_16npf7a86tucn7ei9x1zkccwzqoethz57ajhgt6kfp66hdnht66qqbjhskjq",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000006831fa",
    "signature": "6aa248d86286a2b5b86220345393c22ee044e4df658b84a68e314d9ed4ae4583d824047c2f57089f3ad0fd47fee094f875c218203086b9a31dce677a3c03ba02"
  },
  {
    "internalAccount": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "contract": {
      "abi": "RGNfcfPoQm06RsUvxiEc0LtoCv8Jouy/3J+qGesDyU5QLUaathpqCMtxnYppkUM=",
      "abiLength": 47,
      "abiHash": "e3b27d19ad41ec75971a0910967a88d368f589e9a85d3e2d1f3b7db1de2426c2"
    },
    "owner": "qlc_1cchdoqsiqs1z6jsmp318pd647cdkg8gsb51mghoo74csmsccrzz8m381wsh",
 	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_3mm5zqhnzug3ymmmbpjdmpbc4rgqzwt3aqbnpc69qhunbn1fo6a5168i8uzk",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000004d1eca",
    "signature": "2030955bb654311a23c43ad9e7cf740c1d0533c08c657bf1bd5836db8d0950fa2cce8a4052fbe849c6db2de8a1348565a08eafd9174c77d918d6fd098b44210f"
  },
  {
    "internalAccount": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "contract": {
      "abi": "yVbs6CUx73/YKh9LcQryGE2gxOUhso2ewhfV8V+sT6VzJAoUAekcZ3LgfAt7KeZ0dO7LBHuz+u2L+sM=",
      "abiLength": 59,
      "abiHash": "6985746f764e467a959ced7f25139978e8ee57bcfaff3bc14bd51ad47aab29ee"
    },
    "owner": "qlc_3m176fthnjx4pqrkufty7fot8kcmt9ugkmb9cagpec7b3iwaoxzuuzihh8px",
 	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_19adrdpewyhsrd9q7hhjc869uqsajyt66qt5dr64rqbtwiresbq4odkh4yit",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009d763f",
    "signature": "029e5279928c317b23ba6af41e9e5330cbdfc4b7fd272e94921d0b475fa1591516122c1a19a7af2f93f9f15bc70c77d781792f02633d884d609473fdb81d8905"
  },
  {
    "internalAccount": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "contract": {
      "abi": "NeeipE1zH/sWQLjPqbzcwSFk+AWivOda/OK5FtY82LRz5U6KONz1EMRUsubM1AKU4HasJuCwWvxHw+oujQ7pwZhaVtJ4hxAc36HJULadXJjy",
      "abiLength": 81,
      "abiHash": "b4fcdc674cc456fa24b87cc98209f3edd514515233ae8e6c575511839d2f92f2"
    },
    "owner": "qlc_31nrezhfzgrka9ae1k9ymn9fx4dwpjntdz6x3fss7h58n3ro1tepudnfmxyo",
 	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_1td67f5da5yy3d4qi9fjhpwzubw6hi4du7i8fmhg964664dsnfdpfmqabnra",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000004d6775",
    "signature": "c5e0a359a38ca35046431740ccf746ff4e726fb8a5495fb1290975820c840f30bb939a0184520e0620684e86e37835a627bf9b45b83d047eba269279c8888705"
  },
  {
    "internalAccount": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "contract": {
      "abi": "0jFRlXPYHl2GcEXM7vbUxGp7",
      "abiLength": 18,
      "abiHash": "89fd3c99a475442d27fb8409aaed67ce6e545f757b1abc8375cd69c0a65f9f32"
    },
    "owner": "qlc_3ucju4ug747fye8r9zyasyap64imekmidjfg7ga83g1gk17dbfs3m1gafefa",
 	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_1y74rcprdspfijga3myfh9868p8xjm14k1f75qndzwe1otkwcftw7ypjq88o",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000084273b",
    "signature": "6132b3f7d2b4807724e258e4f194e8dd24572cb347ed2806cb35cfafc1635df1676e04b291c4868c6dd1e1ed5ae42cd23b079264b047a28ea03c078adab91c0c"
  }
]
`
	decimals = []*big.Int{
		util.BigPow(10, 0),
		util.BigPow(10, 3),
		util.BigPow(10, 8),
		util.BigPow(10, 11),
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

func GetTokenById(tokenId types.Hash) (types.TokenInfo, error) {
	if v, ok := tokenCache.Load(tokenId); ok {
		return v.(types.TokenInfo), nil
	}

	return types.TokenInfo{}, fmt.Errorf("can not find token info by id(%s)", tokenId.String())
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

func GetSmartContracts() []*types.SmartContractBlock {
	size := len(smartContractBlocks)
	b := make([]*types.SmartContractBlock, size)
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

func Tokens() ([]*types.TokenInfo, error) {
	var tis []*types.TokenInfo
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

func GetTokenByName(tokenName string) (*types.TokenInfo, error) {
	var info types.TokenInfo
	tokenCache.Range(func(key, value interface{}) bool {
		i := value.(types.TokenInfo)
		if i.TokenName == tokenName {
			info = i
		}
		return true
	})
	return &info, nil
}

func StateBlock() *types.StateBlock {
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

	return sb
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

	b0 := createBlock(types.Open, *ac1, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(100000000000))}, types.Hash(ac1.Address()), ac1.Address()) // genesis
	blocks = append(blocks, b0)

	b1 := createBlock(types.Send, *ac1, b0.GetHash(), token, types.Balance{Int: big.NewInt(int64(40000000000))}, types.Hash(ac2.Address()), ac1.Address()) //a1 send
	blocks = append(blocks, b1)

	b2 := createBlock(types.Open, *ac2, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(60000000000))}, b1.GetHash(), ac1.Address()) //a2 open
	blocks = append(blocks, b2)

	b3 := createBlock(types.Change, *ac2, b2.GetHash(), token, types.Balance{Int: big.NewInt(int64(60000000000))}, types.ZeroHash, ac2.Address()) //a2 change
	blocks = append(blocks, b3)

	b4 := createBlock(types.Send, *ac2, b3.GetHash(), token, types.Balance{Int: big.NewInt(int64(30000000000))}, types.Hash(ac1.Address()), ac2.Address()) //a2 send
	blocks = append(blocks, b4)

	b5 := createBlock(types.Receive, *ac1, b1.GetHash(), token, types.Balance{Int: big.NewInt(int64(70000000000))}, b4.GetHash(), ac1.Address()) //a1 receive
	blocks = append(blocks, b5)

	//b6 := createBlock(*ac2, b4.GetHash(), token, types.Balance{Int: big.NewInt(int64(20000000000))}, types.Hash(ac3.Address()), ac2.Address()) //a2 send
	//blocks = append(blocks, b6)
	//
	//b7 := createBlock(*ac3, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(10000000000))}, b6.GetHash(), ac1.Address()) //a3 open
	//blocks = append(blocks, b7)
	//
	//token2 := Hash()
	//b8 := createBlock(*ac3, types.ZeroHash, token2, types.Balance{Int: big.NewInt(int64(1000000000000))}, types.Hash(ac3.Address()), ac1.Address()) //new token
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

func createBlock(t types.BlockType, ac types.Account, pre types.Hash, token types.Hash, balance types.Balance, link types.Hash, rep types.Address) *types.StateBlock {
	blk := new(types.StateBlock)
	blk.Type = t
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
