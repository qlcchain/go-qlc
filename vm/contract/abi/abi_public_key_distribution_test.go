package abi

import (
	"bytes"
	"crypto/ed25519"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestPackAndUnpack(t *testing.T) {
	_, err := abi.JSONToABIContract(strings.NewReader(JsonPublicKeyDistribution))
	if err != nil {
		t.Fatal(err)
	}

	data, err := PublicKeyDistributionABI.PackMethod(MethodNamePKDVerifierRegister, common.OracleTypeWeChat, "123@gmail.com")
	if err != nil {
		t.Fatal(err)
	}

	reg := new(VerifierRegInfo)
	err = PublicKeyDistributionABI.UnpackMethod(reg, MethodNamePKDVerifierRegister, data)
	if err != nil {
		t.Fatal(err)
	}

	if reg.VType != common.OracleTypeWeChat || reg.VInfo != "123@gmail.com" {
		t.Fatal()
	}

	pk := make([]byte, 32)
	verifiers := make([]types.Address, 0)
	codes := make([]types.Hash, 0)

	typ := common.OracleTypeWeChat
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

	data, err = PublicKeyDistributionABI.PackMethod(MethodNamePKDPublish, typ, id, pk, verifiers, codes, fee)
	if err != nil {
		t.Fatal(err)
	}

	var info PublishInfo
	err = PublicKeyDistributionABI.UnpackMethod(&info, MethodNamePKDPublish, data)
	if err != nil {
		t.Fatal(err)
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

func addTestVerifierInfo(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	data, err := PublicKeyDistributionABI.PackVariable(VariableNamePKDVerifierInfo, vInfo, true)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		return err
	}

	err = ctx.SaveStorage()
	if err != nil {
		return err
	}

	return nil
}

func delTestVerifierInfo(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	data, err := PublicKeyDistributionABI.PackVariable(VariableNamePKDVerifierInfo, vInfo, false)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeEmail
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeEmail
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeEmail
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeEmail
	vi := "123@gmail.com"

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	vt2 := common.OracleTypeWeChat
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

	vs, err := GetVerifiersByType(ctx, common.OracleTypeWeChat)
	if err != nil || len(vs) != 2 {
		t.Fatal()
	}

	if vs[0].VType != common.OracleTypeWeChat || vs[1].VType != common.OracleTypeWeChat {
		t.Fatal()
	}
}

func TestGetVerifiersByAccount(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeEmail
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
		t.Fatal(err)
	}

	if vs[0].Account != account2 {
		t.Fatal()
	}
}

func TestGetVerifierInfoByAccountAndType(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeEmail
	vi := "123@gmail.com"

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	vs, err := GetVerifierInfoByAccountAndType(ctx, account, vt)
	if err != nil || vs.VInfo != vi {
		t.Fatal()
	}
}

func TestDeleteVerifier(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeEmail
	vi := "123@gmail.com"

	err := addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckVerifierExist(ctx, account, vt) {
		t.Fatal()
	}

	err = delTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal(err)
	}

	if CheckVerifierExist(ctx, account, vt) {
		t.Fatal()
	}
}

func addTestOracleInfo(ctx *vmstore.VMContext, account types.Address, ot uint32, id types.Hash, pk []byte, code string, hash types.Hash) error {
	data, err := PublicKeyDistributionABI.PackVariable(VariableNamePKDOracleInfo, code)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, PKDStorageTypeOracle)
	key = append(key, util.BE_Uint32ToBytes(ot)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, hash[:]...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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
	ot2 := common.OracleTypeEmail
	id2 := mock.Hash()
	code2 := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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

	ot2 := common.OracleTypeWeChat
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	pblk := mock.StateBlock()
	pblk.Previous = types.ZeroHash
	err := l.AddStateBlock(pblk)
	if err != nil {
		t.Fatal(err)
	}

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	hash := pblk.GetHash()
	pk := make([]byte, ed25519.PublicKeySize)
	err = random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = addTestOracleInfo(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)
	blk := mock.StateBlock()
	blk.Previous = hash
	blk.Data, _ = PublicKeyDistributionABI.PackMethod(MethodNamePKDPublish, ot, id, pk[:], vs, cs, fee.Int)
	err = l.AddStateBlock(blk)
	if err != nil {
		t.Fatal(err)
	}

	info := GetOracleInfoByHash(ctx, hash)
	if len(info) != 1 || info[0].OID != id || info[0].OType != ot || !bytes.Equal(pk, info[0].PubKey) {
		t.Fatal()
	}
}

