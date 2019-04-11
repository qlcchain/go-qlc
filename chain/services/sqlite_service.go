package services

import (
	"errors"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type SqliteService struct {
	common.ServiceLifecycle
	relation *relation.Relation
	logger   *zap.SugaredLogger
}

func NewSqliteService(cfg *config.Config, eb event.EventBus) (*SqliteService, error) {
	r, err := relation.NewRelation(cfg, eb)
	if err != nil {
		return nil, err
	}
	return &SqliteService{relation: r, logger: log.NewLogger("sqlite_service")}, nil
}

func (r *SqliteService) Init() error {
	if !r.PreInit() {
		return errors.New("pre init fail.")
	}
	defer r.PostInit()

	return nil
}

func (r *SqliteService) Start() error {
	if !r.PreStart() {
		return errors.New("pre start fail.")
	}
	defer r.PostStart()
	if err := r.relation.SetEvent(); err != nil {
		return err
	}
	return nil
}

func (r *SqliteService) Stop() error {
	if !r.PreStop() {
		return errors.New("pre stop fail")
	}
	defer r.PostStop()
	if err := r.relation.UnsubscribeEvent(); err != nil {
		return err
	}
	if err := r.relation.Close(); err != nil {
		r.logger.Error(err)
		return err
	}
	return nil
}

func (r *SqliteService) Status() int32 {
	return r.State()
}
