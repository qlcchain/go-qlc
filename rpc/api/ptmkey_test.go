package api

import (
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
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
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
	pa := NewPtmKeyApi(cfgFile, l)
	tests := []struct {
		name string
		args args
		want *PtmKeyApi
	}{
		// TODO: Add test cases.
		{"OK", args{cfgFile, l}, pa},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPtmKeyApi(tt.args.cfgFile, tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPtmKeyApi() = %v, want %v", got, tt.want)
			}
		})
	}
}
func addTestPtmKey(t *testing.T, store ledger.Store, ctx *vmstore.VMContext, account types.Address, vBtype uint16, vKey []byte) error {
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

	pa := NewPtmKeyApi(cfgFile, l)
	account := mock.Address()
	account2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pk := &PtmKeyUpdateParam{
		account,
		btype,
		key,
	}
	pk2 := &PtmKeyUpdateParam{
		account2,
		btype,
		key,
	}

	type fields struct {
		logger *zap.SugaredLogger
		l      ledger.Store
		cc     *chainctx.ChainContext
		ctx    *vmstore.VMContext
		pu     *contract.PtmKeyUpdate
		pdb    *contract.PtmKeyDelete
	}
	type args struct {
		param *PtmKeyUpdateParam
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		//want    *types.StateBlock
		wantErr bool
	}{
		// TODO: Add test cases.
		{"badaccount", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pk2}, true},
		{"povsysnfailed", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pk}, true},
		{"povheadfailed", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pk}, true},
		{"accounttokenfailed", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pk}, true},
		{"OK", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pk}, false},
	}
	step := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PtmKeyApi{
				logger: tt.fields.logger,
				l:      tt.fields.l,
				cc:     tt.fields.cc,
				ctx:    tt.fields.ctx,
				pu:     tt.fields.pu,
				pdb:    tt.fields.pdb,
			}

			switch step {
			case 2:
				pa.cc.Init(nil)
				pa.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
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
			_, err := p.GetPtmKeyUpdateBlock(tt.args.param)
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

	pa := NewPtmKeyApi(cfgFile, l)
	addr1 := mock.Address()
	addr2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	btypeint := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pd1 := &PtmKeyDeleteParam{
		addr1,
		btype,
	}
	pd2 := &PtmKeyDeleteParam{
		addr2,
		btype,
	}
	type fields struct {
		logger *zap.SugaredLogger
		l      ledger.Store
		cc     *chainctx.ChainContext
		ctx    *vmstore.VMContext
		pu     *contract.PtmKeyUpdate
		pdb    *contract.PtmKeyDelete
	}
	type args struct {
		param *PtmKeyDeleteParam
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		//want    *types.StateBlock
		wantErr bool
	}{
		// TODO: Add test cases.
		{"badaccount", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pd2}, true},
		{"povsysnfailed", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pd1}, true},
		{"povheadfailed", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pd1}, true},
		{"accounttokenfailed", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pd1}, true},
		{"getblocknil", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pd1}, true},
		{"OK", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{pd1}, false},
	}
	step := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PtmKeyApi{
				logger: tt.fields.logger,
				l:      tt.fields.l,
				cc:     tt.fields.cc,
				ctx:    tt.fields.ctx,
				pu:     tt.fields.pu,
				pdb:    tt.fields.pdb,
			}
			switch step {
			case 2:
				pa.cc.Init(nil)
				pa.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
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
			case 5:
				addTestPtmKey(t, l, pa.ctx, addr1, btypeint, []byte(key))
			}
			step = step + 1
			_, err := p.GetPtmKeyDeleteBlock(tt.args.param)
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

	pa := NewPtmKeyApi(cfgFile, l)
	account := mock.Address()
	account2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	btypeint := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pks := make([]*PtmKeyUpdateParam, 0)
	pk := &PtmKeyUpdateParam{
		account,
		btype,
		key,
	}
	pks = append(pks, pk)

	type fields struct {
		logger *zap.SugaredLogger
		l      ledger.Store
		cc     *chainctx.ChainContext
		ctx    *vmstore.VMContext
		pu     *contract.PtmKeyUpdate
		pdb    *contract.PtmKeyDelete
	}
	type args struct {
		account types.Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*PtmKeyUpdateParam
		wantErr bool
	}{
		// TODO: Add test cases.
		{"OK", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{account}, pks, false},
		{"badaccount", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{account2}, nil, true},
	}
	addTestPtmKey(t, l, pa.ctx, account, btypeint, []byte(key))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PtmKeyApi{
				logger: tt.fields.logger,
				l:      tt.fields.l,
				cc:     tt.fields.cc,
				ctx:    tt.fields.ctx,
				pu:     tt.fields.pu,
				pdb:    tt.fields.pdb,
			}
			got, err := p.GetPtmKeyByAccount(tt.args.account)
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccount() = %v, want %v", got, tt.want)
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

	pa := NewPtmKeyApi(cfgFile, l)
	account := mock.Address()
	account2 := mock.Address()
	btype := common.PtmKeyVBtypeStrDefault
	btypeint := common.PtmKeyVBtypeDefault
	key := "/vkgO5TfnsvKZGDc2KT1yxD5fxGNre65SPPuh3hyg0M="
	pks := make([]*PtmKeyUpdateParam, 0)
	pk := &PtmKeyUpdateParam{
		account,
		btype,
		key,
	}
	pks = append(pks, pk)

	type fields struct {
		logger *zap.SugaredLogger
		l      ledger.Store
		cc     *chainctx.ChainContext
		ctx    *vmstore.VMContext
		pu     *contract.PtmKeyUpdate
		pdb    *contract.PtmKeyDelete
	}
	type args struct {
		account types.Address
		Btype   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*PtmKeyUpdateParam
		wantErr bool
	}{
		// TODO: Add test cases.
		{"OK", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{account, btype}, pks, false},
		{"badaccount", fields{pa.logger, pa.l, pa.cc, pa.ctx, pa.pu, pa.pdb}, args{account2, btype}, nil, true},
	}
	addTestPtmKey(t, l, pa.ctx, account, btypeint, []byte(key))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PtmKeyApi{
				logger: tt.fields.logger,
				l:      tt.fields.l,
				cc:     tt.fields.cc,
				ctx:    tt.fields.ctx,
				pu:     tt.fields.pu,
				pdb:    tt.fields.pdb,
			}
			got, err := p.GetPtmKeyByAccountAndBtype(tt.args.account, tt.args.Btype)
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccountAndBtype() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtmKeyApi.GetPtmKeyByAccountAndBtype() = %v, want %v", got, tt.want)
			}
		})
	}
}
