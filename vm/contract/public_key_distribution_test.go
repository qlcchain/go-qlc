package contract

import (
	"crypto/ed25519"
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger"
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

	vk := mock.Hash()
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierRegister, common.OracleTypeEmail, "test@126.com", vk[:])
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

	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDVerifierRegister, common.OracleTypeInvalid, "test@126.com", vk[:])
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
	vk := mock.Hash()
	vr.SetStorage(ctx, blk.Address, common.OracleTypeEmail, "test@126.com", vk[:])
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
	kt := common.PublicKeyTypeED25519
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	vs := make([]types.Address, 0)
	cs := make([]types.Hash, 0)
	fee := big.NewInt(5e8)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, kt, pk, vs, cs, fee)
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrVerifierNum {
		t.Fatal(err)
	}

	vs = append(vs, mock.Address())
	cs = append(cs, mock.Hash())
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, kt, pk, vs, cs, fee)
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	pt = common.OracleTypeEmail
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, kt, pk, vs, cs, fee)
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
	kt := common.PublicKeyTypeED25519
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	hash := mock.Hash()
	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := types.NewBalance(5e8)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, pt, id, kt, pk, hash)
	_, _, err = up.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	pt = common.OracleTypeEmail
	p := new(Publish)
	p.SetStorage(ctx, blk.Address, pt, id, kt, pk, vs, cs, fee, hash)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDUnPublish, pt, id, kt, pk, hash)
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
	kt := common.PublicKeyTypeED25519
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
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
	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	vr := new(VerifierRegister)
	vk := mock.Hash()
	vr.SetStorage(ctx, blk.Address, common.OracleTypeEmail, "test@126.com", vk[:])
	p := new(Publish)
	vs := []types.Address{blk.Address}
	codeComb := append([]byte(types.NewHexBytesFromData(pk).String()), []byte(code)...)
	codeHash, _ := types.Sha256HashData(codeComb)
	cs := []types.Hash{codeHash}
	fee := common.PublishCost
	p.SetStorage(ctx, blk.Address, ot, id, kt, pk, vs, cs, fee, hash)
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
	kt := common.PublicKeyTypeED25519
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	vs := []types.Address{blk.Address}
	cs := []types.Hash{mock.Hash()}
	fee := common.PublishCost
	blk.Data, _ = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	gap, _, _ = o.DoGap(ctx, blk)
	if gap != common.ContractDPKIGapPublish {
		t.Fatal(gap)
	}

	p := new(Publish)
	p.SetStorage(ctx, blk.Address, ot, id, kt, pk, vs, cs, fee, hash)
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
	vk := mock.Hash()
	vr.SetStorage(ctx, blk.Address, common.OracleTypeWeChat, "wcid12345", vk[:])
	am := mock.AccountMeta(blk.Address)
	am.CoinOracle = common.MinVerifierPledgeAmount
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, _, err = vh.ProcessSend(ctx, blk)
	if err != ErrGetVerifier {
		t.Fatal(err)
	}

	ht = []uint32{common.OracleTypeEmail, common.OracleTypeWeChat}
	vr.SetStorage(ctx, blk.Address, common.OracleTypeEmail, "test@126.com", vk[:])
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
	kt := common.PublicKeyTypeED25519
	pk := make([]byte, ed25519.PublicKeySize)
	err = random.Bytes(pk)
	if err != nil {
		t.Fatal(err)
	}

	blk.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, pt, id, kt, pk, vs, cs, fee.Int)
	if err != nil {
		t.Fatal(err)
	}

	err = p.DoSendOnPov(ctx, csdb, 100, blk)
	if err != nil {
		t.Fatal(err)
	}

	kh := common.PublicKeyWithTypeHash(kt, pk)
	pubInfoKey := &abi.PublishInfoKey{
		PType:  pt,
		PID:    id,
		PubKey: kh,
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
	kt := common.PublicKeyTypeED25519
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)

	p := new(Publish)
	pa := mock.Address()
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	err = p.SetStorage(ctx, pa, ot, id, kt, pk, vs, cs, common.PublishCost, hash)
	if err != nil {
		t.Fatal(err)
	}

	blk1.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk1)
	if err != nil {
		t.Fatal(err)
	}

	blk2 := mock.StateBlockWithoutWork()
	blk2.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk2)
	if err != nil {
		t.Fatal(err)
	}

	blk3 := mock.StateBlockWithoutWork()
	blk3.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk3)
	if err != nil {
		t.Fatal(err)
	}

	blk4 := mock.StateBlockWithoutWork()
	blk4.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk4)
	if err != nil {
		t.Fatal(err)
	}

	blk5 := mock.StateBlockWithoutWork()
	blk5.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk5)
	if err != nil {
		t.Fatal(err)
	}

	blk6 := mock.StateBlockWithoutWork()
	blk6.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 10, blk6)
	if err != nil {
		t.Fatal(err)
	}

	blk7 := mock.StateBlockWithoutWork()
	blk7.Data, err = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDOracle, ot, id, kt, pk, code, hash)
	if err != nil {
		t.Fatal()
	}

	err = o.DoSendOnPov(ctx, csdb, 100, blk7)
	if err != nil {
		t.Fatal(err)
	}

	kh := common.PublicKeyWithTypeHash(kt, pk)
	pubInfoKey := &abi.PublishInfoKey{
		PType:  ot,
		PID:    id,
		PubKey: kh,
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

func TestPKDReward_DoGap(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	r := new(PKDReward)

	ctx := vmstore.NewVMContext(l)

	blk := mock.StateBlockWithoutWork()

	_ = l.DropAllPovBlocks()
	mockPovHeight := uint64(1440*3 - 1)
	mockPKDRewardAddPovBlock(t, l, mockPovHeight, blk, 100)

	param1 := new(dpki.PKDRewardParam)
	param1.Account = blk.Address
	param1.Beneficial = blk.Address
	param1.EndHeight = 1439
	param1.RewardAmount = big.NewInt(100000000)

	blk.Data, _ = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDReward,
		param1.Account, param1.Beneficial, param1.EndHeight, param1.RewardAmount)

	gapType, gapData, err := r.DoGap(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
	if gapType != common.ContractNoGap {
		t.Fatal("gapType is not NoGap", gapType, gapData)
	}
	if gapData != nil {
		t.Fatal("gapData is not nil", gapData, gapData)
	}

	param2 := new(dpki.PKDRewardParam)
	param2.Account = blk.Address
	param2.Beneficial = blk.Address
	param2.EndHeight = mockPovHeight - 1
	param2.RewardAmount = big.NewInt(100000000)
	blk.Data, _ = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDReward,
		param2.Account, param2.Beneficial, param2.EndHeight, param2.RewardAmount)

	gapType, gapData, err = r.DoGap(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
	if gapType != common.ContractRewardGapPov {
		t.Fatal("gapType is not GapPov", gapType, gapData)
	}
	if gapData == nil {
		t.Fatal("gapData is nil")
	}
}

func TestPKDReward_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	r := new(PKDReward)

	ctx := vmstore.NewVMContext(l)

	sendBlk := mock.StateBlockWithoutWork()
	sendBlk.Token = cfg.GasToken()

	am := mock.AccountMeta(sendBlk.Address)
	am.CoinOracle = common.MinVerifierPledgeAmount
	l.AddAccountMeta(am, l.Cache().GetCache())

	_ = l.DropAllPovBlocks()
	mockPKDRewardAddPovBlock(t, l, 1439, sendBlk, 100)
	mockPKDRewardAddPovBlock(t, l, 2879, sendBlk, 200)
	mockPKDRewardAddPovBlock(t, l, 4319, sendBlk, 300)

	param1 := new(dpki.PKDRewardParam)
	param1.Account = sendBlk.Address
	param1.Beneficial = sendBlk.Address
	param1.EndHeight = 1439
	param1.RewardAmount = big.NewInt(100 * 100000000)

	sendBlk.Data, _ = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDReward,
		param1.Account, param1.Beneficial, param1.EndHeight, param1.RewardAmount)

	pendKey, pendInfo, err := r.ProcessSend(ctx, sendBlk)
	if err != nil {
		t.Fatal(err)
	}
	if pendKey == nil {
		t.Fatal("ProcessSend pendKey is nil")
	}
	if pendKey.Address != param1.Account {
		t.Fatal("ProcessSend pendKey address not equal", pendKey.Address, param1.Account)
	}
	if pendInfo == nil {
		t.Fatal("ProcessSend pendInfo is nil")
	}
	if pendInfo.Amount.Compare(types.NewBalanceFromBigInt(param1.RewardAmount)) != types.BalanceCompEqual {
		t.Fatal("ProcessSend pendInfo amount not equal", pendInfo.Amount, param1.RewardAmount)
	}

	recvBlk := new(types.StateBlock)
	retConBlks, err := r.DoReceive(ctx, recvBlk, sendBlk)
	if err != nil {
		t.Fatal(err)
	}
	if len(retConBlks) == 0 {
		t.Fatal("return contract blocks is nil")
	}
	if retConBlks[0].Block.Address != param1.Beneficial {
		t.Fatal("recvBlk address not equal", retConBlks[0].Block.Address, param1.Beneficial)
	}
	if retConBlks[0].Amount.Compare(types.NewBalanceFromBigInt(param1.RewardAmount)) != types.BalanceCompEqual {
		t.Fatal("recvBlk amount not equal", retConBlks[0].Amount, param1.RewardAmount)
	}
}

