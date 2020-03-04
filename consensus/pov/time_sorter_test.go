/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package pov

import (
	"sort"
	"testing"
)

func TestTimeSorter_Swap(t *testing.T) {
	dt := []uint32{1583127611, 1583127614, 1583127610}
	sort.Sort(TimeSorter(dt))
	t.Log(dt)
}
