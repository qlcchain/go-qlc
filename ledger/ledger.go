package ledger

import (
	"errors"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/ledger/genesis"
)

var (
	ErrBadWork          = errors.New("bad work")
	ErrBlockSignature   = errors.New("bad block signature")
	ErrBadGenesis       = errors.New("genesis block in store doesn't match the given block")
	ErrMissingPrevious  = errors.New("previous block does not exist")
	ErrMissingLink      = errors.New("link block does not exist")
	ErrUnchecked        = errors.New("block was added to the unchecked list")
	ErrFork             = errors.New("a fork was detected")
	ErrNotFound         = errors.New("item not found in the store")
	ErrZeroSpend        = errors.New("zero spend not allowed")
	ErrTokenBalance     = errors.New("err token balance for receive block")
	ErrChangeNotAllowed = errors.New("balance change not allowed in change representative block")
	ErrTokenType        = errors.New("err token type for change representation")
)

var (
	chain_token_type = "125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36"
)

type Ledger struct {
	opts LedgerOptions
	db   db.Store
}

type LedgerOptions struct {
	Genesis genesis.Genesis
}

var log = common.NewLogger("ledger")

func NewLedger(store db.Store, opts LedgerOptions) (*Ledger, error) {
	ledger := Ledger{opts: opts, db: store}

	// initialize the store with the genesis block if needed
	if err := ledger.setGenesis(&opts.Genesis.Block); err != nil {
		return nil, err
	}

	log.Info("new ledger created")
	return &ledger, nil
}

func (l *Ledger) setGenesis(blk *types.StateBlock) error {
	hash := blk.Hash()

	// make sure the work value is valid
	if !blk.Valid(l.opts.Genesis.WorkThreshold) {
		return ErrBadWork
	}

	// make sure the signature of this block is valid
	if blk.Address.Verify(hash[:], blk.Signature[:]) {
		return ErrBlockSignature
	}

	return l.db.Update(func(txn db.StoreTxn) error {
		if strings.EqualFold(blk.Token.String(), chain_token_type) {

			empty, err := txn.Empty()
			if err != nil {
				return err
			}

			if !empty {
				// if the database is not empty, check if it has the same genesis
				// block as the one in the given options
				found, err := txn.HasBlock(hash)
				if err != nil {
					return err
				}
				if !found {
					return ErrBadGenesis
				}
			} else {
				log.Info("add qlc chain genesis block, ", blk.Hash())
				if err := txn.AddBlock(blk); err != nil {
					return err
				}

				if err := l.updateAccountMeta(txn, blk, blk.Balance); err != nil {
					return err
				}

				if err := txn.AddRepresentation(blk.Representative, blk.Balance); err != nil {
					return err
				}

				return txn.AddFrontier(&types.Frontier{
					Address: blk.Address,
					Hash:    hash,
				})
			}
		} else {
			found, err := txn.HasBlock(hash)
			if err != nil {
				return err
			}
			if !found {
				log.Info("add genesis block, ", blk.Hash())
				if err := txn.AddBlock(blk); err != nil {
					return err
				}

				if err := l.updateAccountMeta(txn, blk, blk.Balance); err != nil {
					return err
				}

				return txn.AddFrontier(&types.Frontier{
					Address: blk.Address,
					Hash:    hash,
				})
			}
		}
		return nil
	})
}

func (l *Ledger) addOpenBlock(txn db.StoreTxn, blk *types.StateBlock) error {
	log.Info("add open block ...")
	hash := blk.Hash()
	// account doesn't exist
	// obtain the pending transaction info
	pending, err := txn.GetPending(blk.Address, blk.Link)
	if err != nil {
		return ErrMissingLink
	}

	if !blk.Balance.Equal(pending.Amount) {
		log.Info("pending amount/block balance, ", pending.Amount, blk.Balance)
		return ErrTokenBalance
	}
	if err = l.updateAccountMeta(txn, blk, pending.Amount); err != nil {
		return err
	}

	// delete the pending transaction
	if err := txn.DeletePending(blk.Address, blk.Link); err != nil {
		return err
	}

	// update representative voting weight
	if strings.Compare(blk.Token.String(), chain_token_type) == 0 {
		if err := txn.AddRepresentation(blk.Representative, pending.Amount); err != nil {
			return err
		}
	}

	// add a frontier for this address
	frontier := types.Frontier{
		Address: blk.Address,
		Hash:    hash,
	}
	if err := txn.AddFrontier(&frontier); err != nil {
		return err
	}

	// finally, add the block
	return txn.AddBlock(blk)
}

