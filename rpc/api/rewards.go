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
	am   *types.AccountMeta
	tm   *types.TokenMeta
}

func NewRewardsApi(l *ledger.Ledger) *RewardsApi {
	return &RewardsApi{ledger: l, logger: log.NewLogger("api_rewards"),
		rewards: &contract.AirdropRewords{}, confidantRewards: &contract.ConfidantRewards{}}
}

type RewardsParam struct {
	Id     string         `json:"Id"`
	Amount types.Balance  `json:"amount"`
	Self   *types.Address `json:"self"`
	To     *types.Address `json:"to"`
}

func (r *RewardsApi) GetUnsignedRewardData(param *RewardsParam) (types.Hash, error) {
	return r.generateHash(param, cabi.MethodNameUnsignedAirdropRewards, func(param *RewardsParam) (bytes []byte, e error) {
		return hex.DecodeString(param.Id)
	})
}

func (r *RewardsApi) GetUnsignedConfidantData(param *RewardsParam) (types.Hash, error) {
	return r.generateHash(param, cabi.MethodNameUnsignedConfidantRewards, func(param *RewardsParam) (bytes []byte, e error) {
		h := types.HashData([]byte(param.Id))
		return h[:], nil
	})
}

func (r *RewardsApi) generateHash(param *RewardsParam, methodName string, fn func(param *RewardsParam) ([]byte, error)) (types.Hash, error) {
	if len(param.Id) == 0 || param.Amount.IsZero() || param.Self.IsZero() || param.To.IsZero() {
		return types.ZeroHash, errors.New("invalid param")
	}

	id, err := fn(param)
	if err != nil {
		return types.ZeroHash, err
	}
	h1 := new(types.Hash)
	h2 := new(types.Hash)

	if _, err := r.ledger.HasTokenMeta(*param.Self, common.GasToken()); err != nil {
		return types.ZeroHash, err
	} else {
		tm, _ := r.ledger.GetTokenMeta(*param.Self, common.GasToken())
		h1 = &tm.Header
	}

	if b, err := r.ledger.HasTokenMeta(*param.To, common.GasToken()); err != nil {
		return types.ZeroHash, err
	} else {
		if b {
			tm, _ := r.ledger.GetTokenMeta(*param.Self, common.GasToken())
			h2 = &tm.Header
		} else {
			h2 = &types.ZeroHash
		}
	}

	if data, err := cabi.RewardsABI.PackMethod(methodName, id, param.To, h1[:], h2[:], param.Amount.Int); err == nil {
		return types.HashData(data), nil
	} else {
		return types.ZeroHash, err
	}
}

func (r *RewardsApi) GetSendRewardBlock(param *RewardsParam, sign *types.Signature) (*types.StateBlock, error) {
	if p, err := r.verifySign(param, sign, cabi.MethodNameUnsignedAirdropRewards, func(param *RewardsParam) (bytes []byte, e error) {
		h := types.HashData([]byte(param.Id))
		return h[:], nil
	}); err == nil {
		return r.generateSend(p, cabi.MethodNameAirdropRewards)
	} else {
		return nil, err
	}
}

func (r *RewardsApi) GetSendConfidantBlock(param *RewardsParam, sign *types.Signature) (*types.StateBlock, error) {
	if p, err := r.verifySign(param, sign, cabi.MethodNameUnsignedConfidantRewards, func(param *RewardsParam) (bytes []byte, e error) {
		h := types.HashData([]byte(param.Id))
		return h[:], nil
	}); err == nil {
		return r.generateSend(p, cabi.MethodNameConfidantRewards)
	} else {
		return nil, err
	}
}

