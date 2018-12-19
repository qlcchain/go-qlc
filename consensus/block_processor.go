package consensus

import (
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
)

var chain_token_type = "125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36"

type process_result byte

const (
	progress process_result = iota
	bad_signature
	old
	fork
	gap_previous
	gap_source
	balance_mismatch
	unreceivable
	other
)

type BlockProcessor struct {
	blocks chan types.Block
	quitCh chan bool
	dp     *Dpos
}

func NewBlockProcessor() *BlockProcessor {
	return &BlockProcessor{
		blocks: make(chan types.Block, 16384),
		quitCh: make(chan bool, 1),
	}
}
func (bp *BlockProcessor) SetDpos(dp *Dpos) {
	bp.dp = dp
}
func (bp *BlockProcessor) Start() {
	go bp.process_blocks()
}
func (bp *BlockProcessor) process_blocks() {
	for {
		select {
		case <-bp.quitCh:
			logger.Info("Stopped process_blocks.")
			return
		case block := <-bp.blocks:
			bp.process_receive_one(block)
		}
	}
}
func (bp *BlockProcessor) process_receive_one(block types.Block) process_result {
	if block.GetType() == types.State {
		state := interface{}(block).(*types.StateBlock)
		return bp.checkStateBasicInfo(*state)
	} else if block.GetType() == types.SmartContract {
		return other
	} else {
		return other
	}
}

