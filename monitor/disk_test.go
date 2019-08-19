/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"strings"
	"testing"

	"github.com/rcrowley/go-metrics"
)

func TestDiskInfo(t *testing.T) {
	stat, err := DiskInfo()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(stat.String())
	}
}

func TestCaptureRuntimeDiskStatsOnce(t *testing.T) {
	r := SystemRegistry
	CaptureRuntimeDiskStatsOnce(r)
	r.Each(func(s string, i interface{}) {
		if strings.Contains(s, "ioCounterStats") {
			switch v := i.(type) {
			case metrics.GaugeFloat64:
				t.Log(s, v.Value())
			case metrics.Timer:
				t.Log(s, v.Count())
			case metrics.Gauge:
				t.Log(s, v.Value())
			case metrics.Histogram:
				t.Log(s, v.Count())
			default:
				t.Log("unknown ", s, v)
			}
		}
	})
}
