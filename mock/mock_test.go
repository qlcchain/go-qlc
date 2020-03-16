/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mock

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestMockHash(t *testing.T) {
	hash := Hash()
	if hash.IsZero() {
		t.Fatal("create hash failed.")
	} else {
		t.Log(hash.String())
	}
}

func TestMockAddress(t *testing.T) {
	address := Address()
	if !types.IsValidHexAddress(address.String()) {
		t.Fatal("mock address failed")
	}
}

func TestAccountMeta_Token(t *testing.T) {
	addr := Address()
	am := AccountMeta(addr)
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
	addr := Address()
	am := AccountMeta(addr)
	bytes, err := json.Marshal(am)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(bytes))

	tm := TokenMeta(addr)
	tm.Type = types.Hash{}

	am.Tokens[0] = tm

	t.Log(strings.Repeat("*", 20))
	bytes2, err2 := json.Marshal(am)
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
	err := json.Unmarshal([]byte(qlc), &block1)
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

	t.Log(util.ToString(block1))
}

func TestMockGenesisScBlock(t *testing.T) {
	var sb types.SmartContractBlock
	abi := []byte("6060604052341561000F57600080FD5B336000806101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908373FFFFFFFFFFFFFFFFFFFF")
	hash, _ := types.HashBytes(abi)
	sb.Abi = types.ContractAbi{Abi: abi, AbiLength: 64, AbiHash: hash}
	sb.Address = Address()
	sb.Owner = Address()
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
	bytes, err := json.Marshal(&sb)
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
		i := rand.Intn(100)
		abi := make([]byte, i)
		_ = random.Bytes(abi)
		hash, _ := types.HashBytes(abi)
		sb.Abi = types.ContractAbi{Abi: abi, AbiLength: uint64(i), AbiHash: hash}
		sb.Address = Address()
		sb.Owner = Address()
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
		sbstring := util.ToString(&sb)
		var buff strings.Builder
		buff.WriteString(strconv.Itoa(index) + strings.Repeat("*", 50) + "\n")
		tmp1 := jsonPrettyPrint(sbstring)
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
		s := util.ToString(block)

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

func TestStateBlock(t *testing.T) {
	b := StateBlock()
	if valid := b.IsValid(); !valid {
		t.Fatal("state block is invalid")
	}
}

func TestStateBlockWithoutWork(t *testing.T) {
	b := StateBlockWithoutWork()
	if b == nil {
		t.Fatal("mock blocks error")
	}
}

func TestAccountMock(t *testing.T) {
	account := Account()
	h := Hash()
	sign := account.Sign(h)
	if !account.Address().Verify(h[:], sign[:]) {
		t.Fatal("account verify error")
	}
}

func TestBlockChain(t *testing.T) {
	//defer func() { os.RemoveAll(filepath.Join(config.QlcTestDataDir(), "blocks.json")) }()
	blocks, err := BlockChain(true)
	if err != nil {
		t.Fatal(err)
	}

	if len(blocks) == 0 {
		t.Fatal("create blocks error")
	}
	//r, err := json.Marshal(blocks)
	//if err != nil {
	//	t.Fatal(err)
	//} else {
	//	t.Log(string(r))
	//}
}

var seedStr = []string{
	"DB68096C0E2D2954F59DA5DAAE112B7B6F72BE35FC96327FE0D81FD0CE5794A9",
	"E4935D4D9DEF9D12BC2059C34848C444DBC462FBFF592428C218BF0BC174D065",
	"192C4DDFD9BCC7DF27701197FBC972E6D07854F718BE09AD1D5E9338C435C1A9",
	"AAB43DFBFCC6702504F03F64B66B0482A206EFDD5990D0C5FFCE164EBB088E06",
	"FBEA7F04DC9AD25E2CBC05FAEF6CEF98DF08CF04582937832F67B3883075244A",
	"0578B09D725C77432886632364FDE29D3DAFB4A7748B7801FBD6D79BBF013B73",
}

func TestNewRandomMac(t *testing.T) {
	for i := 0; i < 10; i++ {
		mac := NewRandomMac()
		t.Log(mac.String())
	}
}

func TestSmartContractBlock(t *testing.T) {
	if blk := SmartContractBlock(); blk == nil {
		t.Fatal()
	}
}

func TestStateBlockWithAddress(t *testing.T) {
	a := Address()
	if blk := StateBlockWithAddress(a); blk.Address != a {
		t.Fatal()
	}
}

func TestMockChain(t *testing.T) {
	t.Skip()
	dir := filepath.Join(config.QlcTestDataDir(), "mock", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []*types.StateBlock
	priv1, _ := hex.DecodeString("7098c089e66bd66476e3b88df8699bcd4dacdd5e1e5b41b3c598a8a36d851184d992a03b7326b7041f689ae727292d761b329a960f3e4335e0a7dcf2c43c4bcf")
	// qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te
	priv2, _ := hex.DecodeString("31ee4e16826569dc631b969e71bd4c46d5c0df0daeca6933f46586f36f49537cd929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393")
	priv3, _ := hex.DecodeString("8be0696a2d51dec8e2859dcb8ce2fd7ce7412eb9d6fa8a2089be8e8f1eeb4f0e458779381a8d21312b071729344a0cb49dc1da385993e19d58b5578da44c0df0")
	ac1 := types.NewAccount(priv1)
	ac2 := types.NewAccount(priv2)
	ac3 := types.NewAccount(priv3)

	tuples := []*types.Tuple{{
		First:  config.GasToken(),
		Second: config.GenesisMintageHash(),
	}, {
		First:  config.ChainToken(),
		Second: config.GenesisBlockHash(),
	}}
	for _, tuple := range tuples {
		token := tuple.First.(types.Hash)
		genesis := tuple.Second.(types.Hash)
		b0 := createBlock(types.Open, *ac1, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(1e15))}, genesis, ac1.Address()) //a1 open
		blocks = append(blocks, b0)

		b1 := createBlock(types.Send, *ac1, b0.GetHash(), token, types.Balance{Int: big.NewInt(int64(1e14))}, types.Hash(ac2.Address()), ac1.Address()) //a1 send
		blocks = append(blocks, b1)

		b2 := createBlock(types.Open, *ac2, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(1e14))}, b1.GetHash(), ac1.Address()) //a2 open
		blocks = append(blocks, b2)

		b3 := createBlock(types.Open, *ac3, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(1e14))}, b2.GetHash(), ac1.Address()) //a2 open
		blocks = append(blocks, b3)
	}

	fmt.Println(util.ToIndentString(blocks))
}
