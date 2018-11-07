package ledger

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/ledger/genesis"
)

var (
	ErrBadWork         = errors.New("bad work")
	ErrBadGenesis      = errors.New("genesis block in store doesn't match the given block")
	ErrMissingPrevious = errors.New("previous block does not exist")
	ErrMissingSource   = errors.New("source block does not exist")
	ErrUnchecked       = errors.New("block was added to the unchecked list")
	ErrFork            = errors.New("a fork was detected")
	ErrNotFound        = errors.New("item not found in the store")
)

type Ledger struct {
	opts LedgerOptions
	db   db.Store
}

type LedgerOptions struct {
	Genesis genesis.Genesis
}

func NewLedger(store db.Store, opts LedgerOptions) (*Ledger, error) {
	ledger := Ledger{opts: opts, db: store}

	// initialize the store with the genesis block if needed
	if err := ledger.setGenesis(&opts.Genesis.Block); err != nil {
		return nil, err
	}

	return &ledger, nil
}

func (l *Ledger) setGenesis(blk *types.StateBlock) error {
	hash := blk.Hash()

	// make sure the work value is valid
	if !blk.Valid(l.opts.Genesis.WorkThreshold) {
		return errors.New("bad work for genesis block")
	}

	// make sure the signature of this block is valid
	if !blk.Address.Verify(hash[:], blk.Signature[:]) {
		//return errors.New("bad signature for genesis block")
	}

	return l.db.Update(func(txn db.StoreTxn) error {
		empty, err := txn.Empty()
		if err != nil {
			return err
		}

		if !empty {
			// if the database is not empty, check if it has the same genesis  block as the one in the given options
			found, err := txn.HasBlock(hash)
			if err != nil {
				return err
			}
			if !found {
				return ErrBadGenesis
			}
		} else {
			if err := txn.AddBlock(blk); err != nil {
				return err
			}

			accountmeta := types.AccountMeta{
				Address: blk.Address,
				Tokens: []*types.TokenMeta{
					&types.TokenMeta{
						Type:      blk.Token,
						Header:    blk.Hash(),
						RepBlock:  blk.Hash(),
						OpenBlock: blk.Hash(),
						Balance:   blk.Balance,
					},
				},
			}

			if err := txn.AddAccountMeta(&accountmeta); err != nil {
				return err
			}

		}

		return nil
	})
}

func (l *Ledger) addOpenBlock(txn db.StoreTxn, blk *types.StateBlock) error {
	// account doesn't exist
	// obtain the pending transaction info
	pending, err := txn.GetPending(blk.Address, blk.Link)
	if err != nil {
		return ErrMissingSource
	}

	// add address info
	accountmeta := types.AccountMeta{
		Address: blk.Address,
		Tokens: []*types.TokenMeta{
			&types.TokenMeta{
				Type:      blk.Token,
				Header:    blk.Hash(),
				RepBlock:  blk.Hash(),
				OpenBlock: blk.Hash(),
				Balance:   blk.Balance,
			},
		},
	}
	if err := txn.AddAccountMeta(&accountmeta); err != nil {
		return err
	}

	// delete the pending transaction
	if err := txn.DeletePending(blk.Address, blk.Link); err != nil {
		return err
	}

	// update representative voting weight
	if err := txn.AddRepresentation(blk.Representative, pending.Amount); err != nil {
		return err
	}

	// add a frontier for this address
	// do...

	// finally, add the block
	return txn.AddBlock(blk)
}

func (l *Ledger) addStateBlock(txn db.StoreTxn, blk *types.StateBlock) error {
	hash := blk.Hash()

	// make sure the signature of this block is valid
	if !blk.Address.Verify(hash[:], blk.Signature[:]) {
		return errors.New("bad block signature")
	}

	// obtain account information if possible
	_, err := txn.GetAccountMeta(blk.Address)
	if err == badger.ErrKeyNotFound {
		// todo: check for key not found error
		if !blk.IsOpen() {
			l.addOpenBlock(txn, blk)
		}
		return err
	}
	return nil
}

func (l *Ledger) addBlock(txn db.StoreTxn, blk types.Block) error {
	hash := blk.Hash()

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
		// do...
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
	err := l.addBlock(txn, blk)
	return err
}

func (l *Ledger) AddBlock(blk types.Block) error {
	return l.db.Update(func(txn db.StoreTxn) error {
		err := l.processBlock(txn, blk)
		if err != nil && err != ErrUnchecked {
			fmt.Printf("try add err: %s\n", err)
		}
		return nil
	})
}

func (l *Ledger) CountBlocks() (uint64, error) {
	panic("implement me")
}

func (l *Ledger) CountUncheckedBlocks() (uint64, error) {
	panic("implement me")
}

func (l *Ledger) GetBalance(address types.Address, token types.Hash) (types.Balance, error) {
	panic("implement me")
}
