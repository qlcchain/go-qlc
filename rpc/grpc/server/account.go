package grpcServer

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/go-qlc/log"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	"go.uber.org/zap"
)

type AccountApi struct {
	logger *zap.SugaredLogger
}

func NewAccountApi() *AccountApi {
	return &AccountApi{logger: log.NewLogger("rpc/account")}
}

func (AccountApi) Create(context.Context, *pb.CreateRequest) (*pb.CreateResponse, error) {
	panic("implement me")
}

func (AccountApi) ForPublicKey(context.Context, *pb.String) (*pbtypes.Address, error) {
	panic("implement me")
}

func (AccountApi) NewSeed(context.Context, *empty.Empty) (*pb.String, error) {
	panic("implement me")
}

func (AccountApi) NewAccounts(context.Context, *pb.UInt32) (*pb.AccountsResponse, error) {
	panic("implement me")
}

func (AccountApi) PublicKey(context.Context, *pb.String) (*pb.String, error) {
	panic("implement me")
}

func (AccountApi) Validate(context.Context, *pb.String) (*pb.Boolean, error) {
	panic("implement me")
}
