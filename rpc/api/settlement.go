/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

type SettlementAPI struct {
	logger         *zap.SugaredLogger
	l              *ledger.Ledger
	createContract *contract.CreateContract
	signContract   *contract.SignContract
	cdrContract    *contract.ProcessCDR
	cc             *context.ChainContext
}

// SignContractParam for confirm contract which created by PartyA
type SignContractParam struct {
	cabi.SignContract
	Address types.Address // PartyB address
}

func NewSettlement(l *ledger.Ledger, cc *context.ChainContext) *SettlementAPI {
	return &SettlementAPI{
		logger:         log.NewLogger("rpc/settlement"),
		l:              l,
		createContract: &contract.CreateContract{},
		signContract:   &contract.SignContract{},
		cdrContract:    &contract.ProcessCDR{},
		cc:             cc,
	}
}

// ToAddress convert CreateContractParam to smart contract address
func (s *SettlementAPI) ToAddress(param *cabi.CreateContractParam) (types.Address, error) {
	return param.Address()
}

// GetContractRewardsBlock generate create contract rewords block by contract send block hash
// @param send contract send block hash
// @return contract rewards block
func (s *SettlementAPI) GetContractRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, func(tx *types.StateBlock) (*types.StateBlock, error) {
		rev := &types.StateBlock{
			Timestamp: common.TimeNow().Unix(),
		}
		ctx := vmstore.NewVMContext(s.l)

		if r, err := s.createContract.DoReceive(ctx, rev, tx); err == nil {
			if len(r) > 0 {
				return r[0].Block, nil
			} else {
				return nil, errors.New("fail to generate create contract reward block")
			}
		} else {
			return nil, err
		}
	})
}

// GetCreateContractBlock
// generate ContractSend block to call smart contract for generating settlement contract as PartyA
// @param param smart contract params
// @return state block to be processed
func (s *SettlementAPI) GetCreateContractBlock(param *cabi.CreateContractParam) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if param == nil {
		return nil, errors.New("invalid input param")
	}

	if isVerified, err := param.Verify(); err != nil {
		return nil, err
	} else if !isVerified {
		return nil, errors.New("invalid input param")
	}
	ctx := vmstore.NewVMContext(s.l)

	if tm, err := ctx.GetTokenMeta(param.PartyA, common.GasToken()); err != nil {
		return nil, err
	} else {
		balance, err := param.Balance()
		if err != nil {
			return nil, err
		}
		if tm.Balance.Compare(balance) == types.BalanceCompSmaller {
			return nil, fmt.Errorf("not enough balance, [%s] of [%s]", balance.String(), tm.Balance.String())
		}

		if singedData, err := param.ToABI(); err == nil {
			sb := &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        param.PartyA,
				Balance:        tm.Balance.Sub(balance),
				Vote:           types.ZeroBalance,
				Network:        types.ZeroBalance,
				Oracle:         types.ZeroBalance,
				Storage:        types.ZeroBalance,
				Previous:       param.Previous,
				Link:           types.Hash(types.SettlementAddress),
				Representative: tm.Representative,
				Data:           singedData,
				Timestamp:      common.TimeNow().Unix(),
			}

			h := ctx.Cache.Trie().Hash()
			if h != nil {
				povHeader, err := s.l.GetLatestPovHeader()
				if err != nil {
					return nil, fmt.Errorf("get pov header error: %s", err)
				}
				sb.PoVHeight = povHeader.GetHeight()
				sb.Extra = *h
			}
			return sb, nil
		} else {
			return nil, err
		}
	}
}

// GetSignContractBlock
// generate ContractSend block to call smart contract for signing settlement contract as PartyB
// smart contract address = hash(*abi.CreateContractParam)
// param.SignatureB = sign(smart contract address + param.ConfirmDate)
// @param param sign settlement contract param created by PartyA
// @return state block(without signature) to be processed
func (s *SettlementAPI) GetSignContractBlock(param *SignContractParam) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if param == nil {
		return nil, errors.New("invalid input param")
	}

	if isVerified, err := param.SignContract.Verify(param.Address); err != nil {
		return nil, err
	} else if !isVerified {
		return nil, errors.New("invalid input param")
	}
	ctx := vmstore.NewVMContext(s.l)

	if storage, err := ctx.GetStorage(types.SettlementAddress[:], param.ContractAddress[:]); err != nil {
		return nil, err
	} else {
		if len(storage) > 0 {
			if cp, err := cabi.ParseContractParam(storage); err != nil {
				return nil, err
			} else {
				// verify partyB
				if cp.PartyB != param.Address {
					return nil, fmt.Errorf("invalid partyB, exp: %s, act: %s", cp.PartyB.String(), param.Address.String())
				}
			}
		} else {
			return nil, fmt.Errorf("invalid saved contract data of %s", param.ContractAddress.String())
		}
	}

	if tm, err := ctx.GetTokenMeta(param.Address, common.GasToken()); err != nil {
		return nil, err
	} else {
		if singedData, err := param.SignContract.ToABI(); err == nil {
			sb := &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        param.Address,
				Balance:        tm.Balance,
				Vote:           types.ZeroBalance,
				Network:        types.ZeroBalance,
				Oracle:         types.ZeroBalance,
				Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(types.SettlementAddress),
				Representative: tm.Representative,
				Data:           singedData,
				Timestamp:      common.TimeNow().Unix(),
			}

			h := ctx.Cache.Trie().Hash()
			if h != nil {
				povHeader, err := s.l.GetLatestPovHeader()
				if err != nil {
					return nil, fmt.Errorf("get pov header error: %s", err)
				}
				sb.PoVHeight = povHeader.GetHeight()
				sb.Extra = *h
			}
			return sb, nil
		} else {
			return nil, err
		}
	}
}

