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
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	errInvalidParam = errors.New("invalid input param")
)

type SettlementAPI struct {
	logger            *zap.SugaredLogger
	l                 ledger.Store
	createContract    *contract.CreateContract
	signContract      *contract.SignContract
	cdrContract       *contract.ProcessCDR
	addPreStop        *contract.AddPreStop
	removePreStop     *contract.RemovePreStop
	updatePreStop     *contract.UpdatePreStop
	addNextStop       *contract.AddNextStop
	removeNextStop    *contract.RemoveNextStop
	updateNextStop    *contract.UpdateNextStop
	terminateContract *contract.TerminateContract
	cc                *context.ChainContext
}

// SignContractParam for confirm contract which created by PartyA
type SignContractParam struct {
	ContractAddress types.Address `json:"contractAddress"`
	Address         types.Address `json:"address"`
}

func NewSettlement(l ledger.Store, cc *context.ChainContext) *SettlementAPI {
	return &SettlementAPI{
		logger:         log.NewLogger("rpc/settlement"),
		l:              l,
		createContract: &contract.CreateContract{},
		signContract:   &contract.SignContract{},
		cdrContract:    &contract.ProcessCDR{},
		addPreStop:     &contract.AddPreStop{},
		removePreStop:  &contract.RemovePreStop{},
		updatePreStop:  &contract.UpdatePreStop{},
		addNextStop:    &contract.AddNextStop{},
		removeNextStop: &contract.RemoveNextStop{},
		updateNextStop: &contract.UpdateNextStop{},
		cc:             cc,
	}
}

// ToAddress convert CreateContractParam to smart contract address
func (s *SettlementAPI) ToAddress(param *cabi.CreateContractParam) (types.Address, error) {
	return param.Address()
}

// GetSettlementRewardsBlock generate create contract rewords block by contract send block hash
// @param send contract send block hash
// @return contract rewards block
func (s *SettlementAPI) GetSettlementRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	if send == nil || send.IsZero() {
		return nil, ErrParameterNil
	}
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	blk, err := s.l.GetStateBlock(*send)
	if err != nil {
		return nil, err
	}

	rev := &types.StateBlock{
		Timestamp: common.TimeNow().Unix(),
	}
	povHeader, err := s.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	rev.PoVHeight = povHeader.GetHeight()

	ctx := vmstore.NewVMContext(s.l)

	if c, ok, err := contract.GetChainContract(types.SettlementAddress, blk.Data); c != nil && ok && err == nil {
		if r, err := c.DoReceive(ctx, rev, blk); err == nil {
			if len(r) > 0 {
				return r[0].Block, nil
			} else {
				return nil, errors.New("fail to generate contract reward block")
			}
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("can not find settlement contract function from block, %s", send.String())
	}
}

type CreateContractParam struct {
	PartyA    cabi.Contractor        `json:"partyA"`
	PartyB    cabi.Contractor        `json:"partyB"`
	Services  []cabi.ContractService `json:"services"`
	StartDate int64                  `json:"startDate"`
	EndDate   int64                  `json:"endDate"`
}

