/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package sync

import (
	"runtime"
	"testing"
)

func HammerMutex(m *Mutex, loops int, cdone chan bool) {
	for i := 0; i < loops; i++ {
		m.Lock()
		m.Unlock()
	}
	cdone <- true
}

func TestMutex_Lock(t *testing.T) {
	if n := runtime.SetMutexProfileFraction(1); n != 0 {
		t.Logf("got mutexrate %d expected 0", n)
	}
	defer runtime.SetMutexProfileFraction(0)
	m := NewMutex()
	c := make(chan bool)
	for i := 0; i < 10; i++ {
		go HammerMutex(m, 1000, c)
	}
	for i := 0; i < 10; i++ {
		<-c
	}
}

func TestMutex_TryLock(t *testing.T) {
	m := NewMutex()
	m.Lock()
	if isLocked := m.TryLock(); isLocked {
		t.Fatal("try lock after locked failed")
	}
	m.Unlock()
	if isLocked := m.TryLock(); !isLocked {
		t.Fatal("try lock without locked failed")
	}

	if locked := m.IsLocked(); !locked {
		t.Fatal("check lock status failed")
	}
}
