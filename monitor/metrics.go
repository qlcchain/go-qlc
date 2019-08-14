/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"github.com/rcrowley/go-metrics"
	"time"
)

var (
	SystemRegistry      = metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "/system/")
	PerformanceRegistry metrics.Registry
)

func init() {
	PerformanceRegistry = metrics.NewPrefixedRegistry("/performance/")
	metrics.RegisterRuntimeMemStats(SystemRegistry)
	metrics.RegisterDebugGCStats(SystemRegistry)
	RegisterRuntimeCpuStats(SystemRegistry)
	RegisterDiskStats(SystemRegistry)
	RegisterRuntimeNetworkStats(SystemRegistry)
}

func Duration(invocation time.Time, name string) {
	elapsed := time.Since(invocation)
	timer := metrics.GetOrRegisterTimer(name, PerformanceRegistry)
	timer.Update(elapsed)
}
