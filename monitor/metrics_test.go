/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"github.com/rcrowley/go-metrics"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
	"math/big"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	t.Skip()
	go influxdb.InfluxDB(
		metrics.DefaultRegistry, // metrics registry
		time.Second*10,          // interval
		"http://localhost:8086", // the InfluxDB url
		"mydb",                  // your InfluxDB database
		"myuser",                // your InfluxDB user
		"mypassword",            // your InfluxDB password
	)
}

func TestDuration(t *testing.T) {
	defer Duration(time.Now(), "TestDuration")

	x := big.NewInt(2)
	y := big.NewInt(1)
	for one := big.NewInt(1); x.Sign() > 0; x.Sub(x, one) {
		y.Mul(y, x)
	}
}
