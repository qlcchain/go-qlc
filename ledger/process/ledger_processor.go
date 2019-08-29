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
	"sort"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"github.com/yireyun/go-queue"
	"go.uber.org/zap"
)

type checkBlock func(*LedgerVerifier, *types.StateBlock, byte) (ProcessResult, error)

type LedgerVerifier struct {
	l             *ledger.Ledger
	checkBlockFns map[types.BlockType]checkBlock
	logger        *zap.SugaredLogger
}

const (
	blockCheck byte = iota
	blockCheckCache
)

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
			r, err := fn(lv, b, blockCheck)
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

func (lv *LedgerVerifier) BlockCheckCache(block types.Block) (ProcessResult, error) {
	if b, ok := block.(*types.StateBlock); ok {
		if fn, ok := lv.checkBlockFns[b.Type]; ok {
			r, err := fn(lv, b, blockCheckCache)
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

func checkStateBlock(lv *LedgerVerifier, block *types.StateBlock, checkType byte) (ProcessResult, error) {
	hash := block.GetHash()
	address := block.GetAddress()

	lv.logger.Debug("check block ", hash)

	if checkType == blockCheck {
		blockExist, err := lv.l.HasStateBlockConfirmed(hash)
		if err != nil {
			return Other, err
		}

		if blockExist {
			return Old, nil
		}
	}
	if checkType == blockCheckCache {
		if err := checkReceiveBlockRepeat(lv, block); err != Progress {
			return err, nil
		}
		blockExist, err := lv.l.HasStateBlock(hash)
		if err != nil {
			return Other, err
		}

		if blockExist {
			return Old, nil
		}
	}

	if block.GetType() == types.ContractSend {
		if block.GetLink() == types.Hash(types.RewardsAddress) {
			return Progress, nil
		}
	}
	if block.GetType() == types.ContractReward {
		var linkBlk *types.StateBlock
		var err error
		if checkType == blockCheck {
			linkBlk, err = lv.l.GetStateBlockConfirmed(block.GetLink())
			if err != nil {
				return GapSource, nil
			}
		}
		if checkType == blockCheckCache {
			linkBlk, err = lv.l.GetStateBlock(block.GetLink())
			if err != nil {
				return GapSource, nil
			}
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

func checkSendBlock(lv *LedgerVerifier, block *types.StateBlock, checkType byte) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block, checkType)
	if err != nil || result != Progress {
		return result, err
	}

	// check previous
	if checkType == blockCheck {
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check fork
			if tm, err := lv.l.GetTokenMetaConfirmed(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
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
	}
	// check previous
	if checkType == blockCheckCache {
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
	}

	return Progress, nil
}

func checkReceiveBlock(lv *LedgerVerifier, block *types.StateBlock, checkType byte) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block, checkType)
	if err != nil || result != Progress {
		return result, err
	}

	// check previous
	if checkType == blockCheck {
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check fork
			if tm, err := lv.l.GetTokenMetaConfirmed(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
				return Fork, nil
			}
			pendingKey := types.PendingKey{
				Address: block.Address,
				Hash:    block.Link,
			}

			//check receive link
			if b, err := lv.l.HasStateBlockConfirmed(block.Link); !b && err == nil {
				return GapSource, nil
			}

			//check pending
			if pending, err := lv.l.GetPending(pendingKey); err == nil {
				if tm, err := lv.l.GetTokenMetaConfirmed(block.Address, block.Token); err == nil {
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
	}
	if checkType == blockCheckCache {
		//if err := checkReceiveBlockRepeat(lv, block); err != Progress {
		//	return err, nil
		//}
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
	}

	return Progress, nil
}

func checkChangeBlock(lv *LedgerVerifier, block *types.StateBlock, checkType byte) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block, checkType)
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
	if checkType == blockCheck {
		if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
			return GapPrevious, nil
		} else {
			//check fork
			if tm, err := lv.l.GetTokenMetaConfirmed(block.Address, block.Token); err == nil && previous.GetHash() != tm.Header {
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
	}
	// check previous
	if checkType == blockCheckCache {
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
	}

	return Progress, nil
}

func checkOpenBlock(lv *LedgerVerifier, block *types.StateBlock, checkType byte) (ProcessResult, error) {
	result, err := checkStateBlock(lv, block, checkType)
	if err != nil || result != Progress {
		return result, err
	}

	//check previous
	if !block.Previous.IsZero() {
		return Other, fmt.Errorf("open block previous is not zero")
	}

	//check link
	if checkType == blockCheck {
		if b, _ := lv.l.HasStateBlockConfirmed(block.Link); !b {
			return GapSource, nil
		} else {
			//check fork
			if _, err := lv.l.GetTokenMetaConfirmed(block.Address, block.Token); err == nil {
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
	}
	//check link
	if checkType == blockCheckCache {
		//if err := checkReceiveBlockRepeat(lv, block); err != Progress {
		//	return err, nil
		//}
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
	}

	return Progress, nil
}

func checkContractSendBlock(lv *LedgerVerifier, block *types.StateBlock, checkType byte) (ProcessResult, error) {
	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}
	result, err := checkSendBlock(lv, block, checkType)
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

func checkContractReceiveBlock(lv *LedgerVerifier, block *types.StateBlock, checkType byte) (ProcessResult, error) {
	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}

	result, err := checkStateBlock(lv, block, checkType)
	if err != nil || result != Progress {
		return result, err
	}
	var input *types.StateBlock
	// check previous
	if checkType == blockCheck {
		if !block.IsOpen() {
			// check previous
			if previous, err := lv.l.GetStateBlockConfirmed(block.Previous); err != nil {
				return GapPrevious, nil
			} else {
				//check fork
				if tm, err := lv.l.GetTokenMetaConfirmed(block.Address, block.GetToken()); err == nil && previous.GetHash() != tm.Header {
					return Fork, nil
				}
			}
		} else {
			//check fork
			if _, err := lv.l.GetTokenMetaConfirmed(block.Address, block.Token); err == nil {
				return Fork, nil
			}
		}
		//check smart c exist
		input, err = lv.l.GetStateBlockConfirmed(block.GetLink())
		if err != nil {
			return GapSource, nil
		}
	}
	// check previous
	if checkType == blockCheckCache {
		//if err := checkReceiveBlockRepeat(lv, block); err != Progress {
		//	return err, nil
		//}
		if !block.IsOpen() {
			// check previous
			if previous, err := lv.l.GetStateBlock(block.Previous); err != nil {
				return GapPrevious, nil
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
		input, err = lv.l.GetStateBlock(block.GetLink())
		if err != nil {
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

func checkReceiveBlockRepeat(lv *LedgerVerifier, block *types.StateBlock) ProcessResult {
	r := Progress
	if block.IsReceiveBlock() {
		var repeatedFound error
		err := lv.l.GetBlockCaches(func(b *types.StateBlock) error {
			if block.GetLink() == b.GetLink() && block.GetHash() != b.GetHash() {
				r = ReceiveRepeated
				return repeatedFound
			}
			return nil
		})
		if err != nil && err != repeatedFound {
			return Other
		}
	}
	return r
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
	am, err := lv.l.GetAccountMetaConfirmed(block.GetAddress(), txn)
	if err != nil && err != ledger.ErrAccountNotFound {
		return fmt.Errorf("get account meta error: %s", err)
	}
	tm, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken(), txn)
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
	//amCache, err := lv.l.GetAccountMetaCache(block.GetAddress(), txn)
	//if err != nil && err != ledger.ErrAccountNotFound {
	//	return fmt.Errorf("get account meta cache error: %s", err)
	//}
	//if amCache != nil {
	//	tmCache := amCache.Token(block.GetToken())
	//	if tmCache != nil && tm != nil {
	//		if tmCache.Header == tm.Header {
	//			err = lv.updateAccountMetaCache(block, amCache, txn)
	//			if err != nil {
	//				return fmt.Errorf("update AccountMeta Cache error: %s", err)
	//			}
	//		}
	//	}
	//}
	if err := lv.updateContractData(block, txn); err != nil {
		return fmt.Errorf("update contract data error: %s", err)
	}
	return nil
}

func (lv *LedgerVerifier) updatePending(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	switch block.Type {
	case types.Send:
		preBlk, err := lv.l.GetStateBlockConfirmed(block.Previous)
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
			tm.BlockCount = tm.BlockCount + 1
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

func (lv *LedgerVerifier) updateAccountMetaCache(block *types.StateBlock, am *types.AccountMeta, txn db.StoreTxn) error {
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
			tm.BlockCount = tm.BlockCount + 1
			tm.Modified = common.TimeNow().UTC().Unix()
		} else {
			am.Tokens = append(am.Tokens, tmNew)
		}
		if err := lv.l.AddOrUpdateAccountMetaCache(am, txn); err != nil {
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
		if err := lv.l.AddAccountMetaCache(&account, txn); err != nil {
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

func (lv *LedgerVerifier) Rollback(hash types.Hash) error {
	lv.l.RollbackChan <- hash
	return nil
}

func (lv *LedgerVerifier) RollbackBlock(hash types.Hash) error {
	if b, err := lv.l.HasBlockCache(hash); b && err == nil {
		lv.logger.Errorf("process rollback cache block: %s", hash.String())
		return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
			err = lv.rollbackBlockCache(hash, txn)
			if err != nil {
				lv.logger.Error(err)
				return err
			}
			return nil
		})
	}

	if b, err := lv.l.HasStateBlockConfirmed(hash); !b || err != nil {
		lv.logger.Errorf("rollback block not found: %s", hash.String())
		return nil
	}
	lv.logger.Errorf("process rollback  block: %s", hash.String())
	// get blocks to roll back
	rollbackMap := make(map[types.Address]*types.StateBlock)
	relatedBlocks := queue.NewQueue(512)
	relatedBlocks.Put(hash)
	lv.logger.Debugf("put block to queue %s ", hash.String())

	for {
		if v, ok, _ := relatedBlocks.Get(); ok {
			// get oldest block
			oldestHash := v.(types.Hash)
			oldestBlock, err := lv.l.GetStateBlockConfirmed(oldestHash)
			if err != nil {
				return fmt.Errorf("can not get block  %s", oldestHash.String())
			}
			lv.logger.Debugf("get block from  queue %s (%s) ,%s  ", oldestBlock.GetHash().String(), oldestBlock.GetType().String(), oldestBlock.Address.String())

			// put oldest block to rollbackMap
			if rBlock, ok := rollbackMap[oldestBlock.GetAddress()]; ok {
				lv.logger.Debugf("get %s from rollback of %s ", rBlock.GetHash().String(), oldestBlock.GetAddress().String())
				if t, err := lv.blockOrderCompare(oldestBlock, rBlock); t && err == nil {
					lv.logger.Debugf("put block to rollback %s (%s), %s ", oldestBlock.GetHash().String(), oldestBlock.GetType().String(), oldestBlock.Address.String())
					rollbackMap[oldestBlock.GetAddress()] = oldestBlock
				} else if err != nil {
					return err
				}
			} else {
				lv.logger.Debugf("put block  to rollback %s (%s), %s ", oldestBlock.GetHash().String(), oldestBlock.GetType().String(), oldestBlock.Address.String())
				rollbackMap[oldestBlock.GetAddress()] = oldestBlock
			}

			// get header block
			tm, err := lv.l.GetTokenMetaConfirmed(oldestBlock.GetAddress(), oldestBlock.GetToken())
			if err != nil {
				return fmt.Errorf("can not get account of block %s", oldestHash.String())
			}
			headerHash := tm.Header
			headerBlock, err := lv.l.GetStateBlockConfirmed(headerHash)
			if err != nil {
				return fmt.Errorf("can not get header block %s", headerHash.String())
			}

			curBlock := headerBlock
			// put link block to rollbackMap
			for {
				//if curBlock.IsOpen() {
				//	break
				//}

				if curBlock.IsSendBlock() {
					linkHash, err := lv.l.GetLinkBlock(curBlock.GetHash())
					// link not found is not error ,may be send block has created but receiver block has not created
					if err != nil && err != ledger.ErrLinkNotFound {
						return fmt.Errorf("can not get link hash %s", curBlock.GetHash().String())
					}
					if !linkHash.IsZero() {
						linkBlock, err := lv.l.GetStateBlockConfirmed(linkHash)
						if err != nil {
							return fmt.Errorf("can not get link block %s", linkHash.String())
						}
						if rBlock, ok := rollbackMap[linkBlock.GetAddress()]; ok {
							lv.logger.Debugf("get link %s from rollback of %s ", rBlock.GetHash().String(), linkBlock.GetAddress().String())
							if t, err := lv.blockOrderCompare(linkBlock, rBlock); t && err == nil {
								lv.logger.Debugf("put block to queue %s (%s) ,%s ", linkBlock.GetHash().String(), linkBlock.GetType().String(), linkBlock.Address.String())
								relatedBlocks.Put(linkBlock.GetHash())
							} else if err != nil {
								return err
							}
						} else {
							lv.logger.Debugf("put block to queue %s (%s), %s ", linkBlock.GetHash().String(), linkBlock.GetType().String(), linkBlock.Address.String())
							relatedBlocks.Put(linkBlock.GetHash())
						}
					}
				}

				if curBlock.GetHash() == oldestHash {
					break
				}

				curHash := curBlock.GetPrevious()
				curBlock, err = lv.l.GetStateBlockConfirmed(curHash)
				if err != nil {
					return fmt.Errorf("can not get previous block %s", curHash.String())
				}
			}
		} else {
			break
		}
	}
	// roll back blocks in badger
	return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
		err := lv.rollbackBlocks(rollbackMap, txn)
		if err != nil {
			lv.logger.Error(err)
		}
		return err
	})
}

// rollback cache blocks
func (lv *LedgerVerifier) rollbackBlockCache(hash types.Hash, txn db.StoreTxn) error {
	block, err := lv.l.GetBlockCache(hash, txn)
	if err != nil {
		return fmt.Errorf("get cache block (%s) err: %s", hash.String(), err)
	}

	// get all blocks of the address
	blocks := make([]*types.StateBlock, 0)
	err = lv.l.GetBlockCaches(func(b *types.StateBlock) error {
		if block.GetAddress() == b.GetAddress() && block.GetToken() == b.GetToken() {
			blocks = append(blocks, b)
		}
		return nil
	})
	if err != nil {
		lv.logger.Error("get block cache error")
	}

	// if receive repeated , rollback later block
	rollBlock := block
	if rollBlock.IsReceiveBlock() {
		for _, b := range blocks {
			if rollBlock.GetLink() == b.GetLink() && rollBlock.GetHash() != b.GetHash() {
				if b.Timestamp > rollBlock.Timestamp {
					rollBlock = b
				}
				break
			}
		}
	}

	// get rollback blocks
	tempBlocks := make([]*types.StateBlock, 0)
	for _, b := range blocks {
		if b.Timestamp >= rollBlock.Timestamp {
			tempBlocks = append(tempBlocks, b)
		}
	}

	// sort
	sort.Slice(tempBlocks, func(i, j int) bool {
		return tempBlocks[i].Timestamp < tempBlocks[j].Timestamp
	})

	rollBlocks := make([]*types.StateBlock, 0)
	rollBlocks = append(rollBlocks, rollBlock)
	for {
		findPre := false
		for _, b := range tempBlocks {
			if b.GetPrevious() == rollBlock.GetHash() {
				rollBlocks = append(rollBlocks, b)
				rollBlock = b
				findPre = true
				break
			}
		}
		if !findPre {
			break
		}
	}

	// delete blocks
	if err := lv.rollbackCacheData(rollBlocks, txn); err != nil {
		lv.logger.Error(err)
		return err
	}
	return nil
}

func (lv *LedgerVerifier) rollbackCacheData(blocks []*types.StateBlock, txn db.StoreTxn) error {
	if len(blocks) == 0 {
		return nil
	}
	for _, block := range blocks {
		if block.IsReceiveBlock() {
			pendingKey := types.PendingKey{
				Address: block.GetAddress(),
				Hash:    block.GetLink(),
			}
			if err := lv.l.UpdatePending(&pendingKey, types.PendingNotUsed, txn); err != nil {
				return err
			}
		}

		if err := lv.l.DeleteBlockCache(block.GetHash(), txn); err != nil {
			return fmt.Errorf("delete BlockCache fail(%s), hash(%s)", err, block.GetHash().String())
		}
		lv.l.EB.Publish(common.EventRollbackUnchecked, block.GetHash())
		lv.logger.Errorf("rollback delete cache block %s (previous: %s, type: %s,  address: %s)", block.GetHash().String(), block.GetPrevious().String(), block.GetType(), block.GetAddress().String())
	}

	blk := blocks[0]
	address := blk.GetAddress()
	lv.logger.Debug("delete token cache, ", address, blk.GetToken())
	err := lv.l.DeleteTokenMetaCache(address, blk.GetToken(), txn)
	if err == nil {
		ac, err := lv.l.GetAccountMetaCache(address, txn)
		if err != nil {
			return err
		}
		if len(ac.Tokens) == 0 {
			if err := lv.l.DeleteAccountMetaCache(address, txn); err != nil {
				return err
			}
			lv.logger.Errorf("rollback delete account %s", address.String())
		}
	}

	return nil
}

// rollback confirmed blocks
func (lv *LedgerVerifier) rollbackBlocks(rollbackMap map[types.Address]*types.StateBlock, txn db.StoreTxn) error {
	sendBlocks, err := lv.sendBlocksInRollback(rollbackMap, txn)
	if err != nil {
		return err
	}

	for address, oldestBlock := range rollbackMap {
		tm, err := lv.l.GetTokenMetaConfirmed(address, oldestBlock.GetToken())
		if err != nil {
			return fmt.Errorf("get tokenmeta error: %s (%s)", err, oldestBlock.GetHash().String())
		}
		headerHash := tm.Header

		hashCur := headerHash
		blockCur, err := lv.l.GetStateBlockConfirmed(hashCur)
		if err != nil {
			return fmt.Errorf("get block error: %s (%s)", err, hashCur.String())
		}

		for {
			blockType := blockCur.GetType()
			blockPre := new(types.StateBlock)
			if !blockCur.IsOpen() {
				blockPre, err = lv.l.GetStateBlockConfirmed(blockCur.Previous, txn)
				if err != nil {
					return fmt.Errorf("get previous block %s : %s", blockCur.Previous.String(), err)
				}
			}

			switch blockType {
			case types.Open:
				if err := lv.rollBackTokenDel(tm, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), open(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, nil, false, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), open(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(types.Hash{}, blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), open(%s)", err, hashCur)
				}
				if _, ok := sendBlocks[blockCur.GetLink()]; !ok {
					if err := lv.rollBackPendingAdd(blockCur, tm.Balance, blockCur.GetToken(), txn); err != nil {
						return fmt.Errorf("rollback pending fail(%s), open(%s)", err, hashCur)
					}
				}
			case types.Send:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, blockPre, true, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), send(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingDel(blockCur, txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), send(%s)", err, hashCur)
				}
			case types.Receive:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), receive(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), receive(%s)", err, hashCur)
				}
				if err := lv.rollBackRep(blockCur.GetRepresentative(), blockCur, blockPre, false, blockCur.GetToken(), txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), receive(%s)", err, hashCur)
				}
				if _, ok := sendBlocks[blockCur.GetLink()]; !ok {
					if err := lv.rollBackPendingAdd(blockCur, blockCur.GetBalance().Sub(blockPre.GetBalance()), blockCur.GetToken(), txn); err != nil {
						return fmt.Errorf("rollback pending fail(%s), receive(%s)", err, hashCur)
					}
				}
			case types.Change:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), change(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), change(%s)", err, hashCur)
				}
				if err := lv.rollBackRepChange(blockPre.GetRepresentative(), blockCur.GetRepresentative(), blockCur, txn); err != nil {
					return fmt.Errorf("rollback representative fail(%s), change(%s)", err, hashCur)
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
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), ContractReward(%s)", err, hashCur)
				}
				if _, ok := sendBlocks[blockCur.GetLink()]; !ok {
					if err := lv.rollBackPendingAdd(blockCur, types.ZeroBalance, types.ZeroHash, txn); err != nil {
						return fmt.Errorf("rollback pending fail(%s), ContractReward(%s)", err, hashCur)
					}
				}
				if err := lv.rollBackContractData(blockCur, txn); err != nil {
					return fmt.Errorf("rollback contract data fail(%s), ContractReward(%s)", err, blockCur.String())
				}
			case types.ContractSend:
				if err := lv.rollBackToken(tm, blockPre, txn); err != nil {
					return fmt.Errorf("rollback token fail(%s), ContractSend(%s)", err, hashCur)
				}
				if err := lv.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
					return fmt.Errorf("rollback frontier fail(%s), ContractSend(%s)", err, hashCur)
				}
				if err := lv.rollBackPendingDel(blockCur, txn); err != nil {
					return fmt.Errorf("rollback pending fail(%s), ContractSend(%s)", err, hashCur)
				}
				if err := lv.rollBackContractData(blockCur, txn); err != nil {
					return fmt.Errorf("rollback contract data fail(%s), ContractSend(%s)", err, blockCur.String())
				}
			}

			// rollback Block
			if err := lv.l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block error: %s, %s", err, hashCur)
			}
			lv.l.EB.Publish(common.EventRollbackUnchecked, hashCur)
			lv.logger.Errorf("rollback delete block %s (previous: %s, type: %s,  address: %s) ", hashCur.String(), blockCur.GetPrevious().String(), blockCur.GetType(), blockCur.GetAddress().String())

			if err := lv.checkBlockCache(blockCur, txn); err != nil {
				lv.logger.Errorf("roll back block cache error : %s", err)
				return err
			}

			if hashCur == oldestBlock.GetHash() {
				break
			}

			hashCur = blockCur.GetPrevious()
			blockCur, err = lv.l.GetStateBlockConfirmed(hashCur, txn)
			if err != nil {
				return fmt.Errorf("get previous block error %s, %s ", err, hashCur.String())
			}
		}
	}
	return nil
}

