/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/monitor/influxdb"

	"github.com/rcrowley/go-metrics"
)

func TestMetrics(t *testing.T) {
	t.Skip()
	ctx, _ := context.WithCancel(context.Background())

	go influxdb.InfluxDB(
		ctx,
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
