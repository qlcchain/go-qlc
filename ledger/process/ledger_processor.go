/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package process

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type LedgerVerifier struct {
	l               *ledger.Ledger
	blockCheck      map[types.BlockType]blockCheck
	cacheBlockCheck map[types.BlockType]blockCheck
	syncBlockCheck  map[types.BlockType]blockCheck
	logger          *zap.SugaredLogger
}

func NewLedgerVerifier(l *ledger.Ledger) *LedgerVerifier {
	return &LedgerVerifier{
		l:               l,
		blockCheck:      newBlockCheck(),
		cacheBlockCheck: newCacheBlockCheck(),
		syncBlockCheck:  newSyncBlockCheck(),
		logger:          log.NewLogger("ledger_verifier"),
	}
}

func (lv *LedgerVerifier) Process(block types.Block) (ProcessResult, error) {
	if b, ok := block.(*types.StateBlock); ok {
		if r, err := lv.BlockCheck(b); r != Progress || err != nil {
			return r, err
		}
		if err := lv.BlockProcess(b); err != nil {
			return Other, err
		}
		return Progress, nil
	}
	return Other, errors.New("invalid block")
}

func (lv *LedgerVerifier) BlockCheck(block *types.StateBlock) (ProcessResult, error) {
	lv.logger.Info("check block, ", block.GetHash())
	if c, ok := lv.blockCheck[block.Type]; ok {
		r, err := c.Check(lv, block)
		if err != nil {
			lv.logger.Error(fmt.Sprintf("error:%s, block:%s", err.Error(), block.GetHash().String()))
		}
		if r != Progress {
			if r == UnReceivable {
				if _, ok := lv.l.VerifiedData[block.GetHash()]; ok {
					return Progress, nil
				}
			}
			lv.logger.Debugf(fmt.Sprintf("process result:%s, block:%s, type: %s", r.String(), block.GetHash().String(), block.GetType().String()))
		}
		return r, err
	} else {
		return Other, fmt.Errorf("unsupport block type %s", block.Type.String())
	}

}

func newBlockCheck() map[types.BlockType]blockCheck {
	r := make(map[types.BlockType]blockCheck)
	r[types.Open] = &openBlockCheck{}
	r[types.Send] = &sendBlockCheck{}
	r[types.Receive] = &receiveBlockCheck{}
	r[types.Change] = &changeBlockCheck{}
	r[types.Online] = &onlineBlockCheck{}
	r[types.ContractSend] = &contractSendBlockCheck{}
	r[types.ContractReward] = &contractReceiveBlockCheck{}
	return r
}

type sendBlockCheck struct {
	blockBaseInfoCheck
	blockForkCheck
	blockBalanceCheck
}

