package contract

import (
	_ "errors"
	_ "fmt"
	_ "net"
	"reflect"
	_ "strconv"
	_ "strings"
	"testing"

	"github.com/qlcchain/go-qlc/common"
	_ "github.com/qlcchain/go-qlc/common/statedb"
	_ "github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestPtmKeyUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)
	blk := mock.StateBlockWithoutWork()
	pbc := new(PtmKeyUpdate)
	btype := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hygbb="
	btype2 := common.PtmKeyVBtypeInvaild
	key2 := "/vkgO5TfnsvKe65SPPuh3hygbb="

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
		{"badtoken", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, true},
		{"baddata", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, true},
		{"badbtype", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, true},
		{"badvkey", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, true},
		{"OK", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, false},
	}
	step := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pku := &PtmKeyUpdate{
				BaseContract: tt.fields.BaseContract,
			}
			switch step {
			case 1:
				blk.Token = mock.Hash()
				blk.Token = cfg.ChainToken()
			case 2:
				blk.Data, _ = abi.PtmKeyABI.PackMethod(abi.MethodNamePtmKeyUpdate, btype2, key)
			case 3:
				blk.Data, _ = abi.PtmKeyABI.PackMethod(abi.MethodNamePtmKeyUpdate, btype, key2)
			case 4:
				blk.Data, _ = abi.PtmKeyABI.PackMethod(abi.MethodNamePtmKeyUpdate, btype, key)
			}
			step = step + 1
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
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()
	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)
	account := mock.Address()
	pbc := new(PtmKeyUpdate)
	btype := common.PtmKeyVBtypeDefault
	btype2 := common.PtmKeyVBtypeInvaild
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hygaa="
	key2 := "/vkgO5TfnsvKZGDc2Kaa"

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
		{"OK", fields{pbc.BaseContract}, args{ctx, account, btype, []byte(key)}, false},
		{"badbyte", fields{pbc.BaseContract}, args{ctx, account, btype2, []byte(key)}, true},
		{"badkey", fields{pbc.BaseContract}, args{ctx, account, btype, []byte(key2)}, true},
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
func addTestPtmKeyStorage(store ledger.Store, ctx *vmstore.VMContext, account types.Address, vBtype uint16, vKey []byte) error {
	data, err := abi.PtmKeyABI.PackVariable(abi.VariableNamePtmKeyStorageVar, string(vKey[:]), true)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, account[:]...)
	key = append(key, util.BE_Uint16ToBytes(vBtype)...)
	err = ctx.SetStorage(contractaddress.PtmKeyKVAddress[:], key, data)
	if err != nil {
		return err
	}

	err = store.SaveStorage(vmstore.ToCache(ctx))
	if err != nil {
		return err
	}

	return nil
}
func TestPtmKeyDelete_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)
	account := mock.Address()
	blk := mock.StateBlockWithoutWork()
	blk.Address = account
	pbc := new(PtmKeyDelete)
	btype := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hygdd="

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
		{"badtoken", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, true},
		{"baddata", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, true},
		{"nofound", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, true},
		{"OK", fields{pbc.BaseContract}, args{ctx, blk}, nil, nil, false},
	}
	step := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkd := &PtmKeyDelete{
				BaseContract: tt.fields.BaseContract,
			}
			switch step {
			case 1:
				blk.Token = mock.Hash()
				blk.Token = cfg.ChainToken()
			case 2:
				blk.Data, _ = abi.PtmKeyABI.PackMethod(abi.MethodNamePtmKeyDelete, btype)
			case 3:
				addTestPtmKeyStorage(l, ctx, account, btype, []byte(key))
			}
			step = step + 1
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
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()
	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)
	account := mock.Address()
	pbc := new(PtmKeyDelete)
	btype := common.PtmKeyVBtypeDefault
	btype2 := common.PtmKeyVBtypeInvaild
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	key2 := "/vkgO5TfnsvKZGDc2K"

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
		{"OK", fields{pbc.BaseContract}, args{ctx, account, btype, []byte(key)}, false},
		{"badbyte", fields{pbc.BaseContract}, args{ctx, account, btype2, []byte(key)}, true},
		{"badkey", fields{pbc.BaseContract}, args{ctx, account, btype, []byte(key2)}, true},
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
