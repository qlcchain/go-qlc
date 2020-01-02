package abi

import (
	"bytes"
	"crypto/ed25519"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"testing"
	"time"
)

func addTestOracleInfo(ctx *vmstore.VMContext, account types.Address, ot uint32, id types.Hash, pk []byte, code string, hash types.Hash) error {
	data, err := OracleABI.PackVariable(VariableNameOracleInfo, code, hash)
	if err != nil {
		return err
	}

	key := append(util.BE_Uint32ToBytes(ot), id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.OracleAddress[:], key, data)
	if err != nil {
		return err
	}

	err = ctx.SaveStorage()
	if err != nil {
		return err
	}

	return nil
}

func TestGetAllOracleInfo(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := types.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(types.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	account2 := mock.Address()
	ot2 := types.OracleTypeEmail
	id2 := mock.Hash()
	code2 := util.RandomFixedStringWithSeed(types.RandomCodeLen, time.Now().UnixNano())
	hash2 := mock.Hash()
	pk2 := make([]byte, ed25519.PublicKeySize)
	err = random.Bytes(pk2)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account2, ot2, id2, pk2, code2, hash2)
	if err != nil {
		t.Fatal(err)
	}

	info := GetAllOracleInfo(ctx)
	if len(info) != 2 {
		t.Fatal()
	}
}

func TestGetOracleInfoByAccount(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := types.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(types.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	account2 := mock.Address()
	err = addTestOracleInfo(ctx, account2, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetOracleInfoByAccount(ctx, account2)
	if len(info) != 1 || info[0].Account != account2 {
		t.Fatal()
	}
}

func TestGetOracleInfoByType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := types.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(types.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	ot2 := types.OracleTypeWeChat
	err = addTestOracleInfo(ctx, account, ot2, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetOracleInfoByType(ctx, ot2)
	if len(info) != 1 || info[0].OType != ot2 {
		t.Fatal()
	}
}

func TestGetOracleInfoByAccountAndType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := types.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(types.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetOracleInfoByAccountAndType(ctx, account, ot)
	if len(info) != 1 || info[0].Account != account || info[0].OType != ot {
		t.Fatal()
	}
}

func TestGetOracleInfoByTypeAndID(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := types.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(types.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetOracleInfoByTypeAndID(ctx, ot, id)
	if len(info) != 1 || info[0].OID != id || info[0].OType != ot {
		t.Fatal()
	}
}

func TestGetOracleInfoByTypeAndIDAndPk(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := types.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(types.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetOracleInfoByTypeAndIDAndPk(ctx, ot, id, pk)
	if len(info) != 1 || info[0].OID != id || info[0].OType != ot || !bytes.Equal(pk, info[0].PubKey) {
		t.Fatal()
	}
}

func TestGetOracleInfoByHash(t *testing.T) {
	hash1 := types.Hash{1, 2, 3}
	hash2 := types.Hash{1, 2, 3}

	if hash1 != hash2 {
		t.Fatal()
	}

	hash3 := types.Hash{1, 3, 2}
	if hash1 == hash3 {
		t.Fatal()
	}
}
