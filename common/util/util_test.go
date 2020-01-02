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

func TestVerifyEmailFormat(t *testing.T) {
	e1 := "11@qq.com"
	e2 := "2222@com"
	e3 := " abc.d@qlink.online"

	if !VerifyEmailFormat(e1) {
		t.Fatal()
	}

	if VerifyEmailFormat(e2) {
		t.Fatal()
	}

	if !VerifyEmailFormat(e3) {
		t.Fatal()
	}
}
