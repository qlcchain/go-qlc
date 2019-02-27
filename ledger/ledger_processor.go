package ledger

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
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

	l.logger.Info("process block, ", hash)

	// make sure smart contract token exist
	// ...

	if !block.IsValid() {
		l.logger.Info("invalid work")
		return BadWork, nil
	}

	blockExist, err := l.HasStateBlock(hash, txn)
	if err != nil {
		return Other, err
	}
	if blockExist == true {
		l.logger.Info("block already exist")
		return Old, nil
	}

	// make sure the signature of this block is valid
	signature := block.GetSignature()
	if !address.Verify(hash[:], signature[:]) {
		l.logger.Info("bad signature")
		return BadSignature, nil
	}

	if pre.IsZero() && bytes.EqualFold(address[:], link[:]) {
		l.logger.Info("genesis block")
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
			l.logger.Info("fork: token meta exist, but pre hash is zero")
			return Fork, nil
		}
		preExist, err := l.HasStateBlock(block.GetPrevious(), txn)
		if err != nil {
			return Other, err
		}
		if !preExist {
			l.logger.Info("gap previous: token meta exist, but pre block not exist")
			return GapPrevious, nil
		}
		if block.GetPrevious() != tm.Header {
			l.logger.Info("fork: pre block exist, but pre hash not equal account token's header hash")
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
			l.logger.Info("gap previous: token meta not exist, but pre hash is not zero")
			return GapPrevious, nil
		}
		if link.IsZero() {
			l.logger.Info("gap source: token meta not exist, but link hash is zero")
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
				l.logger.Info("gap source: open or receive block, but link block is not exist")
				return GapSource, nil
			}
			PendingKey := types.PendingKey{
				Address: address,
				Hash:    link,
			}
			pending, err := l.GetPending(PendingKey, txn)
			if err != nil {
				if err == ErrPendingNotFound {
					l.logger.Info("unreceivable: open or receive block, but pending not exist")
					return UnReceivable, nil
				}
				return Other, err
			}
			if !pending.Amount.Equal(transferAmount) {
				l.logger.Infof("balance mismatch: open or receive block, but pending amount(%s) not equal transfer amount(%s)", pending.Amount, transferAmount)
				return BalanceMismatch, nil
			}
		} else { //change
			if !transferAmount.Equal(types.ZeroBalance) {
				l.logger.Info("balance mismatch: change block, but transfer amount is not 0 ")
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
	l.logger.Info("add block, ", hash)
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
		l.logger.Info("add pending, ", pendingkey)
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
			l.logger.Info("delete pending, ", pendingkey)
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
			l.logger.Infof("sub rep %s from %s ", tm.Balance, tm.Representative)
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
		l.logger.Infof("add rep %s to %s ", block.GetBalance(), block.GetRepresentative())
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
			l.logger.Info("delete frontier, ", *frontier)
			if err := l.DeleteFrontier(frontier.HeaderBlock, txn); err != nil {
				return err
			}
		}
		frontier.OpenBlock = tm.OpenBlock
	} else {
		frontier.OpenBlock = hash
	}
	l.logger.Info("add frontier,", *frontier)
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
		l.logger.Info("update tokenmeta, ", *token)
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
			l.logger.Info("add tokenmeta,", token)
			if err := l.AddTokenMeta(address, &tm, txn); err != nil {
				return err
			}
		} else {
			account := types.AccountMeta{
				Address: address,
				Tokens:  []*types.TokenMeta{&tm},
			}
			l.logger.Info("add accountmeta,", token)
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

func (l *Ledger) GenerateSendBlock(source types.Address, token types.Hash, to types.Address, amount types.Balance, prk ed25519.PrivateKey) (types.Block, error) {
	tm, err := l.GetTokenMeta(source, token)
	if err != nil {
		return nil, err
	}
	balance, err := l.TokenBalance(source, token)
	if err != nil {
		return nil, err
	}

	acc := types.NewAccount(prk)
	if balance.Compare(amount) == types.BalanceCompBigger {
		newBalance := balance.Sub(amount)
		sendBlock, err := types.NewBlock(types.State)
		if err != nil {
			return nil, err
		}
		if sb, ok := sendBlock.(*types.StateBlock); ok {
			sb.Address = source
			sb.Token = token
			sb.Link = to.ToHash()
			sb.Balance = newBalance
			sb.Previous = tm.Header
			sb.Representative = tm.Representative
			sb.Signature = acc.Sign(sb.GetHash())
			//sb.Work, _ = s.GetWork(source)
			//if !sb.IsValid() {
			//	sb.Work = s.generateWork(sb.Root())
			//}
			sb.Work = l.generateWork(sb.Root())
		}
		return sendBlock, nil
	} else {
		return nil, fmt.Errorf("not enought balance(%s) of %s", balance, amount)
	}
}

func (l *Ledger) GenerateReceiveBlock(sendBlock types.Block, prk ed25519.PrivateKey) (types.Block, error) {
	hash := sendBlock.GetHash()
	var state *types.StateBlock
	ok := false
	if state, ok = sendBlock.(*types.StateBlock); !ok {
		return nil, fmt.Errorf("invalid state sendBlock(%s)", hash.String())
	}

	// block not exist
	if exist, err := l.HasStateBlock(hash); !exist || err != nil {
		return nil, fmt.Errorf("sendBlock(%s) does not exist", hash.String())
	}
	sendTm, err := l.Token(hash)
	if err != nil {
		return nil, err
	}
	rxAccount := types.Address(state.Link)
	acc := types.NewAccount(prk)
	if err != nil {
		return nil, err
	}
	info, err := l.GetPending(types.PendingKey{Address: rxAccount, Hash: hash})
	if err != nil {
		return nil, err
	}
	receiveBlock, _ := types.NewBlock(types.State)
	has, err := l.HasAccountMeta(rxAccount)
	if err != nil {
		return nil, err
	}
	if has {
		rxAm, err := l.GetAccountMeta(rxAccount)
		if err != nil {
			return nil, err
		}
		rxTm := rxAm.Token(state.Token)
		if sb, ok := receiveBlock.(*types.StateBlock); ok {
			sb.Address = rxAccount
			sb.Balance = rxTm.Balance.Add(info.Amount)
			sb.Previous = rxTm.Header
			sb.Link = hash
			sb.Representative = rxTm.Representative
			sb.Token = rxTm.Type
			sb.Extra = types.Hash{}
			sb.Signature = acc.Sign(sb.GetHash())
			//sb.Work, _ = s.GetWork(rxAccount)
			//if !sb.IsValid() {
			//	sb.Work = s.generateWork(sb.Root())
			//}
			sb.Work = l.generateWork(sb.Root())
		}
	} else {
		if sb, ok := receiveBlock.(*types.StateBlock); ok {
			sb.Address = rxAccount
			sb.Balance = info.Amount
			sb.Previous = types.Hash{}
			sb.Link = hash
			sb.Representative = sendTm.Representative
			sb.Token = sendTm.Type
			sb.Extra = types.Hash{}
			sb.Signature = acc.Sign(sb.GetHash())
			//sb.Work, _ = s.GetWork(rxAccount)
			//if !sb.IsValid() {
			//	sb.Work = s.generateWork(sb.Root())
			//}
			sb.Work = l.generateWork(sb.Root())
		}
	}
	return receiveBlock, nil
}

func (l *Ledger) GenerateChangeBlock(account types.Address, representative types.Address, prk ed25519.PrivateKey) (types.Block, error) {
	if b, err := l.HasAccountMeta(account); err != nil || !b {
		return nil, fmt.Errorf("account[%s] is not exist", account.String())
	}

	if _, err := l.GetAccountMeta(representative); err != nil {
		return nil, fmt.Errorf("invalid representative[%s]", representative.String())
	}

	//get latest chain token block
	hash := l.Latest(account, mock.GetChainTokenType())

	l.logger.Info(hash)
	if hash.IsZero() {
		return nil, fmt.Errorf("account [%s] does not have the main chain account", account.String())
	}

	block, err := l.GetStateBlock(hash)

	if err != nil {
		return nil, err
	}
	changeBlock, err := types.NewBlock(types.State)
	if err != nil {
		return nil, err
	}
	tm, err := l.GetTokenMeta(account, mock.GetChainTokenType())
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

		//sb.Work, _ = s.GetWork(account)
		//if !sb.IsValid() {
		//	sb.Work = s.generateWork(sb.Root())
		//	_ = s.setWork(account, sb.Work)
		//}
		sb.Work = l.generateWork(sb.Root())
	}
	return changeBlock, nil
}
