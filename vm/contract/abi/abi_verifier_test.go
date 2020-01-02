package abi

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"testing"
)

func addTestVerifierInfo(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	data, err := VerifierABI.PackVariable(VariableNameVerifierInfo, vInfo)
	if err != nil {
		return err
	}

	vrKey := append(util.BE_Uint32ToBytes(vType), account[:]...)
	err = ctx.SetStorage(types.VerifierAddress[:], vrKey, data)
	if err != nil {
		return err
	}

	err = ctx.SaveStorage()
	if err != nil {
		return err
	}

	return nil
}

func TestCheckVerifierExist(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := types.OracleTypeEmail
	vi := "123@gmail.com"

	if CheckVerifierExist(ctx, account, vt) {
		t.Fatal()
	}

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckVerifierExist(ctx, account, vt) {
		t.Fatal()
	}
}

func TestCheckVerifierInfoExist(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := types.OracleTypeEmail
	vi := "123@gmail.com"

	if CheckVerifierInfoExist(ctx, account, vt, vi) {
		t.Fatal()
	}

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckVerifierInfoExist(ctx, account, vt, vi) {
		t.Fatal()
	}
}

func TestGetAllVerifiers(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := types.OracleTypeEmail
	vi := "123@gmail.com"

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	account2 := mock.Address()
	err = addTestVerifierInfo(ctx, account2, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	vs, err := GetAllVerifiers(ctx)
	if err != nil || len(vs) != 2 {
		t.Fatal()
	}

	if vs[0].Account != account && vs[0].Account != account2 {
		t.Fatal()
	}

	if vs[1].Account != account && vs[1].Account != account2 {
		t.Fatal()
	}
}

func TestGetVerifiersByType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := types.OracleTypeEmail
	vi := "123@gmail.com"

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	vt2 := types.OracleTypeWeChat
	vi2 := "1234"
	err = addTestVerifierInfo(ctx, account, vt2, vi2)
	if err != nil {
		t.Fatal(err)
	}

	account3 := mock.Address()
	vi3 := "1234123"
	err = addTestVerifierInfo(ctx, account3, vt2, vi3)
	if err != nil {
		t.Fatal(err)
	}

	vs, err := GetVerifiersByType(ctx, types.OracleTypeWeChat)
	if err != nil || len(vs) != 2 {
		t.Fatal()
	}

	if vs[0].VType != types.OracleTypeWeChat || vs[1].VType != types.OracleTypeWeChat {
		t.Fatal()
	}
}

func TestGetVerifiersByAccount(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := types.OracleTypeEmail
	vi := "123@gmail.com"

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	account2 := mock.Address()
	err = addTestVerifierInfo(ctx, account2, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	vs, err := GetVerifiersByAccount(ctx, account2)
	if err != nil || len(vs) != 1 {
		t.Fatal()
	}

	if vs[0].Account != account2 {
		t.Fatal()
	}
}

func TestGetVerifierInfoByAccountAndType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := types.OracleTypeEmail
	vi := "123@gmail.com"

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	vit, err := GetVerifierInfoByAccountAndType(ctx, account, vt)
	if err != nil || vit != vi {
		t.Fatal()
	}
}
