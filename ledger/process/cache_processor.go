package process

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func (lv *LedgerVerifier) BlockCacheCheck(block types.Block) (ProcessResult, error) {
	if b, ok := block.(*types.StateBlock); ok {
		lv.logger.Info("check cache block, ", b.GetHash())
		if fn, ok := lv.checkCacheBlockFns[b.Type]; ok {
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

func checkCacheStateBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	hash := block.GetHash()
	address := block.GetAddress()

	lv.logger.Debug("check block ", hash)
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

	if block.GetType() == types.ContractSend {
		if types.IsNoSignContractAddress(types.Address(block.GetLink())) {
			return Progress, nil
		}
	}
	if block.GetType() == types.ContractReward {
		linkBlk, err := lv.l.GetStateBlock(block.GetLink())
		if err != nil {
			return GapSource, nil
		}
		if types.IsNoSignContractAddress(types.Address(linkBlk.GetLink())) {
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

func checkCacheSendBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkCacheStateBlock(lv, block)
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

func checkCacheReceiveBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkCacheStateBlock(lv, block)
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
		if pending, err := lv.l.GetPending(&pendingKey); err == nil {
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

func checkCacheChangeBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkCacheStateBlock(lv, block)
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

func checkCacheOpenBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	result, err := checkCacheStateBlock(lv, block)
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
		if pending, err := lv.l.GetPending(&pendingKey); err == nil {
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

func checkCacheContractSendBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}
	result, err := checkCacheSendBlock(lv, block)
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
		switch v := c.(type) {
		case contract.ChainContractV1:
			if err := v.DoSend(vmCtx, clone); err == nil {
				if bytes.EqualFold(block.Data, clone.Data) {
					return Progress, nil
				} else {
					lv.logger.Errorf("data not equal: %s, %s", block.Data, clone.Data)
					return InvalidData, nil
				}
			} else {
				lv.logger.Errorf("v1 ProcessSend error, block: %s, err: ", block.GetHash(), err)
				return Other, err
			}
		case contract.ChainContractV2:
			if types.IsRewardContractAddress(types.Address(block.GetLink())) {
				h, err := v.DoGapPov(vmCtx, clone)
				if err != nil {
					lv.logger.Errorf("do gapPov error: %s", err)
					return Other, err
				}
				if h != 0 {
					return GapPovHeight, nil
				}
			}
			if _, _, err := v.ProcessSend(vmCtx, clone); err == nil {
				if bytes.EqualFold(block.Data, clone.Data) {
					return Progress, nil
				} else {
					lv.logger.Errorf("data not equal: %s, %s", block.Data, clone.Data)
					return InvalidData, nil
				}
			} else {
				lv.logger.Errorf("v2 ProcessSend error, block: %s, err: ", block.GetHash(), err)
				return Other, err
			}
		default:
			return Other, fmt.Errorf("unsupported chain contract %s", reflect.TypeOf(v))
		}
	} else {
		//call vm.Run();
		return Other, fmt.Errorf("can not find chain contract %s", address.String())
	}
}

func checkCacheContractReceiveBlock(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}

	result, err := checkCacheStateBlock(lv, block)
	if err != nil || result != Progress {
		return result, err
	}
	var input *types.StateBlock
	// check previous
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

	address := types.Address(input.GetLink())

	if c, ok, err := contract.GetChainContract(address, input.Data); ok && err == nil {
		clone := block.Clone()
		//TODO:verify extra hash and commit to db
		vmCtx := vmstore.NewVMContext(lv.l)
		switch v := c.(type) {
		case contract.InternalContract:
			if g, e := v.DoReceive(vmCtx, clone, input); e == nil {
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
		default:
			return Other, fmt.Errorf("unsupported chain contract %s", reflect.TypeOf(v))
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

func (lv *LedgerVerifier) BlockCacheProcess(block types.Block) error {
	blk := block.(*types.StateBlock)
	am, err := lv.l.GetAccountMeta(blk.GetAddress())
	if err != nil && err != ledger.ErrAccountNotFound {
		return fmt.Errorf("get account meta cache error: %s", err)
	}
	return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
		if state, ok := block.(*types.StateBlock); ok {
			lv.logger.Info("process cache block, ", state.GetHash())
			err := lv.processCacheBlock(state, am, txn)
			if err != nil {
				lv.logger.Error(fmt.Sprintf("%s, cache block:%s", err.Error(), state.GetHash().String()))
				return err
			}
			return nil
		} else if _, ok := block.(*types.SmartContractBlock); ok {
			return errors.New("smart contract block")
		}
		return errors.New("invalid block")
	})
}

func (lv *LedgerVerifier) processCacheBlock(block *types.StateBlock, am *types.AccountMeta, txn db.StoreTxn) error {
	if err := lv.l.AddBlockCache(block, txn); err != nil {
		return err
	}
	if err := lv.updateAccountMetaCache(block, am, txn); err != nil {
		return fmt.Errorf("update account meta cache error: %s", err)
	}
	if err := lv.updatePendingKind(block, txn); err != nil {
		return fmt.Errorf("update pending kind error: %s", err)
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

func (lv *LedgerVerifier) updatePendingKind(block *types.StateBlock, txn db.StoreTxn) error {
	if block.IsReceiveBlock() {
		pendingKey := &types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		if err := lv.l.UpdatePending(pendingKey, types.PendingUsed, txn); err != nil {
			return err
		}
	}
	return nil
}
