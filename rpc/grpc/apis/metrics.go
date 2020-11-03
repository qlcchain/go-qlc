package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
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

func (m *MetricsAPI) GetCPUInfo(context.Context, *empty.Empty) (*pb.InfoStats, error) {
	r, err := m.metrics.GetCPUInfo()
	if err != nil {
		return nil, err
	}
	return toInfoStats(r), nil
}

func (m *MetricsAPI) GetAllCPUTimeStats(context.Context, *empty.Empty) (*pb.TimesStats, error) {
	r, err := m.metrics.GetAllCPUTimeStats()
	if err != nil {
		return nil, err
	}
	return toTimesStats(r), nil
}

func (m *MetricsAPI) GetCPUTimeStats(context.Context, *empty.Empty) (*pb.TimesStats, error) {
	r, err := m.metrics.GetCPUTimeStats()
	if err != nil {
		return nil, err
	}
	return toTimesStats(r), nil
}

func (m *MetricsAPI) DiskInfo(context.Context, *empty.Empty) (*pb.UsageStat, error) {
	r, err := m.metrics.DiskInfo()
	if err != nil {
		return nil, err
	}
	return toUsageStat(r), nil
}

func (m *MetricsAPI) GetNetworkInterfaces(context.Context, *empty.Empty) (*pb.IOCountersStats, error) {
	r, err := m.metrics.GetNetworkInterfaces()
	if err != nil {
		return nil, err
	}
	return toIOCountersStats(r), nil
}

func toInfoStats(stats []cpu.InfoStat) *pb.InfoStats {
	ss := make([]*pb.InfoStat, 0)
	for _, s := range stats {
		st := &pb.InfoStat{
			Cpu:        s.CPU,
			VendorID:   s.VendorID,
			Family:     s.Family,
			Model:      s.Model,
			Stepping:   s.Stepping,
			PhysicalId: s.PhysicalID,
			CoreId:     s.CoreID,
			Cores:      s.Cores,
			ModelName:  s.ModelName,
			Mhz:        s.Mhz,
			CacheSize:  s.CacheSize,
			Flags:      s.Flags,
			Microcode:  s.Microcode,
		}
		ss = append(ss, st)
	}
	return &pb.InfoStats{Stats: ss}
}

func toTimesStats(stats []cpu.TimesStat) *pb.TimesStats {
	ss := make([]*pb.TimesStat, 0)
	for _, s := range stats {
		st := &pb.TimesStat{
			Cpu:       s.CPU,
			User:      s.User,
			System:    s.System,
			Idle:      s.Idle,
			Nice:      s.Nice,
			Iowait:    s.Iowait,
			Irq:       s.Irq,
			Softirq:   s.Softirq,
			Steal:     s.Steal,
			Guest:     s.Guest,
			GuestNice: s.GuestNice,
		}
		ss = append(ss, st)
	}
	return &pb.TimesStats{Stats: ss}
}

func toUsageStat(disk *disk.UsageStat) *pb.UsageStat {
	return &pb.UsageStat{
		Path:              disk.Path,
		Fstype:            disk.Fstype,
		Total:             disk.Total,
		Free:              disk.Free,
		Used:              disk.Used,
		UsedPercent:       disk.UsedPercent,
		InodesTotal:       disk.InodesTotal,
		InodesUsed:        disk.InodesUsed,
		InodesFree:        disk.InodesFree,
		InodesUsedPercent: disk.InodesUsedPercent,
	}
}

func toIOCountersStats(stats []net.IOCountersStat) *pb.IOCountersStats {
	states := make([]*pb.IOCountersStat, 0)
	for _, s := range stats {
		st := &pb.IOCountersStat{
			Name:        s.Name,
			BytesSent:   s.BytesSent,
			BytesRecv:   s.BytesRecv,
			PacketsSent: s.PacketsSent,
			PacketsRecv: s.PacketsRecv,
			Errin:       s.Errin,
			Errout:      s.Errout,
			Dropin:      s.Dropin,
			Dropout:     s.Dropout,
			Fifoin:      s.Fifoin,
			Fifoout:     s.Fifoout,
		}
		states = append(states, st)
	}
	return &pb.IOCountersStats{Stats: states}
}
