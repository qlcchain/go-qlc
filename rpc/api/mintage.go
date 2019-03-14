/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"go.uber.org/zap"
)

type MintageApi struct {
	logger  *zap.SugaredLogger
	ledger  *ledger.Ledger
	mintage *contract.Mintage
}

func NewMintageApi(ledger *ledger.Ledger) *MintageApi {
	return &MintageApi{ledger: ledger, logger: log.NewLogger("api_mintage"), mintage: &contract.Mintage{}}
}

type MintageParams struct {
	SelfAddr     types.Address `json:"selfAddr"`
	PrevHash     types.Hash    `json:"prevHash"`
	TokenName    string        `json:"tokenName"`
	TokenSymbol  string        `json:"tokenSymbol"`
	TotalSupply  string        `json:"totalSupply"`
	Decimals     uint8         `json:"decimals"`
	PledgeAmount big.Int       `json:"pledgeAmount"`
}

func (m *MintageApi) GetMintageData(param MintageParams) ([]byte, error) {
	tokenId := cabi.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	return cabi.ABIMintage.PackMethod(cabi.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol, totalSupply, param.Decimals)
}

func (m *MintageApi) GetMintageBlock(param MintageParams) (*types.StateBlock, error) {
	tokenId := cabi.NewTokenHash(param.SelfAddr, param.PrevHash, param.TokenName)
	totalSupply, err := util.StringToBigInt(&param.TotalSupply)
	if err != nil {
		return nil, err
	}
	data, err := cabi.ABIMintage.PackMethod(cabi.MethodNameMintage, tokenId, param.TokenName, param.TokenSymbol, totalSupply, param.Decimals)
	if err != nil {
		return nil, err
	}

	am, err := m.ledger.GetAccountMeta(param.SelfAddr)
	if err != nil {
		return nil, err
	}

	tm := am.Token(common.QLCChainToken)
	if tm == nil {
		return nil, fmt.Errorf("%s do not hava any chain token", param.SelfAddr.String())
	}

	minPledgeAmount := types.Balance{Int: big.NewInt(1e12)}

	if tm.Balance.Compare(minPledgeAmount) == types.BalanceCompSmaller {
		return nil, fmt.Errorf("not enough balance %s, expect %s", tm.Balance, tm.Balance)
	}
	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.SelfAddr,
		Balance:        tm.Balance.Sub(minPledgeAmount),
		Previous:       tm.Header,
		Link:           types.Hash(types.MintageAddress),
		Representative: tm.Representative,
		Data:           data,
		Timestamp:      time.Now().UTC().Unix(),
	}

	err = m.mintage.DoSend(m.ledger, send)
	if err != nil {
		return nil, err
	}

	return send, nil
}

func (m *MintageApi) GetRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	reward := &types.StateBlock{}

	blocks, err := m.mintage.DoReceive(m.ledger, reward, input)
	if err != nil {
		return nil, err
	}
	if len(blocks) > 0 {
		return reward, nil
	}

	return nil, errors.New("can not generate reward block")
}

func (m *MintageApi) GetWithdrawMintageData(tokenId types.Hash) ([]byte, error) {
	return cabi.ABIMintage.PackMethod(cabi.MethodNameMintageWithdraw, tokenId)
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

func (m *MintageApi) GetTokenInfoList(count int, index int) (*TokenInfoList, error) {
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
