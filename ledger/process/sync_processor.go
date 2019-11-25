package process

import (
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

func (lv *LedgerVerifier) BlockSyncCheck(block types.Block) (ProcessResult, error) {
	if b, ok := block.(*types.StateBlock); ok {
		lv.logger.Info("check sync block, ", b.GetHash())
		if c, ok := lv.syncBlockCheck[b.Type]; ok {
			r, err := c.Check(lv, b)
			if err != nil {
				lv.logger.Error(fmt.Sprintf("error:%s, sync block:%s", err.Error(), b.GetHash().String()))
			}
			if r != Progress {
				lv.logger.Infof(fmt.Sprintf("check sync result:%s, (%s)", r.String(), b.GetHash().String()))
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

func newSyncBlockCheck() map[types.BlockType]blockCheck {
	c := make(map[types.BlockType]blockCheck)
	c[types.Open] = &syncOpenBlockCheck{}
	c[types.Send] = &sendBlockCheck{}
	c[types.Receive] = &syncReceiveBlockCheck{}
	c[types.Change] = &changeBlockCheck{}
	c[types.Online] = &changeBlockCheck{}
	c[types.ContractSend] = &contractSendBlockCheck{}
	c[types.ContractReward] = &syncContractReceiveBlockCheck{}
	return c
}

type syncReceiveBlockCheck struct {
	syncBlockBaseInfoCheck
	blockForkCheck
}

func (c *syncReceiveBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type syncOpenBlockCheck struct {
	syncBlockBaseInfoCheck
	blockForkCheck
}

func (c *syncOpenBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//check previous
	if !block.Previous.IsZero() {
		return Other, fmt.Errorf("open block previous is not zero")
	}

	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type syncContractReceiveBlockCheck struct {
	syncBlockBaseInfoCheck
	blockForkCheck
}

func (c *syncContractReceiveBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//ignore chain genesis block
	if common.IsGenesisBlock(block) {
		return Progress, nil
	}
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

func (lv *LedgerVerifier) BlockSyncProcess(block types.Block) error {
	return lv.l.BatchUpdate(func(txn db.StoreTxn) error {
		if state, ok := block.(*types.StateBlock); ok {
			lv.logger.Info("process sync block, ", state.GetHash())
			err := lv.processSyncBlock(state, txn)
			if err != nil {
				lv.logger.Error(fmt.Sprintf("%s, sync block:%s", err.Error(), state.GetHash().String()))
				return err
			}
			return nil
		} else if _, ok := block.(*types.SmartContractBlock); ok {
			return errors.New("smart contract block")
		}
		return errors.New("invalid block")
	})
}

func (lv *LedgerVerifier) processSyncBlock(block *types.StateBlock, txn db.StoreTxn) error {
	if err := lv.l.AddSyncStateBlock(block, txn); err != nil {
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
	if err := lv.l.AddSyncCacheBlock(block, txn); err != nil {
		return fmt.Errorf("add sync block error: %s", err)
	}
	if err := lv.updateRepresentative(block, am, tm, txn); err != nil {
		return fmt.Errorf("update representative error: %s", err)
	}
	if err := lv.updateFrontier(block, tm, txn); err != nil {
		return fmt.Errorf("update frontier error: %s", err)
	}
	if err := lv.updateAccountMeta(block, am, txn); err != nil {
		return fmt.Errorf("update account meta error: %s", err)
	}
	return nil
}

func (lv *LedgerVerifier) BlockSyncDoneProcess(block *types.StateBlock) error {
	txn := lv.l.Store.NewTransaction(true)
	if block.IsSendBlock() {
		if _, err := lv.l.GetLinkBlock(block.GetHash(), txn); err == ledger.ErrLinkNotFound {
			lv.logger.Info("sync done, process send block, ", block.GetHash())
			hash := block.GetHash()
			switch block.Type {
			case types.Send:
				preBlk, err := lv.l.GetStateBlockConfirmed(block.Previous, txn)
				if err != nil {
					lv.logger.Errorf("block(%s) sync done error: %s", block.GetHash(), err)
					return err
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
				lv.logger.Info("sync done, add pending, ", pendingKey)
				if err := lv.l.AddPending(&pendingKey, &pending, txn); err != nil {
					lv.logger.Errorf("block(%s) sync done error: %s", block.GetHash(), err)
					return err
				}
			case types.ContractSend:
				if c, ok, err := contract.GetChainContract(types.Address(block.Link), block.Data); ok && err == nil {
					switch v := c.(type) {
					case contract.ChainContractV1:
						if pendingKey, pendingInfo, err := v.DoPending(block); err == nil && pendingKey != nil {
							lv.logger.Info("sync done, add pending contract1, ", pendingKey)
							if err := lv.l.AddPending(pendingKey, pendingInfo, txn); err != nil {
								lv.logger.Errorf("block(%s) sync done error: %s", block.GetHash(), err)
								return err
							}
						}
					case contract.ChainContractV2:
						vmCtx := vmstore.NewVMContext(lv.l)
						if pendingKey, pendingInfo, err := v.ProcessSend(vmCtx, block); err == nil && pendingKey != nil {
							lv.logger.Info("sync done, add pending contract2, ", pendingKey)
							if err := lv.l.AddPending(pendingKey, pendingInfo, txn); err != nil {
								lv.logger.Errorf("block(%s) sync done error: %s", block.GetHash(), err)
								return err
							}
						}
					default:
						lv.logger.Errorf("unsupported chain contract %s", reflect.TypeOf(v))
						return errors.New("unsupported chain contract")
					}
				}
			}
		} else {
			if err != nil {
				lv.logger.Info("sync done, process send block error, ", block.GetHash())
			}
		}
	}

	if block.IsReceiveBlock() {
		// if send block sync done in last time, it will create pending
		pendingKey := types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		if pi, err := lv.l.GetPending(&pendingKey, txn); pi != nil && err == nil {
			lv.logger.Info("sync done, delete pending, ", pendingKey)
			if err := lv.l.DeletePending(&pendingKey, txn); err != nil {
				lv.logger.Errorf("block(%s) sync done error: %s", block.GetHash(), err)
				return err
			}
		}
	}

	if block.IsContractBlock() {
		lv.logger.Info("sync done, process contract block, ", block.GetHash())
		if err := lv.updateContractData(block, txn); err != nil {
			return fmt.Errorf(" update contract data error(%s): %s", block.GetHash().String(), err)
		}
	}

	if err := lv.l.DeleteSyncCacheBlock(block.GetHash(), txn); err != nil {
		lv.logger.Errorf("block(%s) sync done error: %s", block.GetHash(), err)
		return err
	}

	if err := txn.Commit(nil); err != nil {
		lv.logger.Errorf("block(%s) sync done error: %s", block.GetHash(), err)
		return err
	}

	return nil
}
