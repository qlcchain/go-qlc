package apis

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
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
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestNewPtmKeyApi(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()
	type args struct {
		cfgFile string
		l       ledger.Store
	}
	pa := NewPtmKeyAPI(cfgFile, l)
	tests := []struct {
		name string
		args args
		want *PtmKeyAPI
	}{
		// TODO: Add test cases.
		{"OK", args{cfgFile, l}, pa},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPtmKeyAPI(tt.args.cfgFile, tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPtmKeyApi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func addTestPtmKey(store ledger.Store, ctx *vmstore.VMContext, account types.Address, vBtype uint16, vKey []byte) error {
	data, err := abi.PtmKeyABI.PackVariable(abi.VariableNamePtmKeyStorageVar, string(vKey[:]), true)
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

func TestPtmKeyApi_GetPtmKeyUpdateBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	account := mock.Address()
	account2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pk := &api.PtmKeyUpdateParam{
		Account: account,
		Btype:   btype,
		Pubkey:  key,
	}
	pk2 := &api.PtmKeyUpdateParam{
		Account: account2,
		Btype:   btype,
		Pubkey:  key,
	}

	type args struct {
		param *api.PtmKeyUpdateParam
	}
	tests := []struct {
		name string
		args args
		//want    *types.StateBlock
		wantErr bool
	}{
		// TODO: Add test cases.
		{"badaccount", args{pk2}, true},
		{"povsysnfailed", args{pk}, true},
		{"povheadfailed", args{pk}, true},
		{"accounttokenfailed", args{pk}, true},
		{"OK", args{pk}, false},
	}
	step := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPtmKeyAPI(cfgFile, l)
			cc := chainctx.NewChainContext(cfgFile)
			switch step {
			case 2:
				cc.Init(nil)
				cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
				time.Sleep(time.Second)
				//t.Errorf("step 1")
			case 3:
				pb, td := mock.GeneratePovBlock(nil, 0)
				l.AddPovBlock(pb, td)
				l.SetPovLatestHeight(pb.Header.BasHdr.Height)
				l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
				//t.Errorf("step 2")
			case 4:
				am := mock.AccountMeta(account)
				am.CoinOracle = common.MinVerifierPledgeAmount
				l.AddAccountMeta(am, l.Cache().GetCache())
				am.Tokens[0].Type = config.ChainToken()
				l.UpdateAccountMeta(am, l.Cache().GetCache())
				//t.Errorf("step 3")
			}
			step = step + 1
			_, err := p.GetPtmKeyUpdateBlock(context.Background(), toPtmKeyUpdateParam(tt.args.param))
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyUpdateBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPtmKeyApi_GetPtmKeyDeleteBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	addr1 := mock.Address()
	addr2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	pd1 := &api.PtmKeyDeleteParam{
		Account: addr1,
		Btype:   btype,
	}
	pd2 := &api.PtmKeyDeleteParam{
		Account: addr2,
		Btype:   btype,
	}
	type args struct {
		param *api.PtmKeyDeleteParam
	}
	tests := []struct {
		name string
		args args
		//want    *types.StateBlock
		wantErr bool
	}{
		// TODO: Add test cases.
		{"badaccount", args{pd2}, true},
		{"povsysnfailed", args{pd1}, true},
		{"povheadfailed", args{pd1}, true},
		{"accounttokenfailed", args{pd1}, true},
		{"getblocknil", args{pd1}, true},
		//{"OK", args{pd1}, false},
	}
	step := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPtmKeyAPI(cfgFile, l)
			cc := chainctx.NewChainContext(cfgFile)

			switch step {
			case 2:
				cc.Init(nil)
				cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
				time.Sleep(time.Second)
				//t.Errorf("step 1")
			case 3:
				pb, td := mock.GeneratePovBlock(nil, 0)
				l.AddPovBlock(pb, td)
				l.SetPovLatestHeight(pb.Header.BasHdr.Height)
				l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
				//t.Errorf("step 2")
			case 4:
				am := mock.AccountMeta(addr1)
				am.CoinOracle = common.MinVerifierPledgeAmount
				l.AddAccountMeta(am, l.Cache().GetCache())
				am.Tokens[0].Type = config.ChainToken()
				l.UpdateAccountMeta(am, l.Cache().GetCache())
			}
			step = step + 1
			_, err := p.GetPtmKeyDeleteBlock(context.Background(), &pb.PtmKeyByAccountAndBtypeParam{
				Account: toAddressValue(tt.args.param.Account),
				Btype:   tt.args.param.Btype,
			})
			fmt.Println("===", err, err != nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyDeleteBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPtmKeyApi_GetPtmKeyByAccount(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()
	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)

	account := mock.Address()
	//account2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	btypeint := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pks := make([]*api.PtmKeyUpdateParam, 0)
	pk := &api.PtmKeyUpdateParam{
		Account: account,
		Btype:   btype,
		Pubkey:  key,
	}
	pks = append(pks, pk)

	type args struct {
		account types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*api.PtmKeyUpdateParam
		wantErr bool
	}{
		// TODO: Add test cases.
		{"OK", args{account}, pks, false},
		//{"badaccount", args{account2}, nil, true},
	}
	addTestPtmKey(l, ctx, account, btypeint, []byte(key))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPtmKeyAPI(cfgFile, l)

			got, err := p.GetPtmKeyByAccount(context.Background(), toAddress(tt.args.account))
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			rt, err := toOriginPtmKeyUpdateParams(got)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(rt, tt.want) {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccount() = %v, want %v", rt, tt.want)
			}
		})
	}
}

func TestPtmKeyApi_GetPtmKeyByAccountAndBtype(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()
	ctx := vmstore.NewVMContext(l, &contractaddress.PtmKeyKVAddress)

	account := mock.Address()
	//account2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	btypeint := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pks := make([]*api.PtmKeyUpdateParam, 0)
	pk := &api.PtmKeyUpdateParam{
		Account: account,
		Btype:   btype,
		Pubkey:  key,
	}
	pks = append(pks, pk)

	type args struct {
		account types.Address
		Btype   string
	}
	tests := []struct {
		name    string
		args    args
		want    []*api.PtmKeyUpdateParam
		wantErr bool
	}{
		// TODO: Add test cases.
		{"OK", args{account, btype}, pks, false},
		//{"badaccount", args{account2, btype}, nil, true},
	}
	addTestPtmKey(l, ctx, account, btypeint, []byte(key))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPtmKeyAPI(cfgFile, l)

			got, err := p.GetPtmKeyByAccountAndBtype(context.Background(), &pb.PtmKeyByAccountAndBtypeParam{
				Account: toAddressValue(tt.args.account),
				Btype:   tt.args.Btype,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccountAndBtype() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			rt, err := toOriginPtmKeyUpdateParams(got)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(rt, tt.want) {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccountAndBtype() = %v, want %v", rt, tt.want)
			}
		})
	}
}

func toOriginPtmKeyUpdateParams(params *pb.PtmKeyUpdateParams) ([]*api.PtmKeyUpdateParam, error) {
	rs := make([]*api.PtmKeyUpdateParam, 0)
	for _, rp := range params.GetParams() {
		addr, err := toOriginAddressByValue(rp.GetAccount())
		if err != nil {
			return nil, err
		}
		rt := &api.PtmKeyUpdateParam{
			Account: addr,
			Btype:   rp.GetBtype(),
			Pubkey:  rp.GetPubkey(),
		}
		rs = append(rs, rt)
	}
	return rs, nil
}
