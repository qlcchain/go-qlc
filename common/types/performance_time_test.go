/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"testing"
)

func TestPerformanceTime_String(t *testing.T) {
	p := NewPerformanceTime()
	t.Log(p.String())
}
