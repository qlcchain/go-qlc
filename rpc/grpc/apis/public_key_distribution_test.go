package apis

import (
	"context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/contract/dpki"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"math/big"
	"testing"
	"time"
)

func addTestVerifierInfo(t *testing.T, ctx *vmstore.VMContext, l *ledger.Ledger, account types.Address, vType uint32, vInfo string, vKey []byte) {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDVerifierInfo, vInfo, vKey, true)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.PKDStorageTypeVerifier)
	key = append(key, util.BE_Uint32ToBytes(vType)...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(contractaddress.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		t.Fatal(err)
	}

	err = l.SaveStorage(vmstore.ToCache(ctx))
	if err != nil {
		t.Fatal(err)
	}
}

func addTestVerifierState(t *testing.T, l *ledger.Ledger, povHeight uint64, accounts []types.Address, rwdCnt uint) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PubKeyDistributionAddress)
	if err != nil {
		t.Fatal(err)
	}

	for _, account := range accounts {
		ps := types.NewPovVerifierState()
		ps.TotalVerify = uint64(rwdCnt)
		ps.TotalReward = types.NewBigNumFromInt(int64(rwdCnt * 100000000))
		ps.ActiveHeight["email"] = povHeight
		err = dpki.PovSetVerifierState(csdb, account[:], ps)
		if err != nil {
			t.Fatal(err)
		}
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

	povBlk.Header.CbTx.StateHash = gsdb.GetCurHash()
	mock.UpdatePovHash(povBlk)

	err = l.AddPovBlock(povBlk, povTd)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(povBlk.GetHeight(), povBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.SetPovLatestHeight(povBlk.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
}

func addTestPublishInfo(t *testing.T, ctx *vmstore.VMContext, store *ledger.Ledger, account types.Address, pt uint32, id types.Hash, kt uint16, pk []byte,
	vs []types.Address, cs []types.Hash, fee types.Balance, hash types.Hash) {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDPublishInfo, account, vs, cs, fee.Int, true, kt, pk)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	kh := common.PublicKeyWithTypeHash(kt, pk)
	key = append(key, abi.PKDStorageTypePublisher)
	key = append(key, util.BE_Uint32ToBytes(pt)...)
	key = append(key, id[:]...)
	key = append(key, kh...)
	key = append(key, hash[:]...)
	err = ctx.SetStorage(contractaddress.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		t.Fatal(err)
	}

	err = store.SaveStorage(vmstore.ToCache(ctx))
	if err != nil {
		t.Fatal(err)
	}
}

func addTestOracleInfo(t *testing.T, ctx *vmstore.VMContext, store *ledger.Ledger, account types.Address, ot uint32, id types.Hash, kt uint16, pk []byte, code string, hash types.Hash) {
	data, err := abi.PublicKeyDistributionABI.PackVariable(abi.VariableNamePKDOracleInfo, code, kt, pk)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	kh := common.PublicKeyWithTypeHash(kt, pk)
	key = append(key, abi.PKDStorageTypeOracle)
	key = append(key, util.BE_Uint32ToBytes(ot)...)
	key = append(key, id[:]...)
	key = append(key, kh...)
	key = append(key, hash[:]...)
	key = append(key, account[:]...)
	err = ctx.SetStorage(contractaddress.PubKeyDistributionAddress[:], key, data)
	if err != nil {
		t.Fatal(err)
	}

	err = store.SaveStorage(vmstore.ToCache(ctx))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublicKeyDistributionApi_GetVerifierRegisterBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	param := new(api.VerifierRegParam)
	param.Account = mock.Address()
	param.VType = "email"
	param.VInfo = "123@test.com"
	param.VKey = "123"

	_, err := pkd.GetVerifierRegisterBlock(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	_, err = pkd.GetVerifierRegisterBlock(context.Background(), toVerifierRegParam(param))
	if err == nil {
		t.Fatal()
	}

	pkd.cc.Init(nil)
	pkd.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	_, err = pkd.GetVerifierRegisterBlock(context.Background(), toVerifierRegParam(param))
	if err == nil {
		t.Fatal()
	}

	param.VKey = mock.Hash().String()
	_, err = pkd.GetVerifierRegisterBlock(context.Background(), toVerifierRegParam(param))
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Account)
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierRegisterBlock(context.Background(), toVerifierRegParam(param))
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Type = config.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierRegisterBlock(context.Background(), toVerifierRegParam(param))
	if err == nil {
		t.Fatal()
	}

	am.CoinOracle = common.MinVerifierPledgeAmount
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetVerifierRegisterBlock(context.Background(), toVerifierRegParam(param))
	if err == nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	_, err = pkd.GetVerifierRegisterBlock(context.Background(), toVerifierRegParam(param))
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

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	param := new(api.VerifierUnRegParam)
	param.Account = mock.Address()
	param.VType = "email"

	_, err := pkd.GetVerifierUnregisterBlock(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	_, err = pkd.GetVerifierUnregisterBlock(context.Background(), &pb.VerifierUnRegParam{
		Account: toAddressValue(param.Account),
		Type:    param.VType,
	})
	if err == nil {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetAllVerifiers(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	vk := mock.Hash()
	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	account := mock.Address()
	_, err := pkd.GetAllVerifiers(context.Background(), nil)
	if err != nil {
		t.Fatal()
	}

	addTestVerifierInfo(t, ctx, l, account, common.OracleTypeEmail, "123@test.com", vk[:])
	vs, _ := pkd.GetAllVerifiers(context.Background(), nil)
	if len(vs.GetParams()) != 1 || vk.String() != vs.GetParams()[0].Key {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetVerifiersByType(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	vk := mock.Hash()
	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	account := mock.Address()
	addTestVerifierInfo(t, ctx, l, account, common.OracleTypeEmail, "123@test.com", vk[:])
	_, err := pkd.GetVerifiersByType(context.Background(), toString("wechat"))
	if err == nil {
		t.Fatal()
	}

	vs, _ := pkd.GetVerifiersByType(context.Background(), toString("wechat"))
	if vs != nil && len(vs.GetParams()) != 0 {
		t.Fatal()
	}

	vs, err = pkd.GetVerifiersByType(context.Background(), toString("email"))
	if vs == nil || len(vs.GetParams()) != 1 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetActiveVerifiers(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	vs, _ := pkd.GetActiveVerifiers(context.Background(), toString("invalid"))
	if len(vs.GetParams()) != 0 {
		t.Fatal()
	}

	vs, _ = pkd.GetActiveVerifiers(context.Background(), toString("email"))
	if len(vs.GetParams()) != 0 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetVerifiersByAccount(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	vk := mock.Hash()
	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	account := mock.Address()

	addTestVerifierInfo(t, ctx, l, account, common.OracleTypeEmail, "test@123.com", vk[:])
	vs, _ := pkd.GetVerifiersByAccount(context.Background(), toAddress(account))
	if len(vs.GetParams()) != 1 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetVerifierStateByBlockHeight(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	account := mock.Address()

	vs, _ := pkd.GetVerifierStateByBlockHeight(context.Background(), &pb.VerifierStateByBlockHeightRequest{
		Height:  100,
		Address: toAddressValue(account),
	})
	if vs != nil {
		t.Fatal()
	}

	addTestVerifierState(t, l, 100, []types.Address{account}, 1000)
	vs, _ = pkd.GetVerifierStateByBlockHeight(context.Background(), &pb.VerifierStateByBlockHeightRequest{
		Height:  100,
		Address: toAddressValue(account),
	})
	if vs == nil {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetAllVerifierStatesByBlockHeight(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	vs, _ := pkd.GetAllVerifierStatesByBlockHeight(context.Background(), toUInt64(100))
	if vs != nil {
		t.Fatal()
	}

	addTestVerifierState(t, l, 100, []types.Address{mock.Address(), mock.Address(), mock.Address(), mock.Address(), mock.Address()}, 1000)

	vs, _ = pkd.GetAllVerifierStatesByBlockHeight(context.Background(), toUInt64(100))
	if vs == nil || vs.GetVerifierNum() != 5 {
		t.Fatal(vs)
	}
}

func toPublishParam(param *api.PublishParam) *pb.PublishParam {
	return &pb.PublishParam{
		Account:   toAddressValue(param.Account),
		Type:      param.PType,
		Id:        param.PID,
		PubKey:    param.PubKey,
		KeyType:   param.KeyType,
		Fee:       toBalanceValue(param.Fee),
		Verifiers: toAddressValues(param.Verifiers),
		Codes:     toHashValues(param.Codes),
		Hash:      param.Hash,
	}
}

func TestPublicKeyDistributionApi_GetPublishBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	param := new(api.PublishParam)
	param.Account = mock.Address()
	param.Fee = common.PublishCost
	param.Verifiers = []types.Address{mock.Address()}
	param.PType = "invalid"
	param.KeyType = "ed25519"
	param.PubKey = mock.Hash().String()

	_, err := pkd.GetPublishBlock(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	pkd.cc.Init(nil)
	pkd.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	param.PType = "email"
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	param.Verifiers = []types.Address{mock.Address(), mock.Address(), mock.Address()}
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Account)
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Type = config.GasToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Balance = common.PublishCost
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	vk := mock.Hash()
	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	for _, v := range param.Verifiers {
		addTestVerifierInfo(t, ctx, l, v, common.OracleTypeEmail, "123@test.com", vk[:])
	}
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	addTestVerifierState(t, l, 100, []types.Address{mock.Address()}, 100)
	preBlk := mock.StateBlockWithoutWork()
	preBlk.Balance = common.PublishCost
	l.AddStateBlock(preBlk)
	am.Tokens[0].Header = preBlk.GetHash()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetPublishBlock(context.Background(), toPublishParam(param))
	if err != nil {
		t.Fatal(err)
	}
}

func toUnPublishParam(param *api.UnPublishParam) *pb.UnPublishParam {
	return &pb.UnPublishParam{
		Account: toAddressValue(param.Account),
		Type:    param.PType,
		Id:      param.PID,
		PubKey:  param.PubKey,
		KeyType: param.KeyType,
		Hash:    param.Hash,
	}
}

func TestPublicKeyDistributionApi_GetUnPublishBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	param := new(api.UnPublishParam)
	param.Hash = "123"
	param.Account = mock.Address()
	param.PType = "email"
	param.PID = "123@test.com"
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	param.KeyType = "ed25519"
	param.PubKey = pk.String()

	_, err := pkd.GetUnPublishBlock(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	_, err = pkd.GetUnPublishBlock(context.Background(), toUnPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	pkd.cc.Init(nil)
	pkd.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	_, err = pkd.GetUnPublishBlock(context.Background(), toUnPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	param.Hash = mock.Hash().String()
	_, err = pkd.GetUnPublishBlock(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	pt := common.OracleTypeEmail
	id, _ := types.Sha256HashData([]byte(param.PID))
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	fee := common.PublishCost
	hash, _ := types.NewHash(param.Hash)
	addTestPublishInfo(t, ctx, l, param.Account, pt, id, kt, pk[:], vs, cs, fee, hash)
	_, err = pkd.GetUnPublishBlock(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Account)
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetUnPublishBlock(context.Background(), toUnPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Type = config.GasToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetUnPublishBlock(context.Background(), toUnPublishParam(param))
	if err == nil {
		t.Fatal()
	}

	addTestVerifierState(t, l, 100, []types.Address{mock.Address()}, 100)
	_, err = pkd.GetUnPublishBlock(context.Background(), toUnPublishParam(param))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublicKeyDistributionApi_GetPubKeyByTypeAndID(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	pType := "email"
	pID := "123@test.com"
	ps, _ := pkd.GetPubKeyByTypeAndID(context.Background(), &pb.TypeAndIDParam{
		PType: pType,
		PID:   pID,
	})
	if len(ps.GetStates()) != 0 {
		t.Fatal()
	}

	pt := common.OracleTypeEmail
	id, _ := types.Sha256HashData([]byte(pID))
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	fee := common.PublishCost
	hash := mock.Hash()
	account := mock.Address()
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	addTestPublishInfo(t, ctx, l, account, pt, id, kt, pk[:], vs, cs, fee, hash)
	ps, _ = pkd.GetPubKeyByTypeAndID(context.Background(), &pb.TypeAndIDParam{
		PType: pType,
		PID:   pID,
	})
	if len(ps.GetStates()) != 1 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetRecommendPubKey(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	pType := "email"
	pID := "123@test.com"
	ps, _ := pkd.GetRecommendPubKey(context.Background(), &pb.TypeAndIDParam{
		PType: pType,
		PID:   pID,
	})
	if ps != nil {
		t.Fatal()
	}

	pt := common.OracleTypeEmail
	id, _ := types.Sha256HashData([]byte(pID))
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	fee := common.PublishCost
	hash := mock.Hash()
	account := mock.Address()
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	addTestPublishInfo(t, ctx, l, account, pt, id, kt, pk[:], vs, cs, fee, hash)

	account2 := mock.Address()
	hash2 := mock.Hash()
	addTestPublishInfo(t, ctx, l, account2, pt, id, kt, pk[:], vs, cs, fee, hash2)

	ps, _ = pkd.GetRecommendPubKey(context.Background(), &pb.TypeAndIDParam{
		PType: pType,
		PID:   pID,
	})
	if ps == nil {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetPublishInfosByType(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	pType := "email"
	ps, _ := pkd.GetPublishInfosByType(context.Background(), toString(pType))
	if len(ps.GetStates()) != 0 {
		t.Fatal()
	}

	pt := common.OracleTypeEmail
	id := mock.Hash()
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	fee := common.PublishCost
	hash := mock.Hash()
	account := mock.Address()
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	addTestPublishInfo(t, ctx, l, account, pt, id, kt, pk[:], vs, cs, fee, hash)

	pt2 := common.OracleTypeWeChat
	id2 := mock.Hash()
	hash2 := mock.Hash()
	addTestPublishInfo(t, ctx, l, account, pt2, id2, kt, pk[:], vs, cs, fee, hash2)

	ps, _ = pkd.GetPublishInfosByType(context.Background(), toString(pType))
	if len(ps.GetStates()) != 1 || ps.GetStates()[0].Type != "email" {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetPublishInfosByAccountAndType(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	account := mock.Address()

	pt := common.OracleTypeEmail
	id := mock.Hash()
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	fee := common.PublishCost
	hash := mock.Hash()
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	addTestPublishInfo(t, ctx, l, account, pt, id, kt, pk[:], vs, cs, fee, hash)

	pt2 := common.OracleTypeWeChat
	id2 := mock.Hash()
	hash2 := mock.Hash()
	addTestPublishInfo(t, ctx, l, account, pt2, id2, kt, pk[:], vs, cs, fee, hash2)

	ps, _ := pkd.GetPublishInfosByAccountAndType(context.Background(), &pb.AccountAndTypeParam{
		PType:   "",
		Account: toAddressValue(account),
	})
	if len(ps.GetStates()) != 2 {
		t.Fatal()
	}

	ps, _ = pkd.GetPublishInfosByAccountAndType(context.Background(), &pb.AccountAndTypeParam{
		PType:   "email",
		Account: toAddressValue(account),
	})
	if len(ps.GetStates()) != 1 {
		t.Fatal()
	}

	ps, _ = pkd.GetPublishInfosByAccountAndType(context.Background(), &pb.AccountAndTypeParam{
		PType:   "weChat",
		Account: toAddressValue(account),
	})
	if len(ps.GetStates()) != 1 {
		t.Fatal()
	}

	ps, _ = pkd.GetPublishInfosByAccountAndType(context.Background(), &pb.AccountAndTypeParam{
		PType:   "invalid",
		Account: toAddressValue(account),
	})
	if len(ps.GetStates()) != 0 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetOracleBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)
	param := new(api.OracleParam)
	param.Account = mock.Address()
	param.Hash = "123"

	_, err := pkd.GetOracleBlock(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err == nil {
		t.Fatal()
	}

	pkd.cc.Init(nil)
	pkd.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err == nil {
		t.Fatal()
	}

	hash := mock.Hash()
	param.Hash = hash.String()
	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err == nil {
		t.Fatal()
	}

	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	param.KeyType = "ed25519"
	param.PubKey = pk.String()
	param.OType = "email"
	param.OID = "123@test.com"
	param.Code = util.RandomFixedString(common.RandomCodeLen)

	pt := common.OracleTypeEmail
	id, _ := types.Sha256HashData([]byte(param.OID))
	vs := []types.Address{param.Account}
	codeComb := append([]byte(param.PubKey), []byte(param.Code)...)
	codeHash, _ := types.Sha256HashData(codeComb)
	cs := []types.Hash{codeHash}
	fee := common.PublishCost
	addTestPublishInfo(t, ctx, l, mock.Address(), pt, id, kt, pk[:], vs, cs, fee, hash)
	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Account)
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Type = config.GasToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err == nil {
		t.Fatal()
	}

	am.Tokens[0].Balance = common.OracleCost
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err == nil {
		t.Fatal()
	}

	addTestVerifierState(t, l, 100, []types.Address{param.Account}, 1000)
	vk := mock.Hash()
	addTestVerifierInfo(t, ctx, l, param.Account, common.OracleTypeEmail, "123@test.com", vk[:])
	preBlk := mock.StateBlockWithoutWork()
	preBlk.Balance = common.OracleCost
	l.AddStateBlock(preBlk)
	am.Tokens[0].Header = preBlk.GetHash()
	am.CoinOracle = common.MinVerifierPledgeAmount
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = pkd.GetOracleBlock(context.Background(), toOracleParam(param))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublicKeyDistributionApi_GetOracleInfosByType(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	account := mock.Address()
	ot := common.OracleTypeEmail
	id := mock.Hash()
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	addTestOracleInfo(t, ctx, l, account, ot, id, kt, pk[:], code, hash)

	os, _ := pkd.GetOracleInfosByType(context.Background(), toString("weChat"))
	if len(os.GetParams()) != 0 {
		t.Fatal()
	}

	os, _ = pkd.GetOracleInfosByType(context.Background(), toString("email"))
	if len(os.GetParams()) != 1 {
		t.Fatal()
	}

	os, _ = pkd.GetOracleInfosByType(context.Background(), toString(""))
	if len(os.GetParams()) != 1 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetOracleInfosByTypeAndID(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	account := mock.Address()
	ot := common.OracleTypeEmail
	oid := "123@test.com"
	id, _ := types.Sha256HashData([]byte(oid))
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	addTestOracleInfo(t, ctx, l, account, ot, id, kt, pk[:], code, hash)

	os, _ := pkd.GetOracleInfosByTypeAndID(context.Background(), &pb.TypeAndIDParam{
		PType: "email",
		PID:   oid,
	})
	if len(os.GetParams()) != 1 {
		t.Fatal()
	}

	os, _ = pkd.GetOracleInfosByTypeAndID(context.Background(), &pb.TypeAndIDParam{
		PType: "email",
		PID:   "invalid",
	})
	if len(os.GetParams()) != 0 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetOracleInfosByAccountAndType(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	account := mock.Address()
	ot := common.OracleTypeEmail
	oid := "123@test.com"
	id, _ := types.Sha256HashData([]byte(oid))
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	hash := mock.Hash()
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()
	addTestOracleInfo(t, ctx, l, account, ot, id, kt, pk[:], code, hash)

	os, _ := pkd.GetOracleInfosByAccountAndType(context.Background(), &pb.AccountAndTypeParam{
		PType:   "email",
		Account: toAddressValue(account),
	})
	if len(os.GetParams()) != 1 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetOracleInfosByHash(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	account := mock.Address()
	ot := common.OracleTypeEmail
	oid := "123@test.com"
	id, _ := types.Sha256HashData([]byte(oid))
	code := util.RandomFixedStringWithSeed(common.RandomCodeLen, time.Now().UnixNano())
	kt := common.PublicKeyTypeED25519
	pk := mock.Hash()

	publishPrev := mock.StateBlockWithoutWork()
	l.AddStateBlock(publishPrev)

	publish := mock.StateBlockWithoutWork()
	vs := []types.Address{mock.Address()}
	cs := []types.Hash{mock.Hash()}
	publish.Data, _ = abi.PublicKeyDistributionABI.PackMethod(abi.MethodNamePKDPublish, ot, id, kt, pk[:], vs, cs, common.PublishCost.Int)
	publish.Previous = publishPrev.GetHash()
	l.AddStateBlock(publish)

	hash := publish.Previous
	addTestOracleInfo(t, ctx, l, account, ot, id, kt, pk[:], code, hash)

	os, _ := pkd.GetOracleInfosByHash(context.Background(), toHash(hash))
	if len(os.GetParams()) != 1 {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_GetVerifierHeartBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l, &contractaddress.PubKeyDistributionAddress)
	pkd := NewPublicKeyDistributionAPI(cfgFile, l)

	var vt []string
	account := mock.Address()

	blk, _ := pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk != nil {
		t.Fatal()
	}

	vt = append(vt, "email")
	blk, _ = pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk != nil {
		t.Fatal()
	}

	pkd.cc.Init(nil)
	pkd.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	blk, _ = pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(account)
	am.CoinOracle = common.MinVerifierPledgeAmount
	l.AddAccountMeta(am, l.Cache().GetCache())
	blk, _ = pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk != nil {
		t.Fatal()
	}

	vk := mock.Hash()
	addTestVerifierInfo(t, ctx, l, account, common.OracleTypeEmail, "123@test.com", vk[:])
	blk, _ = pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk != nil {
		t.Fatal()
	}

	am.Tokens[0].Type = config.GasToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	blk, _ = pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk != nil {
		t.Fatal()
	}

	am.Tokens[0].Balance = common.OracleCost
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	blk, _ = pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk != nil {
		t.Fatal()
	}

	addTestVerifierState(t, l, 100, []types.Address{mock.Address()}, 1000)
	blk, _ = pkd.GetVerifierHeartBlock(context.Background(), &pb.VerifierHeartBlockRequest{
		Account: toAddressValue(account),
		VTypes:  vt,
	})
	if blk == nil {
		t.Fatal()
	}
}

func TestPublicKeyDistributionApi_Reward(t *testing.T) {
	tearDone, md := setupTestCasePov(t)
	defer tearDone(t)

	md.eb.Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(10 * time.Millisecond)

	pkd := NewPublicKeyDistributionAPI(md.cc.ConfigFile(), md.l)

	account := mock.Account()

	// mock account meta
	am := mock.AccountMeta(account.Address())
	am.Tokens = append(am.Tokens, mock.TokenMeta2(account.Address(), config.GasToken()))
	md.l.AddAccountMeta(am, md.l.Cache().GetCache())

	// mock trie state in global db
	gsdb := statedb.NewPovGlobalStateDB(md.l.DBStore(), types.ZeroHash)
	csdb, _ := gsdb.LookupContractStateDB(contractaddress.PubKeyDistributionAddress)

	ps := types.NewPovVerifierState()
	ps.TotalReward = types.NewBigNumFromInt(100000000)
	dpki.PovSetVerifierState(csdb, account.Address().Bytes(), ps)

	gsdb.CommitToTrie()
	txn := md.l.DBStore().Batch(true)
	gsdb.CommitToDB(txn)
	err := md.l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	psMockBlk1, psMockTd1 := mock.GeneratePovBlock(nil, 0)
	psMockBlk1.Header.BasHdr.Height = 1439
	psMockBlk1.Header.CbTx.StateHash = gsdb.GetCurHash()

	mock.UpdatePovHash(psMockBlk1)

	err = md.l.AddPovBlock(psMockBlk1, psMockTd1)
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.AddPovBestHash(psMockBlk1.GetHeight(), psMockBlk1.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.SetPovLatestHeight(psMockBlk1.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	psMockBlk2, psMockTd2 := mock.GeneratePovBlock(psMockBlk1, 0)
	psMockBlk2.Header.BasHdr.Height = 4320
	psMockBlk2.Header.CbTx.StateHash = gsdb.GetCurHash()

	mock.UpdatePovHash(psMockBlk2)

	err = md.l.AddPovBlock(psMockBlk2, psMockTd2)
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.AddPovBestHash(psMockBlk2.GetHeight(), psMockBlk2.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	err = md.l.SetPovLatestHeight(psMockBlk2.GetHeight())
	if err != nil {
		t.Fatal(err)
	}

	param := new(api.PKDRewardParam)
	param.Account = account.Address()
	param.Beneficial = param.Account
	param.EndHeight = 1439
	param.RewardAmount = big.NewInt(100000000)

	data, err := pkd.PackRewardData(context.Background(), &pb.PKDRewardParam{
		Account:      toAddressValue(param.Account),
		Beneficial:   toAddressValue(param.Beneficial),
		EndHeight:    param.EndHeight,
		RewardAmount: param.RewardAmount.Int64(),
	})
	if err != nil {
		t.Fatal(err)
	}
	param2, err := pkd.UnpackRewardData(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}

	sendBlk, err := pkd.GetRewardSendBlock(context.Background(), param2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pkd.GetRewardRecvBlock(context.Background(), sendBlk)
	if err != nil {
		t.Fatal(err)
	}

	_, _ = pkd.GetRewardHistory(context.Background(), toAddress(account.Address()))
	_, _ = pkd.GetAvailRewardInfo(context.Background(), toAddress(account.Address()))
}
