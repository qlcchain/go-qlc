/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import "sync/atomic"

//go:generate stringer -type=ServiceStatus
type ServiceStatus int32

const (
	Origin ServiceStatus = iota
	Initialing
	Inited
	Starting
	Started
	Stopping
	Stopped
)

type InterceptCall interface {
	RpcCall(kind uint, in, out interface{})
}

//Service action and status
type Service interface {
	Init() error
	Start() error
	Stop() error
	Status() int32
}

type ServiceLifecycle struct {
	Status int32 // ServiceStatus
}

func (s *ServiceLifecycle) PreInit() bool {
	return atomic.CompareAndSwapInt32(&s.Status, 0, 1)
}

func (s *ServiceLifecycle) PostInit() bool {
	return atomic.CompareAndSwapInt32(&s.Status, 1, 2)
}

func (s *ServiceLifecycle) PreStart() bool {
	return atomic.CompareAndSwapInt32(&s.Status, 2, 3)
}

func (s *ServiceLifecycle) PostStart() bool {
	return atomic.CompareAndSwapInt32(&s.Status, 3, 4)
}

func (s *ServiceLifecycle) PreStop() bool {
	return atomic.CompareAndSwapInt32(&s.Status, 4, 5)
}

func (s *ServiceLifecycle) PostStop() bool {
	return atomic.CompareAndSwapInt32(&s.Status, 5, 6)
}

func (s *ServiceLifecycle) Reset() bool {
	return atomic.CompareAndSwapInt32(&s.Status, 6, 0)
}

func (s *ServiceLifecycle) Stopped() bool {
	return s.Status == 6 || s.Status == 5
}

func (s *ServiceLifecycle) State() int32 {
	return s.Status
}

func (s *ServiceLifecycle) String() string {
	return ServiceStatus(s.Status).String()
}
