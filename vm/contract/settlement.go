// +build testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bluele/gcache"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/sync"
	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi/settlement"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	ErrNotImplement    = errors.New("not implemented")
	SettlementContract = NewChainContract(
		map[string]Contract{
			cabi.MethodNameCreateContract:    &CreateContract{},
			cabi.MethodNameSignContract:      &SignContract{},
			cabi.MethodNameProcessCDR:        &ProcessCDR{},
			cabi.MethodNameAddPreStop:        &AddPreStop{},
			cabi.MethodNameUpdatePreStop:     &UpdatePreStop{},
			cabi.MethodNameRemovePreStop:     &RemovePreStop{},
			cabi.MethodNameAddNextStop:       &AddNextStop{},
			cabi.MethodNameUpdateNextStop:    &UpdateNextStop{},
			cabi.MethodNameRemoveNextStop:    &RemoveNextStop{},
			cabi.MethodNameTerminateContract: &TerminateContract{},
			cabi.MethodNameRegisterAsset:     &RegisterAsset{},
		},
		cabi.SettlementABI,
		cabi.JsonSettlement,
	)
)

type internalContract struct {
}

func (i internalContract) GetDescribe() Describe {
	return Describe{pending: true, signature: true, specVer: SpecVer2}
}

func (i internalContract) GetTargetReceiver(_ *vmstore.VMContext, _ *types.StateBlock) (types.Address, error) {
	return types.ZeroAddress, nil
}

func (i internalContract) GetFee(_ *vmstore.VMContext, _ *types.StateBlock) (types.Balance, error) {
	return types.ZeroBalance, nil
}

func (i internalContract) GetRefundData() []byte {
	return []byte{1}
}

func (i internalContract) DoPending(_ *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, ErrNotImplement
}

func (i internalContract) DoSend(_ *vmstore.VMContext, _ *types.StateBlock) error {
	return ErrNotImplement
}

func (i internalContract) DoGap(_ *vmstore.VMContext, _ *types.StateBlock) (common.ContractGapType, interface{}, error) {
	return common.ContractNoGap, nil, nil
}

func (i internalContract) DoSendOnPov(_ *vmstore.VMContext, _ *statedb.PovContractStateDB, _ uint64, _ *types.StateBlock) error {
	return ErrNotImplement
}

func (i internalContract) DoReceiveOnPov(_ *vmstore.VMContext, _ *statedb.PovContractStateDB, _ uint64, _ *types.StateBlock, _ *types.StateBlock) error {
	return ErrNotImplement
}

type CreateContract struct {
	internalContract
}

func (c *CreateContract) GetTargetReceiver(_ *vmstore.VMContext, blk *types.StateBlock) (types.Address, error) {
	param := new(cabi.CreateContractParam)
	err := param.FromABI(blk.GetData())
	if err != nil {
		return types.ZeroAddress, err
	} else {
		return param.PartyB.Address, nil
	}
}

func (c *CreateContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		// verify send block data
		param := new(cabi.CreateContractParam)
		err := param.FromABI(data)
		if err != nil {
			return err
		}
		address, err := param.Address()
		if err != nil {
			return err
		}
		if b, err := cabi.GetContractParam(ctx, &address); err == nil && len(b) > 0 {
			if _, err := cabi.ParseContractParam(b); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid send block[%s] data", input.GetHash().String())
		}

		return nil
	})
}

