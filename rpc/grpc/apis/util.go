package apis

import (
	"context"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

type UtilApi struct {
	util   *api.UtilAPI
	logger *zap.SugaredLogger
}

func NewUtilApi(l ledger.Store) *UtilApi {
	return &UtilApi{
		util:   api.NewUtilAPI(l),
		logger: log.NewLogger("grpc_util"),
	}
}

func (u *UtilApi) Decrypt(ctx context.Context, param *pb.DecryptRequest) (*pb.String, error) {
	r, err := u.util.Decrypt(param.GetCryptograph(), param.GetPassphrase())
	if err != nil {
		return nil, err
	}
	return &pb.String{
		Value: r,
	}, nil
}

func (u *UtilApi) Encrypt(ctx context.Context, param *pb.EncryptRequest) (*pb.String, error) {
	r, err := u.util.Encrypt(param.GetRaw(), param.GetPassphrase())
	if err != nil {
		return nil, err
	}
	return &pb.String{
		Value: r,
	}, nil
}

func (u *UtilApi) RawToBalance(ctx context.Context, param *pb.RawBalance) (*pb.Float, error) {
	amount := toOriginBalanceByValue(param.GetBalance())
	token := param.GetTokenName()
	r, err := u.util.RawToBalance(amount, param.GetUnit(), &token)
	if err != nil {
		return nil, err
	}
	b, _ := r.Float32()
	return &pb.Float{
		Value: b,
	}, nil
}

func (u *UtilApi) BalanceToRaw(ctx context.Context, param *pb.RawBalance) (*pb.Int64, error) {
	amount := toOriginBalanceByValue(param.GetBalance())
	token := toStringPoint(param.GetTokenName())
	r, err := u.util.BalanceToRaw(amount, param.GetUnit(), token)
	if err != nil {
		return nil, err
	}
	return &pb.Int64{
		Value: r.Int64(),
	}, nil
}
