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
	"github.com/shirou/gopsutil/disk"
)

var (
	registerDiskOnce = sync.Once{}
	ioCounterStats   map[string]IOCounterStats
)

type IOCounterStats struct {
	ReadCount        metrics.Gauge
	MergedReadCount  metrics.Gauge
	WriteCount       metrics.Gauge
	MergedWriteCount metrics.Gauge
	ReadBytes        metrics.Gauge
	WriteBytes       metrics.Gauge
	ReadTime         metrics.Gauge
	WriteTime        metrics.Gauge
	IopsInProgress   metrics.Gauge
	IoTime           metrics.Gauge
	WeightedIO       metrics.Gauge
}

func RegisterDiskStats(r metrics.Registry) {
	registerDiskOnce.Do(func() {
		stats, err := disk.IOCounters()
		if err == nil && len(stats) > 0 {
			ioCounterStats = make(map[string]IOCounterStats)
			for name := range stats {
				counterStat := IOCounterStats{}
				counterStat.ReadCount = metrics.NewGauge()
				counterStat.MergedReadCount = metrics.NewGauge()
				counterStat.WriteCount = metrics.NewGauge()
				counterStat.MergedWriteCount = metrics.NewGauge()
				counterStat.ReadBytes = metrics.NewGauge()
				counterStat.WriteBytes = metrics.NewGauge()
				counterStat.ReadTime = metrics.NewGauge()
				counterStat.WriteTime = metrics.NewGauge()
				counterStat.IopsInProgress = metrics.NewGauge()
				counterStat.IoTime = metrics.NewGauge()
				counterStat.WeightedIO = metrics.NewGauge()

				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].ReadCount", name), counterStat.ReadCount)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].MergedReadCount", name), counterStat.MergedReadCount)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].WriteCount", name), counterStat.WriteCount)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].MergedWriteCount", name), counterStat.MergedWriteCount)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].ReadBytes", name), counterStat.ReadBytes)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].WriteBytes", name), counterStat.WriteBytes)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].ReadTime", name), counterStat.ReadTime)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].WriteTime", name), counterStat.WriteTime)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].IopsInProgress", name), counterStat.IopsInProgress)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].IoTime", name), counterStat.IoTime)
				r.Register(fmt.Sprintf("runtime.ioCounterStats[%s].WeightedIO", name), counterStat.WeightedIO)
				ioCounterStats[name] = counterStat
			}
		}
	})
}

func CaptureRuntimeDiskStatsOnce(r metrics.Registry) {
	stats, err := disk.IOCounters()
	if err == nil && len(stats) > 0 {
		for name, status := range stats {
			ioCounterStats[name].ReadCount.Update(int64(status.ReadCount))
			ioCounterStats[name].MergedReadCount.Update(int64(status.MergedReadCount))
			ioCounterStats[name].WriteCount.Update(int64(status.WriteCount))
			ioCounterStats[name].MergedWriteCount.Update(int64(status.MergedWriteCount))
			ioCounterStats[name].ReadBytes.Update(int64(status.ReadBytes))
			ioCounterStats[name].WriteBytes.Update(int64(status.WriteBytes))
			ioCounterStats[name].ReadTime.Update(int64(status.ReadTime))
			ioCounterStats[name].WriteTime.Update(int64(status.WriteTime))
			ioCounterStats[name].IopsInProgress.Update(int64(status.IopsInProgress))
			ioCounterStats[name].IoTime.Update(int64(status.IoTime))
			ioCounterStats[name].WeightedIO.Update(int64(status.WeightedIO))
		}
	}
}

func CaptureRuntimeDiskStats(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			CaptureRuntimeDiskStatsOnce(SystemRegistry)
		}
	}
}

func DiskInfo() (*disk.UsageStat, error) {
	return disk.Usage("/")
}