func (c *CreateContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	// check token is QGAS
	if block.Token != cfg.GasToken() {
		return nil, nil, fmt.Errorf("invalid token: %s", block.Token.String())
	}

	param := new(cabi.CreateContractParam)
	err := param.FromABI(block.Data)
	if err != nil {
		return nil, nil, err
	}

	if b, err := param.Verify(); err == nil && b {
		// check balance
		amount, err := ctx.CalculateAmount(block)
		if err != nil {
			return nil, nil, err
		}

		if exp, err := param.Balance(); err != nil {
			return nil, nil, err
		} else {
			if amount.Compare(exp) == types.BalanceCompSmaller {
				return nil, nil, fmt.Errorf("not enough balance, exp: %s, act: %s", exp.String(), amount.String())
			}
		}

		// make sure that the same block only process once
		address, err := param.Address()
		if err != nil {
			return nil, nil, err
		}
		storage, err := cabi.GetContractParam(ctx, &address)
		if err != nil && err != vmstore.ErrStorageNotFound {
			return nil, nil, err
		}

		if len(storage) > 0 {
			// verify saved data
			param2, err := cabi.ParseContractParam(storage)
			if err != nil {
				return nil, nil, err
			}
			if _, err := param2.Equal(param); err != nil {
				return nil, nil, errors.New("invalid saved create contract data")
			}
		} else {
			if data, err := param.ToContractParam().ToABI(); err == nil {
				if err := cabi.SaveContractParam(ctx, &address, data); err != nil {
					return nil, nil, err
				}
			} else {
				return nil, nil, err
			}
		}

		return &types.PendingKey{
				Address: param.PartyA.Address,
				Hash:    block.GetHash(),
			}, &types.PendingInfo{
				Source: types.Address(block.Link),
				Amount: types.ZeroBalance,
				Type:   block.Token,
			}, nil
	} else {
		return nil, nil, err
	}
}

type SignContract struct {
	internalContract
}

func (s *SignContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		// verify send block data
		param := new(cabi.SignContractParam)
		err := param.FromABI(input.Data)
		if err != nil {
			return err
		}
		if b, err := cabi.GetContractParam(ctx, &param.ContractAddress); err == nil && len(b) > 0 {
			if _, err := cabi.ParseContractParam(b); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid send block[%s] data", input.GetHash().String())
		}
		return nil
	})
}

func (s *SignContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.SignContractParam)
	err := param.FromABI(block.Data)
	if err != nil {
		return nil, nil, err
	}

	// verify block data
	if _, err := param.Verify(); err != nil {
		return nil, nil, err
	}

	if storage, err := cabi.GetContractParam(ctx, &param.ContractAddress); err != nil {
		return nil, nil, err
	} else {
		if len(storage) > 0 {
			if cp, err := cabi.ParseContractParam(storage); err != nil {
				return nil, nil, err
			} else {
				// active contract
				if err := cp.DoActive(block.Address); err != nil {
					return nil, nil, err
				}
				cp.ConfirmDate = param.ConfirmDate

				if data, err := cp.ToABI(); err == nil {
					// save confirm data
					if err := cabi.SaveContractParam(ctx, &param.ContractAddress, data); err == nil {
						return &types.PendingKey{
								Address: block.Address,
								Hash:    block.GetHash(),
							}, &types.PendingInfo{
								Source: types.Address(block.Link),
								Amount: types.ZeroBalance,
								Type:   block.Token,
							}, nil
					} else {
						return nil, nil, err
					}
				} else {
					return nil, nil, err
				}
			}
		} else {
			return nil, nil, fmt.Errorf("invalid saved contract data of %s", param.ContractAddress.String())
		}
	}
}

var lockerCache = newLocker()

func PrintLocker() string {
	all := lockerCache.cache.GetALL(true)
	sb := strings.Builder{}
	for k, v := range all {
		sb.WriteString(fmt.Sprintf("%s=>%t\n", hex.EncodeToString([]byte(k.(string))), v.(*sync.Mutex).IsLocked()))
	}
	return sb.String()
}

type ProcessCDR struct {
	internalContract
}

