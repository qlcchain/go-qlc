package apis

import (
	"context"
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type NetAPI struct {
	net    *api.NetApi
	logger *zap.SugaredLogger
}

func NewNetApi(l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *NetAPI {
	return &NetAPI{
		net:    api.NewNetApi(l, eb, cc),
		logger: log.NewLogger("grpc_net"),
	}
}

func (n *NetAPI) OnlineRepresentatives(context.Context, *empty.Empty) (*pbtypes.Addresses, error) {
	r := n.net.OnlineRepresentatives()
	return toAddresses(r), nil
}

func (n *NetAPI) OnlineRepsInfo(context.Context, *empty.Empty) (*pb.OnlineRepTotal, error) {
	r := n.net.OnlineRepsInfo()
	return toOnlineRepTotal(r), nil
}

func (n *NetAPI) ConnectPeersInfo(ctx context.Context, param *pb.Offset) (*pb.PeerInfos, error) {
	count := int(param.GetCount())
	offset := int(param.GetOffset())
	r, err := n.net.ConnectPeersInfo(count, &offset)
	if err != nil {
		return nil, err
	}
	return toPeerInfos(r), nil
}

func (n *NetAPI) GetOnlinePeersInfo(ctx context.Context, param *pb.Offset) (*pb.PeerInfos, error) {
	count := int(param.GetCount())
	offset := int(param.GetOffset())
	r, err := n.net.GetOnlinePeersInfo(count, &offset)
	if err != nil {
		return nil, err
	}
	return toPeerInfos(r), nil
}

func (n *NetAPI) GetAllPeersInfo(ctx context.Context, param *pb.Offset) (*pb.PeerInfos, error) {
	count := int(param.GetCount())
	offset := int(param.GetOffset())
	r, err := n.net.GetAllPeersInfo(count, &offset)
	if err != nil {
		return nil, err
	}
	return toPeerInfos(r), nil
}

func (n *NetAPI) PeersCount(context.Context, *empty.Empty) (*pb.PeersCountResponse, error) {
	r, err := n.net.PeersCount()
	if err != nil {
		return nil, err
	}
	return &pb.PeersCountResponse{
		Count: r,
	}, nil
}

func (n *NetAPI) GetBandwidthStats(context.Context, *empty.Empty) (*pb.EventBandwidthStats, error) {
	r := n.net.GetBandwidthStats()
	return &pb.EventBandwidthStats{
		TotalIn:  r.TotalIn,
		TotalOut: r.TotalOut,
		RateIn:   r.RateIn,
		RateOut:  r.RateOut,
	}, nil
}

func (n *NetAPI) Syncing(context.Context, *empty.Empty) (*wrappers.BoolValue, error) {
	r := n.net.Syncing()
	return &wrappers.BoolValue{
		Value: r,
	}, nil
}

func (n *NetAPI) GetPeerId(context.Context, *empty.Empty) (*wrappers.StringValue, error) {
	r := n.net.GetPeerId()
	return &wrappers.StringValue{
		Value: r,
	}, nil
}

func toOnlineRepTotal(rep *api.OnlineRepTotal) *pb.OnlineRepTotal {
	return &pb.OnlineRepTotal{}
}

func toPeerInfos(peers []*types.PeerInfo) *pb.PeerInfos {
	return &pb.PeerInfos{}
}
