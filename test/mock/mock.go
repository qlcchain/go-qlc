/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mock

import (
	"fmt"
	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/log"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"sync"
	"time"
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
    "token": "705d1a62cdc6a7c35ff3a0ba59b8683e5e33818abfde93c98009d721762a9d52",
    "balance": "00000000000000060000000000000000",
    "link": "d5ba6c7bb3f4f6545e08b03d6da1258840e0395080378a890601991a2a9e3163",
    "representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "type": "State",
    "address": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000000e408",
    "signature": "fb481e2aac799c6ec4d6f53a525277d185ab84373225199152fba3cc63ef4d067d5c5ab22842d4ffc8f6b88f34f859884a303a32e95146b54429019eddf1fd03"
  },
  {
    "token": "8d827353240b5edb8de3abb4ca8988efcd777b0b2fa6498e6e2d6a37b21d83e5",
    "balance": "00000000000000070000000000000000",
    "link": "6f6b4becca470084032121e7923a668da0bbc3ad1ef0c46513cdcf67080bc9b0",
    "representative": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "type": "State",
    "address": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000c662c2",
    "signature": "e4d28c5646f41ef7d31381ae08120c8cc420989c8256add7f92415e444bf5c461aae0ab045a8a08211bd081e6a3245ba74d8d77f5df37d1b7496fdd20f283a0d"
  },
  {
    "token": "592ddd5f8cbd0e5074b1185e03a8558eeed5c8b6ebe578419ec3ac566c12e2ae",
    "balance": "00000000000000050000000000000000",
    "link": "e4087493c2572dd52ab4ab67ae91c72dff25ca2efececec8205ff1764d75ddbb",
    "representative": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "type": "State",
    "address": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000ce1618",
    "signature": "228a38f7e8c9b4ec72b5635e76ab5a25ea2680879eef2c37bfc507b7b878f2e75b605e376a754d569ffef8c74ba0593e16ec7b41eea3adccf4fcce88cdc56e04"
  },
  {
    "token": "ae21ba8acc1828e6c47943f1121386cb5ead014a366fb3f10fec585fe0641a23",
    "balance": "00000000000000090000000000000000",
    "link": "2f51df2104bae5a0feaaaa575f9ae458ac7353dfdf86f393a8d068bb5bcea95d",
    "representative": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "type": "State",
    "address": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000075bd2e",
    "signature": "6f9c85728d3d29cf09975d62ab93978e74b0c20d9ba0e9e26370a4d60a0217ff7e77a50b8cd4277e9f341e667f310b591a6b3507ec5f59972dd180c94389040d"
  },
  {
    "token": "f93e94e35e5992ac847ee1e586d88eacd1c30a9ba84fc1da02d4362681f1907f",
    "balance": "00000000000000080000000000000000",
    "link": "eee2111fbc1a25c3afb02760b3fdd499d20ca507dec145e2e7f9c07a1d9b9b30",
    "representative": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "type": "State",
    "address": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000008027a6",
    "signature": "4394b1ad3d429976ab2500da186006e428019ea71363bddc39b7afbfa33a82d7619b92296fe5537fe99b970e0db0857b68ea0f8ca416438c554eaa806f19fc0e"
  },
  {
    "token": "7f3909f36647ae2520c5ac0077ee0133f0485a25df7b58a44a001447b4e98b3a",
    "balance": "00000000000000090000000000000000",
    "link": "7b6f0909957f2d5b25a01836b1527599dfd4453705397e7c89967dee27e15286",
    "representative": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "type": "State",
    "address": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009fc06f",
    "signature": "3d9d6310b1dcb678976a2fd3ecd14ef5fffbeeea9d82447bbd821b222b9146f157844833a72e0a0f37a36699b3e92a6a555e2692c10050cfc9f38ac29ab74706"
  }
]
`
	scBlockstring = `[
  {
    "internalAccount": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "contract": {
      "abi": "VBPLA0t2DwbasyBqQYLnb8RZ1iVHgUyxK3t5xwbX7ePkC3ozIScdQERrEn5XX5FgYC/32hy70SUyzIEGRn+lv4NlGwIKxHVW3Ikp0vwfC3eN",
      "abiLength": 81,
      "abiHash": "5cc0e822f0c26404b375f85375688e4d5cc2de7c743fc7fca55d90ffaaab2721"
    },
    "issuer": "qlc_1tss8rbi5w11q17umcoxauuitrr5wd76pjkf9eu71c4f3uqw4t7n7p6rxpkd",
    "type": "SmartContract",
    "address": "qlc_36bir9qr5iswkhzn45dn6c1gxjpyqxwqj7qfodz4j11i53bdwrxorepkm8dk",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000003519ef",
    "signature": "eabc576d2ab01101aacc56849218ac40d15df38540da4258426c302d77a4e4980d44378b7cd81b88fdf140a5fff57c881977ab72b66a19b77d1bc50517307207"
  },
  {
    "internalAccount": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "contract": {
      "abi": "2u2Iw74dnuQRkc4t9Dhv7B7VJ2EHjhW6MsMv9esZ2BogvFKd4/dRi0jdhAfAYT7gxSjCon3g5eFyt/Ig7cWfvJugq30OwhlryhZFMRKAbREg3jUnY7rf",
      "abiLength": 87,
      "abiHash": "06cc5b829332033c98a2e4f39dbec71d1e37a8fe08d45d399fbf2fb25ee1cacc"
    },
    "issuer": "qlc_1ntwb1t4ddyzo8m9chjsso8hzs83di66jim6deun67n5emfanutby9x57ww1",
    "type": "SmartContract",
    "address": "qlc_1mswgm6opoefuh6dpktwe4j68cogbmbusktehexze3iata4qerzhwkwjzr9e",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000001284498",
    "signature": "6a32eb6c30595a8a589f7ada93d8dbf1bedf2c8025e762cb3a67765df341cb97378cd3b0737abee5df3518377737d4e2e3c4d7838edbfda020151cfd0c82cd07"
  },
  {
    "internalAccount": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "contract": {
      "abi": "DR/N92U5L6D+N0WEavnYc1m91wWZkT7qY9I31mKn0vHntkQ/sBFplEknkWrTPuA=",
      "abiLength": 47,
      "abiHash": "d2dad8fa925bbb4a07474c0a5aa9229a0587f9f2cb9effe07858885990a0760b"
    },
    "issuer": "qlc_3j9saxnba5t1i5sfd418qexpq8sdf3s6a4hihp1owe7j3agiy3puwioq5mw8",
    "type": "SmartContract",
    "address": "qlc_1dafhe75wz3r33kxpe9ycc8krd3xbhaeccz7jfns3uwf5c8o1c9esryxh4h5",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009919a2",
    "signature": "d14eb5b1416177da80500062efaae4780d2e04c826fa24014b28ee5d4a83ad3efacca4d1d6b4bba66619f96561399ab9d0110927b3f6f51e7879bb35d2d1830b"
  },
  {
    "internalAccount": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "contract": {
      "abi": "ZEd0YfvsYNCrmj/TnURy2Fdq07liSMoAK0L/6VAwCsOvTecT8G35Z6CHMIMSP3u9ISoi1qmvGxGaxIU=",
      "abiLength": 59,
      "abiHash": "5f368d2094e52ad626df9064cf5e0f1284a1fb668fadabb930aa19e449d9fd5e"
    },
    "issuer": "qlc_3aupdo1cdjdyh6i9ssy91h41zxbb3kkruq6nh5bewhthgoete7azwd8pb8ox",
    "type": "SmartContract",
    "address": "qlc_1k6upokk5z6rnrpgezebphsc73qs1ofyfiqggtcj48d153yxht3yiyfxhnpy",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000861712",
    "signature": "b8dc494172c75a616a1271b89e0721b32a571ac27935b8d059f59ae5da07adf0df61a347ce0800503c719735878a7a6e59ae9dd350723b8410fb48ba8c34920e"
  },
  {
    "internalAccount": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "contract": {
      "abi": "POleSX9PF2xMhkjXPxGC74Pj3UtelPdVfL2EhZOd/sog8As8YTDTRRZfXZwf4ehc0T1r3M1BPoPUFT/ufgNUtRX0n1s03lMCTnSjZZEli0fE",
      "abiLength": 81,
      "abiHash": "0fedca3707b384e35cf3edb1097f3a2ea5d264c40e16dba2118230e3353b6a27"
    },
    "issuer": "qlc_11za3oeptb4ka8p338peih6j7359gzgdbmooywrp7kfo5it4fdb6ybyrikdf",
    "type": "SmartContract",
    "address": "qlc_3usq3t7gng38b4rd8zbp5hdboypymztifeupurw8wgubnxhr9ampen6sz6ue",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000fabb65",
    "signature": "bb3e7d136b9e4f19c5b7f80189aa2a87cdf41c2a9de402a97381c3874c159ee8cd8ac4bf5756fa3a678d33f185a9731ab8044450f4f8de9e3f770a2f3af56e01"
  },
  {
    "internalAccount": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "contract": {
      "abi": "ldsVbV9SbzRoiiDLU/mxYvSx",
      "abiLength": 18,
      "abiHash": "3de094fb7b46f419ad9f814493932ce86a0186a5c4444dd5a60a9bb9676f8439"
    },
    "issuer": "qlc_1euz8fsw1qi86y6xzj9zjze13656554cxkwk648fkmzojbkerq77ik1aj84p",
    "type": "SmartContract",
    "address": "qlc_1qr3hnh6w76ttqh8obn3hedqj5uhq9ys7ymkskyenbwuxwez5ubk3isw18km",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000f1d99c",
    "signature": "28e0ae8f1486cae3c838c8c03ad607d74cbaeee53b436cdb4b9bca4c1dc7a9ffe98cf1a360adac25c069fe3c091804d5898ff32e86a3f461c206ef75f75bfd00"
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
	s2, _ := random.Intn(math.MaxInt64)
	t := types.TokenMeta{
		//TokenAccount: MockAddress(),
		Type:           MockHash(),
		BelongTo:       addr,
		Balance:        types.ParseBalanceInts(uint64(s1), uint64(s2)),
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

// ParseBalance parses the given balance string.
func ParseBalance(s string, unit string) (types.Balance, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return types.ZeroBalance, err
	}

	// zero is a special case
	if d.Equals(decimal.Zero) {
		return types.ZeroBalance, nil
	}

	d = d.Mul(units[unit])
	c := d.Coefficient()
	f := bigPow(10, int64(d.Exponent()))
	i := c.Mul(c, f)

	bytes := i.Bytes()
	balanceBytes := make([]byte, types.BalanceSize)
	copy(balanceBytes[len(balanceBytes)-len(bytes):], bytes)

	var balance types.Balance
	if err := balance.UnmarshalBinary(balanceBytes); err != nil {
		return types.ZeroBalance, err
	}

	return balance, nil
}

func bigPow(base int64, exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(base), big.NewInt(exp), nil)
}

// UnitString returns a decimal representation of this uint128 converted to the
// given unit.
func UnitString(b types.Balance, unit string, precision int32) string {
	d := decimal.NewFromBigInt(b.BigInt(), 0)
	return d.DivRound(units[unit], types.BalanceMaxPrecision).Truncate(precision).String()
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
