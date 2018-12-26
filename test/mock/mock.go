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
    "token": "82ef3d813f77358376617fc0595f61856b1b9545d1925a9a3993081be3ffb4f6",
    "balance": "00000000000000060000000000000000",
    "link": "d5ba6c7bb3f4f6545e08b03d6da1258840e0395080378a890601991a2a9e3163",
    "representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "type": "State",
    "address": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000000e408",
    "signature": "8c1e43009208a25a45e0bdb12bfeb9d83dbcfef586c93eb608b7b06bc3220f9cd821974ce7e561cb7b82fbc42afc3eced43f3c5c4e77a82c867a3dfe4184ee0f"
  },
  {
    "token": "defdc73413c311f5b107c7cc50169d75868b5cff8fe7e44a752fed241f0ba538",
    "balance": "00000000000000070000000000000000",
    "link": "6f6b4becca470084032121e7923a668da0bbc3ad1ef0c46513cdcf67080bc9b0",
    "representative": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "type": "State",
    "address": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000c662c2",
    "signature": "420b1eafd520bd23a6b6ab9519e32c89ad10e370752f8cea3e198a3af675269a2d3a5cd487e4e5d5e4f24046b2abfbe6f750387605644b6c65f8f45ec38f5b05"
  },
  {
    "token": "9214f0897cb1cf6b9dcb4ea0a90f4b6a12a66c3faf97ad59a3ab2b144514b213",
    "balance": "00000000000000050000000000000000",
    "link": "e4087493c2572dd52ab4ab67ae91c72dff25ca2efececec8205ff1764d75ddbb",
    "representative": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "type": "State",
    "address": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000ce1618",
    "signature": "afec1c6704c6f578a51a7d6b7d4e5e0cc670ce62ce63f2f38683ff46ef4b03dd6bfbf0db158d885d8206ac81f67ba5b96070650cbd1a6caa8e6750f5d70fb906"
  },
  {
    "token": "7fce3ecdfda63b626df2d9eb21e34754a9038f178515de2ebbd14f4321c42247",
    "balance": "00000000000000090000000000000000",
    "link": "2f51df2104bae5a0feaaaa575f9ae458ac7353dfdf86f393a8d068bb5bcea95d",
    "representative": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "type": "State",
    "address": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000075bd2e",
    "signature": "61cf0d9df407483428b4b29702d5ec89fb5b45ef74a29f57dcd2a95396a79532c9520f33c8d0fbf8d615e5071c4ad9420523bb046856b9154fa1f40a3adbd20e"
  },
  {
    "token": "fdb1b95fbbe53c6b781b2b9e2c0c2d69229a0901b1c477d41e8597e33b81eb7c",
    "balance": "00000000000000080000000000000000",
    "link": "eee2111fbc1a25c3afb02760b3fdd499d20ca507dec145e2e7f9c07a1d9b9b30",
    "representative": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "type": "State",
    "address": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000008027a6",
    "signature": "63a74f0a9f1ce4f66dbd6408ceba0e7b35f05f293bba7d1e18467f5ea2bf48423a96f53dc915a7912884d0fd82962c1b4162336759555598ada2061b80f0d709"
  },
  {
    "token": "290dc0db17094a16078cf1bd20bf986425205add5a4f3d40619040cd739a2577",
    "balance": "00000000000000090000000000000000",
    "link": "7b6f0909957f2d5b25a01836b1527599dfd4453705397e7c89967dee27e15286",
    "representative": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "type": "State",
    "address": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000009fc06f",
    "signature": "a573cfb4ddbd5ee44adf37359c86b5cfcfd928969f0f864de65c831880e38774d3173b46e5a53357c8d1f44f7f8039c674f57f1227c2c85c35c93a06c55e9b05"
  }
]
`
	scBlockstring = `[
  {
    "internalAccount": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "contract": {
      "abi": "l1gV/nAejd9GTIy/z5XWKQ+Kqv4EseN4F25lL3VjttV4P+9VZEybQ/p9r63LH3zqtuoperg2ST70KRfPMN1zbFpoTgm1D9CLeLJkdNhsze7g",
      "abiLength": 81,
      "abiHash": "82ef3d813f77358376617fc0595f61856b1b9545d1925a9a3993081be3ffb4f6"
    },
    "issuer": "qlc_3rzein3o8xp5i58j6x4mx89yc63g386irsaw8gwmj8ntdyemwcx9xxqtj1jo",
    "type": "SmartContract",
    "address": "qlc_3tmyaihf3smidedmyswmj5f97qg7sn7baws5g6paa83ebwr3x6x7fxaxyos3",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "000000000006e183",
    "signature": "03440da882decbbf14c4f952e81dbca9458a9a4a0586bbf7351d4b62447e6762e10f28bd607bf50fd1eca3d636835ee0e17aed5b07e6f7480ea7d16eee14c40f"
  },
  {
    "internalAccount": "qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc",
    "contract": {
      "abi": "xWE7RWwwlyXFij3ojIzgK2X2N7pi0uAwFc8qnOn9AV0k6IMEXnwUuuyYRfuW0eoCd4tBzNzYJo4RaYPln3z2ZBc6Zc/3baDJXNw0YgmExKhTo1rVrWFa",
      "abiLength": 87,
      "abiHash": "defdc73413c311f5b107c7cc50169d75868b5cff8fe7e44a752fed241f0ba538"
    },
    "issuer": "qlc_3b8gei99gmz9tbaitdsgt5oxoqp4hhy9s9rwxm7ya4tu6ts9ambe9gkfu5y1",
    "type": "SmartContract",
    "address": "qlc_37ktwpruqfp9paoqody9sgpcjiuysw1seatqk1szxrfbhip71g6c8ma6bwoe",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000043fcf",
    "signature": "aaf00308d73ee1a3af21d6ecb613e341e0376f8ada0f73245ce6b065ae0abc5ac3ac967061ba29c00a5d64cb239a02aad89af73e2cbec92f81466ef4fc314f00"
  },
  {
    "internalAccount": "qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii",
    "contract": {
      "abi": "+R6B3OuGJbXF0v+q+9nyBhoDLFaPNvhacM3iL+FH+jXm/4peY0HXJwtgKcGMciU=",
      "abiLength": 47,
      "abiHash": "9214f0897cb1cf6b9dcb4ea0a90f4b6a12a66c3faf97ad59a3ab2b144514b213"
    },
    "issuer": "qlc_14byoic391zdhrou4xekehwmapdwgeqft5j7au673j41mppj1hd79qc6rnee",
    "type": "SmartContract",
    "address": "qlc_3naw6sef5r5t4wj6netricd19jnk7cagcrhgw3rsxpow8z75of7ae6a8q1h5",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000de8dce",
    "signature": "6a67f73f4bced998b4db9115bb2faa4b7a921b72cf352c79e7fe7de3df57b9b40456a0dcb3adcf37f17c8fe770f3cbe3fa605c10e9a11583d51ff6919215bf05"
  },
  {
    "internalAccount": "qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8",
    "contract": {
      "abi": "9LbBuay4ZUVTQ1MuzRoxOqJyHijiVVUTZhhHcUJ4qrg/mlgbQecDxD46kKLdF4pJBmrTC23pKZYFiss=",
      "abiLength": 59,
      "abiHash": "7fce3ecdfda63b626df2d9eb21e34754a9038f178515de2ebbd14f4321c42247"
    },
    "issuer": "qlc_1t9z84cw31z6881rm8smbch6syh8nisqo9eccqgwnmebsac334pw6jns6tea",
    "type": "SmartContract",
    "address": "qlc_3zyqu6jzmoery17styy44uzyxns1zqbw4zfaatyo3pnby177fz3ix95rjdhw",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000001d10392",
    "signature": "55bc1fa1fc14a43db3ef6421077210fd1ebded714dbaad24b941f9d3afb439aaf71f5799fa4a40b37b6d75c3893a0fba5a5580ddc54778ea70588a943449a107"
  },
  {
    "internalAccount": "qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596",
    "contract": {
      "abi": "klrvpM8ZkwRuHb5T2tG5cXElOPU9AczAWkTAWJeCxuQMfLyYvPUV70KwIi/l+lX6Y55zskKL759/AuNaDKYMbwZ6K26jyvzLh9mymeq/pBdD",
      "abiLength": 81,
      "abiHash": "fdb1b95fbbe53c6b781b2b9e2c0c2d69229a0901b1c477d41e8597e33b81eb7c"
    },
    "issuer": "qlc_14axef43qdd51588af5k9nhzng3fg5jr5ns717bx8atomf8ozqbntad67ymj",
    "type": "SmartContract",
    "address": "qlc_3gp9s5ac3bpfsbranougcogug44myoi7e5fgp5z5gc86ugbp4h6iyxb8yysj",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000002da33e",
    "signature": "7e7c117e49e6461e7a9649f67533a4b61f27345d0a7cf1fd9694854e8ad473e3be88d4a11d8599ca8a2db4b0ea010c6c76e6a5d09e2fc1658d5ed3efa9975408"
  },
  {
    "internalAccount": "qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me",
    "contract": {
      "abi": "OAyFw+7PgRyQrUeJZkSgliYg",
      "abiLength": 18,
      "abiHash": "290dc0db17094a16078cf1bd20bf986425205add5a4f3d40619040cd739a2577"
    },
    "issuer": "qlc_1njcg6ud3bkqgx149ff877o5k3tq5154o6gxdedqpk3opqq5snxhffsww1dx",
    "type": "SmartContract",
    "address": "qlc_3a5p8erck1neawsjp35ap3xibgwk5qqdsxadexnk6gm9z8h7cdps5grujdyk",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000002b0932",
    "signature": "32ec7cc96fe488b4a54dadd3e75fb6165d98916b05d51be7cf366260cec3cbb045c89b567b23c0592c19ba0fcf73056f19ca2d1fa2f14a6fa572ef87a9dcdb00"
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
		Type:       MockHash(),
		BelongTo:   addr,
		Balance:    types.ParseBalanceInts(uint64(s1), uint64(s2)),
		BlockCount: 1,
		OpenBlock:  MockHash(),
		Header:     MockHash(),
		RepBlock:   MockHash(),
		Modified:   time.Now().Unix(),
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