func (c *sendBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.balance(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type contractSendBlockCheck struct {
	blockBaseInfoCheck
	blockForkCheck
	blockBalanceCheck
	blockContractCheck
}

func (c *contractSendBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
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
	if r, err := c.balance(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.contract(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type receiveBlockCheck struct {
	blockBaseInfoCheck
	blockForkCheck
	blockSourceCheck
	blockPendingCheck
}

func (c *receiveBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.source(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.pending(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type contractReceiveBlockCheck struct {
	blockBaseInfoCheck
	blockForkCheck
	blockPendingCheck
	blockSourceCheck
	blockContractCheck
}

func (c *contractReceiveBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
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
	if r, err := c.pending(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.source(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.contract(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type openBlockCheck struct {
	blockBaseInfoCheck
	blockForkCheck
	blockSourceCheck
	blockPendingCheck
}

func (c *openBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	//check previous
	if !block.Previous.IsZero() {
		return Other, fmt.Errorf("open block previous is not zero")
	}

	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.source(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.pending(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type changeBlockCheck struct {
	blockBaseInfoCheck
	blockForkCheck
	blockBalanceCheck
}

func (c *changeBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	// check link
	if !block.Link.IsZero() {
		return Other, fmt.Errorf("invalid link hash")
	}
	// check chain token
	if block.GetToken() != common.ChainToken() {
		return Other, fmt.Errorf("invalid token %s, common chain token is %s", block.GetToken().String(), common.ChainToken().String())
	}
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.balance(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

type onlineBlockCheck struct {
	blockBaseInfoCheck
	blockForkCheck
	blockBalanceCheck
}

func (c *onlineBlockCheck) Check(lv *LedgerVerifier, block *types.StateBlock) (ProcessResult, error) {
	// check link
	if !block.Link.IsZero() {
		return Other, fmt.Errorf("invalid link hash")
	}
	// check chain token
	if block.GetToken() != common.ChainToken() {
		return Other, fmt.Errorf("invalid token %s, chain token is %s", block.GetToken().String(), common.ChainToken().String())
	}
	if r, err := c.baseInfo(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.fork(lv, block); r != Progress || err != nil {
		return r, err
	}
	if r, err := c.balance(lv, block); r != Progress || err != nil {
		return r, err
	}
	return Progress, nil
}

func (lv *LedgerVerifier) BlockProcess(block *types.StateBlock) error {
	lv.logger.Infof("block  process: %s(%s) ", block.GetHash().String(), block.GetType().String())
	err := lv.l.Cache().BatchUpdate(func(c *ledger.Cache) error {
		err := lv.processStateBlock(block, c)
		if err != nil {
			lv.logger.Error(fmt.Sprintf("block  process error: %s, block:%s", err.Error(), block.GetHash().String()))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	lv.logger.Debug("publish addRelation,", block.GetHash())
	lv.l.EB.Publish(topic.EventAddRelation, block)
	return nil
}

func (lv *LedgerVerifier) processStateBlock(block *types.StateBlock, cache *ledger.Cache) error {
	am, err := lv.l.GetAccountMetaConfirmed(block.GetAddress())
	if err != nil && err != ledger.ErrAccountNotFound {
		return fmt.Errorf("get account meta error: %s, %s", err, am.Address.String())
	}
	tm, err := lv.l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken())
	if err != nil && err != ledger.ErrAccountNotFound && err != ledger.ErrTokenNotFound {
		return fmt.Errorf("get token meta error: %s", err)
	}
	if err := lv.l.UpdateStateBlock(block, cache); err != nil {
		return err
	}
	if err := lv.updatePending(block, tm, cache); err != nil {
		return fmt.Errorf("update pending error: %s", err)
	}
	if err := lv.updateFrontier(block, tm, cache); err != nil {
		return fmt.Errorf("update frontier error: %s", err)
	}
	if err := lv.updateContractData(block, cache); err != nil {
		return fmt.Errorf("update contract data error: %s", err)
	}
	if err := lv.updateRepresentative(block, am, tm, cache); err != nil {
		return fmt.Errorf("update representative error: %s", err)
	}
	if err := lv.updateAccountMeta(block, am, cache); err != nil {
		return fmt.Errorf("update account meta error: %s", err)
	}
	return nil
}

func (lv *LedgerVerifier) updatePending(block *types.StateBlock, tm *types.TokenMeta, cache *ledger.Cache) error {
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
		if err := lv.l.AddPending(&pendingKey, &pending, cache); err != nil {
			return err
		}
	case types.Open, types.Receive:
		pendingKey := types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		lv.logger.Debug("delete pending, ", pendingKey)
		if err := lv.l.DeletePending(&pendingKey, cache); err != nil {
			return err
		}
	case types.ContractSend:
		if c, ok, err := contract.GetChainContract(types.Address(block.Link), block.Data); ok && err == nil {
			d := c.GetDescribe()
			switch d.GetVersion() {
			case contract.SpecVer1:
				if pendingKey, pendingInfo, err := c.DoPending(block); err == nil && pendingKey != nil {
					lv.logger.Debug("contractSend add pending , ", pendingKey)
					if err := lv.l.AddPending(pendingKey, pendingInfo, cache); err != nil {
						return err
					}
				}
			case contract.SpecVer2:
				vmCtx := vmstore.NewVMContext(lv.l)
				if pendingKey, pendingInfo, err := c.ProcessSend(vmCtx, block); err == nil && pendingKey != nil {
					lv.logger.Debug("contractSend add pending , ", pendingKey)
					if err := lv.l.AddPending(pendingKey, pendingInfo, cache); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unsupported chain contract version %d", d.GetVersion())
			}
		}
	case types.ContractReward:
		pendingKey := types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		lv.logger.Debug("contractReward delete pending, ", pendingKey)
		if err := lv.l.DeletePending(&pendingKey, cache); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) updateRepresentative(block *types.StateBlock, am *types.AccountMeta, tm *types.TokenMeta, cache *ledger.Cache) error {
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
			if err := lv.l.SubRepresentation(tm.Representative, oldBenefit, cache); err != nil {
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
		if err := lv.l.AddRepresentation(block.GetRepresentative(), newBenefit, cache); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) updateFrontier(block *types.StateBlock, tm *types.TokenMeta, cache *ledger.Cache) error {
	hash := block.GetHash()
	frontier := &types.Frontier{
		HeaderBlock: hash,
	}
	if tm != nil {
		if frontier, err := lv.l.GetFrontier(tm.Header); err == nil {
			lv.logger.Debug("delete frontier, ", *frontier)
			if err := lv.l.DeleteFrontier(frontier.HeaderBlock, cache); err != nil {
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
	if err := lv.l.AddFrontier(frontier, cache); err != nil {
		return err
	}
	return nil
}

func (lv *LedgerVerifier) updateAccountMeta(block *types.StateBlock, am *types.AccountMeta, cache *ledger.Cache) error {
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
		if err := lv.l.UpdateAccountMeta(am, cache); err != nil {
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
		if err := lv.l.AddAccountMeta(&account, cache); err != nil {
			return err
		}
	}
	return nil
}

func (lv *LedgerVerifier) updateContractData(block *types.StateBlock, cache *ledger.Cache) error {
	if !common.IsGenesisBlock(block) {
		switch block.GetType() {
		case types.ContractReward:
			input, err := lv.l.GetStateBlock(block.GetLink())
			if err != nil {
				return err
			}
			address := types.Address(input.GetLink())
			c, ok, err := contract.GetChainContract(address, input.Data)
			if !ok || err != nil {
				lv.logger.Errorf("invaild contract %s", err)
				return err
			}
			clone := block.Clone()
			vmCtx := vmstore.NewVMContext(lv.l)
			g, err := c.DoReceive(vmCtx, clone, input)
			if err != nil {
				lv.logger.Error("updateContract DoReceive error: ", err)
				return err
			}
			if len(g) > 0 {
				ctx := g[0].VMContext
				if ctx != nil {
					err := ctx.SaveStorage(cache)
					if err != nil {
						lv.logger.Error("save storage error: ", err)
						return err
					}
					err = ctx.SaveTrie(cache)
					if err != nil {
						lv.logger.Error("save trie error: ", err)
						return err
					}
					return nil
				}
			}
			return errors.New("invalid contract data")
		case types.ContractSend:
			c, ok, err := contract.GetChainContract(types.Address(block.Link), block.Data)
			if ok && err == nil {
				d := c.GetDescribe()
				switch d.GetVersion() {
				case contract.SpecVer2:
					vmCtx := vmstore.NewVMContext(lv.l)
					if _, _, err := c.ProcessSend(vmCtx, block); err == nil {
						if err := vmCtx.SaveStorage(cache); err != nil {
							lv.logger.Error("save storage error: ", err)
							return err
						}
						if err = vmCtx.SaveTrie(cache); err != nil {
							lv.logger.Error("save trie error: ", err)
							return err
						}
					} else {
						lv.logger.Errorf("process send error, ", err)
					}
				}
			}
		}
	}
	return nil
}