func (l *Ledger) addSendBlock(txn db.StoreTxn, blk *types.StateBlock) error {
	log.Info("add send block ...")

	hash := blk.Hash()

	// add pending
	token, _ := txn.GetTokenMeta(blk.Address, blk.Token)
	pending := types.PendingInfo{
		Source: blk.Address,
		Type:   blk.Token,
		Amount: token.Balance.Sub(blk.Balance),
	}
	if err := txn.AddPending(types.Address(blk.Link), hash, &pending); err != nil {
		return err
	}

	// update representative voting weight
	if strings.Compare(blk.Token.String(), chain_token_type) == 0 {
		rep, err := l.getRepresentativeAddress(txn, blk.Address, blk.Token)
		if err != nil {
			return err
		}
		if err := txn.SubRepresentation(rep, pending.Amount); err != nil {
			return err
		}
	}

	// update the token
	token.Balance = token.Balance.Sub(pending.Amount)
	token.Header = hash
	if err := txn.UpdateTokenMeta(blk.Address, token); err != nil {
		return err
	}

	// update the frontier of this account
	if err := l.updateFrontier(txn, blk); err != nil {
		return err
	}

	return txn.AddBlock(blk)
}

func (l *Ledger) addReceiveBlock(txn db.StoreTxn, blk *types.StateBlock) error {
	log.Info("add receive block ...")

	hash := blk.Hash()

	// get pending
	pending, err := txn.GetPending(blk.Address, blk.Link)
	if err != nil {
		return ErrMissingLink
	}
	if err := txn.DeletePending(blk.Address, blk.Link); err != nil {
		return err
	}

	// update representative voting weight
	if strings.Compare(blk.Token.String(), chain_token_type) == 0 {
		rep, err := l.getRepresentativeAddress(txn, blk.Address, blk.Token)
		if err != nil {
			return err
		}
		if err := txn.AddRepresentation(rep, pending.Amount); err != nil {
			return err
		}
	}

	// update the token
	token, _ := txn.GetTokenMeta(blk.Address, blk.Token)
	token.Balance = token.Balance.Add(pending.Amount)
	token.Header = hash
	if !token.Balance.Equal(blk.Balance) {
		log.Info("token balance/block balance ", token.Balance, blk.Balance)
		return ErrTokenBalance
	}
	if err := txn.UpdateTokenMeta(blk.Address, token); err != nil {
		return err
	}

	// update the frontier of this account
	if err := l.updateFrontier(txn, blk); err != nil {
		return err
	}

	return txn.AddBlock(blk)
}

func (l *Ledger) addChangeBlock(txn db.StoreTxn, blk *types.StateBlock) error {
	log.Info("add change block ...")

	hash := blk.Hash()

	if !strings.EqualFold(blk.Token.String(), chain_token_type) {
		log.Info("err token type for change representation ")
		return ErrTokenType
	}

	token, _ := txn.GetTokenMeta(blk.Address, blk.Token)
	if !blk.Balance.Equal(token.Balance) {
		log.Infof("block balance:%s, token balance:%s", blk.Balance, token.Balance)
		return ErrChangeNotAllowed
	}

	// update representative voting weight
	rep, err := l.getRepresentativeAddress(txn, blk.Address, blk.Token)
	if err != nil {
		return err
	}
	if err := txn.SubRepresentation(rep, token.Balance); err != nil {
		return err
	}
	if err := txn.AddRepresentation(blk.Representative, token.Balance); err != nil {
		return err
	}

	// update the token info
	token.Header = hash
	token.RepBlock = hash
	if err := txn.UpdateTokenMeta(blk.Address, token); err != nil {
		return err
	}

	// update the frontier of this account
	if err := l.updateFrontier(txn, blk); err != nil {
		return err
	}

	return txn.AddBlock(blk)
}

func (l *Ledger) addStateBlock(txn db.StoreTxn, blk *types.StateBlock) error {
	hash := blk.Hash()

	// make sure the signature of this block is valid
	if blk.Address.Verify(hash[:], blk.Signature[:]) {
		return ErrBlockSignature
	}

	// obtain account information if possible
	token, err := txn.GetTokenMeta(blk.Address, blk.Token)
	if err != nil {
		if err == db.ErrTokenNotFound || err == badger.ErrKeyNotFound {
			// todo: check for key not found error
			if !blk.IsOpen() {
				return l.addOpenBlock(txn, blk)
			}
			return err
		} else {
			return err
		}
	} else {
		// make sure the hash of the previous block is a frontier
		_, err := txn.GetFrontier(blk.PreviousHash)
		if err != nil {
			return ErrFork
		}

		if blk.Link.IsZero() {
			return l.addChangeBlock(txn, blk)
		} else {
			previous_balance := token.Balance
			log.Info("previous balance/current balance, ", previous_balance, blk.Balance)

			switch previous_balance.Compare(blk.Balance) {
			case types.BalanceCompBigger:
				return l.addSendBlock(txn, blk)
			case types.BalanceCompSmaller:
				return l.addReceiveBlock(txn, blk)
			case types.BalanceCompEqual:
				return ErrZeroSpend
			}
		}
	}
	return nil
}

func (l *Ledger) addBlock(txn db.StoreTxn, blk types.Block) error {
	hash := blk.GetHash()

	// make sure the work value is valid
	if !blk.Valid(l.opts.Genesis.WorkThreshold) {
		return ErrBadWork
	}

	// make sure the hash of this block doesn't exist yet
	found, err := txn.HasBlock(hash)
	if err != nil {
		return err
	}
	if found {
		return db.ErrBlockExists
	}

	// make sure the previous/link block exists
	found, err = txn.HasBlock(blk.Root())
	if err != nil {
		return err
	}

	if !found {
		switch b := blk.(type) {
		case *types.StateBlock:
			if b.IsOpen() {
				return ErrMissingPrevious
			}
			return ErrMissingLink
		default:
			return types.ErrBadBlockType
		}
	}
	switch b := blk.(type) {
	case *types.StateBlock:
		err = l.addStateBlock(txn, b)
	case *types.SmartContractBlock:
		// do...
	default:
		return types.ErrBadBlockType

	}

	if err != nil {
		return err
	}

	// flush if needed
	return txn.Flush()
}

