package contract

import (
	"crypto/ed25519"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	qcfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func getLedger() (func(), *ledger.Ledger) {
	dir := filepath.Join(qcfg.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := qcfg.NewCfgManager(dir)
	cfg, _ := cm.Load()
	l := ledger.NewLedger(cm.ConfigFile)

	var mintageBlock, genesisBlock types.StateBlock
	for _, v := range cfg.Genesis.GenesisBlocks {
		_ = json.Unmarshal([]byte(v.Genesis), &genesisBlock)
		_ = json.Unmarshal([]byte(v.Mintage), &mintageBlock)
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: mintageBlock,
			GenesisBlock:        genesisBlock,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}

	return func() {
		l.Close()
		os.RemoveAll(dir)
	}, l
}

func TestVerifierRegister(t *testing.T) {
	clear, l := getLedger()
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

	blk.Token = common.ChainToken()
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
	l.AddAccountMeta(am)
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
	clear, l := getLedger()
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

	blk.Token = common.ChainToken()
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
	clear, l := getLedger()
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

	blk.Token = common.GasToken()
	blk.Data = []byte("test")
	_, _, err = p.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	pt := common.OracleTypeInvalid
	id := mock.Hash()
	pk := make([]byte, ed25519.PublicKeySize)
	random.Bytes(pk)
	vs := []types.Address{mock.Address(), mock.Address()}
	cs := []types.Hash{mock.Hash(), mock.Hash()}
	fee := big.NewInt(5e8)
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
	clear, l := getLedger()
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

	blk.Token = common.GasToken()
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
	clear, l := getLedger()
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

	blk.Token = common.GasToken()
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
	l.AddAccountMeta(am)
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
	clear, l := getLedger()
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
