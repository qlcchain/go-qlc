package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type AccountApi struct {
	account *api.AccountApi
	logger  *zap.SugaredLogger
}

func NewAccountApi() *AccountApi {
	return &AccountApi{
		account: api.NewAccountApi(),
		logger:  log.NewLogger("grpc_account"),
	}
}

func (a *AccountApi) Create(ctx context.Context, para *pb.CreateRequest) (*pb.CreateResponse, error) {
	seedStr := para.GetSeedStr()
	index := para.GetIndex()
	r, err := a.account.Create(seedStr, &index)
	if err != nil {
		return nil, err
	}
	return &pb.CreateResponse{
		Value: r,
	}, nil
}

func (a *AccountApi) ForPublicKey(ctx context.Context, str *pb.String) (*pbtypes.Address, error) {
	pubStr := toOriginString(str)
	r, err := a.account.ForPublicKey(pubStr)
	if err != nil {
		return nil, err
	}
	return toAddress(r), nil
}

func (a *AccountApi) NewSeed(context.Context, *empty.Empty) (*pb.String, error) {
	r, err := a.account.NewSeed()
	if err != nil {
		return nil, err
	}
	return toString(r), nil
}

func (a *AccountApi) NewAccounts(ctx context.Context, count *pb.UInt32) (*pb.AccountsResponse, error) {
	c := toUInt32PointByProto(count)
	r, err := a.account.NewAccounts(c)
	if err != nil {
		return nil, err
	}
	return toAccounts(r), nil
}

func (a *AccountApi) PublicKey(ctx context.Context, addr *pbtypes.Address) (*pb.String, error) {
	address, err := toOriginAddress(addr)
	if err != nil {
		return nil, err
	}
	r := a.account.PublicKey(address)
	return toString(r), nil
}

func (a *AccountApi) Validate(ctx context.Context, str *pb.String) (*pb.Boolean, error) {
	addStr := toOriginString(str)
	r := a.account.Validate(addStr)
	return toBoolean(r), nil
}

func toAccounts(accs []*api.Accounts) *pb.AccountsResponse {
	as := make([]*pb.Account, 0)
	for _, a := range accs {
		at := &pb.Account{
			Seed:       a.Seed,
			PrivateKey: a.PrivateKey,
			PublicKey:  a.PublicKey,
			Address:    a.Address,
		}
		as = append(as, at)
	}
	return &pb.AccountsResponse{Accounts: as}
}
