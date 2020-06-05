package apis

import (
	"context"
	"github.com/qlcchain/go-qlc/common/util"
	"testing"
)

var (
	metricsApi = NewMetricsAPI()
)

func TestMetricsApi_GetDiskInfo(t *testing.T) {
	if stat, err := metricsApi.DiskInfo(context.Background(), nil); err == nil && stat != nil {
		t.Log(stat.String())
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetNetworkInterfaces(t *testing.T) {
	if stats, err := metricsApi.GetNetworkInterfaces(context.Background(), nil); err == nil && stats != nil {
		for idx, stat := range stats.GetStats() {
			t.Log(idx, stat.String())
		}
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetCPUTimeStats(t *testing.T) {
	if stats, err := metricsApi.GetCPUTimeStats(context.Background(), nil); err == nil && stats != nil {
		for idx, stat := range stats.GetStats() {
			t.Log(idx, stat.String())
		}
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetAllCPUTimeStats(t *testing.T) {
	if stats, err := metricsApi.GetAllCPUTimeStats(context.Background(), nil); err == nil && stats != nil {
		for idx, stat := range stats.GetStats() {
			t.Log(idx, stat.String())
		}
	} else {
		t.Fatal(err)
	}
}

func TestMetricsApi_GetCPUInfo(t *testing.T) {
	if info, err := metricsApi.GetCPUInfo(context.Background(), nil); err == nil && info != nil {
		for idx, i := range info.GetStats() {
			t.Log(idx, util.ToIndentString(i))
		}
	} else {
		t.Fatal(err)
	}
}