func (p *ProcessCDR) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		paramList := new(cabi.CDRParamList)
		err := paramList.FromABI(data)
		if err != nil {
			return err
		}
		// verify block data
		if err := paramList.Verify(); err != nil {
			return err
		}

		for _, param := range paramList.Params {
			if err := param.Verify(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *ProcessCDR) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	paramList := new(cabi.CDRParamList)
	err := paramList.FromABI(block.Data)
	if err != nil {
		return nil, nil, err
	}

	if err := paramList.Verify(); err != nil {
		return nil, nil, err
	}
	contractAddress := paramList.ContractAddress
	contract, err := cabi.GetSettlementContract(ctx, &contractAddress)
	if err != nil {
		return nil, nil, err
	}

	// verify sms date and uploading date
	if block.Timestamp <= 0 || block.Timestamp < contract.StartDate || block.Timestamp > contract.EndDate {
		return nil, nil, fmt.Errorf("invalid uploading date, should be in [%s, %s], got %s",
			timeString(contract.StartDate), timeString(contract.EndDate), timeString(block.Timestamp))
	}

	for _, param := range paramList.Params {
		if err := p.save(ctx, block, &contractAddress, contract, param); err != nil {
			return nil, nil, err
		}
	}

	return &types.PendingKey{
			Address: block.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

func (p *ProcessCDR) save(ctx *vmstore.VMContext, block *types.StateBlock, ca *types.Address,
	contract *cabi.ContractParam, param *cabi.CDRParam) error {
	// verify block data
	if err := param.Verify(); err != nil {
		return err
	}

	h, err := param.ToHash()
	if err != nil {
		return err
	}

	addr := block.Address
	key, _ := types.HashBytes(ca[:], h[:])

	// lock by CDR param
	mutex, err := lockerCache.Get(key[:])
	if err != nil {
		return err
	}
	mutex.Lock()
	log.Root.Debugf("lock: %s[%s], %t", key.String(), addr.String(), mutex.IsLocked())
	defer func() {
		mutex.Unlock()
		log.Root.Debugf("unlock: %s[%s], %t", key.String(), addr.String(), mutex.IsLocked())
	}()

	if param.SmsDt <= 0 || param.SmsDt < contract.StartDate || param.SmsDt > contract.EndDate {
		return fmt.Errorf("invalid SMS date, should be in [%s, %s], got %s",
			timeString(contract.StartDate), timeString(contract.EndDate), timeString(param.SmsDt))
	}

	if !(contract.PartyA.Address == addr || contract.PartyB.Address == addr) {
		return fmt.Errorf("%s can not upload CDR data to contract %s", addr.String(), ca.String())
	}

	if storage, err := ctx.GetStorage(ca[:], h[:]); err != nil {
		if err != vmstore.ErrStorageNotFound {
			return err
		} else {
			// 1st upload data
			state := &cabi.CDRStatus{
				Params: map[string][]cabi.CDRParam{addr.String(): {*param}},
				Status: cabi.SettlementStatusStage1,
			}

			if err := cabi.SaveCDRStatus(ctx, ca, &h, state); err != nil {
				return err
			}
		}
	} else {
		if state, err := cabi.ParseCDRStatus(storage); err != nil {
			return err
		} else {
			sr := cabi.SettlementCDR{
				CDRParam: *param,
				From:     addr,
			}
			// update contract status
			if err := state.DoSettlement(sr); err != nil {
				return err
			} else {
				if err := cabi.SaveCDRStatus(ctx, ca, &h, state); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type AddPreStop struct {
	internalContract
}

func (a *AddPreStop) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		stopParam := new(cabi.StopParam)
		return stopParam.FromABI(cabi.MethodNameAddPreStop, data)
	})
}

func (a *AddPreStop) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	stopParam := new(cabi.StopParam)
	err := stopParam.FromABI(cabi.MethodNameAddPreStop, block.Data)
	if err != nil {
		return nil, nil, err
	}

	return handleSend(ctx, block, false, stopParam.ContractAddress, func(param *cabi.ContractParam) (err error) {
		if names, err := cabi.GetPreStopNames(ctx, &block.Address); err != nil {
			return err
		} else {
			if len(names) > 0 {
				if b, _ := verifyStopName(names, stopParam.StopName); b {
					return fmt.Errorf("stop name %s already exist", stopParam.StopName)
				}
			}
		}
		param.PreStops, err = add(param.PreStops, stopParam.StopName)
		return err
	})
}

type RemovePreStop struct {
	internalContract
}

func (r *RemovePreStop) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		stopParam := new(cabi.StopParam)
		return stopParam.FromABI(cabi.MethodNameRemovePreStop, data)
	})
}

