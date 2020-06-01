package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type PermissionAPI struct {
	permission *api.PermissionApi
	logger     *zap.SugaredLogger
}

func NewPermissionAPI(cfgFile string, l ledger.Store) *PermissionAPI {
	return &PermissionAPI{
		permission: api.NewPermissionApi(cfgFile, l),
		logger:     log.NewLogger("grpc_permission"),
	}
}

func (p *PermissionAPI) GetAdminHandoverBlock(ctx context.Context, params *pb.AdminUpdateParam) (*pbtypes.StateBlock, error) {
	admin, err := types.HexToAddress(params.GetAdmin())
	if err != nil {
		return nil, err
	}
	successor, err := types.HexToAddress(params.GetSuccessor())
	if err != nil {
		return nil, err
	}
	r, err := p.permission.GetAdminHandoverBlock(&api.AdminUpdateParam{
		Admin:     admin,
		Successor: successor,
		Comment:   params.GetComment(),
	})
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PermissionAPI) GetAdmin(ctx context.Context, params *empty.Empty) (*pb.AdminUser, error) {
	r, err := p.permission.GetAdmin()
	if err != nil {
		return nil, err
	}
	return &pb.AdminUser{
		Account: r.Account.String(),
		Comment: r.Comment,
	}, nil
}

func (p *PermissionAPI) GetNodeUpdateBlock(ctx context.Context, params *pb.NodeParam) (*pbtypes.StateBlock, error) {
	addr, err := types.HexToAddress(params.GetAdmin())
	if err != nil {
		return nil, err
	}
	r, err := p.permission.GetNodeUpdateBlock(&api.NodeParam{
		Admin:   addr,
		NodeId:  params.GetNodeId(),
		NodeUrl: params.GetNodeUrl(),
		Comment: params.GetComment(),
	})
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PermissionAPI) GetNode(ctx context.Context, params *pb.String) (*pb.NodeInfo, error) {
	r, err := p.permission.GetNode(params.GetValue())
	if err != nil {
		return nil, err
	}
	return &pb.NodeInfo{
		NodeId:  r.NodeId,
		NodeUrl: r.NodeUrl,
		Comment: r.Comment,
	}, nil
}

func (p *PermissionAPI) GetNodesCount(ctx context.Context, params *empty.Empty) (*pb.Int32, error) {
	r := p.permission.GetNodesCount()
	return &pb.Int32{
		Value: int32(r),
	}, nil
}

func (p *PermissionAPI) GetNodes(ctx context.Context, params *pb.Offset) (*pb.NodeInfos, error) {
	count := int(params.GetCount())
	offset := int(params.GetOffset())
	r, err := p.permission.GetNodes(count, offset)
	if err != nil {
		return nil, err
	}
	nodes := make([]*pb.NodeInfo, 0)
	for _, info := range r {
		n := &pb.NodeInfo{
			NodeId:  info.NodeId,
			NodeUrl: info.NodeUrl,
			Comment: info.Comment,
		}
		nodes = append(nodes, n)
	}
	return &pb.NodeInfos{
		Nodes: nodes,
	}, nil
}
