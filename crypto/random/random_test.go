/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package random

import (
	"testing"
)

func TestBytes(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test1", args{data: make([]byte, 32)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Bytes(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Bytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIntn(t *testing.T) {
	i, err := Intn(100)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(i)
	if i >= 100 {
		t.Fatal("err")
	}
}

func TestPerm(t *testing.T) {
	ints, err := Perm(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(ints) != 10 {
		t.Fatal("wrong len")
	}
	for index, val := range ints {
		if val > 10 {
			t.Fatal("wrong val", index, "=>", val)
		}
		t.Log(index, "->", val)
	}
}