func (r *RewardsApi) verifySign(param *RewardsParam, sign *types.Signature, methodName string, fn func(param *RewardsParam) ([]byte, error)) (*sendParam, error) {
	if len(param.Id) == 0 || param.Amount.IsZero() || param.Self.IsZero() || param.To.IsZero() {
		return nil, errors.New("invalid param")
	}

	id, err := fn(param)
	if err != nil {
		return nil, err
	}

	h1 := new(types.Hash)
	h2 := new(types.Hash)

	if _, err := r.ledger.HasTokenMeta(*param.Self, common.GasToken()); err != nil {
		return nil, err
	}

	am, _ := r.ledger.GetAccountMeta(*param.Self)
	tm := am.Token(common.GasToken())
	h1 = &tm.Header

	if b, err := r.ledger.HasTokenMeta(*param.To, common.GasToken()); err != nil {
		return nil, err
	} else {
		if b {
			tm, _ := r.ledger.GetTokenMeta(*param.Self, common.GasToken())
			h2 = &tm.Header
		} else {
			h2 = &types.ZeroHash
		}
	}
	p := &sendParam{
		RewardsParam: &cabi.RewardsParam{
			Id:         id,
			Beneficial: param.To,
			TxHeader:   h1[:],
			RxHeader:   h2[:],
			Amount:     param.Amount.Int,
			Sign:       sign[:],
		},
		self: param.Self,
		am:   am,
		tm:   tm,
	}

	if data, err := cabi.RewardsABI.PackMethod(methodName, id, param.To, h1[:], h2[:], param.Amount.Int); err == nil {
		h := types.HashData(data)
		if param.Self.Verify(h[:], sign[:]) {
			return p, nil
		} else {
			return nil, fmt.Errorf("invalid sign[%s] of hash[%s]", sign.String(), h.String())
		}
	} else {
		return nil, err
	}
}

func (r *RewardsApi) generateSend(param *sendParam, methodName string) (*types.StateBlock, error) {
	if singedData, err := cabi.RewardsABI.PackMethod(methodName, param.Id, param.Beneficial, param.TxHeader, param.RxHeader, param.Amount, param.Sign); err == nil {
		return &types.StateBlock{
			Type:           types.ContractSend,
			Token:          param.tm.Type,
			Address:        *param.self,
			Balance:        param.tm.Balance.Sub(types.Balance{Int: param.Amount}),
			Vote:           param.am.CoinVote,
			Network:        param.am.CoinNetwork,
			Oracle:         param.am.CoinOracle,
			Storage:        param.am.CoinStorage,
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
	blk, err := r.ledger.GetStateBlock(*send)
	if err != nil {
		return nil, err
	}

	rev := &types.StateBlock{}

	if blocks, err := r.rewards.DoReceive(vmstore.NewVMContext(r.ledger), rev, blk); err == nil {
		if len(blocks) > 0 {
			rev.Timestamp = common.TimeNow().UTC().Unix()
			h := blocks[0].VMContext.Cache.Trie().Hash()
			rev.Extra = *h
			return rev, nil
		}
	} else {
		return nil, err
	}

	return nil, errors.New("can not generate airdrop reward block")
}

func (r *RewardsApi) GetReceiveConfidantBlock(send *types.Hash) (*types.StateBlock, error) {
	blk, err := r.ledger.GetStateBlock(*send)
	if err != nil {
		return nil, err
	}

	rev := &types.StateBlock{}

	if blocks, err := r.confidantRewards.DoReceive(vmstore.NewVMContext(r.ledger), rev, blk); err == nil {
		if len(blocks) > 0 {
			rev.Timestamp = common.TimeNow().UTC().Unix()
			h := blocks[0].VMContext.Cache.Trie().Hash()
			rev.Extra = *h
			return rev, nil
		}
	} else {
		return nil, err
	}

	return nil, errors.New("can not generate confidant reward block")
}

func (r *RewardsApi) GetTotalRewards(txId string) (*big.Int, error) {
	return cabi.GetTotalRewards(vmstore.NewVMContext(r.ledger), txId)
}

func (r *RewardsApi) GetConfidantRewords(confidant types.Address) (map[string]*big.Int, error) {
	return cabi.GetConfidantRewords(vmstore.NewVMContext(r.ledger), confidant)
}
