package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

type MetricsAPI struct {
	metrics *api.MetricsApi
	logger  *zap.SugaredLogger
}

func NewMetricsAPI() *MetricsAPI {
	return &MetricsAPI{
		metrics: api.NewMetricsApi(),
		logger:  log.NewLogger("grpc_metrics"),
	}
}

func (m *MetricsAPI) GetCPUInfo(context.Context, *empty.Empty) (*proto.InfoStats, error) {
	r, err := m.metrics.GetCPUInfo()
	if err != nil {
		return nil, err
	}
	return toInfoStats(r), nil
}

func (m *MetricsAPI) GetAllCPUTimeStats(context.Context, *empty.Empty) (*proto.TimesStats, error) {
	r, err := m.metrics.GetAllCPUTimeStats()
	if err != nil {
		return nil, err
	}
	return toTimesStats(r), nil
}

func (m *MetricsAPI) GetCPUTimeStats(context.Context, *empty.Empty) (*proto.TimesStats, error) {
	r, err := m.metrics.GetCPUTimeStats()
	if err != nil {
		return nil, err
	}
	return toTimesStats(r), nil
}

func (m *MetricsAPI) DiskInfo(context.Context, *empty.Empty) (*proto.UsageStat, error) {
	r, err := m.metrics.DiskInfo()
	if err != nil {
		return nil, err
	}
	return toUsageStat(r), nil
}

func (m *MetricsAPI) GetNetworkInterfaces(context.Context, *empty.Empty) (*proto.IOCountersStats, error) {
	r, err := m.metrics.GetNetworkInterfaces()
	if err != nil {
		return nil, err
	}
	return toIOCountersStats(r), nil
}

func toInfoStats(stats []cpu.InfoStat) *proto.InfoStats {
	return &proto.InfoStats{}
}

func toTimesStats(stats []cpu.TimesStat) *proto.TimesStats {
	return &proto.TimesStats{}
}

func toUsageStat(disk *disk.UsageStat) *proto.UsageStat {
	return &proto.UsageStat{}
}

func toIOCountersStats(stats []net.IOCountersStat) *proto.IOCountersStats {
	return &proto.IOCountersStats{}
}
