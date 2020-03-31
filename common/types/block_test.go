/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

func Test_parseString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want BlockType
	}{
		{
			name: "State",
			args: args{
				s: "state",
			},
			want: State,
		}, {
			name: "Send",
			args: args{
				s: "send",
			},
			want: Send,
		}, {
			name: "Receive",
			args: args{
				s: "receive",
			},
			want: Receive,
		}, {
			name: "Open",
			args: args{
				s: "open",
			},
			want: Open,
		}, {
			name: "Change",
			args: args{
				s: "change",
			},
			want: Change,
		}, {
			name: "ContractReward",
			args: args{
				s: "contractreward",
			},
			want: ContractReward,
		}, {
			name: "ContractSend",
			args: args{
				s: "contractsend",
			},
			want: ContractSend,
		}, {
			name: "ContractRefund",
			args: args{
				s: "contractrefund",
			},
			want: ContractRefund,
		}, {
			name: "ContractError",
			args: args{
				s: "contracterror",
			},
			want: ContractError,
		}, {
			name: "SmartContract",
			args: args{
				s: "smartcontract",
			},
			want: SmartContract,
		}, {
			name: "Online",
			args: args{
				s: "online",
			},
			want: Online,
		}, {
			name: "Invalid",
			args: args{
				s: "ada",
			},
			want: Invalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseString(tt.args.s); got != tt.want {
				t.Errorf("parseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockType_String(t *testing.T) {
	tests := []struct {
		name string
		t    BlockType
		want string
	}{
		{
			name: "State",
			t:    State,
			want: "State",
		}, {
			name: "Send",
			t:    Send,
			want: "Send",
		}, {
			name: "Receive",
			t:    Receive,
			want: "Receive",
		}, {
			name: "Change",
			t:    Change,
			want: "Change",
		}, {
			name: "Open",
			t:    Open,
			want: "Open",
		}, {
			name: "ContractReward",
			t:    ContractReward,
			want: "ContractReward",
		}, {
			name: "ContractSend",
			t:    ContractSend,
			want: "ContractSend",
		}, {
			name: "ContractRefund",
			t:    ContractRefund,
			want: "ContractRefund",
		}, {
			name: "ContractError",
			t:    ContractError,
			want: "ContractError",
		}, {
			name: "SmartContract",
			t:    SmartContract,
			want: "SmartContract",
		}, {
			name: "Online",
			t:    Online,
			want: "Online",
		}, {
			name: "<invalid>",
			t:    20,
			want: "<invalid>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockType_Equal(t *testing.T) {
	type args struct {
		t BlockType
	}
	tests := []struct {
		name string
		e    BlockType
		args args
		want bool
	}{
		{
			name: "",
			e:    State,
			args: args{
				t: 0,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Equal(tt.args.t); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockType_UnmarshalJSON(t *testing.T) {
	s := ContractSend
	if data, err := s.MarshalJSON(); err != nil {
		t.Fatal(err)
	} else {
		s1 := BlockType(100)
		if err := s1.UnmarshalJSON(data); err != nil {
			t.Fatal(err)
		} else {
			if !s1.Equal(s) {
				t.Fatalf("exp: %v, act: %v", s, s1)
			}
		}
	}
}

func TestNewBlock(t *testing.T) {
	type args struct {
		t BlockType
	}
	tests := []struct {
		name    string
		args    args
		want    Block
		wantErr bool
	}{
		{
			name: "State",
			args: args{
				t: State,
			},
			want:    &StateBlock{Type: State},
			wantErr: false,
		}, {
			name: "ContractSend",
			args: args{
				t: ContractSend,
			},
			want:    &StateBlock{Type: ContractSend},
			wantErr: false,
		}, {
			name: "ContractRefund",
			args: args{
				t: ContractRefund,
			},
			want:    &StateBlock{Type: ContractRefund},
			wantErr: false,
		}, {
			name: "ContractReward",
			args: args{
				t: ContractReward,
			},
			want:    &StateBlock{Type: ContractReward},
			wantErr: false,
		}, {
			name: "ContractError",
			args: args{
				t: ContractError,
			},
			want:    &StateBlock{Type: ContractError},
			wantErr: false,
		}, {
			name: "Send",
			args: args{
				t: Send,
			},
			want:    &StateBlock{Type: Send},
			wantErr: false,
		}, {
			name: "Receive",
			args: args{
				t: Receive,
			},
			want:    &StateBlock{Type: Receive},
			wantErr: false,
		}, {
			name: "Change",
			args: args{
				t: Change,
			},
			want:    &StateBlock{Type: Change},
			wantErr: false,
		}, {
			name: "Open",
			args: args{
				t: Open,
			},
			want:    &StateBlock{Type: Open},
			wantErr: false,
		}, {
			name: "Online",
			args: args{
				t: Online,
			},
			want:    &StateBlock{Type: Online},
			wantErr: false,
		}, {
			name: "SmartContract",
			args: args{
				t: SmartContract,
			},
			want:    &SmartContractBlock{},
			wantErr: false,
		}, {
			name: "invalid",
			args: args{
				t: BlockType(100),
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "invalid",
			args: args{
				t: Invalid,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBlock(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateBlock_RelatedAddress(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	b.Type = Send
	data, _ := hex.DecodeString("7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3")
	addr, _ := BytesToAddress(data)
	if address := b.ContractAddress(); address == nil {
		t.Fatal("")
	} else {
		if !bytes.EqualFold(address[:], addr[:]) {
			t.Fatalf("exp:%s ,act: %s", addr, address)
		} else {
			t.Log(address)
		}
	}

	b.Type = Receive
	addr2, _ := HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if address := b.ContractAddress(); address == nil {
		t.Fatal("")
	} else {
		if !bytes.EqualFold(address[:], addr2[:]) {
			t.Fatalf("exp:%s ,act: %s", addr2, address)
		} else {
			t.Log(address)
		}
	}
}