func (lv *LedgerVerifier) checkBlockCache(block *types.StateBlock, txn db.StoreTxn) error {
	rollbacks := make([]*types.StateBlock, 0)
	err := lv.l.GetBlockCaches(func(b *types.StateBlock) error {
		if block.GetAddress() == b.GetAddress() && block.GetToken() == b.GetToken() {
			rollbacks = append(rollbacks, b)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := lv.rollbackCacheData(rollbacks, txn); err != nil {
		return err
	}

	if block.IsSendBlock() {
		err := lv.l.GetBlockCaches(func(b *types.StateBlock) error {
			if block.GetHash() == b.GetLink() {
				err = lv.rollbackBlockCache(b.GetHash(), txn)
				if err != nil {
					lv.logger.Error(err)
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// all Send block to rollback
func (lv *LedgerVerifier) sendBlocksInRollback(blocks map[types.Address]*types.StateBlock, txn db.StoreTxn) (map[types.Hash]types.Address, error) {
	sendBlocks := make(map[types.Hash]types.Address)
	for address, oldestBlock := range blocks {
		tm, err := lv.l.GetTokenMetaConfirmed(address, oldestBlock.GetToken(), txn)
		if err != nil {
			return nil, fmt.Errorf("get tokenmeta error: %s (%s)", err, oldestBlock.GetHash().String())
		}

		curHash := tm.Header
		curBlock, err := lv.l.GetStateBlockConfirmed(curHash, txn)
		if err != nil {
			return nil, fmt.Errorf("get block error: %s (%s)", err, curHash.String())
		}
		for {
			if curBlock.IsSendBlock() {
				sendBlocks[curBlock.GetHash()] = curBlock.GetAddress()
			}

			if curBlock.GetHash() == oldestBlock.GetHash() {
				break
			}

			curHash := curBlock.GetPrevious()
			curBlock, err = lv.l.GetStateBlockConfirmed(curHash, txn)
			if err != nil {
				return nil, fmt.Errorf("can not get previous block %s", curHash.String())
			}
		}
	}
	return sendBlocks, nil
}

// if aBlock is created early than bBlock , return true, else return false
func (lv *LedgerVerifier) blockOrderCompare(aBlock, bBlock *types.StateBlock) (bool, error) {
	lv.logger.Debugf("block compare, %s %s %s ", aBlock.GetHash().String(), bBlock.GetHash().String(), aBlock.Address.String())
	if aBlock.GetAddress() != bBlock.GetAddress() || aBlock.GetToken() != bBlock.GetToken() {
		return false, fmt.Errorf("can not compare block, %s %s", aBlock.GetHash().String(), bBlock.GetHash().String())
	}
	tm, _ := lv.l.GetTokenMetaConfirmed(aBlock.GetAddress(), aBlock.GetToken())
	curHash := tm.Header
	for {
		curBlock, err := lv.l.GetStateBlockConfirmed(curHash)
		if err != nil {
			return false, fmt.Errorf("get block error %s", curBlock.String())
		}
		if curBlock.GetHash() == aBlock.GetHash() {
			return false, nil
		}
		if curBlock.GetHash() == bBlock.GetHash() {
			return true, nil
		}
		curHash = curBlock.GetPrevious()
		if curHash.IsZero() {
			return false, fmt.Errorf("can not find blocks when block compare")
		}
	}
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
	ac, err := lv.l.GetAccountMetaConfirmed(token.BelongTo, txn)
	if err != nil {
		return err
	}
	if pre.GetToken() == common.ChainToken() {
		ac.CoinVote = pre.GetVote()
		ac.CoinOracle = pre.GetOracle()
		ac.CoinNetwork = pre.GetNetwork()
		ac.CoinStorage = pre.GetStorage()
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
	tm.Modified = common.TimeNow().Unix()
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
	ac, err := lv.l.GetAccountMetaConfirmed(address, txn)
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
	blockLink, err := lv.l.GetStateBlockConfirmed(blockCur.GetLink(), txn)
	if err != nil {
		return fmt.Errorf("%s %s", err, blockCur.GetLink())
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
