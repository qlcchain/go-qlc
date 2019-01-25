/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/json"
	"time"
)

type PerformanceTime struct {
	Hash Hash      `json:"hash"`
	T0   time.Time `json:"t0"`
	T1   time.Time `json:"t1"`
	T2   time.Time `json:"t2"`
	T3   time.Time `json:"t3"`
}

func NewPerformanceTime() *PerformanceTime {
	now := time.Now()
	p := PerformanceTime{T0: now, T1: now, T2: now, T3: now, Hash: ZeroHash}
	return &p
}

// String implements the fmt.Stringer interface.
func (p *PerformanceTime) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}
