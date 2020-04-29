package abi

import (
	_ "errors"
	"reflect"
	"testing"

	"github.com/qlcchain/go-qlc/common"
	_ "github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	_ "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
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
		{"OK", args{nil, common.PtmKeyVBtypeDefault, []byte("/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M=")}, false},
		{"bad pklen", args{nil, common.PtmKeyVBtypeDefault, []byte("/vkgO5TfnsvKZGDc2KT1yxD5fx")}, true},
		{"bad ptype", args{nil, 2, []byte("/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M=")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PtmKeyInfoCheck(tt.args.ctx, tt.args.pt, tt.args.pk); (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyInfoCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func addTestPtmKey(store ledger.Store, ctx *vmstore.VMContext, account types.Address, vBtype uint16, vKey []byte) error {
	data, err := PtmKeyABI.PackVariable(VariableNamePtmKeyStorageVar, string(vKey[:]), true)
	if err != nil {
		//fmt.Printf("SetStorage:PackVariable err(%s)\n", err)
		return err
	}

	var key []byte
	key = append(key, account[:]...)
	key = append(key, util.BE_Uint16ToBytes(vBtype)...)
	//fmt.Printf("SetStorage:get key(%s) data(%s)\n", string(key[:]), data)
	err = ctx.SetStorage(contractaddress.PtmKeyKVAddress[:], key, data)
	if err != nil {
		//fmt.Printf("SetStorage:err(%s)\n", err)
		return err
	}

	err = store.SaveStorage(vmstore.ToCache(ctx))
	if err != nil {
		return err
	}

	return nil
}
func TestGetPtmKeyByAccountAndBtype(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)
	account := mock.Address()
	account2 := mock.Address()
	pks := make([]*PtmKeyInfo, 0)
	btype := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pk := &PtmKeyInfo{
		account,
		btype,
		key,
	}
	pks = append(pks, pk)

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
		{"OK", args{ctx, account, common.PtmKeyVBtypeDefault}, pks, false},
		{"badaccount", args{ctx, account2, common.PtmKeyVBtypeDefault}, nil, true},
	}
	addTestPtmKey(l, ctx, account, btype, []byte(key))
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)
	account := mock.Address()
	account2 := mock.Address()
	pks := make([]*PtmKeyInfo, 0)
	btype := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0D="
	pk := &PtmKeyInfo{
		account,
		btype,
		key,
	}
	pks = append(pks, pk)
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
		{"OK", args{ctx, account}, pks, false},
		{"badaccount", args{ctx, account2}, nil, true},
	}
	addTestPtmKey(l, ctx, account, btype, []byte(key))
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
