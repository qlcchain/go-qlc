/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mock

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
)

var (
	// qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq
	priv1, _ = hex.DecodeString("7098c089e66bd66476e3b88df8699bcd4dacdd5e1e5b41b3c598a8a36d851184d992a03b7326b7041f689ae727292d761b329a960f3e4335e0a7dcf2c43c4bcf")
	// qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te
	priv2, _ = hex.DecodeString("31ee4e16826569dc631b969e71bd4c46d5c0df0daeca6933f46586f36f49537cd929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393")
	//qlc_1je9h6w3o5b386oig7sb8j71sf6xr9f5ipemw8gojfcqjpk6r5hiu7z3jx3z
	priv3, _ = hex.DecodeString("8be0696a2d51dec8e2859dcb8ce2fd7ce7412eb9d6fa8a2089be8e8f1eeb4f0e458779381a8d21312b071729344a0cb49dc1da385993e19d58b5578da44c0df0")
	Account1 = types.NewAccount(priv1)
	Account2 = types.NewAccount(priv2)
	Account3 = types.NewAccount(priv3)
)

func Hash() types.Hash {
	h := types.Hash{}
	_ = random.Bytes(h[:])
	return h
}

func AccountMeta(addr types.Address) *types.AccountMeta {
	var am types.AccountMeta
	am.Address = addr
	am.CoinNetwork = types.ZeroBalance
	am.CoinBalance = types.ZeroBalance
	am.CoinStorage = types.ZeroBalance
	am.CoinOracle = types.ZeroBalance
	am.CoinBalance = types.ZeroBalance
	am.CoinVote = types.ZeroBalance
	am.Tokens = []*types.TokenMeta{}
	for i := 0; i < 5; i++ {
		t := TokenMeta(addr)
		am.Tokens = append(am.Tokens, t)
	}
	return &am
}

func TokenMeta(addr types.Address) *types.TokenMeta {
	return TokenMeta2(addr, types.ZeroHash)
}

func TokenMeta2(addr types.Address, token types.Hash) *types.TokenMeta {
	if token.IsZero() {
		token = Hash()
	}
	s1, _ := random.Intn(math.MaxInt32)
	i := new(big.Int).SetInt64(int64(s1))
	t := types.TokenMeta{
		//TokenAccount: Address(),
		Type:           token,
		BelongTo:       addr,
		Balance:        types.Balance{Int: i},
		BlockCount:     1,
		OpenBlock:      Hash(),
		Header:         Hash(),
		Representative: Address(),
		Modified:       common.TimeNow().Unix(),
	}

	return &t
}

func Address() types.Address {
	address, _, _ := types.GenerateAddress()

	return address
}

func Account() *types.Account {
	seed, _ := types.NewSeed()
	_, priv, _ := types.KeypairFromSeed(seed.String(), 0)
	return types.NewAccount(priv)
}

func StateBlock() *types.StateBlock {
	a := Account()
	b, _ := types.NewBlock(types.State)
	sb := b.(*types.StateBlock)
	i, _ := random.Intn(math.MaxInt16)
	sb.Type = types.State
	sb.Balance = types.Balance{Int: big.NewInt(int64(i))}
	//sb.Vote = types.ZeroBalance
	//sb.Oracle = types.ZeroBalance
	//sb.Network = types.ZeroBalance
	//sb.Storage = types.ZeroBalance
	sb.Address = a.Address()
	sb.Token = config.ChainToken()
	sb.Previous = Hash()
	sb.Representative = config.GenesisAddress()
	addr := Address()
	sb.Link = addr.ToHash()
	sb.Signature = a.Sign(sb.GetHash())
	var w types.Work
	worker, _ := types.NewWorker(w, sb.Root())

	sb.Work = worker.NewWork()

	return sb
}

func SmartContractBlock() *types.SmartContractBlock {
	var sb types.SmartContractBlock
	account := Account()

	masterAddress := account.Address()
	i := rand.Intn(100)
	abi := make([]byte, i)
	_ = random.Bytes(abi)
	hash, _ := types.HashBytes(abi)
	sb.Abi = types.ContractAbi{Abi: abi, AbiLength: uint64(i), AbiHash: hash}
	sb.Address = Address()
	sb.Owner = Address()
	sb.InternalAccount = masterAddress
	var w types.Work
	h := sb.GetHash()
	worker, _ := types.NewWorker(w, h)

	sb.Work = worker.NewWork()
	sb.Signature = account.Sign(h)

	return &sb
}

