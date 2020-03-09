package api

import (
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func addTestVerifierInfo(ctx *vmstore.VMContext, account types.Address, vType uint32, vInfo string) error {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDVerifierInfo, vInfo, true)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PKDStorageTypeVerifier)
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

func TestNewPublicKeyDistributionApi_sortPublishInfo(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	hashes := []types.Hash{mock.Hash(), mock.Hash(), mock.Hash(), mock.Hash(), mock.Hash()}

	pubs := []*PublishInfoState{
		{
			PublishParam: &PublishParam{
				Hash: hashes[0].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address()},
				VerifiedHeight: 0,
				VerifiedStatus: types.PovPublishStatusInit,
				BonusFee:       nil,
				PublishHeight:  100,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[1].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address(), mock.Address(), mock.Address()},
				VerifiedHeight: 20,
				VerifiedStatus: types.PovPublishStatusVerified,
				BonusFee:       nil,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[2].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address(), mock.Address(), mock.Address()},
				VerifiedHeight: 10,
				VerifiedStatus: types.PovPublishStatusVerified,
				BonusFee:       nil,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[3].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address(), mock.Address()},
				VerifiedHeight: 10,
				VerifiedStatus: types.PovPublishStatusVerified,
				BonusFee:       nil,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[4].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{},
				VerifiedHeight: 0,
				VerifiedStatus: types.PovPublishStatusInit,
				BonusFee:       nil,
				PublishHeight:  50,
			},
		},
	}

	p := NewPublicKeyDistributionApi(cfgFile, l)
	p.sortPublishInfo(pubs)

	if pubs[0].Hash != hashes[1].String() || pubs[1].Hash != hashes[2].String() || pubs[2].Hash != hashes[3].String() ||
		pubs[3].Hash != hashes[4].String() || pubs[4].Hash != hashes[0].String() {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetVerifierRegisterBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionApi(cfgFile, l)
	param := new(VerifierRegParam)
	param.Account = mock.Address()
	param.VType = "email"
	param.VInfo = "123@test.com"

	_, err := pkd.GetVerifierRegisterBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, err = pkd.GetVerifierRegisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	pkd.cc.Init(nil)
	pkd.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	_, err = pkd.GetVerifierRegisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Account)
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierRegisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Type = config.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierRegisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	am.CoinOracle = common.MinVerifierPledgeAmount
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierRegisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	_, err = pkd.GetVerifierRegisterBlock(param)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublicKeyDistributionApi_GetVerifierUnregisterBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionApi(cfgFile, l)
	param := new(VerifierUnRegParam)
	param.Account = mock.Address()
	param.VType = "email"

	_, err := pkd.GetVerifierUnregisterBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, err = pkd.GetVerifierUnregisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	pkd.cc.Init(nil)
	pkd.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	_, err = pkd.GetVerifierUnregisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	ctx := vmstore.NewVMContext(l)
	addTestVerifierInfo(ctx, param.Account, common.OracleTypeEmail, "123@test.com")
	_, err = pkd.GetVerifierUnregisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Account)
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierUnregisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Type = config.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierUnregisterBlock(param)
	if err == nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	_, err = pkd.GetVerifierUnregisterBlock(param)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublicKeyDistributionApi_GetAllVerifiers(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionApi(cfgFile, l)
	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	_, err := pkd.GetAllVerifiers()
	if err != nil {
		t.Fatal()
	}

	addTestVerifierInfo(ctx, account, common.OracleTypeEmail, "123@test.com")
	vs, _ := pkd.GetAllVerifiers()
	if vs == nil || len(vs) != 1 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetVerifiersByType(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionApi(cfgFile, l)
	ctx := vmstore.NewVMContext(l)
	account := mock.Address()
	addTestVerifierInfo(ctx, account, common.OracleTypeEmail, "123@test.com")
	_, err := pkd.GetVerifiersByType("wechat")
	if err == nil {
		t.Fatal()
	}

	vs, _ := pkd.GetVerifiersByType("weChat")
	if vs != nil && len(vs) != 0 {
		t.Fatal()
	}

	vs, err = pkd.GetVerifiersByType("email")
	if vs == nil || len(vs) != 1 {
		t.Fatal()
	}
}