// GetCreateContractBlock
// generate ContractSend block to call smart contract for generating settlement contract as PartyA
// @param param smart contract params
// @return state block to be processed
func (s *SettlementAPI) GetCreateContractBlock(param *CreateContractParam) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if param == nil {
		return nil, errInvalidParam
	}

	now := time.Now().Unix()

	// WTF
	//if param.StartDate < now {
	//	return nil, fmt.Errorf("invalid start date, should bigger than %d, got: %d", now, param.StartDate)
	//}
	//
	//if param.EndDate < now {
	//	return nil, fmt.Errorf("invalid start end, should bigger than %d, got: %d", now, param.EndDate)
	//}

	ctx := vmstore.NewVMContext(s.l)

	addr := param.PartyA.Address
	if tm, err := ctx.Ledger.GetTokenMeta(addr, config.GasToken()); err != nil {
		return nil, err
	} else {
		createParam := &cabi.CreateContractParam{
			PartyA:    param.PartyA,
			PartyB:    param.PartyB,
			Previous:  tm.Header,
			Services:  param.Services,
			SignDate:  now,
			StartDate: param.StartDate,
			EndDate:   param.EndDate,
		}
		if isVerified, err := createParam.Verify(); err != nil {
			return nil, err
		} else if !isVerified {
			return nil, errInvalidParam
		}

		balance, err := createParam.Balance()
		if err != nil {
			return nil, err
		}
		if tm.Balance.Compare(balance) == types.BalanceCompSmaller {
			return nil, fmt.Errorf("not enough balance, [%s] of [%s]", balance.String(), tm.Balance.String())
		}

		if singedData, err := createParam.ToABI(); err == nil {
			povHeader, err := s.l.GetLatestPovHeader()
			if err != nil {
				return nil, fmt.Errorf("get pov header error: %s", err)
			}
			sb := &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        addr,
				Balance:        tm.Balance.Sub(balance),
				Vote:           types.ZeroBalance,
				Network:        types.ZeroBalance,
				Oracle:         types.ZeroBalance,
				Storage:        types.ZeroBalance,
				Previous:       createParam.Previous,
				Link:           types.Hash(types.SettlementAddress),
				Representative: tm.Representative,
				Data:           singedData,
				PoVHeight:      povHeader.GetHeight(),
				Timestamp:      common.TimeNow().Unix(),
			}

			if _, _, err := s.createContract.ProcessSend(ctx, sb); err != nil {
				return nil, err
			}
			h := ctx.Cache.Trie().Hash()
			if h != nil {
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
// @param param sign settlement contract param created by PartyA
// @return state block(without signature) to be processed
func (s *SettlementAPI) GetSignContractBlock(param *SignContractParam) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if param == nil {
		return nil, errInvalidParam
	}

	signParam := &cabi.SignContractParam{
		ContractAddress: param.ContractAddress,
		ConfirmDate:     time.Now().Unix(),
	}

	if isVerified, err := signParam.Verify(); err != nil {
		return nil, err
	} else if !isVerified {
		return nil, errInvalidParam
	}
	ctx := vmstore.NewVMContext(s.l)

	if tm, err := ctx.Ledger.GetTokenMeta(param.Address, config.GasToken()); err != nil {
		return nil, err
	} else {
		if singedData, err := signParam.ToABI(); err == nil {
			povHeader, err := s.l.GetLatestPovHeader()
			if err != nil {
				return nil, fmt.Errorf("get pov header error: %s", err)
			}
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
				PoVHeight:      povHeader.GetHeight(),
				Timestamp:      common.TimeNow().Unix(),
			}
			if _, _, err := s.signContract.ProcessSend(ctx, sb); err != nil {
				return nil, err
			}

			h := ctx.Cache.Trie().Hash()
			if h != nil {
				sb.Extra = *h
			}

			return sb, nil
		} else {
			return nil, err
		}
	}
}

type StopParam struct {
	cabi.StopParam
	Address types.Address
}

type UpdateStopParam struct {
	cabi.UpdateStopParam
	Address types.Address
}

func (s *SettlementAPI) handleStopAction(addr types.Address, verifier func() error, abi func() ([]byte, error), c contract.Contract) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if addr.IsZero() {
		return nil, errors.New("invalid address")
	}

	if err := verifier(); err != nil {
		return nil, err
	}

	ctx := vmstore.NewVMContext(s.l)

	if tm, err := ctx.Ledger.GetTokenMeta(addr, config.GasToken()); err != nil {
		return nil, err
	} else {
		if singedData, err := abi(); err == nil {
			povHeader, err := s.l.GetLatestPovHeader()
			if err != nil {
				return nil, fmt.Errorf("get pov header error: %s", err)
			}
			sb := &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        addr,
				Balance:        tm.Balance,
				Vote:           types.ZeroBalance,
				Network:        types.ZeroBalance,
				Oracle:         types.ZeroBalance,
				Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(types.SettlementAddress),
				Representative: tm.Representative,
				Data:           singedData,
				PoVHeight:      povHeader.GetHeight(),
				Timestamp:      common.TimeNow().Unix(),
			}

			if _, _, err := c.ProcessSend(ctx, sb); err != nil {
				return nil, err
			}

			h := ctx.Cache.Trie().Hash()
			if h != nil {
				sb.Extra = *h
			}

			return sb, nil
		} else {
			return nil, err
		}
	}
}

func (s *SettlementAPI) GetAddPreStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || len(param.StopName) == 0 {
			return errInvalidParam
		}
		return nil
	}, func() (bytes []byte, err error) {
		p := cabi.StopParam{
			StopName:        param.StopName,
			ContractAddress: param.ContractAddress,
		}
		return p.ToABI(cabi.MethodNameAddPreStop)
	}, s.addPreStop)
}

func (s *SettlementAPI) GetRemovePreStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || len(param.StopName) == 0 {
			return errInvalidParam
		}
		return nil
	}, func() (bytes []byte, err error) {
		p := cabi.StopParam{
			StopName:        param.StopName,
			ContractAddress: param.ContractAddress,
		}
		return p.ToABI(cabi.MethodNameRemovePreStop)
	}, s.removePreStop)
}