func StateBlockWithoutWork() *types.StateBlock {
	sb := new(types.StateBlock)
	a := Account()
	i, _ := random.Intn(math.MaxInt16)
	sb.Type = types.Open
	sb.Balance = types.Balance{Int: big.NewInt(int64(i))}
	//sb.Vote = types.NewBalance(0)
	//sb.Network = types.NewBalance(0)
	//sb.Oracle = types.NewBalance(0)
	//sb.Storage = types.NewBalance(0)
	sb.Address = a.Address()
	sb.Token = config.ChainToken()
	sb.Previous = types.ZeroHash
	sb.Representative = config.GenesisAddress()
	sb.Timestamp = common.TimeNow().Unix()
	//addr := Address()
	sb.Link = Hash()
	message := Hash()
	sb.Message = &message
	return sb
}

func StateBlockWithAddress(addr types.Address) *types.StateBlock {
	sb := new(types.StateBlock)
	//	a := mock.Account()
	i, _ := random.Intn(math.MaxInt16)
	sb.Type = types.Open
	sb.Balance = types.Balance{Int: big.NewInt(int64(i))}
	sb.Address = addr
	sb.Token = config.ChainToken()
	sb.Previous = types.ZeroHash
	sb.Representative = config.GenesisAddress()
	sb.Timestamp = common.TimeNow().Unix()
	//addr := Address()
	sb.Link = Hash()
	message := Hash()
	sb.Message = &message
	return sb
}

func BlockChain(isGas bool) ([]*types.StateBlock, error) {
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
	fmt.Println(ac1.String())
	//ac3 := Account()

	var token types.Hash
	var genesis types.Hash
	if isGas {
		token = config.GasToken()
		genesis = config.GenesisMintageHash()
	} else {
		token = config.ChainToken()
		genesis = config.GenesisBlockHash()
	}

	b0 := createBlock(types.Open, *ac1, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(100000000000))}, genesis, ac1.Address()) //a1 open
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

func BlockChainWithAccount(isGas bool) ([]*types.StateBlock, *types.Account, *types.Account, error) {
	var blocks []*types.StateBlock
	ac1 := Account()
	ac2 := Account()
	fmt.Println(ac1.String())
	//ac3 := Account()

	var token types.Hash
	var genesis types.Hash
	if isGas {
		token = config.GasToken()
		genesis = config.GenesisMintageHash()
	} else {
		token = config.ChainToken()
		genesis = config.GenesisBlockHash()
	}

	b0 := createBlock(types.Open, *ac1, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(100000000000000))}, genesis, ac1.Address()) //a1 open
	blocks = append(blocks, b0)

	b1 := createBlock(types.Send, *ac1, b0.GetHash(), token, types.Balance{Int: big.NewInt(int64(40000000000000))}, types.Hash(ac2.Address()), ac1.Address()) //a1 send
	blocks = append(blocks, b1)

	b2 := createBlock(types.Open, *ac2, types.ZeroHash, token, types.Balance{Int: big.NewInt(int64(60000000000000))}, b1.GetHash(), ac1.Address()) //a2 open
	blocks = append(blocks, b2)

	b3 := createBlock(types.Change, *ac2, b2.GetHash(), token, types.Balance{Int: big.NewInt(int64(60000000000000))}, types.ZeroHash, ac2.Address()) //a2 change
	blocks = append(blocks, b3)

	b4 := createBlock(types.Send, *ac2, b3.GetHash(), token, types.Balance{Int: big.NewInt(int64(30000000000000))}, types.Hash(ac1.Address()), ac2.Address()) //a2 send
	blocks = append(blocks, b4)

	b5 := createBlock(types.Receive, *ac1, b1.GetHash(), token, types.Balance{Int: big.NewInt(int64(70000000000000))}, b4.GetHash(), ac1.Address()) //a1 receive
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

	return blocks, ac1, ac2, nil
}

func createBlock(t types.BlockType, ac types.Account, pre types.Hash, token types.Hash, balance types.Balance, link types.Hash, rep types.Address) *types.StateBlock {
	blk := new(types.StateBlock)
	blk.Type = t
	blk.Address = ac.Address()
	blk.Previous = pre
	blk.Token = token
	blk.Balance = balance
	//blk.Vote = types.ZeroBalance
	//blk.Oracle = types.ZeroBalance
	//blk.Network = types.ZeroBalance
	//blk.Storage = types.ZeroBalance
	blk.Timestamp = common.TimeNow().Unix()
	blk.Link = link
	blk.Representative = rep
	message := Hash()
	blk.Message = &message
	blk.Sender = []byte("15811110000")
	blk.Receiver = []byte("15800001111")
	blk.Signature = ac.Sign(blk.GetHash())
	var w types.Work
	worker, _ := types.NewWorker(w, blk.Root())
	blk.Work = worker.NewWork()
	return blk
}

type Mac [6]byte

func (m Mac) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
}

func NewRandomMac() Mac {
	var m [6]byte

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		mac_byte := rand.Intn(256)
		m[i] = byte(mac_byte)

		rand.Seed(int64(mac_byte))
	}

	return Mac(m)
}
