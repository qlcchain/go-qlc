package contract

import (
	_ "errors"
	_ "fmt"
	_ "net"
	"reflect"
	_ "strconv"
	_ "strings"
	"testing"

	_ "github.com/qlcchain/go-qlc/common/statedb"
	_ "github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestPtmKeyUpdate_ProcessSend(t *testing.T) {
	type fields struct {
		BaseContract BaseContract
	}
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.PendingKey
		want1   *types.PendingInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pku := &PtmKeyUpdate{
				BaseContract: tt.fields.BaseContract,
			}
			got, got1, err := pku.ProcessSend(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyUpdate.ProcessSend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtmKeyUpdate.ProcessSend() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("PtmKeyUpdate.ProcessSend() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPtmKeyUpdate_SetStorage(t *testing.T) {
	type fields struct {
		BaseContract BaseContract
	}
	type args struct {
		ctx     *vmstore.VMContext
		account types.Address
		VBtype  uint16
		vKey    []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pku := &PtmKeyUpdate{
				BaseContract: tt.fields.BaseContract,
			}
			if err := pku.SetStorage(tt.args.ctx, tt.args.account, tt.args.VBtype, tt.args.vKey); (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyUpdate.SetStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPtmKeyDelete_ProcessSend(t *testing.T) {
	type fields struct {
		BaseContract BaseContract
	}
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.PendingKey
		want1   *types.PendingInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkd := &PtmKeyDelete{
				BaseContract: tt.fields.BaseContract,
			}
			got, got1, err := pkd.ProcessSend(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyDelete.ProcessSend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtmKeyDelete.ProcessSend() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("PtmKeyDelete.ProcessSend() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPtmKeyDelete_SetStorage(t *testing.T) {
	type fields struct {
		BaseContract BaseContract
	}
	type args struct {
		ctx     *vmstore.VMContext
		account types.Address
		vBtype  uint16
		vKey    []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkd := &PtmKeyDelete{
				BaseContract: tt.fields.BaseContract,
			}
			if err := pkd.SetStorage(tt.args.ctx, tt.args.account, tt.args.vBtype, tt.args.vKey); (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyDelete.SetStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
