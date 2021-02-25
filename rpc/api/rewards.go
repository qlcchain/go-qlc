/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type RewardsAPI struct {
	logger           *zap.SugaredLogger
	ledger           ledger.Store
	rewards          *contract.AirdropRewards
	confidantRewards *contract.ConfidantRewards
	cc               *chainctx.ChainContext
}

type sendParam struct {
	*cabi.RewardsParam
	self *types.Address
	//am   *types.AccountMeta
	tm *types.TokenMeta
}

func NewRewardsAPI(l ledger.Store, cc *chainctx.ChainContext) *RewardsAPI {
	api := &RewardsAPI{
		ledger:           l,
		logger:           log.NewLogger("api_rewards"),
		rewards:          &contract.AirdropRewards{},
		confidantRewards: &contract.ConfidantRewards{},
		cc:               cc,
	}

	return api
}

type RewardsParam struct {
	Id     string        `json:"Id"`
	Amount types.Balance `json:"amount"`
	Self   types.Address `json:"self"`
	To     types.Address `json:"to"`
}

func (r *RewardsAPI) GetUnsignedRewardData(param *RewardsParam) (types.Hash, error) {
	if param == nil {
		return types.ZeroHash, ErrParameterNil
	}
	return r.generateHash(param, cabi.MethodNameUnsignedAirdropRewards, func(param *RewardsParam) (bytes []byte, e error) {
		return hex.DecodeString(param.Id)
	})
}

func (r *RewardsAPI) GetUnsignedConfidantData(param *RewardsParam) (types.Hash, error) {
	if param == nil {
		return types.ZeroHash, ErrParameterNil
	}
	return r.generateHash(param, cabi.MethodNameUnsignedConfidantRewards, func(param *RewardsParam) (bytes []byte, e error) {
		h := types.HashData([]byte(param.Id))
		return h[:], nil
	})
}

func (r *RewardsAPI) generateHash(param *RewardsParam, methodName string, fn func(param *RewardsParam) ([]byte, error)) (types.Hash, error) {
	if param == nil {
		return types.ZeroHash, ErrParameterNil
	}
	if len(param.Id) == 0 || param.Amount.IsZero() || param.Self.IsZero() || param.To.IsZero() {
		return types.ZeroHash, errors.New("invalid param")
	}

	id, err := fn(param)
	if err != nil {
		return types.ZeroHash, err
	}
	var txHash types.Hash
	var rxHash types.Hash

	if tm, err := r.ledger.GetTokenMeta(param.Self, config.GasToken()); tm != nil && err == nil {
		txHash = tm.Header
	} else {
		return types.ZeroHash, err
	}

	if b, err := r.ledger.HasTokenMeta(param.To, config.GasToken()); err != nil {
		return types.ZeroHash, err
	} else {
		if b {
			tm, _ := r.ledger.GetTokenMeta(param.To, config.GasToken())
			rxHash = tm.Header
		} else {
			rxHash = types.ZeroHash
		}
	}

	r.logger.Debugf("%s %s %s %s %s", methodName, param.To.String(), txHash.String(), rxHash.String(), param.Amount.Int)

	p2 := &cabi.RewardsParam{
		Beneficial: param.To,
		TxHeader:   txHash,
		RxHeader:   rxHash,
		Amount:     param.Amount.Int,
	}
	p2.Id, _ = types.BytesToHash(id)

	if data, err := p2.ToUnsignedABI(methodName); err == nil {
		return types.HashData(data), nil
	} else {
		return types.ZeroHash, err
	}
}

func (r *RewardsAPI) GetSendRewardBlock(param *RewardsParam, sign *types.Signature) (*types.StateBlock, error) {
	if !r.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if param == nil || sign == nil {
		return nil, ErrParameterNil
	}

	if p, err := r.verifySign(param, sign, cabi.MethodNameUnsignedAirdropRewards, func(param *RewardsParam) (types.Hash, error) {
		bytes, err := hex.DecodeString(param.Id)
		if err != nil {
			return types.ZeroHash, err
		}
		h, err := types.BytesToHash(bytes)
		if err != nil {
			return types.ZeroHash, err
		}
		return h, nil
	}); err == nil {
		return r.generateSend(p, cabi.MethodNameAirdropRewards)
	} else {
		return nil, err
	}
}

func (r *RewardsAPI) GetSendConfidantBlock(param *RewardsParam, sign *types.Signature) (*types.StateBlock, error) {
	if !r.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if param == nil || sign == nil {
		return nil, ErrParameterNil
	}

	if p, err := r.verifySign(param, sign, cabi.MethodNameUnsignedConfidantRewards, func(param *RewardsParam) (types.Hash, error) {
		h := types.HashData([]byte(param.Id))
		return h, nil
	}); err == nil {
		return r.generateSend(p, cabi.MethodNameConfidantRewards)
	} else {
		return nil, err
	}
}