func (l *Ledger) processBlock(txn db.StoreTxn, blk types.Block) error {
	log.Info("processing block, ", blk.GetHash())
	err := l.addBlock(txn, blk)
	switch err {
	case ErrMissingPrevious:
		if err := l.addUncheckedBlock(txn, blk.Root(), blk, types.UncheckedKindPrevious); err != nil {
			return err
		}
		log.Info("missing previous, added uncheckedblock")
		return ErrUnchecked
	case ErrMissingLink:
		var source types.Hash
		switch b := blk.(type) {
		case *types.StateBlock:
			source = b.Link
		default:
			return errors.New("unexpected block type")
		}

		// add to unchecked list
		if err := l.addUncheckedBlock(txn, source, blk, types.UncheckedKindLink); err != nil {
			return err
		}
		log.Info("missing link, added uncheckedblock")
		return ErrUnchecked
	case nil:
		log.Info("added block, ", blk.GetHash())
		// try to process any unchecked child blocks
		if err := l.processUncheckedBlock(txn, blk, types.UncheckedKindPrevious); err != nil {
			return err
		}
		if err := l.processUncheckedBlock(txn, blk, types.UncheckedKindLink); err != nil {
			return err
		}
		return nil
	case db.ErrBlockExists:
		log.Info("block already exists, ", blk.GetHash())
		return nil
	default:
		log.Info(err)
		return err
	}
}

func (l *Ledger) getRepresentativeAddress(txn db.StoreTxn, address types.Address, token types.Hash) (types.Address, error) {
	info, err := txn.GetTokenMeta(address, token)
	if err != nil {
		return types.Address{}, err
	}

	blk, err := txn.GetBlock(info.RepBlock)
	if err != nil {
		return types.Address{}, err
	}

	switch b := blk.(type) {
	case *types.StateBlock:
		return b.Representative, nil
	default:
		return types.Address{}, types.ErrBadBlockType
	}
}

func (l *Ledger) addUncheckedBlock(txn db.StoreTxn, parentHash types.Hash, blk types.Block, kind types.UncheckedKind) error {
	found, err := txn.HasUncheckedBlock(parentHash, kind)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	return txn.AddUncheckedBlock(parentHash, blk, kind)
}

func (l *Ledger) processUncheckedBlock(txn db.StoreTxn, blk types.Block, kind types.UncheckedKind) error {
	hash := blk.GetHash()

	found, err := txn.HasUncheckedBlock(hash, kind)
	if err != nil {
		return err
	}
	if found {
		log.Info("unchecked block found,", kind)
		uncheckedBlk, err := txn.GetUncheckedBlock(hash, kind)
		if err != nil {
			return err
		}

		if err := l.processBlock(txn, uncheckedBlk); err == nil {
			// delete from the unchecked list if successful
			if err := txn.DeleteUncheckedBlock(hash, kind); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Ledger) updateFrontier(txn db.StoreTxn, blk *types.StateBlock) error {
	frontier, _ := txn.GetFrontier(blk.PreviousHash)
	if blk.IsOpen() {
		if err := txn.DeleteFrontier(frontier.Hash); err != nil {
			return err
		}
	}
	frontier = &types.Frontier{
		Address: blk.Address,
		Hash:    blk.Hash(),
	}
	if err := txn.AddFrontier(frontier); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) updateAccountMeta(txn db.StoreTxn, blk *types.StateBlock, balance types.Balance) error {
	if account, err := txn.GetAccountMeta(blk.Address); err == badger.ErrKeyNotFound {
		accountmeta := types.AccountMeta{
			Address: blk.Address,
			Tokens: []*types.TokenMeta{
				&types.TokenMeta{
					Type:      blk.Token,
					Header:    blk.Hash(),
					RepBlock:  blk.Hash(),
					OpenBlock: blk.Hash(),
					Balance:   balance,
				},
			},
		}
		if err := txn.AddAccountMeta(&accountmeta); err != nil {
			return err
		}
		return nil
	} else if err == nil {
		// add token info
		token := types.TokenMeta{
			Type:      blk.Token,
			Header:    blk.Hash(),
			RepBlock:  blk.Hash(),
			OpenBlock: blk.Hash(),
			Balance:   balance,
		}
		account.Tokens = append(account.Tokens, &token)
		if err := txn.UpdateAccountMeta(account); err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

func (l *Ledger) AddBlock(blk types.Block) error {
	return l.db.Update(func(txn db.StoreTxn) error {
		err := l.processBlock(txn, blk)
		if err != nil && err != ErrUnchecked {
			return err
		}
		return nil
	})
}
