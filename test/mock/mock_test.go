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

	_, priv, err := types.KeypairFromSeed("425E747CFCDD993019EB1AAC97FD2F5D3A94835D9A779C9BDC590739EDD1BB45", 0)
	if err != nil {
		t.Fatal(err)
	}
	account := types.NewAccount(priv)

	hash = sb.GetHash()
	t.Log(strings.ToUpper(hash.String()))
	sb.Signature = account.Sign(hash)
	verify := account.Address().Verify(hash[:], sb.Signature[:])
	if !verify {
		t.Fatal("invalid Signature")
	}
	t.Log("verify: ", verify)
	if !sb.IsValid() {
		var work types.Work
		worker, err := types.NewWorker(work, sb.Address.ToHash())
		if err != nil {
			t.Fatal(err)
		}
		work = worker.NewWork()
		sb.Work = work
		t.Log("IsValid: ", sb.IsValid())
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
	t.Skip()
	var keys []token

	var s1 []string
	var s2 []string
	for index, token := range keys {
		var block types.StateBlock
		var sb types.SmartContractBlock
		_, priv, _ := types.KeypairFromSeed(token.scKey, 0)
		account := types.NewAccount(priv)

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

		h := sb.GetHash()
		sign := account.Sign(h)
		verify2 := masterAddress.Verify(h[:], sign[:])
		sb.Signature = sign
		verify1 := masterAddress.Verify(h[:], sb.Signature[:])
		t.Log(h.String(), "=>sign: ", verify1, ", ", verify2)
		if !sb.IsValid() {
			var work types.Work
			worker, err := types.NewWorker(work, sb.Address.ToHash())
			if err != nil {
				t.Fatal(err)
			}
			work = worker.NewWork()
			sb.Work = work
			t.Log(h.String(), "=>valid: ", sb.IsValid())
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
		block.Token = h
		block.Representative = masterAddress
		block.Balance = types.StringToBalance(token.balance)
		block.Link = masterAddress.ToHash()
		bh := block.GetHash()
		block.Signature = account.Sign(bh)
		v1 := masterAddress.Verify(bh[:], block.Signature[:])
		t.Log(bh.String(), "=>sign: ", v1)
		if !block.IsValid() {
			var work types.Work
			worker, err := types.NewWorker(work, masterAddress.ToHash())
			if err != nil {
				t.Fatal(err)
			}
			work = worker.NewWork()
			block.Work = work
			//t.Log(block, block.IsValid())
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
	hash := smartContractBlocks[0].GetHash()
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

func TestGetGenesis(t *testing.T) {
	g := GetGenesis()

	for i := 0; i < len(g); i++ {
		block := g[i]
		t.Logf("%v: %s", &block, block)
	}
}

func TestGetGenesis2(t *testing.T) {
	g := GetGenesis()
	for _, b := range g {
		t.Logf("%v", b)
	}
}

func TestGetChainTokenType(t *testing.T) {
	h := GetChainTokenType()
	h2 := smartContractBlocks[0].GetHash()
	if h != h2 {
		t.Fatal("GetChainTokenType error")
	}
	if genesisBlocks[0].Token != h {
		t.Fatal("genesis error")
	}
}
