/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package process

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

var (
	checkBlockFns = make(map[types.BlockType]checkBlock)
)

func init() {
	checkBlockFns[types.Send] = checkSendBlock
	checkBlockFns[types.Receive] = checkReceiveBlock
	checkBlockFns[types.Change] = checkChangeBlock
	checkBlockFns[types.Open] = checkOpenBlock
	checkBlockFns[types.ContractSend] = checkContractSendBlock
	checkBlockFns[types.ContractReward] = checkContractReceiveBlock
}

type checkBlock func(*LedgerVerifier, *types.StateBlock) (ProcessResult, error)

type LedgerVerifier struct {
	l *ledger.Ledger
	//vmContext *vmstore.VMContext
	logger *zap.SugaredLogger
}

func NewLedgerVerifier(l *ledger.Ledger) *LedgerVerifier {
	return &LedgerVerifier{l: l, logger: log.NewLogger("ledger_verifier")}
}

func (lv *LedgerVerifier) Process(block types.Block) (ProcessResult, error) {
	if r, err := lv.BlockCheck(block); r != Progress || err != nil {
		return r, err
	}
	if err := lv.BlockProcess(block); err != nil {
		return Other, err
	}
	return Progress, nil
}

func (lv *LedgerVerifier) BlockCheck(block types.Block) (ProcessResult, error) {
	if b, ok := block.(*types.StateBlock); ok {
		if fn, ok := checkBlockFns[b.Type]; ok {
			r, err := fn(lv, b)
			if err != nil {
				lv.logger.Error(fmt.Sprintf("error:%s, block:%s", err.Error(), b.GetHash().String()))
			}
			if r != Progress {
				lv.logger.Info(fmt.Sprintf("process result:%s, block:%s", r.String(), b.GetHash().String()))
			}
			return r, err
		} else {
			return Other, fmt.Errorf("unsupport block type %s", b.Type.String())
		}
	} else if _, ok := block.(*types.SmartContractBlock); ok {
		return Other, errors.New("smart contract block")
	}
	return Other, errors.New("invalid block")
}

func checkStateBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	hash := block.GetHash()
	address := block.GetAddress()

	lv.logger.Debug("check block ", hash)

	if !block.IsValid() {
		return BadWork, nil
	}

	blockExist, err := lv.l.HasStateBlock(hash)
	if err != nil {
		return Other, err
	}

	if blockExist {
		return Old, nil
	}

	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		return BadSignature, nil
	}

	return Progress, nil
}

func checkSendBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}

	// check previous
	if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
		return GapPrevious, nil
	} else {
		//check fork
		if tm, err := lv.l.GetTokenMeta(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
			return Fork, nil
		}

		//check balance
		//if previous.Balance.Compare(block.Balance) == types.BalanceCompSmaller {
		//	return BalanceMismatch, nil
		//}
	}

	return Progress, nil
}

func checkReceiveBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}

	// check previous
	if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
		return GapPrevious, nil
	} else {
		//check fork
		if tm, err := lv.l.GetTokenMeta(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
			return Fork, nil
		}
		pendingKey := types.PendingKey{
			Address: block.Address,
			Hash:    block.Link,
		}

		//check receive link
		if b, err := lv.l.HasStateBlock(block.Link); !b && err == nil {
			return GapSource, nil
		}

		//check pending
		if pending, err := lv.l.GetPending(pendingKey); err == nil {
			if tm, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
				transferAmount := block.GetBalance().Sub(tm.Balance)
				if !pending.Amount.Equal(transferAmount) || pending.Type != block.Token {
					return BalanceMismatch, nil
				}
			} else {
				return Other, err
			}
		} else if err == ledger.ErrPendingNotFound {
			return UnReceivable, nil
		} else {
			return Other, err
		}
	}

	return Progress, nil
}

func checkChangeBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}

	// check link
	if !block.Link.IsZero() {
		return Other, fmt.Errorf("invalid link hash")
	}

	// check chain token
	if block.Token != common.ChainToken() {
		return Other, fmt.Errorf("invalid token Id")
	}

	// check previous
	if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
		return GapPrevious, nil
	} else {
		//check fork
		if tm, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil && previous.GetHash() != tm.Header {
			return Fork, nil
		} else {
			//check balance
			if block.Balance.Compare(tm.Balance) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
		}
	}

	return Progress, nil
}

func checkOpenBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}

	//check previous
	if !block.Previous.IsZero() {
		return Other, fmt.Errorf("open block previous is not zero")
	}

	//check link
	if b, _ := lv.l.HasStateBlock(block.Link); !b {
		return GapSource, nil
	} else {
		//check fork
		if _, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
			return Fork, nil
		}

		pendingKey := types.PendingKey{
			Address: block.Address,
			Hash:    block.Link,
		}
		//check pending
		if pending, err := lv.l.GetPending(pendingKey); err == nil {
			if !pending.Amount.Equal(block.Balance) || pending.Type != block.Token {
				return BalanceMismatch, nil
			}
		} else if err == ledger.ErrPendingNotFound {
			return UnReceivable, nil
		} else {
			return Other, err
		}
	}

	return Progress, nil
}

func checkContractSendBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkSendBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}

	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}

	//check smart c exist
	address := types.Address(block.GetLink())

	if !contract.IsChainContract(address) {
		if b, err := lv.l.HasSmartContractBlock(address.ToHash()); !b && err == nil {
			return GapSmartContract, nil
		}
	}

	//verify data
	if c, ok, _ := contract.GetChainContract(address, block.Data); ok {
		clone := block.Clone()
		vmCtx := vmstore.NewVMContext(lv.l)
		if err := c.DoSend(vmCtx, clone); err == nil {
			if bytes.EqualFold(block.Data, clone.Data) {
				return Progress, nil
			} else {
				return InvalidData, nil
			}
		} else {
			return Other, err
		}
	} else {
		//call vm.Run();
		return Other, fmt.Errorf("can not find chain contract %s", address.String())
	}
}

func checkContractReceiveBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}
	//check previous
	//if !block.Previous.IsZero() {
	//	return Other, fmt.Errorf("open block previous is not zero")
	//}

	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}

	//check smart c exist
	send, err := lv.l.GetStateBlock(block.GetLink())
	if err != nil {
		return GapSource, nil
	}
	address := types.Address(send.GetLink())

	//verify data
	input, err := lv.l.GetStateBlock(block.Link)
	if err != nil {
		return Other, err
	}
	if c, ok, err := contract.GetChainContract(address, input.Data); ok && err == nil {
		clone := block.Clone()
		//TODO:verify extra hash and commit to db
		vmCtx := vmstore.NewVMContext(lv.l)
		if g, e := c.DoReceive(vmCtx, clone, input); e == nil {
			if len(g) > 0 {
				amount, _ := lv.l.CalculateAmount(block)
				if bytes.EqualFold(g[0].Block.Data, block.Data) && g[0].Token == block.Token &&
					g[0].Amount.Compare(amount) == types.BalanceCompEqual && g[0].ToAddress == block.Address {
					//save contract data
					ctx := g[0].VMContext
					if ctx != nil {
						err := ctx.SaveStorage()
						if err != nil {
							return InvalidData, nil
						}
						err = ctx.SaveTrie()
						if err != nil {
							return InvalidData, nil
						}
					}
					return Progress, nil
				} else {
					return InvalidData, nil
				}
			} else {
				return Other, fmt.Errorf("can not generate receive block")
			}
		} else {
			return Other, e
		}
	} else {
		//call vm.Run();
		return Other, fmt.Errorf("can not find chain contract %s", address.String())
	}
}

func (lv *LedgerVerifier) BlockProcess(block types.Block) error {
	return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
		if state, ok := block.(*types.StateBlock); ok {
			err := lv.processStateBlock(state, txn)
			if err != nil {
				lv.logger.Error(fmt.Sprintf("%s, block:%s", err.Error(), state.GetHash().String()))
				return err
			}
			return nil
		} else if _, ok := block.(*types.SmartContractBlock); ok {
			return errors.New("smart contract block")
		}
		return errors.New("invalid block")
	})
}

func (lv *LedgerVerifier) processStateBlock(block *types.StateBlock, txn db.StoreTxn) error {
	lv.logger.Debug("process block, ", block.GetHash())
	if err := lv.l.AddStateBlock(block, txn); err != nil {
		return err
	}

	am, err := lv.l.GetAccountMeta(block.GetAddress(), txn)
	if err != nil && err != ledger.ErrAccountNotFound {
		return fmt.Errorf("get account meta error: %s", err)
	}
	tm, err := lv.l.GetTokenMeta(block.GetAddress(), block.GetToken(), txn)
	if err != nil && err != ledger.ErrAccountNotFound && err != ledger.ErrTokenNotFound {
		return fmt.Errorf("get token meta error: %s", err)
	}
	if err := lv.updateRepresentative(block, am, tm, txn); err != nil {
		return fmt.Errorf("update representative error: %s", err)
	}
	if err := lv.updatePending(block, tm, txn); err != nil {
		return fmt.Errorf("update pending error: %s", err)
	}
	if err := lv.updateFrontier(block, tm, txn); err != nil {
		return fmt.Errorf("update frontier error: %s", err)
	}
	if err := lv.updateAccountMeta(block, am, txn); err != nil {
		return fmt.Errorf("update account meta error: %s", err)
	}
	return nil
}

