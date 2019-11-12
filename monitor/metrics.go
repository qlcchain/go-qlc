/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"time"

	"github.com/rcrowley/go-metrics"
)

var (
	SystemRegistry      = metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "/system/")
	PerformanceRegistry = metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "/performance/")
)

func init() {
	metrics.RegisterRuntimeMemStats(SystemRegistry)
	metrics.RegisterDebugGCStats(SystemRegistry)
	RegisterRuntimeCPUStats(SystemRegistry)
	RegisterDiskStats(SystemRegistry)
	RegisterRuntimeNetworkStats(SystemRegistry)
}

func Duration(invocation time.Time, name string) {
	elapsed := time.Since(invocation)
	timer := metrics.GetOrRegisterTimer(name, PerformanceRegistry)
	timer.Update(elapsed)
}