func addTestPublishInfo(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte,
	vs []types.Address, cs []types.Hash, fee types.Balance, hash types.Hash) error {
	data, err := PublicKeyDistributionABI.PackVariable(VariableNamePKDPublishInfo, account, vs, cs, fee.Int, true)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, hash[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		return err
	}

	err = ctx.SaveStorage()
	if err != nil {
		return err
	}

	return nil
}

func delTestPublishInfo(ctx *vmstore.VMContext, account types.Address, pt uint32, id types.Hash, pk []byte,
	vs []types.Address, cs []types.Hash, fee types.Balance, hash types.Hash) error {
	data, err := PublicKeyDistributionABI.PackVariable(VariableNamePKDPublishInfo, account, vs, cs, fee.Int, false)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, pk...)
	key = append(key, hash[:]...)
	err = ctx.SetStorage(types.PubKeyDistributionAddress[:], key, data)
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	pk := make([]byte, ed25519.PublicKeySize)
	err := PublishInfoCheck(ctx, mock.Address(), common.OracleTypeInvalid, mock.Hash(), pk, common.PublishCost)
	if err == nil {
		t.Fatal()
	}

	pk = make([]byte, ed25519.PublicKeySize+1)
	err = PublishInfoCheck(ctx, mock.Address(), common.OracleTypeEmail, mock.Hash(), pk, common.PublishCost)
	if err == nil {
		t.Fatal()
	}

	pk = make([]byte, ed25519.PublicKeySize)
	err = PublishInfoCheck(ctx, mock.Address(), common.OracleTypeEmail, mock.Hash(), pk, types.NewBalance(3e8))
	if err == nil {
		t.Fatal()
	}

	err = PublishInfoCheck(ctx, mock.Address(), common.OracleTypeEmail, mock.Hash(), pk, common.PublishCost)
	if err != nil {
		t.Fatal()
	}

	err = PublishInfoCheck(ctx, mock.Address(), common.OracleTypeEmail, mock.Hash(), pk, types.NewBalance(10e8))
	if err != nil {
		t.Fatal()
	}
}

func TestCheckPublishInfoExist(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	if CheckPublishInfoExist(ctx, account, pt, id, pk, hash) {
		t.Fatal()
	}

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckPublishInfoExist(ctx, account, pt, id, pk, hash) {
		t.Fatal()
	}
}

