package grpcServer

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/grpc/proto"
	"go.uber.org/zap"
)

type AccountApi struct {
	logger *zap.SugaredLogger
}

func NewAccountApi() *AccountApi {
	return &AccountApi{logger: log.NewLogger("rpc/account")}
}

func (AccountApi) Create(context.Context, *proto.CreateRequest) (*proto.CreateResponse, error) {
	panic("implement me")
}

func (AccountApi) ForPublicKey(context.Context, *proto.String) (*proto.Address, error) {
	panic("implement me")
}

func (AccountApi) NewSeed(context.Context, *empty.Empty) (*proto.String, error) {
	panic("implement me")
}

func (AccountApi) NewAccounts(context.Context, *proto.UInt32) (*proto.AccountsResponse, error) {
	panic("implement me")
}

func (AccountApi) PublicKey(context.Context, *proto.String) (*proto.String, error) {
	panic("implement me")
}

func (AccountApi) Validate(context.Context, *proto.String) (*proto.Boolean, error) {
	panic("implement me")
}
