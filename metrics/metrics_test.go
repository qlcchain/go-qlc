/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package metrics

import (
	"github.com/rcrowley/go-metrics"
	"github.com/vrischmann/go-metrics-influxdb"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	go influxdb.InfluxDB(
		metrics.DefaultRegistry, // metrics registry
		time.Second*10,          // interval
		"http://localhost:8086", // the InfluxDB url
		"mydb",                  // your InfluxDB database
		"myuser",                // your InfluxDB user
		"mypassword",            // your InfluxDB password
	)
}
