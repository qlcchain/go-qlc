package ledger

import (
	"bytes"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"time"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/ledger/db"
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
	if err != nil {
		l.logger.Error(err)
		return Other, err
	}
	if r != Progress {
		return r, nil
	}
	if err := l.BlockProcess(block); err != nil {
		l.logger.Error(err)
		return Other, err
	}
	return Progress, nil
}

func (l *Ledger) BlockCheck(block types.Block) (ProcessResult, error) {
	if b, ok := block.(*types.StateBlock); ok {
		return l.checkStateBlock(b)
	} else if _, ok := block.(*types.SmartContractBlock); ok {
		return Other, errors.New("smart contract block")
	}
	return Other, errors.New("invalid block")
}

func (l *Ledger) checkStateBlock(block *types.StateBlock) (ProcessResult, error) {
	hash := block.GetHash()
	pre := block.GetPrevious()
	link := block.GetLink()
	address := block.GetAddress()

	l.logger.Debug("process block ", hash)

	// make sure smart contract token exist
	// ...

	if !block.IsValid() {
		l.logger.Infof("invalid work (%s)", hash)
		return BadWork, nil
	}

	blockExist, err := l.HasStateBlock(hash)
	if err != nil {
		return Other, err
	}

	if blockExist == true {
		l.logger.Infof("block already exist (%s)", hash)
		return Old, nil
	}

	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		l.logger.Infof("bad signature (%s)", hash)
		return BadSignature, nil
	}

	if pre.IsZero() && bytes.EqualFold(address[:], link[:]) {
		l.logger.Info("genesis block")
		//return Progress, nil
	}

	transferAmount := block.GetBalance()
	tmExist, err := l.HasTokenMeta(address, block.GetToken())
	if err != nil {
		return Other, err
	}
	isSend := false
	if tmExist {
		tm, err := l.GetTokenMeta(address, block.GetToken())
		if err != nil {
			return Other, err
		}
		if pre.IsZero() {
			l.logger.Info("fork: token meta exist, but pre hash is zero (%s)", hash)
			return Fork, nil
		}
		preExist, err := l.HasStateBlock(block.GetPrevious())
		if err != nil {
			return Other, err
		}
		if !preExist {
			l.logger.Infof("gap previous: token meta exist, but pre block not exist (%s)", hash)
			return GapPrevious, nil
		}
		if block.GetPrevious() != tm.Header {
			l.logger.Infof("fork: pre block exist, but pre hash not equal account token's header hash (%s)", hash)
			return Fork, nil
		}
		if block.GetBalance().Compare(tm.Balance) == types.BalanceCompSmaller {
			isSend = true
			transferAmount = tm.Balance.Sub(block.GetBalance()) // send
		} else {
			transferAmount = block.GetBalance().Sub(tm.Balance) // receive or change
		}
	} else {
		if !pre.IsZero() {
			l.logger.Infof("gap previous: token meta not exist, but pre hash is not zero (%s) ", hash)
			return GapPrevious, nil
		}
		if link.IsZero() {
			l.logger.Infof("gap source: token meta not exist, but link hash is zero (%s)", hash)
			return GapSource, nil
		}
	}
	if !isSend {
		if !link.IsZero() { // open or receive
			linkExist, err := l.HasStateBlock(link)
			if err != nil {
				return Other, err
			}
			if !linkExist {
				l.logger.Infof("gap source: open or receive block, but link block is not exist (%s)", hash)
				return GapSource, nil
			}
			PendingKey := types.PendingKey{
				Address: address,
				Hash:    link,
			}
			pending, err := l.GetPending(PendingKey)
			if err != nil {
				if err == ErrPendingNotFound {
					l.logger.Infof("unreceivable: open or receive block, but pending not exist (%s)", hash)
					return UnReceivable, nil
				}
				return Other, err
			}
			if !pending.Amount.Equal(transferAmount) {
				l.logger.Infof("balance mismatch: open or receive block, but pending amount(%s) not equal transfer amount(%s) (%s)", pending.Amount, transferAmount, hash)
				return BalanceMismatch, nil
			}
		} else { //change
			if !transferAmount.Equal(types.ZeroBalance) {
				l.logger.Infof("balance mismatch: change block, but transfer amount is not 0 (%s)", hash)
				return BalanceMismatch, nil
			}
		}
	}
	return Progress, nil
}

func (l *Ledger) BlockProcess(block types.Block) error {
	return l.BatchUpdate(func(txn db.StoreTxn) error {
		if state, ok := block.(*types.StateBlock); ok {
			return l.processStateBlock(state, txn)
		} else if _, ok := block.(*types.SmartContractBlock); ok {
			return errors.New("smart contract block")
		}
		return errors.New("invalid block")
	})
}

