package dpos

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
)

type checkPointPos byte

const (
	checkPointReceive checkPointPos = iota
	checkPointBlockProcessStart
	checkPointBlockProcessEnd
	checkPointUncheckProcessStart
	checkPointUncheckProcessEnd
	checkPointBlockConfirmed

	checkPointBlockCheck
	checkPointProcessResult
	checkPointEnd
	checkPointSectionStart
	checkPointSectionEnd
)

const (
	rcv = iota
	bps
	bpe
	up1s
	up1e
	up2s
	up2e
	up3s
	up3e
	bcm
	blBut
)

const (
	bc = iota
	pr
	ed
	ss
	se
	bpBut
)

type PerfType int32

const (
	perfTypeClose PerfType = iota
	perfTypeBlockLife
	perfTypeBlockProcess
	perfTypeAll
	perfTypeExport
)

type perfInfo struct {
	status       atomic.Value
	blockLife    *sync.Map
	blockProcess *sync.Map
}

func (pt PerfType) String() string {
	switch pt {
	case perfTypeClose:
		return "close"
	case perfTypeBlockLife:
		return "blockLife"
	case perfTypeBlockProcess:
		return "blockProcess"
	case perfTypeAll:
		return "blockLife and blockProcess"
	default:
		return "invalid"
	}
}

func (dps *DPoS) setPerf(in interface{}, out interface{}) {
	op := in.(PerfType)
	rsp := out.(map[string]interface{})
	rsp["err"] = nil

	switch op {
	case perfTypeClose:
		dps.pf.status.Store(op)
	case perfTypeBlockLife, perfTypeBlockProcess, perfTypeAll:
		dps.pf.status.Store(op)
		dps.pf.blockLife = new(sync.Map)
		dps.pf.blockProcess = new(sync.Map)
	case perfTypeExport:
		if dps.pf.status.Load() == perfTypeClose {
			rsp["err"] = fmt.Errorf("performance test is closed")
		} else {
			err := dps.perfDataExport()
			if err != nil {
				rsp["err"] = fmt.Errorf("export err(%s)", err)
			}
		}
	default:
		rsp["err"] = fmt.Errorf("invalid param")
	}
}

func (dps *DPoS) getPerf(in interface{}, out interface{}) {
	rsp := out.(map[string]interface{})
	rsp["err"] = nil

	rsp["status"] = fmt.Sprintf("%s", dps.pf.status.Load())
	blNum := 0
	bpNum := 0

	if dps.pf.blockLife != nil {
		dps.pf.blockLife.Range(func(key, value interface{}) bool {
			blNum++
			return true
		})
	}

	if dps.pf.blockProcess != nil {
		dps.pf.blockProcess.Range(func(key, value interface{}) bool {
			bpNum++
			return true
		})
	}

	rsp["blNum"] = blNum
	rsp["bpNum"] = bpNum
}

func (dps *DPoS) perfBlockLifeCheckPointAdd(hash types.Hash, pos checkPointPos) {
	if dps.pf.status.Load() == perfTypeClose || dps.pf.status.Load() == perfTypeBlockProcess {
		return
	}

	var bl []time.Time
	blv, ok := dps.pf.blockLife.Load(hash)
	if !ok {
		bl = make([]time.Time, blBut)
		dps.pf.blockLife.Store(hash, bl)
	} else {
		bl = blv.([]time.Time)
	}

	now := time.Now()
	switch pos {
	case checkPointReceive:
		if bl[rcv].IsZero() {
			bl[rcv] = now
		}
	case checkPointBlockProcessStart:
		if bl[bps].IsZero() {
			bl[bps] = now
		}
	case checkPointBlockProcessEnd:
		if bl[bpe].IsZero() {
			bl[bpe] = now
		}
	case checkPointUncheckProcessStart:
		if bl[up1s].IsZero() {
			bl[up1s] = now
		} else if bl[up2s].IsZero() {
			bl[up2s] = now
		} else if bl[up3s].IsZero() {
			bl[up3s] = now
		}
	case checkPointUncheckProcessEnd:
		if bl[up1e].IsZero() {
			bl[up1e] = now
		} else if bl[up2e].IsZero() {
			bl[up2e] = now
		} else if bl[up3e].IsZero() {
			bl[up3e] = now
		}
	case checkPointBlockConfirmed:
		if bl[bcm].IsZero() {
			bl[bcm] = now
		}
	}
}

func (dps *DPoS) perfBlockProcessCheckPointAdd(hash types.Hash, pos checkPointPos) {
	if dps.pf.status.Load() == perfTypeClose || dps.pf.status.Load() == perfTypeBlockLife {
		return
	}

	var bp []time.Time
	bpv, ok := dps.pf.blockProcess.Load(hash)
	if !ok {
		bp = make([]time.Time, bpBut)
		dps.pf.blockProcess.Store(hash, bp)
	} else {
		bp = bpv.([]time.Time)
	}

	now := time.Now()
	switch pos {
	case checkPointBlockCheck:
		if bp[bc].IsZero() {
			bp[bc] = now
		}
	case checkPointProcessResult:
		if bp[pr].IsZero() {
			bp[pr] = now
		}
	case checkPointEnd:
		if bp[ed].IsZero() {
			bp[ed] = now
		}
	case checkPointSectionStart:
		if bp[ss].IsZero() {
			bp[ss] = now
		}
	case checkPointSectionEnd:
		if bp[se].IsZero() {
			bp[se] = now
		}
	}
}

func (dps *DPoS) perfDataExport() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	blf, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s_bl.csv", time.Now().Format("20060102T150405"))))
	if err != nil {
		return err
	}

	bpf, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s_bp.csv", time.Now().Format("20060102T150405"))))
	if err != nil {
		return err
	}

	_, _ = blf.Write([]byte("hash,all,bpc,up1c,up2c,up3c\n"))
	dps.pf.blockLife.Range(func(hash, val interface{}) bool {
		bl := val.([]time.Time)
		_, _ = blf.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
			hash, bl[bcm].Sub(bl[rcv]), bl[bpe].Sub(bl[bps]), bl[up1e].Sub(bl[up1s]),
			bl[up2e].Sub(bl[up2s]), bl[up3e].Sub(bl[up3s])))
		return true
	})

	_, _ = bpf.Write([]byte("hash,all,bcu,pru,su\n"))
	dps.pf.blockProcess.Range(func(hash, val interface{}) bool {
		bp := val.([]time.Time)
		_, _ = bpf.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s\n",
			hash, bp[ed].Sub(bp[bc]), bp[pr].Sub(bp[bc]), bp[ed].Sub(bp[pr]), bp[se].Sub(bp[ss])))
		return true
	})

	_ = blf.Close()
	_ = bpf.Close()
	return nil
}
