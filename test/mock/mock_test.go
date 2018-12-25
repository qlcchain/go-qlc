/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mock

import (
	"bytes"
	"encoding/json"
	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMockHash(t *testing.T) {
	hash := MockHash()
	if hash.IsZero() {
		t.Fatal("create hash failed.")
	} else {
		t.Log(hash.String())
	}
}

func TestMockAddress(t *testing.T) {
	address := MockAddress()
	if !types.IsValidHexAddress(address.String()) {
		t.Fatal("mock address failed")
	}
}

func TestAccountMeta_Token(t *testing.T) {
	addr := MockAddress()
	am := MockAccountMeta(addr)
	token := am.Token(types.Hash{})
	if token != nil {
		t.Fatal("get token failed")
	}

	if len(am.Tokens) > 0 {
		tt := am.Tokens[0].Type
		tm := am.Token(tt)
		if !reflect.DeepEqual(tm, am.Tokens[0]) {
			t.Fatal("get the first token failed")
		}
	}
}

func TestMockAccountMeta(t *testing.T) {
	addr := MockAddress()
	am := MockAccountMeta(addr)
	bytes, err := jsoniter.Marshal(am)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(bytes))

	tm := MockTokenMeta(addr)
	tm.Type = types.Hash{}

	am.Tokens[0] = tm

	t.Log(strings.Repeat("*", 20))
	bytes2, err2 := jsoniter.Marshal(am)
	if err2 != nil {
		t.Fatal(err2)
	}
	t.Log(string(bytes2))
}

func TestMockGenesisBlock(t *testing.T) {
	qlc := `{
  "type": "State",
  "address": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
  "previous": "0000000000000000000000000000000000000000000000000000000000000000",
  "signature": "ad57aa8819fa6a7811a13ff0684a79afdefb05077bcad4ec7365c32d2a88d78c8c7c54717b40c0888a0692d05bf3771df6d16a1f24ae612172922bbd4d93370f",
  "work": "f3389dd67ced8429",
  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
  "token": "D21C700BCB29D22E5815AF8020416425D700727A465161E5A50F454B8482F367",
  "balance": "00000000000000060000000000000000",
  "link": "d5ba6c7bb3f4f6545e08b03d6da1258840e0395080378a890601991a2a9e3163",
  "representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic"
}`

	var block1 types.StateBlock
	err := jsoniter.Unmarshal([]byte(qlc), &block1)
	if err != nil {
		t.Fatal(err)
	}

	//addr, _ := types.HexToAddress("qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic")
	//if block1.Address != addr {
	//	t.Fatal("addr != address")
	//}
	//rep, _ := types.HexToAddress("qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic")
	//if rep != block1.Representative {
	//	t.Fatal("rep != Representative")
	//}
	//
	//b := types.Balance{Hi: 6, Lo: 0}
	//if block1.Balance != b {
	//	t.Fatalf("b(%d)!=Balance(%d)", b.BigInt(), block1.Balance.BigInt())
	//}
	//link, _ := types.NewHash("D5BA6C7BB3F4F6545E08B03D6DA1258840E0395080378A890601991A2A9E3163")
	//if link != block1.Link {
	//	t.Fatal("link != Link")
	//}
	//h, _ := types.NewHash("125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36")
	//if h != block1.Token {
	//	t.Fatal("h != Token")
	//}
	//var work types.Work
	//_ = work.UnmarshalText([]byte("f3389dd67ced8429"))
	//if work != block1.Work {
	//	t.Fatal("work != Work")
	//}
	var sign types.Signature
	err = sign.Of("AD57AA8819FA6A7811A13FF0684A79AFDEFB05077BCAD4EC7365C32D2A88D78C8C7C54717B40C0888A0692D05BF3771DF6D16A1F24AE612172922BBD4D93370F")
	if err != nil {
		t.Fatal(err)
	}
	if sign != block1.Signature {
		t.Fatal("sign != Signature")
	}
	hash := block1.GetHash()
	account := types.NewAccount([]byte("4870A614A9971DE060ED2997E07FBF0A724A20B83F8B6D7D6FF59D552F403F96"))
	sign1 := account.Sign(hash)
	block1.Signature = sign1
	if !block1.IsValid() {
		var work types.Work
		worker, err := types.NewWorker(work, hash)
		if err != nil {
			t.Fatal(err)
		}
		work = worker.NewWork()
		block1.Work = work
	}

	t.Log(jsoniter.MarshalToString(block1))
}

func TestMockGenesisScBlock(t *testing.T) {
	var sb types.SmartContractBlock
	sb.Type = types.SmartContract
	abi := []byte("6060604052341561000F57600080FD5B336000806101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908373FFFFFFFFFFFFFFFFFFFF")
	hash, _ := types.HashBytes(abi)
	sb.Abi = types.ContractAbi{Abi: abi, AbiLength: 64, AbiHash: hash}
	sb.Address = MockAddress()
	sb.Issuer = MockAddress()
	sb.InternalAccount, _ = types.HexToAddress("qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic")

	account := types.NewAccount([]byte("425E747CFCDD993019EB1AAC97FD2F5D3A94835D9A779C9BDC590739EDD1BB45"))
	hash = sb.GetHash()
	t.Log(strings.ToUpper(hash.String()))
	sign := account.Sign(hash)
	sb.Signature = sign
	if !sb.IsValid() {
		var work types.Work
		worker, err := types.NewWorker(work, hash)
		if err != nil {
			t.Fatal(err)
		}
		work = worker.NewWork()
		sb.Work = work
	}
	bytes, err := jsoniter.Marshal(&sb)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}

