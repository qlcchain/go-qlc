package ledger

import (
	"bytes"
	"time"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/test/mock"
)

type ProcessResult byte

const (
	Progress ProcessResult = iota
	BadWork
	BadSignature
	Old
	Fork
	GapPrevious
	GapSource
	BalanceMismatch
	UnReceivable
	Other
)

func (l *Ledger) Process(block types.Block) (ProcessResult, error) {
	r, err := l.BlockCheck(block)
	if r != Progress {
		return r, err
	}
	if err := l.BlockProcess(block); err != nil {
		return Other, err
	}
	return Progress, nil
}

func (l *Ledger) BlockCheck(block types.Block) (ProcessResult, error) {
	txn, b := l.getTxn(false)
	defer l.releaseTxn(txn, b)

	switch block.GetType() {
	case types.State:
		if state, ok := block.(*types.StateBlock); ok {
			return l.checkStateBlock(state, txn)
		}
	case types.SmartContract:
		return Other, errors.New("smartcontract block")
	default:
	}
	return Other, errors.New("invalid block")
}

func (l *Ledger) checkStateBlock(block *types.StateBlock, txn db.StoreTxn) (ProcessResult, error) {

	hash := block.GetHash()
	pre := block.GetPrevious()
	link := block.GetLink()
	address := block.GetAddress()

	logger.Info("process block, ", hash)

	// make sure smart contract token exist
	// ...

	if !block.IsValid() {
		logger.Info("invalid work")
		return BadWork, nil
	}

	blockExist, err := l.HasStateBlock(hash, txn)
	if err != nil {
		return Other, err
	}
	if blockExist == true {
		logger.Info("block already exist")
		return Old, nil
	}

	// make sure the signature of this block is valid
	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		logger.Info("bad signature")
		return BadSignature, nil
	}

	if pre.IsZero() && bytes.EqualFold(address[:], link[:]) {
		logger.Info("genesis block")
		return Progress, nil
	}

	transferAmount := block.GetBalance()
	tmExist, err := l.HasTokenMeta(address, block.GetToken(), txn)
	if err != nil {
		return Other, err
	}
	isSend := false
	if tmExist { // this tm chain exist , not open
		tm, err := l.GetTokenMeta(address, block.GetToken(), txn)
		if err != nil {
			return Other, err
		}

		if pre.IsZero() {
			logger.Info("fork: token meta exist, but pre hash is zero")
			return Fork, nil
		}
		preExist, err := l.HasStateBlock(block.GetPrevious(), txn)
		if err != nil {
			return Other, err
		}
		if !preExist {
			logger.Info("gap previous: token meta exist, but pre block not exist")
			return GapPrevious, nil
		}
		if block.GetPrevious() != tm.Header {
			logger.Info("fork: pre block exist, but pre hash not equal account token's header hash")
			return Fork, nil
		}
		if block.GetBalance().Compare(tm.Balance) == types.BalanceCompSmaller {
			isSend = true
			transferAmount = tm.Balance.Sub(block.GetBalance()) // send
		} else {
			transferAmount = block.GetBalance().Sub(tm.Balance) // receive or change
		}

	} else { // open
		if !pre.IsZero() {
			logger.Info("gap previous: token meta not exist, but pre hash is not zero")
			return GapPrevious, nil
		}
		if link.IsZero() {
			logger.Info("gap source: token meta not exist, but link hash is zero")
			return GapSource, nil
		}
	}
	if !isSend {
		if !link.IsZero() { // open or receive
			linkExist, err := l.HasStateBlock(link, txn)
			if err != nil {
				return Other, err
			}
			if !linkExist {
				logger.Info("gap source: open or receive block, but link block is not exist")
				return GapSource, nil
			}
			PendingKey := types.PendingKey{
				Address: address,
				Hash:    link,
			}
			pending, err := l.GetPending(PendingKey, txn)
			if err != nil {
				if err == ErrPendingNotFound {
					logger.Info("unreceivable: open or receive block, but pending not exist")
					return UnReceivable, nil
				}
				return Other, err
			}
			if !pending.Amount.Equal(transferAmount) {
				logger.Infof("balance mismatch: open or receive block, but pending amount(%s) not equal transfer amount(%s)", pending.Amount, transferAmount)
				return BalanceMismatch, nil
			}
		} else { //change
			if !transferAmount.Equal(types.ZeroBalance) {
				logger.Info("balance mismatch: change block, but transfer amount is not 0 ")
				return BalanceMismatch, nil
			}
		}
	}
	return Progress, nil
}

func (l *Ledger) BlockProcess(block types.Block) error {
	txn, b := l.getTxn(true)
	defer l.releaseTxn(txn, b)
	//err := l.BatchUpdate(func(txn db.StoreTxn) error {
	switch block.GetType() {
	case types.State:
		if state, ok := block.(*types.StateBlock); ok {
			return l.processStateBlock(state, txn)
		} else {
			return errors.New("invalid block")
		}
	case types.SmartContract:
		return nil
	default:
		return errors.New("invalid block type")
	}
	//})
	//if err != nil {
	//	return err
	//}
	//return nil
}