func (lv *LedgerVerifier) updatePending(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	switch block.Type {
	case types.Send:
		pending := types.PendingInfo{
			Source: block.GetAddress(),
			Type:   block.GetToken(),
			Amount: tm.Balance.Sub(block.GetBalance()),
		}
		pendingKey := types.PendingKey{
			Address: types.Address(block.GetLink()),
			Hash:    hash,
		}
		lv.logger.Debug("add pending, ", pendingKey)
		if err := lv.l.AddPending(&pendingKey, &pending, txn); err != nil {
			return err
		}
	case types.Open, types.Receive:
		pendingKey := types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		lv.logger.Debug("delete pending, ", pendingKey)
		if err := lv.l.DeletePending(&pendingKey, txn); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) updateRepresentative(block *types.StateBlock, am *types.AccountMeta, tm *types.TokenMeta, txn db.StoreTxn) error {
	if block.GetToken() == common.ChainToken() {
		if tm != nil && !tm.Representative.IsZero() {
			oldBenefit := &types.Benefit{
				Vote:    am.GetVote(),
				Network: am.GetNetwork(),
				Oracle:  am.GetOracle(),
				Storage: am.GetStorage(),
				Balance: am.GetBalance(),
				Total:   am.TotalBalance(),
			}
			lv.logger.Debugf("sub rep(%s) from %s ", oldBenefit, tm.Representative)
			if err := lv.l.SubRepresentation(tm.Representative, oldBenefit, txn); err != nil {
				return err
			}
		}
		newBenefit := &types.Benefit{
			Vote:    block.GetVote(),
			Network: block.GetNetwork(),
			Oracle:  block.GetOracle(),
			Storage: block.GetStorage(),
			Balance: block.GetBalance(),
			Total:   block.TotalBalance(),
		}
		lv.logger.Debugf("add rep(%s) to %s ", newBenefit, block.GetRepresentative())
		if err := lv.l.AddRepresentation(block.GetRepresentative(), newBenefit, txn); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) updateFrontier(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	frontier := &types.Frontier{
		HeaderBlock: hash,
	}
	if tm != nil {
		if frontier, err := lv.l.GetFrontier(tm.Header, txn); err == nil {
			lv.logger.Debug("delete frontier, ", *frontier)
			if err := lv.l.DeleteFrontier(frontier.HeaderBlock, txn); err != nil {
				return err
			}
		} else {
			return err
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	lv.logger.Debug("add frontier,", *frontier)
	if err := lv.l.AddFrontier(frontier, txn); err != nil {
		return err
	}
	return nil
}

func (lv *LedgerVerifier) updateAccountMeta(block *types.StateBlock, am *types.AccountMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	rep := block.GetRepresentative()
	address := block.GetAddress()
	token := block.GetToken()
	balance := block.GetBalance()

	tmNew := &types.TokenMeta{
		Type:           token,
		Header:         hash,
		Representative: rep,
		OpenBlock:      hash,
		Balance:        balance,
		BlockCount:     1,
		BelongTo:       address,
		Modified:       time.Now().Unix(),
	}

	if am != nil {
		tm := am.Token(block.GetToken())
		if block.GetToken() == common.ChainToken() {
			am.CoinBalance = balance
			am.CoinOracle = block.GetOracle()
			am.CoinNetwork = block.GetNetwork()
			am.CoinVote = block.GetVote()
			am.CoinStorage = block.GetStorage()
		}
		if tm != nil {
			tm.Header = hash
			tm.Representative = rep
			tm.Balance = balance
			tm.BlockCount = tm.BlockCount + 1
			tm.Modified = time.Now().Unix()
		} else {
			am.Tokens = append(am.Tokens, tmNew)
		}
		if err := lv.l.UpdateAccountMeta(am, txn); err != nil {
			return err
		}
	} else {
		account := types.AccountMeta{
			Address: address,
			Tokens:  []*types.TokenMeta{tmNew},
		}

		if block.GetToken() == common.ChainToken() {
			account.CoinBalance = balance
			account.CoinOracle = block.GetOracle()
			account.CoinNetwork = block.GetNetwork()
			account.CoinVote = block.GetVote()
			account.CoinStorage = block.GetStorage()
		}
		if err := lv.l.AddAccountMeta(&account, txn); err != nil {
			return err
		}
	}
	return nil
}
