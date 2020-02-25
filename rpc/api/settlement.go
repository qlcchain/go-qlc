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

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

type SettlementAPI struct {
	logger            *zap.SugaredLogger
	l                 *ledger.Ledger
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

func NewSettlement(l *ledger.Ledger, cc *context.ChainContext) *SettlementAPI {
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

// GetContractRewardsBlock generate create contract rewords block by contract send block hash
// @param send contract send block hash
// @return contract rewards block
func (s *SettlementAPI) GetContractRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.createContract, "GetContractRewardsBlock")
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
		return nil, errors.New("invalid input param")
	}

	now := time.Now().Unix()

	if param.StartDate < now {
		return nil, fmt.Errorf("invalid start date, should bigger than %d, got: %d", now, param.StartDate)
	}

	if param.EndDate < now {
		return nil, fmt.Errorf("invalid start end, should bigger than %d, got: %d", now, param.EndDate)
	}

	if param.EndDate < param.StartDate {
		return nil, fmt.Errorf("invalid end date, should bigger than %d, got: %d", param.StartDate, param.EndDate)
	}

	ctx := vmstore.NewVMContext(s.l)

	addr := param.PartyA.Address
	if tm, err := ctx.GetTokenMeta(addr, common.GasToken()); err != nil {
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
			return nil, errors.New("invalid input param")
		}

		balance, err := createParam.Balance()
		if err != nil {
			return nil, err
		}
		if tm.Balance.Compare(balance) == types.BalanceCompSmaller {
			return nil, fmt.Errorf("not enough balance, [%s] of [%s]", balance.String(), tm.Balance.String())
		}

		if singedData, err := createParam.ToABI(); err == nil {
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
// @param param sign settlement contract param created by PartyA
// @return state block(without signature) to be processed
func (s *SettlementAPI) GetSignContractBlock(param *SignContractParam) (*types.StateBlock, error) {
	if !s.cc.IsPoVDone() {
		return nil, context.ErrPoVNotFinish
	}

	if param == nil {
		return nil, errors.New("invalid input param")
	}

	signParam := &cabi.SignContractParam{
		ContractAddress: param.ContractAddress,
		ConfirmDate:     time.Now().Unix(),
	}

	if isVerified, err := signParam.Verify(); err != nil {
		return nil, err
	} else if !isVerified {
		return nil, errors.New("invalid input param")
	}
	ctx := vmstore.NewVMContext(s.l)

	if tm, err := ctx.GetTokenMeta(param.Address, common.GasToken()); err != nil {
		return nil, err
	} else {
		if singedData, err := signParam.ToABI(); err == nil {
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

			if _, _, err := s.signContract.ProcessSend(ctx, sb); err != nil {
				return nil, err
			} else {
				return sb, nil
			}
		} else {
			return nil, err
		}
	}
}

// GetSignRewardsBlock generate create contract rewords block by contract send block hash
// @param send contract send block hash
// @return contract rewards block
func (s *SettlementAPI) GetSignRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.signContract, "GetSignRewardsBlock")
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

	if tm, err := ctx.GetTokenMeta(addr, common.GasToken()); err != nil {
		return nil, err
	} else {
		if singedData, err := abi(); err == nil {
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

			if _, _, err := c.ProcessSend(ctx, sb); err != nil {
				return nil, err
			} else {
				return sb, nil
			}
		} else {
			return nil, err
		}
	}
}

func (s *SettlementAPI) GetAddPreStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || len(param.StopName) == 0 {
			return errors.New("invalid input param")
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

func (s *SettlementAPI) GetAddPreStopRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.addPreStop, "GetAddPreStopRewardsBlock")
}

func (s *SettlementAPI) GetRemovePreStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || len(param.StopName) == 0 {
			return errors.New("invalid input param")
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

func (s *SettlementAPI) GetRemovePreStopRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.removePreStop, "GetRemovePreStopRewardsBlock")
}

