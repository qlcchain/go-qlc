package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUnchecked_Serialize(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
	unchecked := &Unchecked{Block: &b, Kind: Synchronized}
	s, err := unchecked.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
	u2 := new(Unchecked)
	if err := u2.Deserialize(s); err != nil {
		t.Fatal(err)
	}
	t.Log(u2)
	if u2.Block.GetHash() != unchecked.Block.GetHash() || u2.Kind != unchecked.Kind {
		t.Fatal()
	}
}

func TestSynchronizedKind_String(t *testing.T) {
	tests := []struct {
		name string
		s    SynchronizedKind
		want string
	}{
		{
			name: "Synchronized",
			s:    Synchronized,
			want: "sync",
		},
		{
			name: "UnSynchronized",
			s:    UnSynchronized,
			want: "unsync",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToSyncKind(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want SynchronizedKind
	}{
		{
			name: "sync",
			args: args{
				str: "sync",
			},
			want: Synchronized,
		}, {
			name: "unsync",
			args: args{
				str: "unsync",
			},
			want: UnSynchronized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToSyncKind(tt.args.str); got != tt.want {
				t.Errorf("StringToSyncKind() = %v, want %v", got, tt.want)
			}
		})
	}
}
