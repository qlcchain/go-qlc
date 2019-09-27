/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/monitor"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
	"go.uber.org/zap"
)

type MetricsApi struct {
	logger *zap.SugaredLogger
}

func NewMetricsApi() *MetricsApi {
	return &MetricsApi{logger: log.NewLogger("api_metrics")}
}

// GetCpuInfo get cpu info by core
func (m *MetricsApi) GetCPUInfo() ([]cpu.InfoStat, error) {
	return monitor.CpuInfo()
}

// GetAllCPUTimeStats get total CPU time status
func (m *MetricsApi) GetAllCPUTimeStats() ([]cpu.TimesStat, error) {
	return cpu.Times(false)
}

// GetAllCPUTimeStats get CPU time status per core
func (m *MetricsApi) GetCPUTimeStats() ([]cpu.TimesStat, error) {
	return cpu.Times(true)
}

// DiskInfo get disk info
func (m *MetricsApi) DiskInfo() (*disk.UsageStat, error) {
	return monitor.DiskInfo()
}

// GetNetworkInterfaces get network interface info
func (m *MetricsApi) GetNetworkInterfaces() ([]net.IOCountersStat, error) {
	return monitor.Interfaces()
}
