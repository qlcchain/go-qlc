package dpos

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/mock"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPerf(t *testing.T) {
	if perfTypeClose.String() != "close" || perfTypeBlockLife.String() != "blockLife" ||
		perfTypeBlockProcess.String() != "blockProcess" || perfTypeAll.String() != "blockLife and blockProcess" {
		t.Fatal()
	}

	dps := getTestDpos()

	out := make(map[string]interface{})
	dps.RPC(common.RpcDPoSSetConsPerf, perfTypeBlockLife, out)

	hash1 := mock.Hash()
	hash2 := mock.Hash()
	dps.perfBlockLifeCheckPointAdd(hash1, checkPointReceive)
	dps.perfBlockLifeCheckPointAdd(hash1, checkPointBlockProcessStart)
	dps.perfBlockLifeCheckPointAdd(hash1, checkPointBlockProcessEnd)
	dps.perfBlockLifeCheckPointAdd(hash1, checkPointUncheckProcessStart)
	dps.perfBlockLifeCheckPointAdd(hash1, checkPointUncheckProcessEnd)
	dps.perfBlockLifeCheckPointAdd(hash1, checkPointBlockConfirmed)
	dps.perfBlockLifeCheckPointAdd(hash2, checkPointReceive)
	dps.perfBlockLifeCheckPointAdd(hash2, checkPointBlockProcessStart)
	dps.perfBlockLifeCheckPointAdd(hash2, checkPointBlockProcessEnd)
	dps.perfBlockLifeCheckPointAdd(hash2, checkPointUncheckProcessStart)
	dps.perfBlockLifeCheckPointAdd(hash2, checkPointUncheckProcessEnd)
	dps.perfBlockLifeCheckPointAdd(hash2, checkPointBlockConfirmed)

	in := make(map[string]interface{})
	rsp := make(map[string]interface{})
	dps.RPC(common.RpcDPoSGetConsPerf, in, rsp)
	if rsp["status"] != perfTypeBlockLife.String() || rsp["blNum"] != 2 {
		t.Fatal()
	}

	dps.RPC(common.RpcDPoSSetConsPerf, perfTypeClose, out)
	in = make(map[string]interface{})
	rsp = make(map[string]interface{})
	dps.RPC(common.RpcDPoSGetConsPerf, in, rsp)
	if rsp["status"] != perfTypeClose.String() {
		t.Fatal()
	}

	dps.RPC(common.RpcDPoSSetConsPerf, perfTypeBlockProcess, out)
	in = make(map[string]interface{})
	rsp = make(map[string]interface{})
	dps.RPC(common.RpcDPoSGetConsPerf, in, rsp)
	if rsp["status"] != perfTypeBlockProcess.String() || rsp["bpNum"] != 0 {
		t.Fatal()
	}

	dps.perfBlockProcessCheckPointAdd(hash1, checkPointBlockCheck)
	dps.perfBlockProcessCheckPointAdd(hash1, checkPointProcessResult)
	dps.perfBlockProcessCheckPointAdd(hash1, checkPointEnd)
	dps.perfBlockProcessCheckPointAdd(hash1, checkPointSectionStart)
	dps.perfBlockProcessCheckPointAdd(hash1, checkPointSectionEnd)
	dps.perfBlockProcessCheckPointAdd(hash1, checkPointSection2Start)
	dps.perfBlockProcessCheckPointAdd(hash1, checkPointSection2End)
	in = make(map[string]interface{})
	rsp = make(map[string]interface{})
	dps.RPC(common.RpcDPoSGetConsPerf, in, rsp)
	if rsp["status"] != perfTypeBlockProcess.String() || rsp["bpNum"] != 1 {
		t.Fatal()
	}

	dps.RPC(common.RpcDPoSSetConsPerf, perfTypeExport, out)
	if out["err"] != nil {
		t.Fatal(out["err"])
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	err = filepath.Walk(dir, func(filename string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(fi.Name()), "_bl.csv") {
			os.Remove(filename)
		}
		if strings.HasSuffix(strings.ToLower(fi.Name()), "_bp.csv") {
			os.Remove(filename)
		}
		return nil
	})
}

func TestGetConsInfo(t *testing.T) {
	dps := getTestDpos()

	out := make(map[string]interface{})
	dps.RPC(common.RpcDPoSConsInfo, nil, out)
	if out["err"] != nil {
		t.Fatal(out["err"])
	}
}
