/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestVmLogs(t *testing.T) {
	h1 := Hash{}
	_ = random.Bytes(h1[:])
	d1 := make([]byte, 10)
	_ = random.Bytes(d1)

	h2 := Hash{}
	_ = random.Bytes(h2[:])
	d2 := make([]byte, 11)
	_ = random.Bytes(d2)
	logs := VmLogs{[]*VmLog{{Topics: []Hash{h1}, Data: d1}, {Topics: []Hash{h2}, Data: d2}}}
	//t.Log(util.ToIndentString(logs))

	buff, err := logs.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	newLogs := new(VmLogs)
	_, err = newLogs.UnmarshalMsg(buff)
	if err != nil {
		t.Fatal(err)
	}

	//t.Log(util.ToIndentString(newLogs))

	if len(logs.Logs) != len(newLogs.Logs) {
		t.Fatal("invalid len", len(logs.Logs), len(newLogs.Logs))
	}

	for i := 0; i < len(logs.Logs); i++ {
		l1 := logs.Logs[i]
		l2 := newLogs.Logs[i]
		if !bytes.EqualFold(l1.Data, l2.Data) {
			t.Fatal("invalid data")
		}

		if len(l1.Topics) != len(l2.Topics) {
			t.Fatal("invalid topic len", len(l1.Topics), len(l2.Topics))
		}

		for j := 0; j < len(l1.Topics); j++ {
			t1 := l1.Topics[j]
			t2 := l2.Topics[j]
			if t1 != t2 {
				t.Fatal("invalid topic", i, j, t1.String(), t2.String())
			}
		}
	}
}