func (l *Ledger) processStateBlock(block *types.StateBlock, txn db.StoreTxn) error {
	hash := block.GetHash()
	l.logger.Debug("add block, ", hash)
	if err := l.AddStateBlock(block, txn); err != nil {
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
	if err := l.updateAccountMeta(block, txn); err != nil {
		return err
	}
	if err := l.updateFrontier(hash, tm, txn); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) updatePending(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	hash := block.GetHash()
	link := block.GetLink()
	if block.GetType() == types.Send { // send
		pending := types.PendingInfo{
			Source: block.GetAddress(),
			Type:   block.GetToken(),
			Amount: tm.Balance.Sub(block.GetBalance()),
		}
		pendingkey := types.PendingKey{
			Address: types.Address(block.GetLink()),
			Hash:    hash,
		}
		l.logger.Debug("add pending, ", pendingkey)
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
			l.logger.Debug("delete pending, ", pendingkey)
			if err := l.DeletePending(pendingkey, txn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Ledger) updateRepresentative(block *types.StateBlock, tm *types.TokenMeta, txn db.StoreTxn) error {
	if block.GetToken() == common.QLCChainToken {
		if tm != nil && !tm.Representative.IsZero() {
			l.logger.Debugf("sub rep %s from %s ", tm.Balance, tm.Representative)
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
		l.logger.Debugf("add rep %s to %s ", block.GetBalance(), block.GetRepresentative())
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
			l.logger.Debug("delete frontier, ", *frontier)
			if err := l.DeleteFrontier(frontier.HeaderBlock, txn); err != nil {
				return err
			}
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	l.logger.Debug("add frontier,", *frontier)
	if err := l.AddFrontier(frontier, txn); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) updateAccountMeta(block *types.StateBlock, txn db.StoreTxn) error {
	hash := block.GetHash()
	rep := block.GetRepresentative()
	address := block.GetAddress()
	token := block.GetToken()
	balance := block.GetBalance()
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
		l.logger.Debug("update tokenmeta, ", *token)
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
			l.logger.Debug("add tokenmeta,", token)
			if err := l.AddTokenMeta(address, &tm, txn); err != nil {
				return err
			}
		} else {
			account := types.AccountMeta{
				Address: address,
				Tokens:  []*types.TokenMeta{&tm},
			}
			l.logger.Debug("add accountmeta,", token)
			if err := l.AddAccountMeta(&account, txn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Ledger) generateWork(hash types.Hash) types.Work {
	var work types.Work
	worker, _ := types.NewWorker(work, hash)
	return worker.NewWork()
	//
	////cache to db
	//_ = s.setWork(hash, work)
}

func (l *Ledger) GenerateSendBlock(source types.Address, token types.Hash, to types.Address, amount types.Balance, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	tm, err := l.GetTokenMeta(source, token)
	if err != nil {
		return nil, err
	}
	//balance, err := l.TokenBalance(source, token)
	//if err != nil {
	//	return nil, err
	//}

	if tm.Balance.Compare(amount) != types.BalanceCompSmaller {
		sb := types.StateBlock{
			Type:           types.Send,
			Address:        source,
			Token:          token,
			Link:           to.ToHash(),
			Balance:        tm.Balance.Sub(amount),
			Previous:       tm.Header,
			Representative: tm.Representative,
		}
		acc := types.NewAccount(prk)
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
		return &sb, nil
	} else {
		return nil, fmt.Errorf("not enought balance(%s) of %s", tm.Balance, amount)
	}
}

func (l *Ledger) GenerateReceiveBlock(sendBlock *types.StateBlock, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	hash := sendBlock.GetHash()
	if !sendBlock.GetType().Equal(types.Send) {
		return nil, fmt.Errorf("(%s) is not send block", hash.String())
	}
	if exist, err := l.HasStateBlock(hash); !exist || err != nil {
		return nil, fmt.Errorf("send block(%s) does not exist", hash.String())
	}
	acc := types.NewAccount(prk)
	rxAccount := types.Address(sendBlock.Link)
	info, err := l.GetPending(types.PendingKey{Address: rxAccount, Hash: hash})
	if err != nil {
		return nil, err
	}
	has, err := l.HasAccountMeta(rxAccount)
	if err != nil {
		return nil, err
	}
	if has {
		rxAm, err := l.GetAccountMeta(rxAccount)
		if err != nil {
			return nil, err
		}
		rxTm := rxAm.Token(sendBlock.GetToken())
		sb := types.StateBlock{
			Type:           types.Receive,
			Address:        rxAccount,
			Balance:        rxTm.Balance.Add(info.Amount),
			Previous:       rxTm.Header,
			Link:           hash,
			Representative: rxTm.Representative,
			Token:          rxTm.Type,
			Extra:          types.ZeroHash,
		}
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
		return &sb, nil
	} else {
		//genesis, err := mock.GetTokenById(mock.GetChainTokenType())
		//if err != nil {
		//	return nil, err
		//}
		sb := &types.StateBlock{
			Type:           types.Open,
			Address:        rxAccount,
			Balance:        info.Amount,
			Previous:       types.ZeroHash,
			Link:           hash,
			Representative: sendBlock.GetRepresentative(), //Representative: genesis.Owner,
			Token:          sendBlock.GetToken(),
			Extra:          types.ZeroHash,
		}
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
		return sb, nil
	}
}

func (l *Ledger) GenerateChangeBlock(account types.Address, representative types.Address, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	if b, err := l.HasAccountMeta(account); err != nil || !b {
		return nil, fmt.Errorf("account[%s] is not exist", account.String())
	}

	if _, err := l.GetAccountMeta(representative); err != nil {
		return nil, fmt.Errorf("invalid representative[%s]", representative.String())
	}

	//get latest chain token block
	hash := l.Latest(account, common.QLCChainToken)

	l.logger.Info(hash)
	if hash.IsZero() {
		return nil, fmt.Errorf("account [%s] does not have the main chain account", account.String())
	}

	block, err := l.GetStateBlock(hash)
	if err != nil {
		return nil, err
	}
	tm, err := l.GetTokenMeta(account, common.QLCChainToken)
	acc := types.NewAccount(prk)
	if sb, ok := changeBlock.(*types.StateBlock); ok {
		sb.Address = account
		sb.Balance = tm.Balance
		sb.Previous = tm.Header
		sb.Link = account.ToHash()
		sb.Representative = representative
		sb.Token = block.Token
		sb.Extra = types.Hash{}
		sb.Signature = acc.Sign(sb.GetHash())

		sb.Work = l.generateWork(sb.Root())
	}
	acc := types.NewAccount(prk)
	sb.Signature = acc.Sign(sb.GetHash())
	sb.Work = l.generateWork(sb.Root())
	return &sb, nil
}