func (bp *BlockProcessor) checkStateBasicInfo(block types.StateBlock) process_result {
	hash := block.GetHash()
	pre := block.GetPrevious()
	link := block.GetLink()
	address := block.GetAddress()
	l := bp.dp.ledger
	logger.Info("---- process block, ", hash)

	// make sure smart contract token exist
	// ...

	// make sure the block does not exist
	block_exist, err := l.HasBlock(hash)
	if err != nil {
		return other
	}
	if block_exist == true {
		logger.Info("block already exist")
		return old
	}

	// make sure the signature of this block is valid
	signature := block.GetSignature()
	if address.Verify(hash[:], signature[:]) {
		logger.Info("bad signature")
		return bad_signature
	}

	is_send := false
	transfer_amount := block.GetBalance()

	tm_exist, err := l.HasTokenMeta(address, block.GetToken())
	if err != nil && err != ledger.ErrAccountNotFound {
		return other
	}
	if tm_exist { // this tm chain exist , not open
		logger.Info("tokenmeta exist")
		tm, err := l.GetTokenMeta(address, block.GetToken())
		if err != nil {
			return other
		}

		if pre.IsZero() {
			logger.Info("fork: tokenmeta exist, but pre is zero")
			return fork
		}
		previous_exist, err := l.HasBlock(block.GetPrevious())
		if err != nil {
			return other
		}
		if !previous_exist {
			logger.Info("gap previous: tokenmeta exist, but pre block not exist")
			return gap_previous
		}
		if block.GetPrevious() != tm.Header {
			logger.Info("fork: pre block exist, but  prehash not equal account headerhash")
			return fork
		}
		if block.GetBalance().Compare(tm.Balance) == types.BalanceCompSmaller {
			is_send = true
			transfer_amount = tm.Balance.Sub(block.GetBalance()) // send
		} else {
			transfer_amount = block.GetBalance().Sub(tm.Balance) // receive or change
		}

	} else { // open
		logger.Info("tokenmeta not exist")
		if !pre.IsZero() {
			logger.Info("fork: tokenmeta not exist, but pre not zero")
			return fork
		}
		if link.IsZero() {
			logger.Info("gap source: tokenmeta not exist, but link is zero ")
			return gap_source
		}
	}
	if !is_send {
		if !link.IsZero() { // open or receive
			link_exist, err := l.HasBlock(link)
			if err != nil {
				return other
			}
			if !link_exist {
				logger.Info("gap source: open or receive block, but link block is not exist")
				return gap_source
			}
			PendingKey := types.PendingKey{
				Address: address,
				Hash:    link,
			}
			pending, err := l.GetPending(PendingKey)
			if err != nil {
				if err == ledger.ErrPendingNotFound {
					logger.Info("unreceivable: open or receive block, but pending not exist")
					return unreceivable
				}
				return other
			}
			if !pending.Amount.Equal(transfer_amount) {
				logger.Infof("balance mismatch: open or receive block, but pending amount(%s) not equal transfer amount(%s)", pending.Amount, transfer_amount)
				return balance_mismatch
			}
		} else { //change
			if !transfer_amount.Equal(types.ParseBalanceInts(0, 0)) {
				logger.Info("balance mismatch: change block, but transfer amount is not 0 ")
				return balance_mismatch
			}
		}
	}
	if err := bp.addBasicInfo(block, is_send); err != nil {
		return other
	}
	return progress
}
func (bp *BlockProcessor) addBasicInfo(block types.StateBlock, is_send bool) error {
	hash := block.GetHash()
	logger.Info("add block, ", hash)
	l := bp.dp.ledger
	if err := l.AddBlock(&block); err != nil {
		return err
	}

	tm, err := l.GetTokenMeta(block.GetAddress(), block.GetToken())
	if err != nil && err != ledger.ErrTokenNotFound && err != ledger.ErrAccountNotFound {
		return err
	}

	if err := bp.updateRepresentative(block, tm, l); err != nil {
		return err
	}

	if err := bp.updatePending(block, tm, is_send, l); err != nil {
		return err
	}

	if err := bp.updateAccountMeta(hash, hash, block.GetAddress(), block.GetToken(), block.GetBalance(), l); err != nil {
		return err
	}

	if err := bp.updateFrontier(hash, tm, l); err != nil {
		return err
	}

	return nil
}
func (bp *BlockProcessor) updatePending(block types.StateBlock, tm *types.TokenMeta, is_send bool, session *ledger.Ledger) error {
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
		if err := session.AddPending(pendingkey, &pending); err != nil {
			return err
		}
	} else if !link.IsZero() { // receive or open
		logger.Info("delete pending")
		pendingkey := types.PendingKey{
			Address: block.GetAddress(),
			Hash:    block.GetLink(),
		}
		if err := session.DeletePending(pendingkey); err != nil {
			return err
		}
	}
	return nil
}
func (bp *BlockProcessor) updateRepresentative(block types.StateBlock, tm *types.TokenMeta, session *ledger.Ledger) error {
	if strings.EqualFold(block.GetToken().String(), chain_token_type) {
		if tm != nil && !tm.RepBlock.IsZero() {
			blk, err := session.GetBlock(tm.RepBlock)
			if err != nil {
				return err
			}
			state := interface{}(blk).(*types.StateBlock)
			logger.Infof("sub rep %s from %s ", tm.Balance, state.GetRepresentative())
			session.SubRepresentation(state.GetRepresentative(), tm.Balance)
		}
		logger.Infof("add rep %s to %s ", block.GetBalance(), block.GetRepresentative())
		if err := session.AddRepresentation(block.GetRepresentative(), block.GetBalance()); err != nil {
			return err
		}
	}
	return nil
}
func (bp *BlockProcessor) updateFrontier(hash types.Hash, tm *types.TokenMeta, session *ledger.Ledger) error {
	frontier := &types.Frontier{
		HeaderBlock: hash,
	}
	if tm != nil {
		if frontier, err := session.GetFrontier(tm.Header); err == nil {
			logger.Info("delete frontier")
			if err := session.DeleteFrontier(frontier.HeaderBlock); err != nil {
				return err
			}
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	logger.Infof("add frontier(%s, %s)", frontier.HeaderBlock, frontier.OpenBlock)
	if err := session.AddFrontier(frontier); err != nil {
		return err
	}
	return nil
}
func (bp *BlockProcessor) updateAccountMeta(hash types.Hash, rep_block types.Hash, address types.Address, token types.Hash, balance types.Balance, session *ledger.Ledger) error {
	tm_exist, err := session.HasTokenMeta(address, token)
	if err != nil && err != ledger.ErrAccountNotFound {
		return err
	}
	if tm_exist {
		token, err := session.GetTokenMeta(address, token)
		if err != nil {
			return err
		}
		token.Header = hash
		token.RepBlock = rep_block
		token.Balance = balance
		token.BlockCount = token.BlockCount + 1
		logger.Info("update tokenmeta")
		session.UpdateTokenMeta(address, token)

	} else {
		account_exist, err := session.HasAccountMeta(address)
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
		}
		if account_exist {
			logger.Info("add tokenmeta")
			if err := session.AddTokenMeta(address, &tm); err != nil {
				return err
			}
		} else {
			account := types.AccountMeta{
				Address: address,
				Tokens:  []*types.TokenMeta{&tm},
			}
			logger.Info("add accountmeta")
			if err := session.AddAccountMeta(&account); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bp *BlockProcessor) Stop() {
	bp.quitCh <- true
}
