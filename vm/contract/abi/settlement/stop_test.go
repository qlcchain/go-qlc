/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"reflect"
	"testing"
)

func TestStopParam_Verify(t *testing.T) {
	type fields struct {
		StopName string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				StopName: "test1",
			},
			wantErr: false,
		}, {
			name: "false",
			fields: fields{
				StopName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &StopParam{
				StopName: tt.fields.StopName,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateStopParam_Verify(t *testing.T) {
	type fields struct {
		StopName string
		NewName  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				StopName: "test1",
				NewName:  "test2",
			},
			wantErr: false,
		}, {
			name: "false",
			fields: fields{
				StopName: "",
				NewName:  "222",
			},
			wantErr: true,
		}, {
			name: "false2",
			fields: fields{
				StopName: "111",
				NewName:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := &UpdateStopParam{
				StopName: tt.fields.StopName,
				New:      tt.fields.NewName,
			}
			if err := z.Verify(); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStopParam_ToABI(t *testing.T) {
	param := StopParam{
		StopName: "test",
	}
	data, err := param.ToABI(MethodNameAddNextStop)
	if err != nil {
		t.Fatal(err)
	}

	p2 := &StopParam{}
	if err = p2.FromABI(MethodNameAddNextStop, data); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid param, %v, %v", &param, p2)
		} else {
			if err := p2.Verify(); err != nil {
				t.Fatalf("verify failed, %s", err)
			}
		}
	}
}

func TestUpdateStopParam_ToABI(t *testing.T) {
	param := UpdateStopParam{
		StopName: "test",
		New:      "hahah",
	}
	data, err := param.ToABI(MethodNameUpdateNextStop)
	if err != nil {
		t.Fatal(err)
	}

	p2 := &UpdateStopParam{}
	if err = p2.FromABI(MethodNameUpdateNextStop, data); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(&param, p2) {
			t.Fatalf("invalid param, %v, %v", &param, p2)
		} else {
			if err := p2.Verify(); err != nil {
				t.Fatalf("verify failed, %s", err)
			}
		}
	}
}