func (r *RewardsAPI) verifySign(param *RewardsParam, sign *types.Signature, methodName string, fn func(param *RewardsParam) (types.Hash, error)) (*sendParam, error) {
	if param == nil || sign == nil {
		return nil, ErrParameterNil
	}
	if len(param.Id) == 0 || param.Amount.IsZero() || param.Self.IsZero() || param.To.IsZero() {
		return nil, errors.New("invalid param")
	}

	id, err := fn(param)
	if err != nil {
		return nil, err
	}

	if _, err := r.ledger.HasTokenMeta(param.Self, config.GasToken()); err != nil {
		r.logger.Errorf("token not found, %s %s", param.Self.String(), config.GasToken().String())
		return nil, err
	}

	am, _ := r.ledger.GetAccountMeta(param.Self)
	tm := am.Token(config.GasToken())
	txHash := tm.Header

	var rxHash types.Hash
	if b, err := r.ledger.HasTokenMeta(param.To, config.GasToken()); err != nil {
		r.logger.Errorf("token not found, %s %s", param.To.String(), config.GasToken().String())
		return nil, err
	} else {
		if b {
			tm, _ := r.ledger.GetTokenMeta(param.To, config.GasToken())
			rxHash = tm.Header
		} else {
			rxHash = types.ZeroHash
		}
	}

	p := &sendParam{
		RewardsParam: &cabi.RewardsParam{
			Id:         id,
			Beneficial: param.To,
			TxHeader:   txHash,
			RxHeader:   rxHash,
			Amount:     param.Amount.Int,
			Sign:       *sign,
		},
		self: &param.Self,
		//am:   am,
		tm: tm,
	}

	r.logger.Debugf("verify %s %s %s %s %s", methodName, param.To.String(), txHash.String(), rxHash.String(), param.Amount.Int)

	if b, err := p.RewardsParam.Verify(param.Self, methodName); b && err == nil {
		return p, nil
	} else {
		return nil, err
	}
}

func (r *RewardsAPI) generateSend(param *sendParam, methodName string) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	povHeader, err := r.ledger.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	if singedData, err := param.ToSignedABI(methodName); err == nil {
		return &types.StateBlock{
			Type:    types.ContractSend,
			Token:   param.tm.Type,
			Address: *param.self,
			Balance: param.tm.Balance.Sub(types.Balance{Int: param.Amount}),
			//Vote:           types.ZeroBalance,
			//Network:        types.ZeroBalance,
			//Oracle:         types.ZeroBalance,
			//Storage:        types.ZeroBalance,
			Previous:       param.tm.Header,
			Link:           types.Hash(contractaddress.RewardsAddress),
			Representative: param.tm.Representative,
			Data:           singedData,
			PoVHeight:      povHeader.GetHeight(),
			Timestamp:      common.TimeNow().Unix(),
		}, nil
	} else {
		return nil, err
	}
}

func (r *RewardsAPI) GetReceiveRewardBlock(send *types.Hash) (*types.StateBlock, error) {
	if send == nil {
		return nil, ErrParameterNil
	}
	if !r.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	blk, err := r.ledger.GetStateBlock(*send)
	if err != nil {
		return nil, err
	}

	rev := &types.StateBlock{Timestamp: common.TimeNow().Unix()}

	var result []*contract.ContractBlock

	vmContext := vmstore.NewVMContext(r.ledger, &contractaddress.RewardsAddress)
	if r.IsAirdropRewards(blk.Data) {
		result, err = r.rewards.DoReceive(vmContext, rev, blk)
	} else {
		result, err = r.confidantRewards.DoReceive(vmContext, rev, blk)
	}
	if err == nil {
		if len(result) > 0 {
			povHeader, err := r.ledger.GetLatestPovHeader()
			if err != nil {
				return nil, fmt.Errorf("get pov header error: %s", err)
			}
			rev.PoVHeight = povHeader.GetHeight()
			h := vmstore.TrieHash(result[0].VMContext)
			if h != nil {
				rev.Extra = h
			}
			return rev, nil
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("can not generate reward block, %s", err)
	}
}

func (r *RewardsAPI) IsAirdropRewards(data []byte) bool {
	if s, err := checkContractMethod(data); err == nil && s == cabi.MethodNameAirdropRewards {
		return true
	}
	return false
}

func (r *RewardsAPI) GetTotalRewards(txId string) (*big.Int, error) {
	return cabi.GetTotalRewards(r.ledger, txId)
}

func (r *RewardsAPI) GetRewardsDetail(txId string) ([]*cabi.RewardsInfo, error) {
	return cabi.GetRewardsDetail(r.ledger, txId)
}

func (r *RewardsAPI) GetConfidantRewards(confidant types.Address) (map[string]*big.Int, error) {
	return cabi.GetConfidantRewords(r.ledger, confidant)
}

func (r *RewardsAPI) GetConfidantRewordsDetail(confidant types.Address) (map[string][]*cabi.RewardsInfo, error) {
	return cabi.GetConfidantRewordsDetail(r.ledger, confidant)
}

func checkContractMethod(data []byte) (string, error) {
	if name, b, err := contract.GetChainContractName(contractaddress.RewardsAddress, data); b {
		return name, nil
	} else {
		return "", err
	}
}
