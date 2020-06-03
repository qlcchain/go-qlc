package apis

import (
	"context"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type PtmKeyAPI struct {
	ptm    *api.PtmKeyApi
	logger *zap.SugaredLogger
}

func NewPtmKeyAPI(cfgFile string, l ledger.Store) *PtmKeyAPI {
	return &PtmKeyAPI{
		ptm:    api.NewPtmKeyApi(cfgFile, l),
		logger: log.NewLogger("grpc_ptm"),
	}
}

func (p *PtmKeyAPI) GetPtmKeyUpdateBlock(ctx context.Context, param *pb.PtmKeyUpdateParam) (*pbtypes.StateBlock, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.ptm.GetPtmKeyUpdateBlock(&api.PtmKeyUpdateParam{
		Account: addr,
		Btype:   param.GetBtype(),
		Pubkey:  param.GetPubkey(),
	})
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PtmKeyAPI) GetPtmKeyDeleteBlock(ctx context.Context, param *pb.PtmKeyByAccountAndBtypeParam) (*pbtypes.StateBlock, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.ptm.GetPtmKeyDeleteBlock(&api.PtmKeyDeleteParam{
		Account: addr,
		Btype:   param.GetBtype(),
	})
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PtmKeyAPI) GetPtmKeyByAccount(ctx context.Context, param *pbtypes.Address) (*pb.PtmKeyUpdateParams, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := p.ptm.GetPtmKeyByAccount(addr)
	if err != nil {
		return nil, err
	}
	return toPtmKeyUpdateParams(r), nil
}

func (p *PtmKeyAPI) GetPtmKeyByAccountAndBtype(ctx context.Context, param *pb.PtmKeyByAccountAndBtypeParam) (*pb.PtmKeyUpdateParams, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.ptm.GetPtmKeyByAccountAndBtype(addr, param.GetBtype())
	if err != nil {
		return nil, err
	}
	return toPtmKeyUpdateParams(r), nil
}

func toPtmKeyUpdateParams(params []*api.PtmKeyUpdateParam) *pb.PtmKeyUpdateParams {
	return &pb.PtmKeyUpdateParams{}
}
