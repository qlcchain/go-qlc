/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
)

type SqliteService struct {
	common.ServiceLifecycle
	Relation *relation.Relation
	logger   *zap.SugaredLogger
	id       string
}

func NewSqliteService(cfgFile string) (*SqliteService, error) {
	r, err := relation.NewRelation(cfgFile)
	if err != nil {
		return nil, err
	}
	return &SqliteService{Relation: r, logger: log.NewLogger("sqlite_service"), id: cfgFile}, nil
}

func (r *SqliteService) Init() error {
	if !r.PreInit() {
		return errors.New("pre init fail")
	}
	defer r.PostInit()

	re := r.Relation
	count1, err := re.BlocksCount()
	if err != nil {
		return err
	}

	chain := context.NewChainContext(r.id)
	service, err := chain.Service(context.LedgerService)
	if err != nil {
		return nil
	}
	ledgerService := service.(*LedgerService)
	count2, err := ledgerService.Ledger.CountStateBlocks()
	if err != nil {
		return err
	}

	if count1 != count2 {
		if err := re.EmptyStore(); err != nil {
			return err
		}
		err = re.BatchUpdate(func(txn *sqlx.Tx) error {
			blocks := make([]*types.StateBlock, 0)
			err := ledgerService.Ledger.GetStateBlocks(func(block *types.StateBlock) error {
				blocks = append(blocks, block)
				if len(blocks) == 199 {
					if err := re.AddBlocks(txn, blocks); err != nil {
						return err
					}
					blocks = blocks[0:0]
				}
				return nil
			})
			if err != nil {
				return err
			}
			if err := re.AddBlocks(txn, blocks); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SqliteService) Start() error {
	if !r.PreStart() {
		return errors.New("pre start fail")
	}
	defer r.PostStart()
	if err := r.Relation.SetEvent(); err != nil {
		return err
	}
	return nil
}

func (r *SqliteService) Stop() error {
	if !r.PreStop() {
		return errors.New("pre stop fail")
	}
	defer r.PostStop()
	if err := r.Relation.UnsubscribeEvent(); err != nil {
		return err
	}
	if err := r.Relation.Close(); err != nil {
		r.logger.Error(err)
		return err
	}
	return nil
}

func (r *SqliteService) Status() int32 {
	return r.State()
}

func (r *SqliteService) RpcCall(kind uint, in, out interface{}) {

}
