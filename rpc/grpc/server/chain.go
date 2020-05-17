package grpcServer

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/grpc/proto"
	"go.uber.org/zap"
)

type ChainApi struct {
	ledger ledger.Store
	logger *zap.SugaredLogger
}

func NewChainApi(l ledger.Store) *ChainApi {
	return &ChainApi{ledger: l, logger: log.NewLogger("api_chain")}
}

func (ChainApi) LedgerSize(context.Context, *empty.Empty) (*proto.LedgerSizeResponse, error) {
	panic("implement me")
}

func (ChainApi) Version(context.Context, *empty.Empty) (*proto.VersionResponse, error) {
	panic("implement me")
}
