/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import "sync/atomic"

//go:generate stringer -type=TaskStatus
type TaskStatus int32

const (
	Origin TaskStatus = iota
	Initialing
	Inited
	Starting
	Started
	Stopping
	Stopped
)

//Task action and status
type Task interface {
	Init() error
	Start() error
	Stop() error
	Status() int32
}

type TaskLifecycle struct {
	Status int32 // TaskStatus
}

func (t *TaskLifecycle) PreInit() bool {
	return atomic.CompareAndSwapInt32(&t.Status, 0, 1)
}

func (t *TaskLifecycle) PostInit() bool {
	return atomic.CompareAndSwapInt32(&t.Status, 1, 2)
}

func (t *TaskLifecycle) PreStart() bool {
	return atomic.CompareAndSwapInt32(&t.Status, 2, 3)
}

func (t *TaskLifecycle) PostStart() bool {
	return atomic.CompareAndSwapInt32(&t.Status, 3, 4)
}

func (t *TaskLifecycle) PreStop() bool {
	return atomic.CompareAndSwapInt32(&t.Status, 4, 5)
}

func (t *TaskLifecycle) PostStop() bool {
	return atomic.CompareAndSwapInt32(&t.Status, 5, 6)
}

func (t *TaskLifecycle) Stopped() bool {
	return t.Status == 6 || t.Status == 5
}

func (t *TaskLifecycle) State() int32 {
	return t.Status
}

func (t *TaskLifecycle) String() string {
	return TaskStatus(t.Status).String()
}
