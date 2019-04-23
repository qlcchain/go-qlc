/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package util

import "testing"

func TestRandomFixedString(t *testing.T) {
	l := 16
	pw := RandomFixedString(l)
	if len(pw) != l {
		t.Fatal("invalid len", len(pw))
	}
	t.Log(pw)
}
