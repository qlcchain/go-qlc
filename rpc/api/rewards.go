/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type RewardsApi struct {
	logger           *zap.SugaredLogger
	ledger           *ledger.Ledger
	rewards          *contract.AirdropRewords
	confidantRewards *contract.ConfidantRewards
}

type sendParam struct {
	*cabi.RewardsParam
	self *types.Address
	//am   *types.AccountMeta
	tm *types.TokenMeta
}

func NewRewardsApi(l *ledger.Ledger) *RewardsApi {
	return &RewardsApi{ledger: l, logger: log.NewLogger("api_rewards"),
		rewards: &contract.AirdropRewords{}, confidantRewards: &contract.ConfidantRewards{}}
}

type RewardsParam struct {
	Id     string        `json:"Id"`
	Amount types.Balance `json:"amount"`
	Self   types.Address `json:"self"`
	To     types.Address `json:"to"`
}

func (r *RewardsApi) GetUnsignedRewardData(param *RewardsParam) (types.Hash, error) {
	if param == nil {
		return types.ZeroHash, ErrParameterNil
	}
	return r.generateHash(param, cabi.MethodNameUnsignedAirdropRewards, func(param *RewardsParam) (bytes []byte, e error) {
		return hex.DecodeString(param.Id)
	})
}

func (r *RewardsApi) GetUnsignedConfidantData(param *RewardsParam) (types.Hash, error) {
	if param == nil {
		return types.ZeroHash, ErrParameterNil
	}
	return r.generateHash(param, cabi.MethodNameUnsignedConfidantRewards, func(param *RewardsParam) (bytes []byte, e error) {
		h := types.HashData([]byte(param.Id))
		return h[:], nil
	})
}

func (r *RewardsApi) generateHash(param *RewardsParam, methodName string, fn func(param *RewardsParam) ([]byte, error)) (types.Hash, error) {
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

	if tm, err := r.ledger.GetTokenMeta(param.Self, common.GasToken()); tm != nil && err == nil {
		txHash = tm.Header
	} else {
		return types.ZeroHash, err
	}

	if b, err := r.ledger.HasTokenMeta(param.To, common.GasToken()); err != nil {
		return types.ZeroHash, err
	} else {
		if b {
			tm, _ := r.ledger.GetTokenMeta(param.To, common.GasToken())
			rxHash = tm.Header
		} else {
			rxHash = types.ZeroHash
		}
	}

	r.logger.Debugf("%s %s %s %s %s", methodName, param.To.String(), txHash.String(), rxHash.String(), param.Amount.Int)

	if data, err := cabi.RewardsABI.PackMethod(methodName, id, param.To, txHash, rxHash, param.Amount.Int); err == nil {
		return types.HashData(data), nil
	} else {
		return types.ZeroHash, err
	}
}

func (r *RewardsApi) GetSendRewardBlock(param *RewardsParam, sign *types.Signature) (*types.StateBlock, error) {
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

func (r *RewardsApi) GetSendConfidantBlock(param *RewardsParam, sign *types.Signature) (*types.StateBlock, error) {
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

func (r *RewardsApi) verifySign(param *RewardsParam, sign *types.Signature, methodName string, fn func(param *RewardsParam) (types.Hash, error)) (*sendParam, error) {
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

	if _, err := r.ledger.HasTokenMeta(param.Self, common.GasToken()); err != nil {
		return nil, err
	}

	am, _ := r.ledger.GetAccountMeta(param.Self)
	tm := am.Token(common.GasToken())
	txHash := tm.Header

	var rxHash types.Hash
	if b, err := r.ledger.HasTokenMeta(param.To, common.GasToken()); err != nil {
		return nil, err
	} else {
		if b {
			tm, _ := r.ledger.GetTokenMeta(param.To, common.GasToken())
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

func (r *RewardsApi) generateSend(param *sendParam, methodName string) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}
	if singedData, err := cabi.RewardsABI.PackMethod(methodName, param.Id, param.Beneficial, param.TxHeader, param.RxHeader, param.Amount, param.Sign); err == nil {
		return &types.StateBlock{
			Type:           types.ContractSend,
			Token:          param.tm.Type,
			Address:        *param.self,
			Balance:        param.tm.Balance.Sub(types.Balance{Int: param.Amount}),
			Vote:           types.ZeroBalance,
			Network:        types.ZeroBalance,
			Oracle:         types.ZeroBalance,
			Storage:        types.ZeroBalance,
			Previous:       param.tm.Header,
			Link:           types.Hash(types.RewardsAddress),
			Representative: param.tm.Representative,
			Data:           singedData,
			Timestamp:      common.TimeNow().UTC().Unix(),
		}, nil
	} else {
		return nil, err
	}

}

func (r *RewardsApi) GetReceiveRewardBlock(send *types.Hash) (*types.StateBlock, error) {
	if send == nil {
		return nil, ErrParameterNil
	}
	blk, err := r.ledger.GetStateBlock(*send)
	if err != nil {
		return nil, err
	}

	rev := &types.StateBlock{}

	var result []*contract.ContractBlock

	vmContext := vmstore.NewVMContext(r.ledger)
	if r.IsAirdropRewards(blk.Data) {
		result, err = r.rewards.DoReceive(vmContext, rev, blk)
	} else {
		result, err = r.confidantRewards.DoReceive(vmContext, rev, blk)
	}
	if err == nil {
		if len(result) > 0 {
			rev.Timestamp = common.TimeNow().UTC().Unix()
			h := result[0].VMContext.Cache.Trie().Hash()
			if h != nil {
				rev.Extra = *h
			}
			return rev, nil
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("can not generate reward block, %s", err)
	}
}

func (r *RewardsApi) IsAirdropRewards(data []byte) bool {
	if s, err := checkContractMethod(data); err == nil && s == cabi.MethodNameAirdropRewards {
		return true
	}
	return false
}

func (r *RewardsApi) GetTotalRewards(txId string) (*big.Int, error) {
	return cabi.GetTotalRewards(vmstore.NewVMContext(r.ledger), txId)
}

func (r *RewardsApi) GetRewardsDetail(txId string) ([]*cabi.RewardsInfo, error) {
	return cabi.GetRewardsDetail(vmstore.NewVMContext(r.ledger), txId)
}

func (r *RewardsApi) GetConfidantRewards(confidant types.Address) (map[string]*big.Int, error) {
	return cabi.GetConfidantRewords(vmstore.NewVMContext(r.ledger), confidant)
}

func (r *RewardsApi) GetConfidantRewordsDetail(confidant types.Address) (map[string][]*cabi.RewardsInfo, error) {
	return cabi.GetConfidantRewordsDetail(vmstore.NewVMContext(r.ledger), confidant)
}

func checkContractMethod(data []byte) (string, error) {
	if name, b, err := contract.GetChainContractName(types.RewardsAddress, data); b {
		return name, nil
	} else {
		return "", err
	}
}