func (r *RemovePreStop) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	stopParam := new(cabi.StopParam)
	err := stopParam.FromABI(cabi.MethodNameRemovePreStop, block.Data)
	if err != nil {
		return nil, nil, err
	}
	return handleSend(ctx, block, false, stopParam.ContractAddress, func(param *cabi.ContractParam) (err error) {
		param.PreStops, err = remove(param.PreStops, stopParam.StopName)
		return err
	})
}

type UpdatePreStop struct {
	internalContract
}

func (u *UpdatePreStop) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		// verify block data
		param := new(cabi.UpdateStopParam)
		err := param.FromABI(cabi.MethodNameUpdatePreStop, data)
		if err != nil {
			return err
		}
		if err := param.Verify(); err != nil {
			return err
		}
		return nil
	})
}

func (u *UpdatePreStop) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	up := new(cabi.UpdateStopParam)
	err := up.FromABI(cabi.MethodNameUpdatePreStop, block.Data)
	if err != nil {
		return nil, nil, err
	}
	return handleSend(ctx, block, false, up.ContractAddress, func(param *cabi.ContractParam) (err error) {
		if names, err := cabi.GetPreStopNames(ctx, &block.Address); err != nil {
			return err
		} else {
			if len(names) > 0 {
				if b, _ := verifyStopName(names, up.New); b {
					return fmt.Errorf("stop name %s already exist", up.New)
				}
			}
		}

		param.PreStops, err = update(param.PreStops, up.StopName, up.New)
		return err
	})
}

type AddNextStop struct {
	internalContract
}

func (a *AddNextStop) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		stopParam := new(cabi.StopParam)
		return stopParam.FromABI(cabi.MethodNameAddNextStop, data)
	})
}

func (a *AddNextStop) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	stopParam := new(cabi.StopParam)
	err := stopParam.FromABI(cabi.MethodNameAddNextStop, block.Data)
	if err != nil {
		return nil, nil, err
	}
	return handleSend(ctx, block, true, stopParam.ContractAddress, func(param *cabi.ContractParam) (err error) {
		if names, err := cabi.GetNextStopNames(ctx, &block.Address); err != nil {
			return err
		} else {
			if len(names) > 0 {
				if b, _ := verifyStopName(names, stopParam.StopName); b {
					return fmt.Errorf("stop name %s already exist", stopParam.StopName)
				}
			}
		}

		param.NextStops, err = add(param.NextStops, stopParam.StopName)
		return err
	})
}

type RemoveNextStop struct {
	internalContract
}

func (r *RemoveNextStop) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		stopParam := new(cabi.StopParam)
		return stopParam.FromABI(cabi.MethodNameRemoveNextStop, data)
	})
}

func (r *RemoveNextStop) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	stopParam := new(cabi.StopParam)
	err := stopParam.FromABI(cabi.MethodNameRemoveNextStop, block.Data)
	if err != nil {
		return nil, nil, err
	}
	return handleSend(ctx, block, true, stopParam.ContractAddress, func(param *cabi.ContractParam) (err error) {
		param.NextStops, err = remove(param.NextStops, stopParam.StopName)
		return err
	})
}

type UpdateNextStop struct {
	internalContract
}

func (u *UpdateNextStop) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		// verify block data
		param := new(cabi.UpdateStopParam)
		err := param.FromABI(cabi.MethodNameUpdateNextStop, data)
		if err != nil {
			return err
		}
		if err := param.Verify(); err != nil {
			return err
		}
		return nil
	})
}

