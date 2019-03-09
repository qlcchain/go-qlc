/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"go.uber.org/zap"
	"sort"
)

type MintageApi struct {
	logger *zap.SugaredLogger
	ledger *ledger.Ledger
}

func NewMintageApi(ledger *ledger.Ledger) *MintageApi {
	return &MintageApi{ledger: ledger, logger: log.NewLogger("rpc/mintage")}
}

type MintageParams struct {
	SelfAddr    types.Address
	PrevHash    types.Hash
	TokenName   string
	TokenSymbol string
	TotalSupply string
	Decimals    uint8
}

func (m *MintageApi) GetMintageData(param MintageParams) ([]byte, error) {
	tokenId := abi.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	return abi.ABIMintage.PackMethod(abi.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol, totalSupply, param.Decimals)
}

func (m *MintageApi) GetWithdrawMintageData(tokenId types.Hash) ([]byte, error) {
	return abi.ABIMintage.PackMethod(abi.MethodNameMintageWithdraw, tokenId)
}

type ApiTokenInfo struct {
	types.TokenInfo
}

type TokenInfoList struct {
	Count int             `json:"totalCount"`
	List  []*ApiTokenInfo `json:"tokenInfos"`
}

type byName []*ApiTokenInfo

func (a byName) Len() int      { return len(a) }
func (a byName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool {
	if a[i].TokenName == a[j].TokenName {
		return a[i].TokenId.String() < a[j].TokenId.String()
	}
	return a[i].TokenName < a[j].TokenName
}

func (m *MintageApi) GetTokenInfoList(index int, count int) (*TokenInfoList, error) {
	tokens, err := m.ledger.ListTokens()
	if err != nil {
		return nil, err
	}
	tokenList := make([]*ApiTokenInfo, 0)
	for _, v := range tokens {
		tokenList = append(tokenList, &ApiTokenInfo{TokenInfo: *v})
	}
	sort.Sort(byName(tokenList))
	length := len(tokens)
	start, end := getRange(index, count, length)
	return &TokenInfoList{length, tokenList[start:end]}, nil
}

func (m *MintageApi) GetTokenInfoById(tokenId types.Hash) (*ApiTokenInfo, error) {
	token, err := m.ledger.GetTokenById(tokenId)
	if err != nil {
		return nil, err
	}
	return &ApiTokenInfo{*token}, nil
}

func getRange(index, count, listLen int) (int, int) {
	start := index * count
	if start >= listLen {
		return listLen, listLen
	}
	end := start + count
	if end >= listLen {
		return start, listLen
	}
	return start, end
}