func (s *SettlementAPI) GetUpdatePreStopBlock(param *UpdateStopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.Address.IsZero() || len(param.StopName) == 0 || len(param.New) == 0 || param.StopName == param.New {
			return errors.New("invalid input param")
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

func (s *SettlementAPI) GetUpdatePreStopRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.updatePreStop, "GetUpdatePreStopRewardsBlock")
}

func (s *SettlementAPI) GetAddNextStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.ContractAddress.IsZero() || len(param.StopName) == 0 {
			return errors.New("invalid input param")
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

func (s *SettlementAPI) GetAddNextStopRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.addNextStop, "GetAddNextStopRewardsBlock")
}

func (s *SettlementAPI) GetRemoveNextStopBlock(param *StopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.ContractAddress.IsZero() || len(param.StopName) == 0 {
			return errors.New("invalid input param")
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

func (s *SettlementAPI) GetRemoveNextStopRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.removeNextStop, "GetRemoveNextStopRewardsBlock")
}

func (s *SettlementAPI) GetUpdateNextStopBlock(param *UpdateStopParam) (*types.StateBlock, error) {
	return s.handleStopAction(param.Address, func() error {
		if param == nil || param.ContractAddress.IsZero() || len(param.StopName) == 0 || len(param.New) == 0 || param.StopName == param.New {
			return errors.New("invalid input param")
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

func (s *SettlementAPI) GetUpdateNextStopRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.updateNextStop, "GetUpdateNextStopRewardsBlock")
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
		if tm, err := ctx.GetTokenMeta(*addr, common.GasToken()); err != nil {
			return nil, err
		} else {
			address, err := c.Address()
			param.ContractAddress = address
			if err != nil {
				return nil, err
			}
			if singedData, err := param.ToABI(); err == nil {
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
}

// GetProcessCDRRewardsBlock generate rewards block by send block hash
func (s *SettlementAPI) GetProcessCDRRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.cdrContract, "GetProcessCDRRewardsBlock")
}

type TerminateParam struct {
	cabi.TerminateParam
	Address types.Address // PartyB address
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
		return nil, errors.New("invalid input param")
	}

	if err := param.Verify(); err != nil {
		return nil, err
	}
	ctx := vmstore.NewVMContext(s.l)

	if tm, err := ctx.GetTokenMeta(param.Address, common.GasToken()); err != nil {
		return nil, err
	} else {
		if singedData, err := param.ToABI(); err == nil {
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

			if _, _, err := s.terminateContract.ProcessSend(ctx, sb); err != nil {
				return nil, err
			} else {
				return sb, nil
			}
		} else {
			return nil, err
		}
	}
}

// GetTerminateRewardsBlock generate create contract rewords block by contract send block hash
// @param send contract send block hash
// @return contract rewards block
func (s *SettlementAPI) GetTerminateRewardsBlock(send *types.Hash) (*types.StateBlock, error) {
	return s.getContractRewardsBlock(send, s.terminateContract, "GetTerminateRewardsBlock")
}

// GetCDRStatus get CDRstatus by settlement smart contract address and CDR hash
// @param addr settlement smart contract address
// @param hash CDR data hash
func (s *SettlementAPI) GetCDRStatus(addr *types.Address, hash types.Hash) (*cabi.CDRStatus, error) {
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

func (s *SettlementAPI) getContractRewardsBlock(send *types.Hash, c contract.Contract, name string) (*types.StateBlock, error) {
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

	rev := &types.StateBlock{
		Timestamp: common.TimeNow().Unix(),
	}
	povHeader, err := s.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	rev.PoVHeight = povHeader.GetHeight()

	ctx := vmstore.NewVMContext(s.l)

	if r, err := c.DoReceive(ctx, rev, blk); err == nil {
		if len(r) > 0 {
			return r[0].Block, nil
		} else {
			return nil, fmt.Errorf("%s: fail to generate contract reward block", name)
		}
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
