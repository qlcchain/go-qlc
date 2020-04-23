package api

import (
	"reflect"
	"testing"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

func TestNewPtmKeyApi(t *testing.T) {
	type args struct {
		cfgFile string
		l       ledger.Store
	}
	tests := []struct {
		name string
		args args
		want *PtmKeyApi
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPtmKeyApi(tt.args.cfgFile, tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPtmKeyApi() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPtmKeyApi_GetPtmKeyUpdateBlock(t *testing.T) {
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
		name    string
		fields  fields
		args    args
		want    *types.StateBlock
		wantErr bool
	}{
		// TODO: Add test cases.
	}
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
			got, err := p.GetPtmKeyUpdateBlock(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyUpdateBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtmKeyApi.GetPtmKeyUpdateBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPtmKeyApi_GetPtmKeyDeleteBlock(t *testing.T) {
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
		name    string
		fields  fields
		args    args
		want    *types.StateBlock
		wantErr bool
	}{
		// TODO: Add test cases.
	}
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
			got, err := p.GetPtmKeyDeleteBlock(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("PtmKeyApi.GetPtmKeyDeleteBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PtmKeyApi.GetPtmKeyDeleteBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPtmKeyApi_GetPtmKeyByAccount(t *testing.T) {
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
	}
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
	}
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
