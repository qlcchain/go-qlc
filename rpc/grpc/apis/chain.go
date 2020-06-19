package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

type ChainAPI struct {
	chain  *api.ChainApi
	logger *zap.SugaredLogger
}

func NewChainAPI(l ledger.Store) *ChainAPI {
	return &ChainAPI{
		chain:  api.NewChainApi(l),
		logger: log.NewLogger("grpc_chain"),
	}
}

func (c *ChainAPI) LedgerSize(context.Context, *empty.Empty) (*pb.LedgerSizeResponse, error) {
	r, err := c.chain.LedgerSize()
	if err != nil {
		return nil, err
	}
	return &pb.LedgerSizeResponse{
		Size: r,
	}, nil
}

func (c *ChainAPI) Version(context.Context, *empty.Empty) (*pb.VersionResponse, error) {
	r, err := c.chain.Version()
	if err != nil {
		return nil, err
	}
	return &pb.VersionResponse{
		Size: r,
	}, nil
}
