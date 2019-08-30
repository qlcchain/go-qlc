/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/shirou/gopsutil/net"
)

var (
	registerNetOnce = sync.Once{}
	netStats        map[string]NetStat
)

type NetStat struct {
	BytesSent   metrics.Gauge
	BytesRecv   metrics.Gauge
	PacketsSent metrics.Gauge
	PacketsRecv metrics.Gauge
	Errin       metrics.Gauge
	Errout      metrics.Gauge
	Dropin      metrics.Gauge
	Dropout     metrics.Gauge
	Fifoin      metrics.Gauge
	Fifoout     metrics.Gauge
}

func RegisterRuntimeNetworkStats(r metrics.Registry) {
	registerNetOnce.Do(func() {
		if interfaces, err := Interfaces(); err == nil {
			netStats = make(map[string]NetStat)
			for _, stat := range interfaces {
				netStat := NetStat{
					BytesSent:   metrics.NewGauge(),
					BytesRecv:   metrics.NewGauge(),
					PacketsSent: metrics.NewGauge(),
					PacketsRecv: metrics.NewGauge(),
					Errin:       metrics.NewGauge(),
					Errout:      metrics.NewGauge(),
					Dropin:      metrics.NewGauge(),
					Dropout:     metrics.NewGauge(),
					Fifoin:      metrics.NewGauge(),
					Fifoout:     metrics.NewGauge(),
				}
				netStats[stat.Name] = netStat
				r.Register(fmt.Sprintf("runtime.netStat[%s].BytesSent", stat.Name), netStat.BytesSent)
				r.Register(fmt.Sprintf("runtime.netStat[%s].BytesRecv", stat.Name), netStat.BytesRecv)
				r.Register(fmt.Sprintf("runtime.netStat[%s].PacketsSent", stat.Name), netStat.PacketsSent)
				r.Register(fmt.Sprintf("runtime.netStat[%s].PacketsRecv", stat.Name), netStat.PacketsRecv)
				r.Register(fmt.Sprintf("runtime.netStat[%s].Errin", stat.Name), netStat.Errin)
				r.Register(fmt.Sprintf("runtime.netStat[%s].Errout", stat.Name), netStat.Errout)
				r.Register(fmt.Sprintf("runtime.netStat[%s].Dropin", stat.Name), netStat.Dropin)
				r.Register(fmt.Sprintf("runtime.netStat[%s].Dropout", stat.Name), netStat.Dropout)
				r.Register(fmt.Sprintf("runtime.netStat[%s].Fifoin", stat.Name), netStat.Fifoin)
				r.Register(fmt.Sprintf("runtime.netStat[%s].Fifoout", stat.Name), netStat.Fifoout)
			}
		}
	})
}

func CaptureRuntimeNetStatsOnce(r metrics.Registry) {
	if stats, err := Interfaces(); err == nil {
		for _, stat := range stats {
			if netStat, ok := netStats[stat.Name]; ok {
				netStat.BytesSent.Update(int64(stat.BytesSent))
				netStat.BytesRecv.Update(int64(stat.BytesRecv))
				netStat.PacketsSent.Update(int64(stat.PacketsSent))
				netStat.PacketsRecv.Update(int64(stat.PacketsRecv))
				netStat.Errin.Update(int64(stat.Errin))
				netStat.Errout.Update(int64(stat.Errout))
				netStat.Dropin.Update(int64(stat.Dropin))
				netStat.Dropout.Update(int64(stat.Dropout))
				netStat.Fifoin.Update(int64(stat.Fifoin))
				netStat.Fifoout.Update(int64(stat.Fifoout))
			}
		}
	}
}

func CaptureRuntimeNetStats(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			CaptureRuntimeNetStatsOnce(SystemRegistry)
		}
	}
}

func Interfaces() ([]net.IOCountersStat, error) {
	return net.IOCounters(true)
}