func (u *UpdateNextStop) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	up := new(cabi.UpdateStopParam)
	err := up.FromABI(cabi.MethodNameUpdateNextStop, block.Data)
	if err != nil {
		return nil, nil, err
	}

	return handleSend(ctx, block, true, up.ContractAddress, func(param *cabi.ContractParam) (err error) {
		if names, err := cabi.GetNextStopNames(ctx, &block.Address); err != nil {
			return err
		} else {
			if len(names) > 0 {
				if b, _ := verifyStopName(names, up.New); b {
					return fmt.Errorf("stop name %s already exist", up.New)
				}
			}
		}

		param.NextStops, err = update(param.NextStops, up.StopName, up.New)
		return err
	})
}

type locker struct {
	cache gcache.Cache
}

func newLocker() *locker {
	return &locker{cache: gcache.New(100).LRU().LoaderFunc(func(key interface{}) (i interface{}, err error) {
		return sync.NewMutex(), nil
	}).Expiration(30 * time.Minute).Build()}
}

func (l *locker) Get(key []byte) (*sync.Mutex, error) {
	if b, err := l.cache.Get(string(key)); err != nil {
		return nil, err
	} else {
		return b.(*sync.Mutex), nil
	}
}

func add(s []string, name string) ([]string, error) {
	if b, _ := verifyStopName(s, name); !b {
		s = append(s, name)
		return s, nil
	}
	return s, fmt.Errorf("name: %s already exist", name)
}

func remove(s []string, name string) ([]string, error) {
	if b, i := verifyStopName(s, name); !b {
		return s, fmt.Errorf("name: %s does not exist", name)
	} else {
		s = append(s[:i], s[i+1:]...)
	}
	return s, nil
}

func update(s []string, old, new string) ([]string, error) {
	if s1, err := remove(s, old); err == nil {
		if s1, err = add(s1, new); err == nil {
			return s1, nil
		} else {
			return s, err
		}
	} else {
		return s, err
	}
}

// verifyStopName verify name is exist in s, if exist, return true
func verifyStopName(s []string, name string) (bool, int) {
	for i, n := range s {
		if n == name {
			return true, i
		}
	}
	return false, 0
}

func handleReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock, fn func(data []byte) error) ([]*ContractBlock, error) {
	if err := fn(input.Data); err != nil {
		return nil, err
	}

	txMeta, _ := ctx.GetAccountMeta(input.Address)
	txToken := txMeta.Token(input.Token)
	txHash := input.GetHash()

	block.Type = types.ContractReward
	block.Address = input.Address
	block.Link = txHash
	block.Token = input.Token
	//block.Extra = types.ZeroHash
	//block.Vote = types.ZeroBalance
	//block.Network = types.ZeroBalance
	//block.Oracle = types.ZeroBalance
	//block.Storage = types.ZeroBalance

	block.Balance = txToken.Balance
	block.Previous = txToken.Header
	block.Representative = input.Representative

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: input.Address,
			BlockType: types.ContractReward,
			Amount:    types.ZeroBalance,
			Token:     input.Token,
			Data:      []byte{},
		},
	}, nil
}

