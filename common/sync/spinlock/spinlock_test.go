/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package spinlock

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTryLock(t *testing.T) {
	var sl SpinLock
	if !sl.TryLock() {
		t.Error("TryLock failed.")
	}
}

func TestLock(t *testing.T) {
	var sl SpinLock
	t.Log(sl.String())
	sl.Lock()

	if sl.TryLock() {
		t.Error("TryLock() returned true when it shouldn't have, w t f mate?")
	}

	sl.Unlock()

	if !sl.TryLock() {
		t.Error("TryLock() returned false when it shouldn't have.")
	}
	t.Log(sl.String())
}

type SpinMap struct {
	SpinLock
	m map[int]bool
}

func (sm *SpinMap) Add(i int) {
	sm.Lock()
	sm.m[i] = true
	sm.Unlock()
}

func (sm *SpinMap) Get(i int) (b bool) {
	sm.Lock()
	b = sm.m[i]
	sm.Unlock()
	return
}

type MutexMap struct {
	sync.Mutex
	m map[int]bool
}

func (mm *MutexMap) Add(i int) {
	mm.Lock()
	mm.m[i] = true
	mm.Unlock()
}

func (mm *MutexMap) Get(i int) (b bool) {
	mm.Lock()
	b = mm.m[i]
	mm.Unlock()
	return
}

type RWMutexMap struct {
	sync.RWMutex
	m map[int]bool
}

func (mm *RWMutexMap) Add(i int) {
	mm.Lock()
	mm.m[i] = true
	mm.Unlock()
}

func (mm *RWMutexMap) Get(i int) (b bool) {
	mm.RLock()
	b = mm.m[i]
	mm.RUnlock()
	return
}

const N = 1e3

var (
	sm   = SpinMap{m: map[int]bool{}}
	mm   = MutexMap{m: map[int]bool{}}
	rwmm = RWMutexMap{m: map[int]bool{}}
)

func BenchmarkSpinL(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(N * 2)
		for i := 0; i < N; i++ {
			go func(i int) {
				sm.Add(i)
				wg.Done()
			}(i)
			go sm.Get(i)

			go func(i int) {
				sm.Get(i - 1)
				sm.Get(i + 1)
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
}

func BenchmarkMutex(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(N * 2)
		for i := 0; i < N; i++ {
			go func(i int) {
				mm.Add(i)
				wg.Done()
			}(i)
			go mm.Get(i)
			go func(i int) {
				mm.Get(i - 1)
				mm.Get(i + 1)
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
}

func BenchmarkRWMutex(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(N * 2)
		for i := 0; i < N; i++ {
			go func(i int) {
				rwmm.Add(i)
				wg.Done()
			}(i)
			go mm.Get(i)
			go func(i int) {
				rwmm.Get(i - 1)
				rwmm.Get(i + 1)
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
}

func testLock(threads, n int, l sync.Locker) time.Duration {
	var wg sync.WaitGroup
	wg.Add(threads)

	var count1 int
	var count2 int

	start := time.Now()
	for i := 0; i < threads; i++ {
		go func() {
			for i := 0; i < n; i++ {
				l.Lock()
				count1++
				count2 += 2
				l.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	dur := time.Since(start)
	if count1 != threads*n {
		panic("mismatch")
	}
	if count2 != threads*n*2 {
		panic("mismatch")
	}
	return dur
}

func TestSpinLock(t *testing.T) {
	n := 1000000
	fmt.Printf("[1] spinlock %4.0fms\n", testLock(1, n, &SpinLock{}).Seconds()*1000)
	fmt.Printf("[1] mutex    %4.0fms\n", testLock(1, n, &sync.Mutex{}).Seconds()*1000)
	fmt.Printf("[4] spinlock %4.0fms\n", testLock(4, n, &SpinLock{}).Seconds()*1000)
	fmt.Printf("[4] mutex    %4.0fms\n", testLock(4, n, &sync.Mutex{}).Seconds()*1000)
	fmt.Printf("[8] spinlock %4.0fms\n", testLock(8, n, &SpinLock{}).Seconds()*1000)
	fmt.Printf("[8] mutex    %4.0fms\n", testLock(8, n, &sync.Mutex{}).Seconds()*1000)
}
