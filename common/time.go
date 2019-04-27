/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/beevik/ntp"
)

var (
	servers = []string{
		"ntp.ntsc.ac.cn",
		"time1.aliyun.com",
		"time2.aliyun.com",
		"time3.aliyun.com",
		"time4.aliyun.com",
		"time5.aliyun.com",
		"time6.aliyun.com",
		"time7.aliyun.com",

		"time1.apple.com",
		"time2.apple.com",
		"time3.apple.com",
		"time4.apple.com",
		"time5.apple.com",
	}
	serverSize     = len(servers)
	availableIndex = 0
	offset         int64
	times          = 5
)

func init() {
	go update()

	go func() {
		t := time.NewTicker(time.Minute)
		for {
			select {
			case <-t.C:
				update()
			}
		}
	}()
}

func TimeNow() time.Time {
	tmp := atomic.LoadInt64(&offset)
	if tmp == 0 {
		return time.Now()
	} else {
		return time.Now().Add(time.Duration(tmp))
	}
}

func update() {
	retry := 0

	for {
		if du, err := delta(servers[availableIndex]); err == nil {
			atomic.StoreInt64(&offset, int64(du))
			break
		} else {
			availableIndex++
			availableIndex = availableIndex % serverSize
			retry++

			if retry > serverSize {
				return
			}
		}
	}
}

func delta(server string) (time.Duration, error) {
	ds := make([]time.Duration, times)
	for i := 0; i < times; i++ {
		if response, err := ntp.Query(server); err == nil {
			ds[i] = response.ClockOffset
		} else {
			return 0, err
		}
	}

	sort.Slice(ds, func(i, j int) bool {
		return ds[i] > ds[j]
	})

	return average(ds[1 : len(ds)-1]), nil
}

func average(ds []time.Duration) time.Duration {
	total := time.Duration(0)
	for _, v := range ds {
		total += v
	}
	return total / time.Duration(len(ds))
}