func (s *SettlementAPI) GetUpdatePreStopBlock(param *UpdateStopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.Address.IsZero() || len(param.StopName) == 0 || len(param.New) == 0 || param.StopName == param.New {
			return errInvalidParam
		}
		return nil
	}, func() (bytes []byte, err error) {
		p := cabi.UpdateStopParam{
			ContractAddress: param.ContractAddress,
			StopName:        param.StopName,
			New:             param.New,
		}
		return p.ToABI(cabi.MethodNameUpdatePreStop)
	}, s.updatePreStop)
}

func (s *SettlementAPI) GetAddNextStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.ContractAddress.IsZero() || len(param.StopName) == 0 {
			return errInvalidParam
		}
		return nil
	}, func() (bytes []byte, err error) {
		p := cabi.StopParam{
			ContractAddress: param.ContractAddress,
			StopName:        param.StopName,
		}
		return p.ToABI(cabi.MethodNameAddNextStop)
	}, s.addNextStop)
}

func (s *SettlementAPI) GetRemoveNextStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.ContractAddress.IsZero() || len(param.StopName) == 0 {
			return errInvalidParam
		}
		return nil
	}, func() (bytes []byte, err error) {
		p := cabi.StopParam{
			ContractAddress: param.ContractAddress,
			StopName:        param.StopName,
		}
		return p.ToABI(cabi.MethodNameRemoveNextStop)
	}, s.removeNextStop)
}

func (s *SettlementAPI) GetUpdateNextStopBlock(param *UpdateStopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.ContractAddress.IsZero() || len(param.StopName) == 0 || len(param.New) == 0 || param.StopName == param.New {
			return errInvalidParam
		}
		return nil
	}, func() (bytes []byte, err error) {
		p := cabi.UpdateStopParam{
			ContractAddress: param.ContractAddress,
			StopName:        param.StopName,
			New:             param.New,
		}
		return p.ToABI(cabi.MethodNameUpdateNextStop)
	}, s.updateNextStop)
}

// SettlementContract settlement contract for RPC
type SettlementContract struct {
	cabi.ContractParam
	Address types.Address `json:"address"`
}

