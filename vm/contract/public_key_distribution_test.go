package contract

import (
	"crypto/ed25519"
	"math/big"
	"testing"
	"time"

	cfg "github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/contract/dpki"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestVerifierRegister(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	vr := new(VerifierRegister)
	blk := mock.StateBlock()

	blk.Token = types.ZeroHash
	_, _, err := vr.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal()
	}

	blk.Token = cfg.ChainToken()
	blk.Data = []byte("test")
	_, _, err = vr.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal()
	}

	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierRegister, common.OracleTypeEmail, "test@126.com")
	_, _, err = vr.ProcessSend(ctx, blk)
	if err != ErrNotEnoughPledge {
		t.Fatal()
	}

	am := mock.AccountMeta(blk.Address)
	am.CoinOracle = common.MinVerifierPledgeAmount
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, _, err = vr.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierRegister, common.OracleTypeInvalid, "test@126.com")
	_, _, err = vr.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal()
	}
}

func TestVerifierUnregister(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	vu := new(VerifierUnregister)
	blk := mock.StateBlock()

	blk.Token = types.ZeroHash
	_, _, err := vu.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal()
	}

	blk.Token = cfg.ChainToken()
	blk.Data = []byte("test")
	_, _, err = vu.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierUnregister, common.OracleTypeInvalid)
	_, _, err = vu.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal()
	}

	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierUnregister, common.OracleTypeEmail)
	vr := new(VerifierRegister)
	vr.SetStorage(ctx, blk.Address, common.OracleTypeEmail, "test@126.com")
	_, _, err = vu.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal()
	}
}

func TestPublish(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	p := new(Publish)
	blk := mock.StateBlock()

	blk.Token = types.ZeroHash
	_, _, err := p.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal()
	}

	blk.Token = cfg.GasToken()
	blk.Data = []byte("test")
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	pt := common.OracleTypeInvalid
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	vs := make([]types.Address, 0)
	cs := make([]types.Hash, 0)
	fee := big.NewInt(5e8)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, pk, vs, cs, fee)
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrVerifierNum {
		t.Fatal(err)
	}

	vs = append(vs, mock.Address())
	cs = append(cs, mock.Hash())
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, pk, vs, cs, fee)
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	pt = common.OracleTypeEmail
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, pk, vs, cs, fee)
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrCalcAmount {
		t.Fatal(err)
	}

	blk.Type = types.Open
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrNotEnoughFee {
		t.Fatal(err)
	}

	blk.Balance = types.NewBalance(5e8)
	_, _, err = p.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnPublish(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	up := new(UnPublish)
	blk := mock.StateBlock()

	blk.Token = types.ZeroHash
	_, _, err := up.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal()
	}

	blk.Token = cfg.GasToken()
	blk.Data = []byte("test")
	_, _, err = up.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	pt := common.OracleTypeInvalid
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	hash := mock.Hash()
	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, pt, id, pk, hash)
	_, _, err = up.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	pt = common.OracleTypeEmail
	p := new(Publish)
	p.SetStorage(ctx, blk.Address, pt, id, pk, vs, cs, fee, hash)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, pt, id, pk, hash)
	_, _, err = up.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOracle(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	o := new(Oracle)
	blk := mock.StateBlock()

	blk.Token = types.ZeroHash
	_, _, err := o.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal()
	}

	blk.Token = cfg.GasToken()
	blk.Data = []byte("test")
	_, _, err = o.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	ot := common.OracleTypeInvalid
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	_, _, err = o.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	blk.Flag = types.BlockFlagNonSync
	_, _, err = o.ProcessSend(ctx, blk)
	if err != ErrNotEnoughPledge {
		t.Fatal(err)
	}

	am := mock.AccountMeta(blk.Address)
	am.CoinOracle = common.MinVerifierPledgeAmount
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, _, err = o.ProcessSend(ctx, blk)
	if err != ErrGetVerifier {
		t.Fatal(err)
	}

	ot = common.OracleTypeEmail
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	vr := new(VerifierRegister)
	vr.SetStorage(ctx, blk.Address, common.OracleTypeEmail, "test@126.com")
	p := new(Publish)
	vs := []types.Address{blk.Address}
	codeComb := append([]byte(types.NewHexBytesFromData(pk).String()), []byte(code)...)
	codeHash, _ := types.Sha256HashData(codeComb)
	cs := []types.Hash{codeHash}
	fee := common.PublishCost
	p.SetStorage(ctx, blk.Address, ot, id, pk, vs, cs, fee, hash)
	_, _, err = o.ProcessSend(ctx, blk)
	if err != ErrCalcAmount {
		t.Fatal(err)
	}

	blk.Type = types.Open
	_, _, err = o.ProcessSend(ctx, blk)
	if err != ErrNotEnoughFee {
		t.Fatal(err)
	}

	blk.Balance = common.OracleCost
	_, _, err = o.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOracle_DoGap(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	o := new(Oracle)
	blk := mock.StateBlock()

	gap, _, _ := o.DoGap(ctx, blk)
	if gap != common.ContractNoGap {
		t.Fatal()
	}

	ot := common.OracleTypeInvalid
	id := mock.Hash()
	hash := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	vs := []types.Address{blk.Address}
	cs := []types.Hash{mock.Hash()}
	fee := common.PublishCost
	blk.Data, _ = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	gap, _, _ = o.DoGap(ctx, blk)
	if gap != common.ContractDPKIGapPublish {
		t.Fatal(gap)
	}

	p := new(Publish)
	p.SetStorage(ctx, blk.Address, ot, id, pk, vs, cs, fee, hash)
	gap, _, _ = o.DoGap(ctx, blk)
	if gap != common.ContractNoGap {
		t.Fatal()
	}
}

