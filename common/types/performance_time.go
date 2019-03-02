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
	Hash Hash  `json:"hash"` //block hash
	T0   int64 `json:"t0"`   //The time when the block message was received for the first time
	T1   int64 `json:"t1"`   //The time when the block was confirmed
	T2   int64 `json:"t2"`   //The time when the block begin consensus
	T3   int64 `json:"t3"`   //The time when the block first consensus complete
}

func NewPerformanceTime() *PerformanceTime {
	now := time.Now().Unix()
	p := PerformanceTime{T0: now, T1: now, T2: now, T3: now, Hash: ZeroHash}
	return &p
}

// String implements the fmt.Stringer interface.
func (p *PerformanceTime) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}