type token struct {
	scKey   string
	key     string
	balance string
}

func TestGenerate(t *testing.T) {
	//t.Skip()
	keys := []token{{"425E747CFCDD993019EB1AAC97FD2F5D3A94835D9A779C9BDC590739EDD1BB45", "4870A614A9971DE060ED2997E07FBF0A724A20B83F8B6D7D6FF59D552F403F96", "00000000000000060000000000000000"},
		{"E8CD816A8EEB7D61DC14C8F7CD46883170E5C740F2F17B37BEF0C27E99A96809", "310D5FDDF18C03056426BFD44912970D07FD6C8B598E320672495EC84032910D", "00000000000000070000000000000000"},
		{"04AA70F95DBC8CF477618404717D92ABA4B4E845C1DE0422FB72BB61FDC6B8AF", "B121C30B72EFA746497750D85C3D9DA29C4B18948B5166F0CFEFEBE2E2D08561", "00000000000000050000000000000000"},
		{"4F4575B728324EA111C2EC9557A01DE9A5CDC33E280DD694E155881DE33655D6", "58A627164BDD262FE6E29B31A3E7894D1BB1CF5F4BDB7E60EDCEAD5AABF22D76", "00000000000000090000000000000000"},
		{"470D9B25ADF80117D816CC7BCFB72A9ECC68DA56E24B38694344FE7C03309B77", "9E7354A2E7908F8A859A0AD108DECBFE6CC83F0F31B48B846E273BA3DBC75894", "00000000000000080000000000000000"},
		{"7FADB303836C615A91DEFB050445ADDFC70E2D2525A2B987FF690E70DA18F694", "D821BF91B19A27B108C4295364FB11F4300FA7C4F44485177BCBF9938075C058", "00000000000000090000000000000000"},
	}

	var s1 []string
	var s2 []string
	for index, token := range keys {
		var block types.StateBlock
		var sb types.SmartContractBlock
		scAccount := types.NewAccount([]byte(token.scKey))
		account := types.NewAccount([]byte(token.key))
		masterAddress := account.Address()
		sb.Type = types.SmartContract
		i := rand.Intn(100)
		abi := make([]byte, i)
		_ = random.Bytes(abi)
		hash, _ := types.HashBytes(abi)
		sb.Abi = types.ContractAbi{Abi: abi, AbiLength: uint64(i), AbiHash: hash}
		sb.Address = MockAddress()
		sb.Issuer = MockAddress()
		sb.InternalAccount = masterAddress

		hash = sb.GetHash()
		sign := scAccount.Sign(hash)
		sb.Signature = sign
		if !sb.IsValid() {
			var work types.Work
			worker, err := types.NewWorker(work, hash)
			if err != nil {
				t.Fatal(err)
			}
			work = worker.NewWork()
			sb.Work = work
		}
		bytes, err := jsoniter.MarshalToString(&sb)
		if err != nil {
			t.Fatal(err)
		}
		var buff strings.Builder
		buff.WriteString(strconv.Itoa(index) + strings.Repeat("*", 50) + "\n")
		tmp1 := jsonPrettyPrint(bytes)
		buff.WriteString("smart contract: " + token.scKey + "\n" + tmp1 + "\n")
		s1 = append(s1, tmp1)

		block.Type = types.State
		block.Address = masterAddress
		block.Token = hash
		block.Representative = masterAddress
		block.Balance, _ = types.ParseBalanceString(token.balance)
		block.Link = masterAddress.ToHash()
		h := block.GetHash()
		block.Signature = account.Sign(h)
		if !block.IsValid() {
			var work types.Work
			worker, err := types.NewWorker(work, hash)
			if err != nil {
				t.Fatal(err)
			}
			work = worker.NewWork()
			block.Work = work
		}
		s, err := jsoniter.MarshalToString(block)
		if err != nil {
			t.Fatal(err)
		}
		tmp2 := jsonPrettyPrint(s)
		buff.WriteString("genesis: " + token.key + "\n" + tmp2 + "\n")
		s2 = append(s2, tmp2)
		t.Log(buff.String())
		buff.Reset()
	}
	t.Log("[" + strings.Join(s1, ",") + "]")
	t.Log("[" + strings.Join(s2, ",") + "]")
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func TestGetTokenById(t *testing.T) {
	hash := genesisBlocks[0].GetHash()
	t.Log(hash)
	ti, err := GetTokenById(hash)
	if err != nil {
		t.Fatal(err)
	}

	if ti.TokenName != "QLC" {
		t.Fatal("err token name")
	}

	if GetChainTokenType() != hash {
		t.Fatal("chain token error")
	}

	t.Log(jsoniter.MarshalToString(ti))
}