func (l *Ledger) processStateBlock(block *types.StateBlock, txn db.StoreTxn) error {
	hash := block.GetHash()
	logger.Info("add block, ", hash)
	if err := l.AddBlock(block, txn); err != nil {
		return err
	}

	tm, err := l.GetTokenMeta(block.GetAddress(), block.GetToken(), txn)
	if err != nil && err != ErrTokenNotFound && err != ErrAccountNotFound {
		return err
	}

	if err := l.updateRepresentative(block, tm, txn); err != nil {
		return err
	}

	if err := l.updatePending(block, tm, txn); err != nil {
		return err
	}

	if err := l.updateAccountMeta(hash, block.GetRepresentative(), block.GetAddress(), block.GetToken(), block.GetBalance(), txn); err != nil {
		return err
	}

	if err := l.updateFrontier(hash, tm, txn); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) updatePending(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	isSend, err := l.isSend(block, txn)
	if err != nil {
		return err
	}
	hash := block.GetHash()
	link := block.GetLink()
	if isSend { // send
		pending := types.PendingInfo{
			Source: block.GetAddress(),
			Type:   block.GetToken(),
			Amount: tm.Balance.Sub(block.GetBalance()),
		}
		pendingkey := types.PendingKey{
			Address: types.Address(block.GetLink()),
			Hash:    hash,
		}
		logger.Info("add pending, ", pendingkey)
		if err := l.AddPending(pendingkey, &pending, txn); err != nil {
			return err
		}
	} else if !link.IsZero() { // not change
		pre := block.GetPrevious()
		address := block.GetAddress()
		if !(pre.IsZero() && bytes.EqualFold(address[:], link[:])) { // not genesis
			pendingkey := types.PendingKey{
				Address: block.GetAddress(),
				Hash:    block.GetLink(),
			}
			logger.Info("delete pending, ", pendingkey)
			if err := l.DeletePending(pendingkey, txn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Ledger) isSend(block *types.StateBlock, txn db.StoreTxn) (bool, error) {
	isSend := false
	tmExist, err := l.HasTokenMeta(block.GetAddress(), block.GetToken(), txn)
	if err != nil {
		return false, err
	}
	if tmExist {
		tm, err := l.GetTokenMeta(block.GetAddress(), block.GetToken(), txn)
		if err != nil {
			return false, err
		}
		if block.GetBalance().Compare(tm.Balance) == types.BalanceCompSmaller {
			isSend = true
		}
	}
	return isSend, nil
}

func (l *Ledger) updateRepresentative(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	if block.GetToken() == mock.GetChainTokenType() {
		if tm != nil && !tm.Representative.IsZero() {
			logger.Infof("sub rep %s from %s ", tm.Balance, tm.Representative)
			if err := l.SubRepresentation(tm.Representative, tm.Balance, txn); err != nil {
				return err
			}
			//blk, err := l.GetStateBlock(tm.Representative, txn)
			//if err != nil {
			//	return err
			//}
			//if state, ok := blk.(*types.StateBlock); ok {
			//	logger.Infof("sub rep %s from %s ", tm.Balance, state.GetRepresentative())
			//	if err := l.SubRepresentation(state.GetRepresentative(), tm.Balance, txn); err != nil {
			//		return err
			//	}
			//} else {
			//	return errors.New("invalid block")
			//}
		}
		logger.Infof("add rep %s to %s ", block.GetBalance(), block.GetRepresentative())
		if err := l.AddRepresentation(block.GetRepresentative(), block.GetBalance(), txn); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) updateFrontier(hash types.Hash, tm *types.TokenMeta, txn db.StoreTxn) error {
	frontier := &types.Frontier{
		HeaderBlock: hash,
	}
	if tm != nil {
		if frontier, err := l.GetFrontier(tm.Header, txn); err == nil {
			logger.Info("delete frontier, ", *frontier)
			if err := l.DeleteFrontier(frontier.HeaderBlock, txn); err != nil {
				return err
			}
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	logger.Info("add frontier,", *frontier)
	if err := l.AddFrontier(frontier, txn); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) updateAccountMeta(hash types.Hash, rep types.Address, address types.Address, token types.Hash, balance types.Balance, txn db.StoreTxn) error {
	tmExist, err := l.HasTokenMeta(address, token, txn)
	if err != nil {
		return err
	}
	if tmExist {
		token, err := l.GetTokenMeta(address, token, txn)
		if err != nil {
			return err
		}
		token.Header = hash
		token.Representative = rep
		token.Balance = balance
		token.BlockCount = token.BlockCount + 1
		token.Modified = time.Now().Unix()
		logger.Info("update tokenmeta, ", *token)
		if err := l.UpdateTokenMeta(address, token, txn); err != nil {
			return err
		}
	} else {
		acExist, err := l.HasAccountMeta(address, txn)
		if err != nil {
			return err
		}
		tm := types.TokenMeta{
			Type:           token,
			Header:         hash,
			Representative: rep,
			OpenBlock:      hash,
			Balance:        balance,
			BlockCount:     1,
			BelongTo:       address,
			Modified:       time.Now().Unix(),
		}
		if acExist {
			logger.Info("add tokenmeta,", token)
			if err := l.AddTokenMeta(address, &tm, txn); err != nil {
				return err
			}
		} else {
			account := types.AccountMeta{
				Address: address,
				Tokens:  []*types.TokenMeta{&tm},
			}
			logger.Info("add accountmeta,", token)
			if err := l.AddAccountMeta(&account, txn); err != nil {
				return err
			}
		}
	}
	return nil
}