func mockPKDRewardAddPovBlock(t *testing.T, l ledger.Store, povHeight uint64, txBlock *types.StateBlock, rwdCnt uint) {
	// blk1
	povBlk1, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk1.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PubKeyDistributionAddress)
	if err != nil {
		t.Fatal(err)
	}

	ps := types.NewPovVerifierState()
	ps.TotalVerify = uint64(rwdCnt)
	ps.TotalReward = types.NewBigNumFromInt(int64(rwdCnt * 100000000))
	if povHeight > 240 {
		ps.ActiveHeight["email"] = povHeight - 240
	}
	err = dpki.PovSetVerifierState(csdb, txBlock.Address.Bytes(), ps)
	if err != nil {
		t.Fatal(err)
	}

	err = gsdb.CommitToTrie()
	if err != nil {
		t.Fatal(err)
	}
	txn := l.DBStore().Batch(true)
	err = gsdb.CommitToDB(txn)
	if err != nil {
		t.Fatal(err)
	}
	err = l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	povBlk1.Header.CbTx.StateHash = gsdb.GetCurHash()

	mock.UpdatePovHash(povBlk1)

	t.Log("add pov block", povBlk1)

	err = l.AddPovBlock(povBlk1, povTd)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(povBlk1.GetHeight(), povBlk1.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.SetPovLatestHeight(povBlk1.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	/*
		{
			gsdb2 := statedb.NewPovGlobalStateDB(l.DBStore(), povBlk1.Header.CbTx.StateHash)
			csdb2, err := gsdb2.LookupContractStateDB(contractaddress.PubKeyDistributionAddress)
			if err != nil {
				t.Fatal(err)
			}

			retPs, err := dpki.PovGetVerifierState(csdb2, txBlock.Address.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			t.Log("PovGetVerifierState", retPs)
		}
	*/
}

func TestPKDReward_GetTargetReceiver(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	r := new(PKDReward)
	ctx := vmstore.NewVMContext(l)

	sendBlk := mock.StateBlockWithoutWork()
	sendBlk.Token = cfg.GasToken()
	param := new(dpki.PKDRewardParam)
	param.Account = sendBlk.Address
	param.Beneficial = mock.Address()
	param.EndHeight = 1439
	param.RewardAmount = big.NewInt(100 * 100000000)

	sendBlk.Data, _ = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDReward,
		param.Account, param.Beneficial, param.EndHeight, param.RewardAmount)
	recv, err := r.GetTargetReceiver(ctx, sendBlk)
	if err != nil || recv != param.Beneficial {
		t.Fatal()
	}
}
