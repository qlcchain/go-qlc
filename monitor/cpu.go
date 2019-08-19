/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"fmt"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/shirou/gopsutil/cpu"
)

var (
	registerCPUOnce = sync.Once{}
	cpuStats        map[string]CpuStat
)

type CpuStat struct {
	User         metrics.GaugeFloat64
	System       metrics.GaugeFloat64
	Idle         metrics.GaugeFloat64
	Nice         metrics.GaugeFloat64
	Iowait       metrics.GaugeFloat64
	Irq          metrics.GaugeFloat64
	Softirq      metrics.GaugeFloat64
	Steal        metrics.GaugeFloat64
	Guest        metrics.GaugeFloat64
	GuestNice    metrics.GaugeFloat64
	Stolen       metrics.GaugeFloat64
	ReadCpuStats metrics.Timer
}

func RegisterRuntimeCpuStats(r metrics.Registry) {
	registerCPUOnce.Do(func() {
		stats, err := cpu.Times(true)
		if err == nil {
			cpuStats = make(map[string]CpuStat)
			for _, stat := range stats {
				cpuStat := CpuStat{}
				cpuStat.User = metrics.NewGaugeFloat64()
				cpuStat.System = metrics.NewGaugeFloat64()
				cpuStat.Idle = metrics.NewGaugeFloat64()
				cpuStat.Nice = metrics.NewGaugeFloat64()
				cpuStat.Iowait = metrics.NewGaugeFloat64()
				cpuStat.Irq = metrics.NewGaugeFloat64()
				cpuStat.Softirq = metrics.NewGaugeFloat64()
				cpuStat.Steal = metrics.NewGaugeFloat64()
				cpuStat.Guest = metrics.NewGaugeFloat64()
				cpuStat.GuestNice = metrics.NewGaugeFloat64()
				cpuStat.Stolen = metrics.NewGaugeFloat64()
				cpuStat.ReadCpuStats = metrics.NewTimer()

				cpuStats[stat.CPU] = cpuStat

				r.Register(fmt.Sprintf("runtime.cpuStat[%s].User", stat.CPU), cpuStat.User)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].System", stat.CPU), cpuStat.System)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Idle", stat.CPU), cpuStat.Idle)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Nice", stat.CPU), cpuStat.Nice)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Iowait", stat.CPU), cpuStat.Iowait)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Irq", stat.CPU), cpuStat.Irq)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Softirq", stat.CPU), cpuStat.Softirq)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Steal", stat.CPU), cpuStat.Steal)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Guest", stat.CPU), cpuStat.Guest)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].GuestNice", stat.CPU), cpuStat.GuestNice)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].Stolen", stat.CPU), cpuStat.Stolen)
				r.Register(fmt.Sprintf("runtime.cpuStat[%s].ReadCpuStats", stat.CPU), cpuStat.ReadCpuStats)
			}
		}
	})
}

func CaptureRuntimeCpuStatsOnce(r metrics.Registry) {
	stats, err := cpu.Times(true)
	if err == nil {
		t := time.Now()
		for _, stat := range stats {
			if cpuStat, ok := cpuStats[stat.CPU]; ok {
				cpuStat.User.Update(stat.User)
				cpuStat.User.Update(stat.User)
				cpuStat.System.Update(stat.System)
				cpuStat.Idle.Update(stat.Idle)
				cpuStat.Nice.Update(stat.Nice)
				cpuStat.Iowait.Update(stat.Iowait)
				cpuStat.Irq.Update(stat.Irq)
				cpuStat.Softirq.Update(stat.Softirq)
				cpuStat.Steal.Update(stat.Steal)
				cpuStat.Guest.Update(stat.Guest)
				cpuStat.GuestNice.Update(stat.GuestNice)
				cpuStat.Stolen.Update(stat.Stolen)
				cpuStat.ReadCpuStats.UpdateSince(t)
			}
		}
	}
}

func CaptureRuntimeCpuStats(r metrics.Registry, d time.Duration) {
	for range time.Tick(d) {
		CaptureRuntimeCpuStatsOnce(r)
	}
}

func CpuInfo() ([]cpu.InfoStat, error) {
	return cpu.Info()
}
