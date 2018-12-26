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
    "token": "199a92277a3f8eb3e8ddb2aeafecf6546cae91ba953a576401c0bddf1e9865ca",
    "balance": "00000000000000060000000000000000",
    "link": "3732344132304238334638423644374436464635394435353246343033463936",
    "representative": "qlc_1fsk8j1m6e4491sneg448s45gj3pas55cgc68ntm6jjn81snegbpac5gh7ro",
    "type": "State",
    "address": "qlc_1fsk8j1m6e4491sneg448s45gj3pas55cgc68ntm6jjn81snegbpac5gh7ro",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000495b22",
    "signature": "4595d7bd47e66414beaadfe4d85dc83fedbac20d64d58af6a51481a92cb125138df164d00fd67ab70e31d977c7c1f65326c133857840bb4e99ee88495369cf0f"
  },
  {
    "token": "0aa34c9b43a7445d96da9ce4834372d3b4c20ff063df89cc62550d88a6613089",
    "balance": "00000000000000070000000000000000",
    "link": "3037464436433842353938453332303637323439354543383430333239313044",
    "representative": "qlc_1e3qas45eisraatmkg478es51fjq8at5kfc7aew5ae3m8awm4e46q3ooio3y",
    "type": "State",
    "address": "qlc_1e3qas45eisraatmkg478es51fjq8at5kfc7aew5ae3m8awm4e46q3ooio3y",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000371627",
    "signature": "206beefe26a7b9f945f71727d3cbb129c6344dc3fddebabd7cbde77068dec2cd6d3a4935315b529092c6727e45b60fdaf6662890f5c3cf83909c20879dc12b08"
  },
  {
    "token": "4b542f291def8b48ec66340e98c54b5d3bb110c029297a54455876a6bc6153ed",
    "balance": "00000000000000050000000000000000",
    "link": "3943344231383934384235313636463043464546454245324532443038353631",
    "representative": "qlc_1gc58j354g3s8iw66fbj8ru6ee45as4nejc4ans6cek681w5cfjjsk3c1szw",
    "type": "State",
    "address": "qlc_1gc58j354g3s8iw66fbj8ru6ee45as4nejc4ans6cek681w5cfjjsk3c1szw",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000006b5fac",
    "signature": "a46d8b746ddfd64ba43bddf605454c9bbaca5ea51345ab5c41922cc609f767a152547a558b575029b5d9c38dc2a57391cdf4f5a41a52d76328fe0e36467f9206"
  },
  {
    "token": "59930056d4fae0fe994daf6ea07478bc945d65ac2aba72e78b01f98d749b5b0e",
    "balance": "00000000000000090000000000000000",
    "link": "3142423143463546344244423745363045444345414435414142463232443736",
    "representative": "qlc_1ec4aarn8jjoart66j448x4mee47aj3ncic68o1n4ik88as6afspyb8j7s8z",
    "type": "State",
    "address": "qlc_1ec4aarn8jjoart66j448x4mee47aj3ncic68o1n4ik88as6afspyb8j7s8z",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000518fb7",
    "signature": "7402a52307e4acfa58e58277b2cb01715de7f0834a3f9cdcf03cf807e8912905fc81ce13539cfff6dd6d2d576e236f09f0c09538327f7df628589ab125688108"
  },
  {
    "token": "b70b717c8677be04dec3a643c6f04c55e660df5af3769c6b6383cd0930caa3a5",
    "balance": "00000000000000080000000000000000",
    "link": "3643433833463046333142343842383436453237334241334442433735383934",
    "representative": "qlc_1fk5aew58jjiarsm4ijn9335if3pans5get4a6snaik58wtmigbn8faf16kg",
    "type": "State",
    "address": "qlc_1fk5aew58jjiarsm4ijn9335if3pans5get4a6snaik58wtmigbn8faf16kg",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000fc6868",
    "signature": "b35008a38a711b59eeeb395afb4fb035b6962ef85138a5e1380553bb6e98de13fe089604240eab73bdd937c497c240ea3843ddae9c454c35d6c5999799307408"
  },
  {
    "token": "4efaf4e434f28d54963ae7a01e866d2d155071ceec3e500ff840ce75decd918a",
    "balance": "00000000000000090000000000000000",
    "link": "3330304641374334463434343835313737424342463939333830373543303538",
    "representative": "qlc_1esi83564ft58j55af3n91tm4fsqab3n6jjs96smie3q8o3m1fbro44y69im",
    "type": "State",
    "address": "qlc_1esi83564ft58j55af3n91tm4fsqab3n6jjs96smie3q8o3m1fbro44y69im",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000000c015a",
    "signature": "82d3967e12dd592baf29d8bdbab8958467e148bcfdf9774504e68abed3ef1c21a78880cb2385daa2d685b1eea7f06d21102f4449621d0e05a1eef364ed679f0f"
  }
]
`
	scBlockstring = `[
  {
    "internalAccount": "qlc_1fsk8j1m6e4491sneg448s45gj3pas55cgc68ntm6jjn81snegbpac5gh7ro",
    "contract": {
      "abi": "tldLHqruPVZt3355hwyhEjIY4+YkMs9IM4vZc36mHYsqK3JF1ccMctcm4qKeIHPsEgU4zKE+BY7K4Gu1oLG8HtuU42c28FYAf+sdCBq9nMzF",
      "abiLength": 81,
      "abiHash": "7be6bf13ebf8a91bfefbdf47a055b7b317b5f2627d403f3ce8ed31f37fde402a"
    },
    "issuer": "qlc_1c8wg9eqjkcoeoc5hf7bbq1jc1hgoyayygt89adpxmbi55h7rmghumw4e78j",
    "type": "SmartContract",
    "address": "qlc_3tsbd1w65ujy6at6kwz81f8sry1ihixewsp44hshmmkymjc4ro6zhqbuxqdk",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000495b22",
    "signature": "358ce17185339de541b398b34933d8e4909c8e2695092a70f70ac902fff9ebac2badc3835d95a07fb8a8a5f1cda7a06bd2491507e13dce3a70c3d9b8b5553500"
  },
  {
    "internalAccount": "qlc_1e3qas45eisraatmkg478es51fjq8at5kfc7aew5ae3m8awm4e46q3ooio3y",
    "contract": {
      "abi": "RXHF5ANW7jdqYEyLqV/oXBPdqFN9qnG9Q5u8cYakVCi0Q1GhLxKX77Ye7iBf8hynPuuTN2WIllms8Z5vQpmQX5ajPSJCEE4cTEq6WRT4A6Pf8T70nKk4",
      "abiLength": 87,
      "abiHash": "7352b3c2ec12a5efe973b9aa776d6ce40a2c1ac2c9111c431c98bcd84a12a846"
    },
    "issuer": "qlc_3shroc3oimccschg4gxdhbaryumqdswd7by8bsrg3ytmq8dzh3p9u3abuyx5",
    "type": "SmartContract",
    "address": "qlc_38zuqywdfspj8ucq31fffo56uobagxeqmhd3wpqmxn6jmudogfrp7afrdnec",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000371627",
    "signature": "c62ef11e9393175efb2995d69d06e4f376c8bd53adf32eec991b15fc5af02992257fa17c0cd300e614680e42808f850a04107bde6cc6d33cab89a083aeb23403"
  },
  {
    "internalAccount": "qlc_1gc58j354g3s8iw66fbj8ru6ee45as4nejc4ans6cek681w5cfjjsk3c1szw",
    "contract": {
      "abi": "U7GBDJV3eDLd4Nna/zVpWippa1gY0EDF1bMKU+ttQvb9n0YxlHm04TAbp7d6NIM=",
      "abiLength": 47,
      "abiHash": "2a12310c2f095353abf4f8dd2d70286f4e679d413695c2ddf2b223e4edfde4a7"
    },
    "issuer": "qlc_3xxyra16x4aox395p5r77kbkmm6fncabtkjh37rjf313aawm3hhsas5nraa8",
    "type": "SmartContract",
    "address": "qlc_3up3sosbrxdob4gr8rc5sjwnc9o7ozqd6597bg69sih31p9cxh5efd33cii6",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000006b5fac",
    "signature": "67362c7658a2fdd5c20330f1ec0ac9afc0255f4619e66f4ed3b76fec1a44d6ff099b61ff8fc08e6ee0cc279cca4a082530359137b426e11603959cdfbc16c009"
  },
  {
    "internalAccount": "qlc_1ec4aarn8jjoart66j448x4mee47aj3ncic68o1n4ik88as6afspyb8j7s8z",
    "contract": {
      "abi": "z5FsM8OwbbhKVwZqXn6HQxDrrjNmLChBKjtE+NjNt2fHefi1mrke3sONoJRml2x2fN7VdMTdEJBDjLg=",
      "abiLength": 59,
      "abiHash": "d515f611993366497945c5710c85911582a8d35bd09b04c93446a2503bb914eb"
    },
    "issuer": "qlc_3bseyncp61pbkwdect4k4jpkc5txa8n9st3f6opofzpuyw81znckcut9kraa",
    "type": "SmartContract",
    "address": "qlc_3tcoy4q1f4ui1rfy1pbpgmhpxjbqfk3dtmbzkk1687men6emrm5rqfsanepq",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000518fb7",
    "signature": "b63bea22d2b8a4ba4b4f14c7750c95a6fe1f39eecb6ba9c10b5e8aaaa6507ada83691a9f78fbdf76ecd415c62918e1722c9702e26da849843bed21037dcda20c"
  },
  {
    "internalAccount": "qlc_1fk5aew58jjiarsm4ijn9335if3pans5get4a6snaik58wtmigbn8faf16kg",
    "contract": {
      "abi": "643z8JAIQAWGh27RsXUE7K7MTqUXidZIP9fYjggSTmcQ6JxB8yDYdNc/xCC5B0n1r7MH1bRlA/YU4Yul0qlDjmHPMaElim/rHaPnlMiM8F8K",
      "abiLength": 81,
      "abiHash": "cea5b990ad7bf9785cc9ab6896853779d06c9716a282c011df8a722e9d874111"
    },
    "issuer": "qlc_1cq44htgau4c9bh4sy58dy8j6mo5abdqxfrpd4ztbrn494g3p8ir56tkhdu8",
    "type": "SmartContract",
    "address": "qlc_3uxxmr7y6fxp8ta7f3o5q5319ry18143tkg6xh566bzgrg7q53db4xi8orcn",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "0000000000fc6868",
    "signature": "23efe1e2e0c4c7df0559a67e8d0212c9cfd0eb99b753caf230c4b10abf6cdfe9336cf141fb3a4b43124ed4f9886b989705cd60da01856e2c53b311a0658eef0c"
  },
  {
    "internalAccount": "qlc_1esi83564ft58j55af3n91tm4fsqab3n6jjs96smie3q8o3m1fbro44y69im",
    "contract": {
      "abi": "05jQsJqNY+cClDzrixYWBd3/",
      "abiLength": 18,
      "abiHash": "1de52ced6c8511a05088c5b24ed59801e21fa6a6ab4f5fee3aa616a8d60179b7"
    },
    "issuer": "qlc_18nszuxspjmt7bz7mr4hq3cnj77d79be9n1w4rf6sobinysgasykxpzg494u",
    "type": "SmartContract",
    "address": "qlc_39rjjq7ke5n33dcmypgmsi3in686dx4zdy8sjjtyz815k645epmnk6qcmz75",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000000c015a",
    "signature": "ee1d25afbeab0d469b3fe680ebbe4af9cce4d0e372b5e402b0843e03f30ff1f91c2a66b2c80655c5a4e0d1d3f8eea796822797c83fd3ac7f10d9134dea231c0a"
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
	var b []types.Block
	for _, sb := range genesisBlocks {
		b = append(b, &sb)
	}
	return b
}

func GetSmartContracts() []types.Block {
	var b []types.Block
	for _, sb := range smartContractBlocks {
		b = append(b, &sb)
	}
	return b
}