func handleSend(ctx *vmstore.VMContext, block *types.StateBlock, isPartyA bool, address types.Address,
	process func(param *cabi.ContractParam) (err error)) (*types.PendingKey, *types.PendingInfo, error) {
	// check token is QGAS
	if block.Token != cfg.GasToken() {
		return nil, nil, fmt.Errorf("invalid token: %s", block.Token.String())
	}

	// make sure that the same block only process once
	//address := types.Address(block.Link)
	storage, err := cabi.GetContractParam(ctx, &address)
	if err != nil {
		return nil, nil, err
	}

	if len(storage) > 0 {
		// verify saved data
		param, err := cabi.ParseContractParam(storage)
		if err != nil {
			return nil, nil, err
		}
		addr := param.PartyB.Address
		if isPartyA {
			addr = param.PartyA.Address
		}

		if addr == block.Address {
			if err := process(param); err != nil {
				return nil, nil, err
			}

			if data, err := param.ToABI(); err == nil {
				if err := cabi.SaveContractParam(ctx, &address, data); err != nil {
					return nil, nil, err
				}
			} else {
				return nil, nil, err
			}
		} else {
			return nil, nil, fmt.Errorf("permission denied, only %s can change, but got %s", addr.String(), block.Address.String())
		}
	} else {
		return nil, nil, errors.New("invalid saved contract data")
	}

	return &types.PendingKey{
			Address: block.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

type TerminateContract struct {
	internalContract
}

func (t *TerminateContract) GetTargetReceiver(ctx *vmstore.VMContext, blk *types.StateBlock) (types.Address, error) {
	param := new(cabi.TerminateParam)
	err := param.FromABI(blk.Data)
	if err != nil {
		return types.ZeroAddress, err
	}

	if b, err := cabi.GetContractParam(ctx, &param.ContractAddress); err == nil && len(b) > 0 {
		if cp, err := cabi.ParseContractParam(b); err != nil {
			return types.ZeroAddress, err
		} else {
			if cp.PartyA.Address == blk.Address {
				return cp.PartyB.Address, nil
			} else {
				return cp.PartyA.Address, nil
			}
		}
	} else {
		return types.ZeroAddress, fmt.Errorf("can not find data of settlement contract[%s]", param.ContractAddress.String())
	}
}

func (t *TerminateContract) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		// verify send block data
		param := new(cabi.TerminateParam)
		err := param.FromABI(input.Data)
		if err != nil {
			return err
		}
		if b, err := cabi.GetContractParam(ctx, &param.ContractAddress); err == nil && len(b) > 0 {
			if _, err := cabi.ParseContractParam(b); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid send block[%s] data", input.GetHash().String())
		}
		return nil
	})
}

func (t *TerminateContract) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	param := new(cabi.TerminateParam)
	err := param.FromABI(block.Data)
	if err != nil {
		return nil, nil, err
	}

	// verify block data
	if err := param.Verify(); err != nil {
		return nil, nil, err
	}

	if storage, err := cabi.GetContractParam(ctx, &param.ContractAddress); err != nil {
		return nil, nil, err
	} else {
		if len(storage) > 0 {
			if cp, err := cabi.ParseContractParam(storage); err != nil {
				return nil, nil, err
			} else {
				if err := cp.DoTerminate(&cabi.Terminator{
					Address: block.Address,
					Request: param.Request,
				}); err != nil {
					return nil, nil, err
				}

				if data, err := cp.ToABI(); err == nil {
					// save confirm data
					if err := cabi.SaveContractParam(ctx, &param.ContractAddress, data); err == nil {
						return &types.PendingKey{
								Address: block.Address,
								Hash:    block.GetHash(),
							}, &types.PendingInfo{
								Source: types.Address(block.Link),
								Amount: types.ZeroBalance,
								Type:   block.Token,
							}, nil
					} else {
						return nil, nil, err
					}
				} else {
					return nil, nil, err
				}
			}
		} else {
			return nil, nil, fmt.Errorf("invalid saved contract data of %s", param.ContractAddress.String())
		}
	}
}

type RegisterAsset struct {
	internalContract
}

func (r *RegisterAsset) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	return handleReceive(ctx, block, input, func(data []byte) error {
		param, err := cabi.ParseAssertParam(input.GetData())
		if err != nil {
			return err
		}
		h, err := param.ToAddress()
		if err != nil {
			return err
		}
		if _, err := cabi.GetAssetParam(ctx, h); err != nil {
			return fmt.Errorf("invalid send block[%s] data", input.GetHash().String())
		}
		return nil
	})
}

func (r *RegisterAsset) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if err := cabi.SaveAssetParam(ctx, block.Data); err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: block.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

func timeString(t int64) string {
	return time.Unix(t, 0).Format(time.RFC3339)
}
