/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/util"
)

var (
	api = NewMetricsApi()
)

func TestMetricsApi_GetDiskInfo(t *testing.T) {
	if stat, err := api.DiskInfo(); err == nil && stat != nil {
		t.Log(stat.String())
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetNetworkInterfaces(t *testing.T) {
	if stats, err := api.GetNetworkInterfaces(); err == nil && stats != nil {
		for idx, stat := range stats {
			t.Log(idx, stat.String())
		}
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetCPUTimeStats(t *testing.T) {
	if stats, err := api.GetCPUTimeStats(); err == nil && stats != nil {
		for idx, stat := range stats {
			t.Log(idx, stat.String())
		}
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetAllCPUTimeStats(t *testing.T) {
	if stats, err := api.GetAllCPUTimeStats(); err == nil && stats != nil {
		for idx, stat := range stats {
			t.Log(idx, stat.String())
		}
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetCPUInfo(t *testing.T) {
	if info, err := api.GetCPUInfo(); err == nil && info != nil {
		for idx, i := range info {
			t.Log(idx, util.ToIndentString(i))
		}
	} else {
		t.Fatal(err)
	}
}
