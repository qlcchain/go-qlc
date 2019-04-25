/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beevik/ntp"
)

func TestTimeNow(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			t.Log(atomic.LoadInt64(&offset), TimeNow().Format(time.RFC3339), time.Now().Format(time.RFC3339))
		}
	}()

	wg.Wait()
}

func TestTime(t *testing.T) {
	for _, server := range servers {
		t.Log("query ", server)
		start := time.Now()
		t1, err := ntp.Time(server)
		end := time.Now()
		t.Log("cost", end.Sub(start))
		if err != nil {
			t.Log(err)
		} else {
			t.Log(t1, time.Now())
		}
	}
}

func TestQuery(t *testing.T) {
	for _, server := range servers {
		start := time.Now()
		resp, err := ntp.Query(server)
		end := time.Now()
		t.Logf("query %s, cost %s", server, end.Sub(start))
		if err != nil {
			t.Log(err)
		} else {
			t.Log(time.Now().Format(time.RFC3339), resp.ClockOffset)
		}
	}
}

func TestDelta(t *testing.T) {
	for _, server := range servers {
		start := time.Now()
		offset, err := delta(server)
		end := time.Now()
		t.Logf("query %s, cost %s\n", server, end.Sub(start))
		if err != nil {
			t.Log(err)
		} else {
			t.Log(time.Now().Format(time.RFC3339), offset)
		}
	}
}

func TestDuration(t *testing.T) {
	duration, err := time.ParseDuration("20.655039ms")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(duration, int64(duration))
}

func Test_update(t *testing.T) {
	update()
	t.Log("offset", atomic.LoadInt64(&offset))
}
