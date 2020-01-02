package abi

import (
	"bytes"
	"crypto/ed25519"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
	"testing"
)

func addTestPublishInfo(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte,
	vs []types.Address, cs []types.Hash, fee types.Balance) error {
	data, err := PublisherABI.PackVariable(VariableNamePublishInfo, vs, cs, fee.Int)
	if err != nil {
		return err
	}

	key := append(util.BE_Uint32ToBytes(pt), id[:]...)
	key = append(key, pk...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PublisherAddress[:], key, data)
	if err != nil {
		return err
	}

	err = ctx.SaveStorage()
	if err != nil {
		return err
	}

	return nil
}

func TestPublishInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	pk := make([]byte, ed25519.PublicKeySize)
	err := PublishInfoCheck(ctx, mock.Address(), types.OracleTypeInvalid, mock.Hash(), pk, types.NewBalance(5e8))
	if err == nil {
		t.Fatal()
	}

	pk = make([]byte, ed25519.PublicKeySize+1)
	err = PublishInfoCheck(ctx, mock.Address(), types.OracleTypeEmail, mock.Hash(), pk, types.NewBalance(5e8))
	if err == nil {
		t.Fatal()
	}

	pk = make([]byte, ed25519.PublicKeySize)
	err = PublishInfoCheck(ctx, mock.Address(), types.OracleTypeEmail, mock.Hash(), pk, types.NewBalance(3e8))
	if err == nil {
		t.Fatal()
	}

	err = PublishInfoCheck(ctx, mock.Address(), types.OracleTypeEmail, mock.Hash(), pk, types.NewBalance(5e8))
	if err != nil {
		t.Fatal()
	}

	err = PublishInfoCheck(ctx, mock.Address(), types.OracleTypeEmail, mock.Hash(), pk, types.NewBalance(10e8))
	if err != nil {
		t.Fatal()
	}
}

func TestPublisherPackAndUnPack(t *testing.T) {
	pk := make([]byte, 32)
	verifiers := make([]types.Address, 0)
	codes := make([]types.Hash, 0)

	acc := mock.Address()
	typ := uint32(1)
	id := mock.Hash()
	_ = random.Bytes(pk)
	fee := big.NewInt(5e8)

	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())

	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())

	data, err := PublisherABI.PackMethod(MethodNamePublish, acc, typ, id, pk, verifiers, codes, fee)
	if err != nil {
		t.Fatal(err)
	}

	var info PublishInfo
	err = PublisherABI.UnpackMethod(&info, MethodNamePublish, data)
	if err != nil {
		t.Fatal(err)
	}

	if info.Account != acc {
		t.Fatal()
	}

	if info.PID != id {
		t.Fatal()
	}

	if info.PType != typ {
		t.Fatal()
	}

	if !bytes.Equal(info.PubKey, pk) {
		t.Fatal()
	}

	if info.Fee.Cmp(fee) != 0 {
		t.Fatal()
	}

	for i, v := range info.Verifiers {
		if v != verifiers[i] {
			t.Fatal()
		}
	}

	for i, c := range info.Codes {
		if c != codes[i] {
			t.Fatal()
		}
	}
}

func TestCheckPublishKeyRegistered(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	err = CheckPublishKeyRegistered(ctx, account, pt, id, pk)
	if err == nil {
		t.Fatal()
	}
}

func TestCheckPublishInfoExist(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	if CheckPublishInfoExist(ctx, account, pt, id) {
		t.Fatal()
	}

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckPublishInfoExist(ctx, account, pt, id) {
		t.Fatal()
	}
}

func TestGetPublishInfoByTypeAndId(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	id2 := mock.Hash()
	err = addTestPublishInfo(ctx, account, pt, id2, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfoByTypeAndId(ctx, pt, id)
	if len(info) != 1 {
		t.Fatal()
	}

	if info[0].PID != id || info[0].PType != pt {
		t.Fatal()
	}

	info2 := GetPublishInfoByTypeAndId(ctx, pt, id2)
	if len(info) != 1 {
		t.Fatal()
	}

	if info2[0].PID != id2 || info2[0].PType != pt {
		t.Fatal()
	}
}

func TestGetAllPublishInfo(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	pt2 := types.OracleTypeWeChat
	id2 := mock.Hash()
	err = addTestPublishInfo(ctx, account, pt2, id2, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	info := GetAllPublishInfo(ctx)
	if len(info) != 2 {
		t.Fatal()
	}

	switch info[0].PType {
	case types.OracleTypeEmail:
		if info[0].PID != id {
			t.Fatal()
		}
	case types.OracleTypeWeChat:
		if info[0].PID != id2 {
			t.Fatal()
		}
	default:
		t.Fatal()
	}

	switch info[1].PType {
	case types.OracleTypeEmail:
		if info[1].PID != id {
			t.Fatal()
		}
	case types.OracleTypeWeChat:
		if info[1].PID != id2 {
			t.Fatal()
		}
	default:
		t.Fatal()
	}
}

func TestGetPublishInfoByType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	pt2 := types.OracleTypeWeChat
	id2 := mock.Hash()
	err = addTestPublishInfo(ctx, account, pt2, id2, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfoByType(ctx, types.OracleTypeWeChat)
	if len(info) != 1 {
		t.Fatal()
	}

	if info[0].PType != types.OracleTypeWeChat {
		t.Fatal()
	}
}

func TestGetPublishInfoByAccount(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	id2 := mock.Hash()
	account2 := mock.Address()
	err = addTestPublishInfo(ctx, account2, pt, id2, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	id3 := mock.Hash()
	err = addTestPublishInfo(ctx, account2, pt, id3, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfoByAccount(ctx, account2)
	if len(info) != 2 {
		t.Fatal()
	}

	if info[0].Account != account2 || info[1].Account != account2 {
		t.Fatal()
	}
}

func TestGetPublishInfoByAccountAndType(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	id2 := mock.Hash()
	pt2 := types.OracleTypeWeChat
	account2 := mock.Address()
	err = addTestPublishInfo(ctx, account2, pt2, id2, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	id3 := mock.Hash()
	pt3 := types.OracleTypeEmail
	err = addTestPublishInfo(ctx, account2, pt3, id3, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfoByAccountAndType(ctx, account2, types.OracleTypeEmail)
	if len(info) != 1 {
		t.Fatal()
	}

	if info[0].Account != account2 || info[0].PType != types.OracleTypeEmail {
		t.Fatal()
	}
}

func TestGetPublishKeyByAccountAndTypeAndID(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := types.OracleTypeEmail
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee)
	if err != nil {
		t.Fatal(err)
	}

	pubKey := GetPublishKeyByAccountAndTypeAndID(ctx, account, pt, id)
	if !bytes.Equal(pubKey, pk) {
		t.Fatal()
	}
}