// GetAllContracts query all settlement contracts
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return all settlement contracts
func (s *SettlementAPI) GetAllContracts(count int, offset *int) ([]*SettlementContract, error) {
	return s.queryContractsByAddress(count, offset, func() (params []*cabi.ContractParam, err error) {
		ctx := vmstore.NewVMContext(s.l)
		return cabi.GetAllSettlementContract(ctx)
	})
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

// GetContractsByStatus query all related settlement contracts info by address and contract status
// @param addr user qlcchain address
// @param status contract status string
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return matched settlement contract
func (s *SettlementAPI) GetContractsByStatus(addr *types.Address, status string, count int, offset *int) ([]*SettlementContract, error) {
	state, err := cabi.ParseContractStatus(status)
	if err != nil {
		return nil, err
	}
	return s.queryContractsByAddress(count, offset, func() (params []*cabi.ContractParam, err error) {
		ctx := vmstore.NewVMContext(s.l)
		return cabi.GetContractsByStatus(ctx, addr, state)
	})
}

// GetExpiredContracts query all expired settlement contracts info by user's address
// @param addr user qlcchain address
// @param status contract status
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
// @return matched settlement contract
func (s *SettlementAPI) GetExpiredContracts(addr *types.Address, count int, offset *int) ([]*SettlementContract, error) {
	return s.queryContractsByAddress(count, offset, func() (params []*cabi.ContractParam, err error) {
		ctx := vmstore.NewVMContext(s.l)
		return cabi.GetExpiredContracts(ctx, addr)
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

// GetContractAddressByPartyANextStop query all settlement contracts as Party A info by address and NextStop
// @param addr user qlcchain address
// @param stopName PartyA nextStop
// @return contract Address
func (s *SettlementAPI) GetContractAddressByPartyANextStop(addr *types.Address, stopName string) (*types.Address, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GetContractsAddressByPartyANextStop(ctx, addr, stopName)
}

// GetContractAddressByPartyBPreStop query all settlement contracts as Party B info by address and PreStop
// @param addr user qlcchain address
// @param stopName PartyB preStop
// @return contract Address
func (s *SettlementAPI) GetContractAddressByPartyBPreStop(addr *types.Address, stopName string) (*types.Address, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GetContractsAddressByPartyBPreStop(ctx, addr, stopName)
}

// GetProcessCDRBlock save CDR data for the settlement
// @param addr user qlc address
// @param param CDR params to be processed
// @return contract send block without signature
func (s *SettlementAPI) GetProcessCDRBlock(addr *types.Address, param *cabi.CDRParam) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if err := param.Verify(); err != nil {
		return nil, err
	}

	ctx := vmstore.NewVMContext(s.l)

	if c, err := cabi.FindSettlementContract(ctx, addr, param); err != nil {
		return nil, err
	} else {
		if tm, err := ctx.Ledger.GetTokenMeta(*addr, config.GasToken()); err != nil {
			return nil, err
		} else {
			address, err := c.Address()
			if err != nil {
				return nil, err
			}
			param.ContractAddress = address

			if !c.IsAvailable() {
				return nil, fmt.Errorf("contract %s is invalid, please check contract status and start/end date",
					address.String())
			}

			if singedData, err := param.ToABI(); err == nil {
				povHeader, err := s.l.GetLatestPovHeader()
				if err != nil {
					return nil, fmt.Errorf("get pov header error: %s", err)
				}
				sb := &types.StateBlock{
					Type:           types.ContractSend,
					Token:          tm.Type,
					Address:        *addr,
					Balance:        tm.Balance,
					Vote:           types.ZeroBalance,
					Network:        types.ZeroBalance,
					Oracle:         types.ZeroBalance,
					Storage:        types.ZeroBalance,
					Previous:       tm.Header,
					Link:           types.Hash(types.SettlementAddress),
					Representative: tm.Representative,
					Data:           singedData,
					PoVHeight:      povHeader.GetHeight(),
					Timestamp:      common.TimeNow().Unix(),
				}

				if _, _, err := s.cdrContract.ProcessSend(ctx, sb); err != nil {
					return nil, err
				}

				h := ctx.Cache.Trie().Hash()
				if h != nil {
					sb.Extra = *h
				}
				return sb, nil
			} else {
				return nil, err
			}
		}
	}
}

type TerminateParam struct {
	cabi.TerminateParam
	Address types.Address
}

// GetTerminateContractBlock
// generate ContractSend block to call smart contract for terminating settlement contract
// @param param sign settlement contract param created by PartyA
// @return state block(without signature) to be processed
func (s *SettlementAPI) GetTerminateContractBlock(param *TerminateParam) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if param == nil {
		return nil, errInvalidParam
	}

	if err := param.Verify(); err != nil {
		return nil, err
	}
	ctx := vmstore.NewVMContext(s.l)

	if tm, err := ctx.Ledger.GetTokenMeta(param.Address, config.GasToken()); err != nil {
		return nil, err
	} else {
		if singedData, err := param.ToABI(); err == nil {
			povHeader, err := s.l.GetLatestPovHeader()
			if err != nil {
				return nil, fmt.Errorf("get pov header error: %s", err)
			}

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
				PoVHeight:      povHeader.GetHeight(),
				Timestamp:      common.TimeNow().Unix(),
			}

			if _, _, err := s.terminateContract.ProcessSend(ctx, sb); err != nil {
				return nil, err
			}

			h := ctx.Cache.Trie().Hash()
			if h != nil {
				sb.Extra = *h
			}

			return sb, nil
		} else {
			return nil, err
		}
	}
}

// GetCDRStatus get CDRstatus by settlement smart contract address and CDR hash
// @param addr settlement smart contract address
// @param hash CDR data hash
func (s *SettlementAPI) GetCDRStatus(addr *types.Address, hash types.Hash) (*cabi.CDRStatus, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GetCDRStatus(ctx, addr, hash)
}

// GetCDRStatus get CDRstatus by settlement smart contract address and CDR data
// @param addr settlement smart contract address
// @param index,sender,destination CDR data
func (s *SettlementAPI) GetCDRStatusByCdrData(addr *types.Address, index uint64, sender, destination string) (*cabi.CDRStatus, error) {
	hash, err := types.HashBytes(util.BE_Uint64ToBytes(index), []byte(sender), []byte(destination))
	if err != nil {
		return nil, err
	}
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GetCDRStatus(ctx, addr, hash)
}

func (s *SettlementAPI) GetCDRStatusByDate(addr *types.Address, start, end int64, count int, offset *int) ([]*cabi.CDRStatus, error) {
	ctx := vmstore.NewVMContext(s.l)
	if status, err := cabi.GetCDRStatusByDate(ctx, addr, start, end); err != nil {
		return nil, err
	} else {
		size := len(status)
		if size > 0 {
			sort.Slice(status, func(i, j int) bool {
				return sortCDRFun(status[i], status[j])
			})
		}
		start, end, err := calculateRange(size, count, offset)
		if err != nil {
			return nil, err
		}
		return status[start:end], nil
	}
}

// GetAllCDRStatus get all cdr status of the specific settlement smart contract
// @param addr settlement smart contract
// @param count max settlement contract records size
// @param offset offset of all settlement contract records(optional)
func (s *SettlementAPI) GetAllCDRStatus(addr *types.Address, count int, offset *int) ([]*cabi.CDRStatus, error) {
	ctx := vmstore.NewVMContext(s.l)

	if status, err := cabi.GetAllCDRStatus(ctx, addr); err != nil {
		return nil, err
	} else {
		size := len(status)
		if size > 0 {
			sort.Slice(status, func(i, j int) bool {
				return sortCDRFun(status[i], status[j])
			})
		}
		start, end, err := calculateRange(size, count, offset)
		if err != nil {
			return nil, err
		}
		return status[start:end], nil
	}
}

// GetSummaryReport generate summary report by smart contract address and start/end date
// @param addr settlement contract address
// @param start report start date (UTC unix time)
// @param end report end data (UTC unix time)
// @return summary report if error not exist
func (s *SettlementAPI) GetSummaryReport(addr *types.Address, start, end int64) (*cabi.SummaryResult, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GetSummaryReport(ctx, addr, start, end)
}

// GenerateInvoices Generate reports for specified contracts based on start and end date
// @param addr user qlcchain address
// @param start report start date (UTC unix time)
// @param end report end data (UTC unix time)
// @return settlement report
func (s *SettlementAPI) GenerateInvoices(addr *types.Address, start, end int64) ([]*cabi.InvoiceRecord, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GenerateInvoices(ctx, addr, start, end)
}

// GenerateInvoicesByContract generate invoice by settlement contract address
// @param addr settlement contract address
// @param start report start date (UTC unix time)
// @param end report end data (UTC unix time)
// @return settlement report
func (s *SettlementAPI) GenerateInvoicesByContract(addr *types.Address, start, end int64) ([]*cabi.InvoiceRecord, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GenerateInvoicesByContract(ctx, addr, start, end)
}

// GenerateMultiPartyInvoice generate multi-party invoice by settlement contract address
// @param addr settlement contract address
// @param start report start date (UTC unix time)
// @param end report end data (UTC unix time)
// @return settlement invoice
func (s *SettlementAPI) GenerateMultiPartyInvoice(addr *types.Address, start, end int64) ([]*cabi.InvoiceRecord, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GenerateMultiPartyInvoice(ctx, addr, start, end)
}

// GetPreStopNames get all previous stop names by user address
func (s *SettlementAPI) GetPreStopNames(addr *types.Address) ([]string, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GetPreStopNames(ctx, addr)
}

// GetNextStopNames get all next stop names by user address
func (s *SettlementAPI) GetNextStopNames(addr *types.Address) ([]string, error) {
	ctx := vmstore.NewVMContext(s.l)
	return cabi.GetNextStopNames(ctx, addr)
}

var sortCDRFun = func(cdr1, cdr2 *cabi.CDRStatus) bool {
	dt1, sender1, _, err := cdr1.ExtractID()
	if err != nil {
		return false
	}
	dt2, sender2, _, err := cdr2.ExtractID()
	if err != nil {
		return false
	}
	if dt1 < dt2 {
		return true
	}

	if dt1 > dt2 {
		return false
	}

	return sender1 < sender2
}

func (s *SettlementAPI) queryContractsByAddress(count int, offset *int, fn func() ([]*cabi.ContractParam, error)) ([]*SettlementContract, error) {
	contracts, err := fn()
	if err != nil {
		return nil, err
	}

	size := len(contracts)

	start, end, err := calculateRange(size, count, offset)
	if err != nil {
		return nil, err
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

func calculateRange(size, count int, offset *int) (start, end int, err error) {
	if size <= 0 {
		return 0, 0, fmt.Errorf("can not find any records, size=%d", size)
	}

	o := 0

	if count <= 0 {
		return 0, 0, fmt.Errorf("invalid count: %d", count)
	}

	if offset != nil && *offset >= 0 {
		o = *offset
	}

	if o >= size {
		return 0, 0, fmt.Errorf("overflow, max:%d,offset:%d", size, o)
	} else {
		start = o
		end = start + count
		if end > size {
			end = size
		}
		return
	}
}