// GetSignRewardsBlock generate create contract rewords block by contract send block hash
// @param send contract send block hash
// @return contract rewards block
func (s *SettlementAPI) GetSignRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, func(tx *types.StateBlock) (*types.StateBlock, error) {
		rev := &types.StateBlock{
			Timestamp: common.TimeNow().Unix(),
		}
		ctx := vmstore.NewVMContext(s.l)

		if r, err := s.signContract.DoReceive(ctx, rev, tx); err == nil {
			if len(r) > 0 {
				return r[0].Block, nil
			} else {
				return nil, errors.New("fail to generate sign contract reward block")
			}
		} else {
			return nil, err
		}
	})
}

// SettlementContract settlement contract for RPC
type SettlementContract struct {
	cabi.ContractParam               //contract params
	Address            types.Address // settlement smart contract address
}

// GetContractsByAddress query all related settlement contracts info by address
// @param addr user qlcchain address
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return all settlement contract
func (s *SettlementAPI) GetContractsByAddress(addr *types.Address, count int, offset *int) ([]*SettlementContract, error) {
	return s.queryContractsByAddress(count, offset, func() (params []*cabi.ContractParam, err error) {
		ctx := vmstore.NewVMContext(s.l)
		return cabi.GetContractsByAddress(ctx, addr)
	})
}

// GetContractsAsPartyA query all settlement contracts as Party A info by address
// @param addr user qlcchain address
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return all settlement contract as PartyA
func (s *SettlementAPI) GetContractsAsPartyA(addr *types.Address, count int, offset *int) ([]*SettlementContract, error) {
	return s.queryContractsByAddress(count, offset, func() (params []*cabi.ContractParam, err error) {
		ctx := vmstore.NewVMContext(s.l)
		return cabi.GetContractsIDByAddressAsPartyA(ctx, addr)
	})
}

// GetContractsAsPartyA query all settlement contracts as Party B info by address
// @param addr user qlcchain address
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return all settlement contract as PartyB
func (s *SettlementAPI) GetContractsAsPartyB(addr *types.Address, count int, offset *int) ([]*SettlementContract, error) {
	return s.queryContractsByAddress(count, offset, func() (params []*cabi.ContractParam, err error) {
		ctx := vmstore.NewVMContext(s.l)
		return cabi.GetContractsIDByAddressAsPartyB(ctx, addr)
	})
}

// TODO: to be confirmed
type CDRParam struct {
	cabi.CDRParam
}

// GetProcessCDRBlock save CDR data for the settlement
// @param addr settlement smart contract address
// @param params CDR params to be processed
// @return contract send block to be processed
func (s *SettlementAPI) GetProcessCDRBlock(addr *types.Address, params []*CDRParam) ([]*types.StateBlock, error) {
	return nil, nil
}

// TODO: report data te be confirmed
type SettlementRecord struct {
	Address                  types.Address // smart contract address
	Customer                 string
	CustomerSr               string
	Country                  string
	Operator                 string
	ServiceId                string
	MCC                      int
	MNC                      int
	Currency                 string
	UnitPrice                float64
	SumOfBillableSMSCustomer int64
	SumOfTOTPrice            float64
}

// TODO: query condition to be confirmed, maybe more interface?
// GenerateReport Generate reports for specified contracts based on start and end date
// @param addr user qlcchain address
// @param start report start date
// @param end report end data
// @return settlement report
func (s *SettlementAPI) GenerateReport(addr *types.Address, start, end time.Time) (*[]SettlementRecord, error) {
	return nil, nil
}

func (s *SettlementAPI) getContractRewardsBlock(send *types.Hash,
	fn func(tx *types.StateBlock) (*types.StateBlock, error)) (*types.StateBlock, error) {
	if send == nil {
		return nil, ErrParameterNil
	}
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	blk, err := s.l.GetStateBlock(*send)
	if err != nil {
		return nil, err
	}

	if rx, err := fn(blk); err == nil {
		povHeader, err := s.l.GetLatestPovHeader()
		if err != nil {
			return nil, fmt.Errorf("get pov header error: %s", err)
		}
		rx.PoVHeight = povHeader.GetHeight()
		return rx, nil
	} else {
		return nil, err
	}
}

func (s *SettlementAPI) queryContractsByAddress(count int, offset *int, fn func() ([]*cabi.ContractParam, error)) ([]*SettlementContract, error) {
	contracts, err := fn()
	if err != nil {
		return nil, err
	}

	size := len(contracts)
	if count <= 0 {
		return nil, fmt.Errorf("invalid count: %d", count)
	}

	start := 0
	if offset != nil && *offset >= 0 {
		start = *offset
	}
	end := start + count
	if end >= size {
		end = size - 1
	}

	var result []*SettlementContract

	for _, c := range contracts[start:end] {
		if address, err := c.Address(); err == nil {
			result = append(result, &SettlementContract{
				ContractParam: *c,
				Address:       address,
			})
		} else {
			s.logger.Error(err)
		}
	}

	return result, nil
}
