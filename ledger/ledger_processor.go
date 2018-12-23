package ledger

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

var chain_token_type = "125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36"

type process_result byte

const (
	Progress process_result = iota
	Bad_signature
	Old
	Fork
	Gap_previous
	Gap_source
	Balance_mismatch
	Unreceivable
	Other
)

func (l *Ledger) Process(block types.Block) process_result {
	process_result := Progress

	err := l.BatchUpdate(func(txn db.StoreTxn) error {
		switch block.GetType() {
		case types.State:
			if state, ok := block.(*types.StateBlock); ok {
				result, err := l.checkStateBasicInfo(state, txn)
				if err != nil {
					return err
				}
				process_result = result
				return nil
			} else {
				return errors.New("unvalid block")
			}
		case types.SmartContract:
			return nil
		default:
			return errors.New("unvalid block type")

		}
	})
	if err != nil {
		return Other
	}
	return process_result
}
func (l *Ledger) checkStateBasicInfo(block *types.StateBlock, txn db.StoreTxn) (process_result, error) {

	hash := block.GetHash()
	pre := block.GetPrevious()
	link := block.GetLink()
	address := block.GetAddress()

	logger.Info("process block, ", hash)

	// make sure smart contract token exist
	// ...

	// make sure the block does not exist
	block_exist, err := l.HasBlock(hash, txn)
	if err != nil {
		return Other, err
	}
	if block_exist == true {
		logger.Info("block already exist")
		return Old, nil
	}

	// make sure the signature of this block is valid
	signature := block.GetSignature()
	if address.Verify(hash[:], signature[:]) {
		logger.Info("bad signature")
		return Bad_signature, nil
	}

	is_send := false
	transfer_amount := block.GetBalance()

	tm_exist, err := l.HasTokenMeta(address, block.GetToken(), txn)
	if err != nil {
		return Other, err
	}
	if tm_exist { // this tm chain exist , not open
		tm, err := l.GetTokenMeta(address, block.GetToken(), txn)
		if err != nil {
			return Other, err
		}

		if pre.IsZero() {
			logger.Info("fork: token meta exist, but pre hash is zero")
			return Fork, nil
		}
		previous_exist, err := l.HasBlock(block.GetPrevious(), txn)
		if err != nil {
			return Other, err
		}
		if !previous_exist {
			logger.Info("gap previous: token meta exist, but pre block not exist")
			return Gap_previous, nil
		}
		if block.GetPrevious() != tm.Header {
			logger.Info("fork: pre block exist, but pre hash not equal account token's header hash")
			return Fork, nil
		}
		if block.GetBalance().Compare(tm.Balance) == types.BalanceCompSmaller {
			is_send = true
			transfer_amount = tm.Balance.Sub(block.GetBalance()) // send
		} else {
			transfer_amount = block.GetBalance().Sub(tm.Balance) // receive or change
		}

	} else { // open
		if !pre.IsZero() {
			logger.Info("gap previous: token meta not exist, but pre hash is not zero")
			return Gap_previous, nil
		}
		if link.IsZero() {
			logger.Info("gap source: token meta not exist, but link hash is zero")
			return Gap_source, nil
		}
	}
	if !is_send {
		if !link.IsZero() { // open or receive
			link_exist, err := l.HasBlock(link, txn)
			if err != nil {
				return Other, err
			}
			if !link_exist {
				logger.Info("gap source: open or receive block, but link block is not exist")
				return Gap_source, nil
			}
			PendingKey := types.PendingKey{
				Address: address,
				Hash:    link,
			}
			pending, err := l.GetPending(PendingKey, txn)
			if err != nil {
				if err == ErrPendingNotFound {
					logger.Info("unreceivable: open or receive block, but pending not exist")
					return Unreceivable, nil
				}
				return Other, err
			}
			if !pending.Amount.Equal(transfer_amount) {
				logger.Infof("balance mismatch: open or receive block, but pending amount(%s) not equal transfer amount(%s)", pending.Amount, transfer_amount)
				return Balance_mismatch, nil
			}
		} else { //change
			if !transfer_amount.Equal(types.ParseBalanceInts(0, 0)) {
				logger.Info("balance mismatch: change block, but transfer amount is not 0 ")
				return Balance_mismatch, nil
			}
		}
	}
	if err := l.addBasicInfo(block, is_send, txn); err != nil {
		return Other, err
	}
	return Progress, nil
}
func (l *Ledger) addBasicInfo(block *types.StateBlock, is_send bool, txn db.StoreTxn) error {
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

	if err := l.updatePending(block, tm, is_send, txn); err != nil {
		return err
	}

	if err := l.updateAccountMeta(hash, hash, block.GetAddress(), block.GetToken(), block.GetBalance(), txn); err != nil {
		return err
	}

	if err := l.updateFrontier(hash, tm, txn); err != nil {
		return err
	}

	return nil
}
func (l *Ledger) updatePending(block *types.StateBlock, tm *types.TokenMeta, is_send bool, txn db.StoreTxn) error {
	hash := block.GetHash()
	link := block.GetLink()
	if is_send { // send
		pending := types.PendingInfo{
			Source: block.GetAddress(),
			Type:   block.GetToken(),
			Amount: tm.Balance.Sub(block.GetBalance()),
		}
		logger.Info("add pending")
		pendingkey := types.PendingKey{
			Address: types.Address(block.GetLink()),
			Hash:    hash,
		}
		if err := l.AddPending(pendingkey, &pending, txn); err != nil {
			return err
		}
	} else if !link.IsZero() { // receive or open
		logger.Info("delete pending")
		pendingkey := types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		if err := l.DeletePending(pendingkey, txn); err != nil {
			return err
		}
	}
	return nil
}
func (l *Ledger) updateRepresentative(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	if strings.EqualFold(block.GetToken().String(), chain_token_type) {
		if tm != nil && !tm.RepBlock.IsZero() {
			blk, err := l.GetBlock(tm.RepBlock, txn)
			if err != nil {
				return err
			}
			if state, ok := blk.(*types.StateBlock); ok {
				logger.Infof("sub rep %s from %s ", tm.Balance, state.GetRepresentative())
				if err := l.SubRepresentation(state.GetRepresentative(), tm.Balance, txn); err != nil {
					return err
				}
			} else {
				return errors.New("invalid block")
			}
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
			logger.Info("delete frontier")
			if err := l.DeleteFrontier(frontier.HeaderBlock, txn); err != nil {
				return err
			}
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	logger.Infof("add frontier(%s, %s)", frontier.HeaderBlock, frontier.OpenBlock)
	if err := l.AddFrontier(frontier, txn); err != nil {
		return err
	}
	return nil
}
func (l *Ledger) updateAccountMeta(hash types.Hash, rep_block types.Hash, address types.Address, token types.Hash, balance types.Balance, txn db.StoreTxn) error {
	tm_exist, err := l.HasTokenMeta(address, token, txn)
	if err != nil {
		return err
	}
	if tm_exist {
		token, err := l.GetTokenMeta(address, token, txn)
		if err != nil {
			return err
		}
		token.Header = hash
		token.RepBlock = rep_block
		token.Balance = balance
		token.BlockCount = token.BlockCount + 1
		logger.Info("update tokenmeta, ", token)
		if err := l.UpdateTokenMeta(address, token, txn); err != nil {
			return err
		}
	} else {
		account_exist, err := l.HasAccountMeta(address, txn)
		if err != nil {
			return err
		}
		tm := types.TokenMeta{
			Type:       token,
			Header:     hash,
			RepBlock:   rep_block,
			OpenBlock:  hash,
			Balance:    balance,
			BlockCount: 1,
			BelongTo:   address,
		}
		if account_exist {
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
