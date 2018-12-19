/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
	"math"
	"time"
)

var (
	ChainTokenType, _ = types.NewHash("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")
)

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
