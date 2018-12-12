/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"testing"
)

func TestTaskLifecycle(t *testing.T) {
	tl := &TaskLifecycle{}
	t.Log(tl.Status, tl.String())
	b := tl.PreInit()
	if !b {
		t.Fatal("PreInit error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PostInit()
	if !b {
		t.Fatal("PostInit error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PreStart()
	if !b {
		t.Fatal("PreStart error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PostStart()
	if !b {
		t.Fatal("PostStart error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PreStop()
	if !b {
		t.Fatal("PreStop error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PostStop()
	if !b {
		t.Fatal("PostStop error")
	}
	t.Log(tl.Status, tl.String())
}