func TestVerifierHeart(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	vh := new(VerifierHeart)
	blk := mock.StateBlock()

	blk.Token = types.ZeroHash
	_, _, err := vh.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal()
	}

	blk.Token = cfg.GasToken()
	blk.Data = []byte("test")
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	ht := []uint32{common.OracleTypeInvalid, common.OracleTypeWeChat}
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierHeart, ht)
	blk.Flag &= ^types.BlockFlagNonSync
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != ErrCalcAmount {
		t.Fatal(err)
	}

	blk.Flag |= types.BlockFlagNonSync
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != ErrNotEnoughPledge {
		t.Fatal(err)
	}

	vr := new(VerifierRegister)
	vr.SetStorage(ctx, blk.Address, common.OracleTypeWeChat, "wcid12345")
	am := mock.AccountMeta(blk.Address)
	am.CoinOracle = common.MinVerifierPledgeAmount
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != ErrGetVerifier {
		t.Fatal(err)
	}

	ht = []uint32{common.OracleTypeEmail, common.OracleTypeWeChat}
	vr.SetStorage(ctx, blk.Address, common.OracleTypeEmail, "test@126.com")
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierHeart, ht)
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != ErrCalcAmount {
		t.Fatal(err)
	}

	blk.Type = types.Open
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != ErrNotEnoughFee {
		t.Fatal(err)
	}

	blk.Balance = common.OracleCost
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifierHeart_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	cs := types.NewPovContractState()
	csdb := statedb.NewPovContractStateDB(l.DBStore(), cs)
	ctx := vmstore.NewVMContext(l)
	vh := new(VerifierHeart)
	blk := mock.StateBlock()
	var err error

	ht := []uint32{common.OracleTypeWeChat}
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierHeart, ht)
	err = vh.DoSendOnPov(ctx, csdb, 100, blk)
	if err != nil {
		t.Fatal(err)
	}

	var vsRawKey []byte
	vsRawKey = append(vsRawKey, blk.Address[:]...)
	vsVal, err := dpki.PovGetVerifierState(csdb, vsRawKey)
	if err != nil {
		t.Fatal(err)
	}

	if vsVal.ActiveHeight[common.OracleTypeToString(common.OracleTypeWeChat)] != 100 {
		t.Fatal()
	}
}

func TestPublish_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	ctx := vmstore.NewVMContext(l)
	p := new(Publish)
	blk := mock.StateBlockWithoutWork()
	var err error

	pt := common.OracleTypeEmail
	id := mock.Hash()
	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := common.PublishCost
	pk := make([]byte, ed25519.PublicKeySize)
	err = random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, pk, vs, cs, fee.Int)
	if err != nil {
		t.Fatal(err)
	}

	err = p.DoSendOnPov(ctx, csdb, 100, blk)
	if err != nil {
		t.Fatal(err)
	}

	pubInfoKey := &abi.PublishInfoKey{
		PType:  pt,
		PID:    id,
		PubKey: pk,
		Hash:   blk.Previous,
	}
	psRawKey := pubInfoKey.ToRawKey()

	ps, _ := dpki.PovGetPublishState(csdb, psRawKey)
	if ps == nil || ps.PublishHeight != 100 {
		t.Fatal()
	}
}

func TestOracle_DoSendOnPov(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	csdb := statedb.NewPovContractStateDB(l.DBStore(), types.NewPovContractState())
	ctx := vmstore.NewVMContext(l)
	o := new(Oracle)
	blk1 := mock.StateBlockWithoutWork()
	var err error

	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)

	p := new(Publish)
	pa := mock.Address()
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	err = p.SetStorage(ctx, pa, ot, id, pk, vs, cs, common.PublishCost, hash)
	if err != nil {
		t.Fatal(err)
	}

	blk1.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk1)
	if err != nil {
		t.Fatal(err)
	}

	blk2 := mock.StateBlockWithoutWork()
	blk2.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk2)
	if err != nil {
		t.Fatal(err)
	}

	blk3 := mock.StateBlockWithoutWork()
	blk3.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk3)
	if err != nil {
		t.Fatal(err)
	}

	blk4 := mock.StateBlockWithoutWork()
	blk4.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk4)
	if err != nil {
		t.Fatal(err)
	}

	blk5 := mock.StateBlockWithoutWork()
	blk5.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk5)
	if err != nil {
		t.Fatal(err)
	}

	blk6 := mock.StateBlockWithoutWork()
	blk6.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk6)
	if err != nil {
		t.Fatal(err)
	}

	blk7 := mock.StateBlockWithoutWork()
	blk7.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 100, blk7)
	if err != nil {
		t.Fatal(err)
	}

	pubInfoKey := &abi.PublishInfoKey{
		PType:  ot,
		PID:    id,
		PubKey: pk,
		Hash:   hash,
	}
	psRawKey := pubInfoKey.ToRawKey()

	ps, _ := dpki.PovGetPublishState(csdb, psRawKey)
	if ps == nil || ps.VerifiedHeight != 10 || ps.VerifiedStatus != types.PovPublishStatusVerified {
		t.Fatal()
	}

	vrs, _ := dpki.PovGetVerifierState(csdb, blk1.Address[:])
	if vrs == nil || vrs.TotalVerify == 0 || vrs.ActiveHeight["email"] != 10 {
		t.Fatal()
	}

	vrs, _ = dpki.PovGetVerifierState(csdb, blk6.Address[:])
	if vrs != nil {
		t.Fatal()
	}

	vrs, _ = dpki.PovGetVerifierState(csdb, blk7.Address[:])
	if vrs != nil {
		t.Fatal()
	}
}
