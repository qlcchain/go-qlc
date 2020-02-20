/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package sync

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	LockedFlag   int32 = 1
	UnlockedFlag int32 = 0
)

type Mutex struct {
	in     sync.Mutex
	status *int32
}

func NewMutex() *Mutex {
	status := UnlockedFlag
	return &Mutex{
		status: &status,
	}
}

func (m *Mutex) Lock() {
	m.in.Lock()
	atomic.StoreInt32(m.status, LockedFlag)
}

func (m *Mutex) Unlock() {
	m.in.Unlock()
	atomic.StoreInt32(m.status, UnlockedFlag)
}

func (m *Mutex) TryLock() bool {
	if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.in)), UnlockedFlag, LockedFlag) {
		atomic.AddInt32(m.status, LockedFlag)
		return true
	}
	return false
}

func (m *Mutex) IsLocked() bool {
	return atomic.LoadInt32(m.status) == LockedFlag
}
