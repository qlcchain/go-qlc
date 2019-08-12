/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package metrics

import (
	"github.com/rcrowley/go-metrics"
)

var (
	SystemRegistry = metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "/system/")
)

func init() {
	metrics.RegisterRuntimeMemStats(SystemRegistry)
	metrics.RegisterDebugGCStats(SystemRegistry)
	RegisterRuntimeCpuStats(SystemRegistry)
	RegisterDiskStats(SystemRegistry)
	RegisterRuntimeNetworkStats(SystemRegistry)
}
