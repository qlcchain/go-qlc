/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package monitor

import (
	"testing"

	"github.com/rcrowley/go-metrics"
)

func TestCpuInfo(t *testing.T) {
	infos, err := CPUInfo()
	if err != nil {
		t.Fatal(err)
	}

	for _, info := range infos {
		t.Log(info.String())
	}
}

func TestCaptureRuntimeCpuStatsOnce(t *testing.T) {
	r := SystemRegistry
	CaptureRuntimeCPUStatsOnce(r)
	r.Each(func(s string, i interface{}) {
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
	})
}