func TestGetPublishInfoByTypeAndId(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	id2 := mock.Hash()
	err = addTestPublishInfo(ctx, account, pt, id2, pk, vs, cs, fee, hash)
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	pt2 := common.OracleTypeWeChat
	id2 := mock.Hash()
	err = addTestPublishInfo(ctx, account, pt2, id2, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetAllPublishInfo(ctx)
	if len(info) != 2 {
		t.Fatal()
	}

	switch info[0].PType {
	case common.OracleTypeEmail:
		if info[0].PID != id {
			t.Fatal()
		}
	case common.OracleTypeWeChat:
		if info[0].PID != id2 {
			t.Fatal()
		}
	default:
		t.Fatal()
	}

	switch info[1].PType {
	case common.OracleTypeEmail:
		if info[1].PID != id {
			t.Fatal()
		}
	case common.OracleTypeWeChat:
		if info[1].PID != id2 {
			t.Fatal()
		}
	default:
		t.Fatal()
	}
}

func TestGetPublishInfoByType(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	pt2 := common.OracleTypeWeChat
	id2 := mock.Hash()
	err = addTestPublishInfo(ctx, account, pt2, id2, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfoByType(ctx, common.OracleTypeWeChat)
	if len(info) != 1 {
		t.Fatal()
	}

	if info[0].PType != common.OracleTypeWeChat {
		t.Fatal()
	}
}

func TestGetPublishInfoByAccount(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	id2 := mock.Hash()
	account2 := mock.Address()
	err = addTestPublishInfo(ctx, account2, pt, id2, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	id3 := mock.Hash()
	err = addTestPublishInfo(ctx, account2, pt, id3, pk, vs, cs, fee, hash)
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	id2 := mock.Hash()
	pt2 := common.OracleTypeWeChat
	account2 := mock.Address()
	err = addTestPublishInfo(ctx, account2, pt2, id2, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	id3 := mock.Hash()
	pt3 := common.OracleTypeEmail
	err = addTestPublishInfo(ctx, account2, pt3, id3, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfoByAccountAndType(ctx, account2, common.OracleTypeEmail)
	if len(info) != 1 {
		t.Fatal()
	}

	if info[0].Account != account2 || info[0].PType != common.OracleTypeEmail {
		t.Fatal()
	}
}

func TestDeletePublishInfo(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckPublishInfoExist(ctx, account, pt, id, pk, hash) {
		t.Fatal(err)
	}

	err = delTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	if CheckPublishInfoExist(ctx, account, pt, id, pk, hash) {
		t.Fatal()
	}
}

func TestVerifierRegInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeInvalid
	vi := "123@test.com"

	err := VerifierRegInfoCheck(ctx, account, vt, vi)
	if err == nil {
		t.Fatal()
	}

	vt = common.OracleTypeEmail
	err = addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal()
	}

	err = VerifierRegInfoCheck(ctx, account, vt, vi)
	if err == nil {
		t.Fatal()
	}

	vt = common.OracleTypeWeChat
	vi = "123456"
	err = VerifierRegInfoCheck(ctx, account, vt, vi)
	if err != nil {
		t.Fatal()
	}

	err = addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal()
	}

	err = VerifierRegInfoCheck(ctx, account, vt, vi)
	if err == nil {
		t.Fatal()
	}
}

func TestVerifierUnRegInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	vt := common.OracleTypeInvalid

	err := VerifierUnRegInfoCheck(ctx, account, vt)
	if err == nil {
		t.Fatal()
	}

	vt = common.OracleTypeWeChat
	err = VerifierUnRegInfoCheck(ctx, account, vt)
	if err == nil {
		t.Fatal()
	}

	vi := "123456"
	err = addTestVerifierInfo(ctx, account, vt, vi)
	if err != nil {
		t.Fatal()
	}

	err = VerifierUnRegInfoCheck(ctx, account, vt)
	if err != nil {
		t.Fatal()
	}
}

func TestVerifierPledgeCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()

	err := VerifierPledgeCheck(ctx, account)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(account)
	am.CoinOracle = common.MinVerifierPledgeAmount
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	err = VerifierPledgeCheck(ctx, account)
	if err != nil {
		t.Fatal()
	}
}

func TestOracleInfoCheck(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeInvalid
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize+1)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	err = OracleInfoCheck(ctx, account, ot, id, pk, code, hash)
	if err == nil {
		t.Fatal()
	}

	ot = common.OracleTypeEmail
	err = OracleInfoCheck(ctx, account, ot, id, pk, code, hash)
	if err == nil {
		t.Fatal()
	}

	pk = pk[:32]
	err = OracleInfoCheck(ctx, account, ot, id, pk, code, hash)
	if err == nil {
		t.Fatal()
	}

	codeComb := append([]byte(types.NewHexBytesFromData(pk).String()), []byte(code)...)
	codeHash, err := types.Sha256HashData(codeComb)
	if err != nil {
		t.Fatal(err)
	}

	account1 := mock.Address()
	code1 := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	vs := []types.Address{account1}
	cs := []types.Hash{codeHash}
	err = addTestPublishInfo(ctx, account, ot, id, pk, vs, cs, common.PublishCost, hash)
	if err != nil {
		t.Fatal()
	}

	err = OracleInfoCheck(ctx, account, ot, id, pk, code1, hash)
	if err == nil {
		t.Fatal()
	}

	vs = []types.Address{account}
	err = addTestPublishInfo(ctx, account, ot, id, pk, vs, cs, common.PublishCost, hash)
	if err != nil {
		t.Fatal()
	}

	err = OracleInfoCheck(ctx, account, ot, id, pk, code1, hash)
	if err == nil {
		t.Fatal()
	}

	err = OracleInfoCheck(ctx, account, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckOracleInfoExist(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
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

	if !CheckOracleInfoExist(ctx, account, ot, id, pk, hash) {
		t.Fatal()
	}
}

func TestGetPublishInfoByKey(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfoByKey(ctx, pt, id, pk, hash)
	if info == nil {
		t.Fatal()
	}
}

func TestGetPublishInfo(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	pt := common.OracleTypeEmail
	id := mock.Hash()
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	err := random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost

	err = addTestPublishInfo(ctx, account, pt, id, pk, vs, cs, fee, hash)
	if err != nil {
		t.Fatal(err)
	}

	info := GetPublishInfo(ctx, pt, id, pk, hash)
	if info == nil {
		t.Fatal()
	}
}
