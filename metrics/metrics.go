/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package metrics

import "github.com/rcrowley/go-metrics"

var (
	systemRegistry = metrics.NewPrefixedChildRegistry(metrics.DefaultRegistry, "/system")
	cpuRegistry    = metrics.NewPrefixedChildRegistry(systemRegistry, "/cpu")
	memoryRegistry = metrics.NewPrefixedChildRegistry(systemRegistry, "/memory")
	diskRegistry   = metrics.NewPrefixedChildRegistry(systemRegistry, "/disk")
)
