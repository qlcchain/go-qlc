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

type checkBlock func(*LedgerVerifier, *types.StateBlock) (ProcessResult, error)

type LedgerVerifier struct {
	l             *ledger.Ledger
	checkBlockFns map[types.BlockType]checkBlock
	logger        *zap.SugaredLogger
}

func NewLedgerVerifier(l *ledger.Ledger) *LedgerVerifier {
	checkBlockFns := make(map[types.BlockType]checkBlock)
	checkBlockFns[types.Send] = checkSendBlock
	checkBlockFns[types.Receive] = checkReceiveBlock
	checkBlockFns[types.Change] = checkChangeBlock
	checkBlockFns[types.Open] = checkOpenBlock
	checkBlockFns[types.ContractSend] = checkContractSendBlock
	checkBlockFns[types.ContractReward] = checkContractReceiveBlock

	return &LedgerVerifier{
		l:             l,
		checkBlockFns: checkBlockFns,
		logger:        log.NewLogger("ledger_verifier"),
	}
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
		if fn, ok := lv.checkBlockFns[b.Type]; ok {
			r, err := fn(lv, b)
			if err != nil {
				lv.logger.Error(fmt.Sprintf("error:%s, block:%s", err.Error(), b.GetHash().String()))
			}
			if r != Progress {
				lv.logger.Debugf(fmt.Sprintf("process result:%s, block:%s", r.String(), b.GetHash().String()))
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

	blockExist, err := lv.l.HasStateBlockConfirmed(hash)
	if err != nil {
		return Other, err
	}

	if blockExist {
		return Old, nil
	}

	if block.GetType() == types.ContractSend {
		if block.GetLink() == types.Hash(types.RewardsAddress) {
			return Progress, nil
		}
	}
	if block.GetType() == types.ContractReward {
		linkBlk, err := lv.l.GetStateBlock(block.GetLink())
		if err != nil {
			return GapSource, nil
		}
		if linkBlk.GetLink() == types.Hash(types.RewardsAddress) {
			return Progress, nil
		}
	}

	if !block.IsValid() {
		return BadWork, errors.New("bad work")
	}

	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		return BadSignature, errors.New("bad signature")
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

		if block.GetType() == types.Send {
			//check balance
			if !(previous.Balance.Compare(block.Balance) == types.BalanceCompBigger) {
				return BalanceMismatch, nil
			}
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
				return BalanceMismatch, nil
			}
		}
		if block.GetType() == types.ContractSend {
			//check totalBalance
			if previous.TotalBalance().Compare(block.TotalBalance()) == types.BalanceCompSmaller {
				return BalanceMismatch, nil
			}
		}
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
				//check vote,network,storage,oracle
				if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
					previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
					previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
					previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
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
			//check vote,network,storage,oracle
			if previous.GetVote().Compare(block.GetVote()) != types.BalanceCompEqual ||
				previous.GetNetwork().Compare(block.GetNetwork()) != types.BalanceCompEqual ||
				previous.GetStorage().Compare(block.GetStorage()) != types.BalanceCompEqual ||
				previous.GetOracle().Compare(block.GetOracle()) != types.BalanceCompEqual {
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
			//check vote,network,storage,oracle
			vote := block.GetVote()
			network := block.GetNetwork()
			storage := block.GetStorage()
			oracle := block.GetOracle()
			if !vote.IsZero() || !network.IsZero() ||
				!storage.IsZero() || !oracle.IsZero() {
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
	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}
	result, err := checkSendBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}
	//check smart c exist
	address := types.Address(block.GetLink())

	if !contract.IsChainContract(address) {
		if b, err := lv.l.HasSmartContractBlock(address.ToHash()); !b && err == nil {
			return GapSmartContract, nil
		}
	}

	//verify data
	if c, ok, err := contract.GetChainContract(address, block.Data); ok && err == nil {
		clone := block.Clone()
		vmCtx := vmstore.NewVMContext(lv.l)
		if err := c.DoSend(vmCtx, clone); err == nil {
			if bytes.EqualFold(block.Data, clone.Data) {
				return Progress, nil
			} else {
				lv.logger.Errorf("data not equal: %s, %s", block.Data, clone.Data)
				return InvalidData, nil
			}
		} else {
			lv.logger.Error("DoSend error")
			return Other, err
		}
	} else {
		//call vm.Run();
		return Other, fmt.Errorf("can not find chain contract %s", address.String())
	}
}

func checkContractReceiveBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}

	result, err := checkStateBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}
	// check previous
	if !block.IsOpen() {
		// check previous
		if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
			if previous, err = lv.l.GetBlockCache(block.Previous); err != nil {
				return GapPrevious, nil
			}
		} else {
			//check fork
			if tm, err := lv.l.GetTokenMeta(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
				return Fork, nil
			}
		}
	} else {
		//check fork
		if _, err := lv.l.GetTokenMeta(block.Address, block.Token); err == nil {
			return Fork, nil
		}
	}

	//check smart c exist
	input, err := lv.l.GetStateBlock(block.GetLink())
	if err != nil {
		if input, err = lv.l.GetBlockCache(block.Link); err != nil {
			return GapSource, nil
		}
	}
	address := types.Address(input.GetLink())

	if c, ok, err := contract.GetChainContract(address, input.Data); ok && err == nil {
		clone := block.Clone()
		//TODO:verify extra hash and commit to db
		vmCtx := vmstore.NewVMContext(lv.l)
		if g, e := c.DoReceive(vmCtx, clone, input); e == nil {
			if len(g) > 0 {
				amount, err := lv.l.CalculateAmount(block)
				if err != nil {
					lv.logger.Error("calculate amount error:", err)
				}
				if bytes.EqualFold(g[0].Block.Data, block.Data) && g[0].Token == block.Token &&
					g[0].Amount.Compare(amount) == types.BalanceCompEqual && g[0].ToAddress == block.Address {
					return Progress, nil
				} else {
					lv.logger.Errorf("data from contract, %s, %s, %s, %s, data from block, %s, %s, %s, %s",
						g[0].Block.Data, g[0].Token, g[0].Amount, g[0].ToAddress, block.Data, block.Token, amount, block.Address)
					return InvalidData, nil
				}
			} else {
				return Other, fmt.Errorf("can not generate receive block")
			}
		} else {
			if address == types.MintageAddress && e == vmstore.ErrStorageNotFound {
				return GapTokenInfo, nil
			} else {
				lv.logger.Error("DoReceive error ", e)
				return Other, e
			}
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

func (lv *LedgerVerifier) BlockCacheProcess(block types.Block) error {
	return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
		if state, ok := block.(*types.StateBlock); ok {
			err := lv.addBlockCache(state, txn)
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

func (lv *LedgerVerifier) addBlockCache(block *types.StateBlock, txn db.StoreTxn) error {
	lv.logger.Debug("process block cache, ", block.GetHash())
	if err := lv.l.AddBlockCache(block, txn); err != nil {
		return err
	}

	am, err := lv.l.GetAccountMeta(block.GetAddress(), txn)
	if err != nil && err != ledger.ErrAccountNotFound {
		return fmt.Errorf("get account meta error: %s", err)
	}
	if err := lv.updateAccountMetaCacheHeader(block, am, txn); err != nil {
		return fmt.Errorf("update account meta cache header error: %s", err)
	}
	return nil
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
	if err := lv.updateContractData(block, txn); err != nil {
		return fmt.Errorf("update contract data error: %s", err)
	}
	return nil
}

func (lv *LedgerVerifier) updatePending(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	switch block.Type {
	case types.Send:
		preBlk, err := lv.l.GetStateBlock(block.Previous)
		if err != nil {
			return errors.New("previous block not found")
		}
		pending := types.PendingInfo{
			Source: block.GetAddress(),
			Type:   block.GetToken(),
			Amount: preBlk.Balance.Sub(block.GetBalance()),
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
	case types.ContractSend:
		if c, ok, err := contract.GetChainContract(types.Address(block.Link), block.Data); ok && err == nil {
			if pendingKey, pendingInfo, err := c.DoPending(block); err == nil && pendingKey != nil {
				lv.logger.Debug("contractSend add pending , ", pendingKey)
				if err := lv.l.AddPending(pendingKey, pendingInfo, txn); err != nil {
					return err
				}
			}
		}
	case types.ContractReward:
		pendingKey := types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		lv.logger.Debug("contractReward delete pending, ", pendingKey)
		if err := lv.l.DeletePending(&pendingKey, txn); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) updateRepresentative(block *types.StateBlock, am *types.AccountMeta, tm *types.TokenMeta, txn db.StoreTxn) error {
	if block.GetToken() == common.ChainToken() {
		if tm != nil && !tm.Representative.IsZero() && tm.ConfirmHeader != types.ZeroHash {
			preBlk, err := lv.l.GetStateBlock(block.Previous)
			if err != nil {
				return err
			}
			oldBenefit := &types.Benefit{
				Vote:    preBlk.GetVote(),
				Network: preBlk.GetNetwork(),
				Oracle:  preBlk.GetOracle(),
				Storage: preBlk.GetStorage(),
				Balance: preBlk.GetBalance(),
				Total:   preBlk.TotalBalance(),
			}
			lv.logger.Debugf("sub rep(%s) from %s ", oldBenefit, preBlk.Representative)
			if err := lv.l.SubRepresentation(preBlk.Representative, oldBenefit, txn); err != nil {
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
	if tm != nil && tm.ConfirmHeader != types.ZeroHash {
		if frontier, err := lv.l.GetFrontier(tm.ConfirmHeader, txn); err == nil {
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
		ConfirmHeader:  hash,
		Representative: rep,
		OpenBlock:      hash,
		Balance:        balance,
		BlockCount:     1,
		BelongTo:       address,
		Modified:       common.TimeNow().UTC().Unix(),
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
			if tm.ConfirmHeader != types.ZeroHash {
				tm.Header = hash
				tm.ConfirmHeader = hash
				tm.Representative = rep
				tm.Balance = balance
				tm.BlockCount = tm.BlockCount + 1
				tm.Modified = common.TimeNow().UTC().Unix()
			} else {
				tm.Type = token
				tm.Header = hash
				tm.ConfirmHeader = hash
				tm.Representative = rep
				tm.OpenBlock = hash
				tm.Balance = balance
				tm.BlockCount = 1
				tm.BelongTo = address
				tm.Modified = common.TimeNow().UTC().Unix()
			}
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

func (lv *LedgerVerifier) updateAccountMetaCacheHeader(block *types.StateBlock, am *types.AccountMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	rep := block.GetRepresentative()
	address := block.GetAddress()
	token := block.GetToken()
	balance := block.GetBalance()

	tmNew := &types.TokenMeta{
		Type:           token,
		Header:         hash,
		ConfirmHeader:  types.ZeroHash,
		Representative: rep,
		OpenBlock:      hash,
		Balance:        balance,
		BlockCount:     1,
		BelongTo:       address,
		Modified:       common.TimeNow().UTC().Unix(),
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
			tm.Modified = common.TimeNow().UTC().Unix()
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

func (lv *LedgerVerifier) updateContractData(block *types.StateBlock, txn db.StoreTxn) error {
	if !common.IsGenesisBlock(block) && block.GetType() == types.ContractReward {
		input, err := lv.l.GetStateBlock(block.GetLink())
		if err != nil {
			return nil
		}
		address := types.Address(input.GetLink())
		c, ok, err := contract.GetChainContract(address, input.Data)
		if !ok || err != nil {
			return fmt.Errorf("invaild contract %s", err)
		}
		clone := block.Clone()
		vmCtx := vmstore.NewVMContext(lv.l)
		g, err := c.DoReceive(vmCtx, clone, input)
		if err != nil {
			return err
		}
		if len(g) > 0 {
			ctx := g[0].VMContext
			if ctx != nil {
				err := ctx.SaveStorage(txn)
				if err != nil {
					lv.logger.Error("save storage error: ", err)
					return err
				}
				err = ctx.SaveTrie(txn)
				if err != nil {
					lv.logger.Error("save trie error: ", err)
					return err
				}
				return nil
			}
		}
		return errors.New("invaild contract data")
	}
	return nil
}

// TODO: implement
func (lv *LedgerVerifier) Rollback(hash types.Hash) error {
	return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
		err := lv.processRollback(hash, true, txn)
		if err != nil {
			lv.logger.Error(err)
		}
		return err
	})
}

func (lv *LedgerVerifier) processRollback(hash types.Hash, isRoot bool, txn db.StoreTxn) error {
	lv.l.EB.Publish(common.EventRollbackUnchecked, hash)

	tm, err := lv.l.Token(hash, txn)
	if err != nil {
		return fmt.Errorf("get block(%s) token err : %s", hash.String(), err)
	}

	blockHead, err := lv.l.GetStateBlock(tm.Header, txn)
	if err != nil {
		return fmt.Errorf("get header block %s : %s", tm.Header.String(), err)
	}

	blockCur := blockHead
	for {
		hashCur := blockCur.GetHash()
		//blockType, err := l.JudgeBlockKind(hashCur, txn)
		blockType := blockCur.GetType()

		blockPre := new(types.StateBlock)
		if !blockCur.IsOpen() {
			blockPre, err = lv.l.GetStateBlock(blockCur.Previous, txn)
			if err != nil {
				return fmt.Errorf("get previous block %s : %s", blockCur.Previous.String(), err)
			}
		}
		switch blockType {
		case types.Open:

			if err := lv.rollBackTokenDel(tm, txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), open(%s)", err, hashCur)
			}
			if b, err := lv.l.HasBlockCache(blockCur.GetHash()); !b && err == nil {
				if err := lv.rollBackFrontier(types.Hash{}, blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), open(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, nil, false, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), open(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingAdd(blockCur, tm.Balance, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), open(%s)", err, hashCur)
				}
			}
			lv.logger.Debug("---delete open block, ", hashCur)
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), open(%s)", err, hashCur)
			}
			if hashCur != hash {
				if err := lv.processRollback(blockCur.GetLink(), false, txn); err != nil {
					return err
				}
			}
		case types.Send:
			if hashCur != hash || isRoot {
				linkHash, err := lv.l.GetLinkBlock(blockCur.GetHash(), txn)
				// link not found is not error ,may be send block has created but receiver block has not created
				if err != nil && err != ledger.ErrLinkNotFound {
					return fmt.Errorf("get block(%s)'s link : %s", blockCur.GetHash().String(), err)
				}
				if linkHash != types.ZeroHash {
					if err := lv.processRollback(linkHash, false, txn); err != nil {
						return err
					}
				}
			}
			if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), send(%s)", err, hashCur)
			}
			if b, err := lv.l.HasBlockCache(blockCur.GetHash()); !b && err == nil {
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, blockPre, true, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingDel(blockCur, txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), send(%s)", err, hashCur)
				}
			}
			lv.logger.Debug("---delete send block, ", hashCur)
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), send(%s)", err, hashCur)
			}

		case types.Receive:
			if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), receive(%s)", err, hashCur)
			}
			if b, err := lv.l.HasBlockCache(blockCur.GetHash()); !b && err == nil {
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), receive(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, blockPre, false, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), receive(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingAdd(blockCur, blockCur.GetBalance().Sub(blockPre.GetBalance()), blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), receive(%s)", err, hashCur)
				}
			}
			lv.logger.Debug("---delete receive block, ", hashCur)
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), receive(%s)", err, hashCur)
			}
			if hashCur != hash {
				if err := lv.processRollback(blockCur.GetLink(), false, txn); err != nil {
					return err
				}
			}
		case types.Change:
			if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), change(%s)", err, hashCur)
			}
			if b, err := lv.l.HasBlockCache(blockCur.GetHash()); !b && err == nil {
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), change(%s)", err, hashCur)
				}
				if err := lv.rollBackRepChange(blockPre.GetRepresentative(), blockCur.GetRepresentative(), blockCur, txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), change(%s)", err, hashCur)
				}
			}
			lv.logger.Debug("---delete change block, ", hashCur)
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), change(%s)", err, hashCur)
			}
		case types.ContractReward:
			previousHash := blockCur.GetPrevious()
			if previousHash.IsZero() {
				if err := lv.rollBackTokenDel(tm, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), ContractReward(%s)", err, hashCur)
				}
			} else {
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), ContractReward(%s)", err, hashCur)
				}
			}
			if b, err := lv.l.HasBlockCache(blockCur.GetHash()); !b && err == nil {
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), ContractReward(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingAdd(blockCur, types.ZeroBalance, types.ZeroHash, txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), ContractReward(%s)", err, hashCur)
				}
			}
			if err := lv.rollBackContractData(blockCur, txn); err != nil {
				return fmt.Errorf("rollback contract data fail(%s), ContractReward(%s)", err, blockCur.String())
			}

			lv.logger.Debug("---delete ContractReward block, ", hashCur)
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), ContractReward(%s)", err, hashCur)
			}
			if hashCur != hash {
				if err := lv.processRollback(blockCur.GetLink(), false, txn); err != nil {
					return err
				}
			}
		case types.ContractSend:
			if hashCur != hash || isRoot {
				linkHash, err := lv.l.GetLinkBlock(blockCur.GetHash(), txn)
				if err != nil && err != ledger.ErrLinkNotFound {
					return fmt.Errorf("get block(%s) link error: %s", blockCur.GetHash().String(), err)
				}
				if linkHash != types.ZeroHash {
					if err := lv.processRollback(linkHash, false, txn); err != nil {
						return err
					}
				}
			}

			if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), ContractSend(%s)", err, hashCur)
			}
			if b, err := lv.l.HasBlockCache(blockCur.GetHash()); !b && err == nil {
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), ContractSend(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingDel(blockCur, txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), ContractSend(%s)", err, hashCur)
				}
			}
			if err := lv.rollBackContractData(blockCur, txn); err != nil {
				return fmt.Errorf("rollback contract data fail(%s), ContractSend(%s)", err, blockCur.String())
			}
			lv.logger.Debug("---delete ContractSend block, ", hashCur)
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), ContractSend(%s)", err, hashCur)
			}
		}

		if hashCur == hash {
			break
		}

		preHash := blockCur.GetPrevious()
		blockCur, err = lv.l.GetStateBlock(preHash, txn)
		if err != nil {
			return fmt.Errorf("get previous block %s : %s", preHash.String(), err)
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackFrontier(pre types.Hash, cur types.Hash, txn db.StoreTxn) error {
	frontier, err := lv.l.GetFrontier(cur, txn)
	if err != nil {
		return err
	}
	lv.logger.Debug("delete frontier, ", frontier)
	if err := lv.l.DeleteFrontier(cur, txn); err != nil {
		return err
	}
	if !pre.IsZero() {
		frontier.HeaderBlock = pre
		lv.logger.Debug("add frontier, ", frontier)
		if err := lv.l.AddFrontier(frontier, txn); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackToken(token *types.TokenMeta, pre *types.StateBlock, txn db.StoreTxn) error {
	ac, err := lv.l.GetAccountMeta(token.BelongTo, txn)
	if err != nil {
		return err
	}
	ac.CoinVote = pre.GetVote()
	ac.CoinOracle = pre.GetOracle()
	ac.CoinNetwork = pre.GetNetwork()
	ac.CoinStorage = pre.GetStorage()
	if pre.GetToken() == common.ChainToken() {
		ac.CoinBalance = pre.GetBalance()
	}
	tm := ac.Token(pre.GetToken())
	if tm == nil {
		return fmt.Errorf("can not get token %s from account %s", pre.GetToken().String(), ac.Address.String())
	}
	tm.Balance = pre.GetBalance()
	tm.Header = pre.GetHash()
	tm.Representative = pre.GetRepresentative()
	tm.BlockCount = tm.BlockCount - 1
	tm.Modified = common.TimeNow().UTC().Unix()
	lv.logger.Debug("update token, ", tm.BelongTo, tm.Type)
	if err := lv.l.UpdateAccountMeta(ac, txn); err != nil {
		return err
	}
	return nil
}

func (lv *LedgerVerifier) rollBackTokenDel(tm *types.TokenMeta, txn db.StoreTxn) error {
	address := tm.BelongTo
	lv.logger.Debug("delete token, ", address, tm.Type)
	if err := lv.l.DeleteTokenMeta(address, tm.Type, txn); err != nil {
		return err
	}
	ac, err := lv.l.GetAccountMeta(address, txn)
	if err != nil {
		return err
	}
	if len(ac.Tokens) == 0 {
		if err := lv.l.DeleteAccountMeta(address, txn); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackRep(representative types.Address, blockCur, blockPre *types.StateBlock, isSend bool, token types.Hash, txn db.StoreTxn) error {
	if token == common.ChainToken() {
		if isSend {
			diff := &types.Benefit{
				Vote:    blockPre.GetVote().Sub(blockCur.GetVote()),
				Network: blockPre.GetNetwork().Sub(blockCur.GetNetwork()),
				Oracle:  blockPre.GetOracle().Sub(blockCur.GetOracle()),
				Storage: blockPre.GetStorage().Sub(blockCur.GetStorage()),
				Balance: blockPre.GetBalance().Sub(blockCur.GetBalance()),
				Total:   blockPre.TotalBalance().Sub(blockCur.TotalBalance()),
			}
			lv.logger.Debugf("add rep(%s) to %s", diff, representative)
			if err := lv.l.AddRepresentation(representative, diff, txn); err != nil {
				return err
			}
		} else {
			diff := new(types.Benefit)
			if blockPre == nil {
				diff = &types.Benefit{
					Vote:    blockCur.GetVote(),
					Network: blockCur.GetNetwork(),
					Oracle:  blockCur.GetOracle(),
					Storage: blockCur.GetStorage(),
					Balance: blockCur.GetBalance(),
					Total:   blockCur.TotalBalance(),
				}
			} else {
				diff = &types.Benefit{
					Vote:    blockCur.GetVote().Sub(blockPre.GetVote()),
					Network: blockCur.GetNetwork().Sub(blockPre.GetNetwork()),
					Oracle:  blockCur.GetOracle().Sub(blockPre.GetOracle()),
					Storage: blockCur.GetStorage().Sub(blockPre.GetStorage()),
					Balance: blockCur.GetBalance().Sub(blockPre.GetBalance()),
					Total:   blockCur.TotalBalance().Sub(blockPre.TotalBalance()),
				}
			}
			lv.logger.Debugf("sub rep %s from %s", diff, representative)
			if err := lv.l.SubRepresentation(representative, diff, txn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (lv *LedgerVerifier) rollBackRepChange(preRepresentation types.Address, curRepresentation types.Address, blockCur *types.StateBlock, txn db.StoreTxn) error {
	diff := &types.Benefit{
		Vote:    blockCur.GetVote(),
		Network: blockCur.GetNetwork(),
		Oracle:  blockCur.GetOracle(),
		Storage: blockCur.GetStorage(),
		Balance: blockCur.GetBalance(),
		Total:   blockCur.TotalBalance(),
	}
	lv.logger.Debugf("add rep(%s) to %s", diff, preRepresentation)
	if err := lv.l.AddRepresentation(preRepresentation, diff, txn); err != nil {
		return err
	}
	lv.logger.Debugf("sub rep(%s) from %s", diff, curRepresentation)
	if err := lv.l.SubRepresentation(curRepresentation, diff, txn); err != nil {
		return err
	}
	return nil
}

func (lv *LedgerVerifier) rollBackPendingAdd(blockCur *types.StateBlock, amount types.Balance, token types.Hash, txn db.StoreTxn) error {
	blockLink, err := lv.l.GetStateBlock(blockCur.GetLink(), txn)
	if err != nil {
		return err
	}

	if blockCur.GetType() == types.ContractReward {
		if c, ok, err := contract.GetChainContract(types.Address(blockLink.Link), blockLink.Data); ok && err == nil {
			if pendingKey, pendingInfo, err := c.DoPending(blockLink); err == nil && pendingKey != nil {
				lv.logger.Debug("add contractreward pending , ", pendingKey)
				if err := lv.l.AddPending(pendingKey, pendingInfo, txn); err != nil {
					return err
				}
			}
		}
		return nil
	} else {
		pendingkey := types.PendingKey{
			Address: blockCur.GetAddress(),
			Hash:    blockLink.GetHash(),
		}
		pendinginfo := types.PendingInfo{
			Source: blockLink.GetAddress(),
			Amount: amount,
			Type:   token,
		}
		lv.logger.Debug("add pending, ", pendingkey, pendinginfo)
		if err := lv.l.AddPending(&pendingkey, &pendinginfo, txn); err != nil {
			return err
		}
		return nil
	}
}

func (lv *LedgerVerifier) rollBackPendingDel(blockCur *types.StateBlock, txn db.StoreTxn) error {
	if blockCur.GetType() == types.ContractSend {
		if c, ok, err := contract.GetChainContract(types.Address(blockCur.Link), blockCur.Data); ok && err == nil {
			if pendingKey, _, err := c.DoPending(blockCur); err == nil && pendingKey != nil {
				lv.logger.Debug("delete contractsend pending , ", pendingKey)
				if err := lv.l.DeletePending(pendingKey, txn); err != nil {
					return err
				}
			}
		}
		return nil
	} else {
		address := types.Address(blockCur.Link)
		hash := blockCur.GetHash()
		pendingkey := types.PendingKey{
			Address: address,
			Hash:    hash,
		}
		lv.logger.Debug("delete pending ,", pendingkey)
		if err := lv.l.DeletePending(&pendingkey, txn); err != nil {
			return err
		}
		return nil
	}
}

func (lv *LedgerVerifier) rollBackContractData(block *types.StateBlock, txn db.StoreTxn) error {
	extra := block.GetExtra()
	if !extra.IsZero() {
		//t := trie.NewTrie(lv.l.Store, &extra, nil)
		//
		//nodes := t.NewNodeIterator(func(node *trie.TrieNode) bool {
		//	return true
		//})
		//
		//for node := range nodes {
		//	trieData, err := node.Serialize()
		//	if err != nil {
		//		return err
		//	}
		//
		//	vmContext := vmstore.NewVMContext(lv.l)
		//	h := node.Hash()
		//	contractData, _ := vmContext.GetStorage(h[:], nil)
		//	if contractData != nil {
		//		if !bytes.Equal(contractData, trieData) {
		//			return errors.New("contract data invalid")
		//		}
		//		//move contract data to new table
		//		if err := lv.l.AddTemporaryData(h[:], contractData, txn); err != nil {
		//			return err
		//		}
		//		//delete trie and contract data
		//		if err := t.DeleteNode(h[:], txn); err != nil {
		//			return err
		//		}
		//		if err := vmContext.DeleteStorage(h[:], txn); err != nil {
		//			return err
		//		}
		//	}
		//}
	}
	return nil
}
