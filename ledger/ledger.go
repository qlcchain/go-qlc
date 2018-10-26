package ledger

import (
	"errors"

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

	// make sure the signature of this block is valid

	return l.db.Update(func(txn db.StoreTxn) error {
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

//func (l *Ledger) processBlock(txn db.StoreTxn, blk ){
//
//}

func (l *Ledger) AddBlock(blk *types.Block) {

}
