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
	"github.com/qlcchain/go-qlc/common"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
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
	sb.Address = a.Address()
	sb.Token = common.QLCChainToken
	sb.Previous = Hash()
	sb.Representative = common.GenesisAccountAddress
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

	token := common.QLCChainToken

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
