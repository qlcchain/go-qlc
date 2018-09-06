/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import "testing"

func TestNewLogger(t *testing.T) {
	logger1 := NewLogger("test1")
	logger1.Debug("debug1")
}
