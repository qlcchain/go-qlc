package abi

import (
	_ "errors"
	"reflect"
	"testing"

	_ "github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	_ "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestPtmKeyInfoCheck(t *testing.T) {
	type args struct {
		ctx *vmstore.VMContext
		pt  uint16
		pk  []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PtmKeyInfoCheck(tt.args.ctx, tt.args.pt, tt.args.pk); (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyInfoCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPtmKeyByAccountAndBtype(t *testing.T) {
	type args struct {
		ctx     *vmstore.VMContext
		account types.Address
		vBtype  uint16
	}
	tests := []struct {
		name    string
		args    args
		want    []*PtmKeyInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPtmKeyByAccountAndBtype(tt.args.ctx, tt.args.account, tt.args.vBtype)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPtmKeyByAccountAndBtype() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPtmKeyByAccountAndBtype() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPtmKeyByAccount(t *testing.T) {
	type args struct {
		ctx     *vmstore.VMContext
		account types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*PtmKeyInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPtmKeyByAccount(tt.args.ctx, tt.args.account)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPtmKeyByAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPtmKeyByAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
