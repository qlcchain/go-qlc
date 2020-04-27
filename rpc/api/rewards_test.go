/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract"
)

func setupRewardsTestCase(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *RewardsAPI) {
	//t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := chainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewRewardsAPI(l, cc)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	for i := range blocks {
		block := blocks[i]
		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
	}

	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
	}, verifier, api
}

func TestRewardsAPI_GetUnsignedRewardData(t *testing.T) {
	teardownTestCase, _, api := setupRewardsTestCase(t)
	defer teardownTestCase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	type fields struct {
		logger           *zap.SugaredLogger
		ledger           ledger.Store
		rewards          *contract.AirdropRewards
		confidantRewards *contract.ConfidantRewards
		cc               *chainctx.ChainContext
	}
	type args struct {
		param *RewardsParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.Hash
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				logger:           api.logger,
				ledger:           api.ledger,
				rewards:          api.rewards,
				confidantRewards: api.confidantRewards,
				cc:               api.cc,
			},
			args: args{
				param: &RewardsParam{
					Id:     mock.Hash().String(),
					Amount: types.Balance{Int: big.NewInt(100)},
					Self:   a1,
					To:     a2,
				},
			},
			want:    types.Hash{},
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				logger:           api.logger,
				ledger:           api.ledger,
				rewards:          api.rewards,
				confidantRewards: api.confidantRewards,
				cc:               api.cc,
			},
			args: args{
				param: nil,
			},
			want:    types.Hash{},
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				logger:           api.logger,
				ledger:           api.ledger,
				rewards:          api.rewards,
				confidantRewards: api.confidantRewards,
				cc:               api.cc,
			},
			args: args{
				param: &RewardsParam{
					Id:     "",
					Amount: types.Balance{Int: big.NewInt(100)},
					Self:   a1,
					To:     a2,
				},
			},
			want:    types.Hash{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RewardsAPI{
				logger:           tt.fields.logger,
				ledger:           tt.fields.ledger,
				rewards:          tt.fields.rewards,
				confidantRewards: tt.fields.confidantRewards,
				cc:               tt.fields.cc,
			}
			_, err := r.GetUnsignedRewardData(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUnsignedRewardData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetUnsignedRewardData() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestRewardsAPI_GetUnsignedConfidantData(t *testing.T) {
	teardownTestCase, _, api := setupRewardsTestCase(t)
	defer teardownTestCase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	type fields struct {
		logger           *zap.SugaredLogger
		ledger           ledger.Store
		rewards          *contract.AirdropRewards
		confidantRewards *contract.ConfidantRewards
		cc               *chainctx.ChainContext
	}
	type args struct {
		param *RewardsParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.Hash
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				logger:           api.logger,
				ledger:           api.ledger,
				rewards:          api.rewards,
				confidantRewards: api.confidantRewards,
				cc:               api.cc,
			},
			args: args{
				param: &RewardsParam{
					Id:     mock.Hash().String(),
					Amount: types.Balance{Int: big.NewInt(100)},
					Self:   a1,
					To:     a2,
				},
			},
			want:    types.Hash{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RewardsAPI{
				logger:           tt.fields.logger,
				ledger:           tt.fields.ledger,
				rewards:          tt.fields.rewards,
				confidantRewards: tt.fields.confidantRewards,
				cc:               tt.fields.cc,
			}
			_, err := r.GetUnsignedConfidantData(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUnsignedConfidantData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetUnsignedConfidantData() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestNewRewardsAPI(t *testing.T) {
	teardownTestCase, verifier, api := setupRewardsTestCase(t)
	defer teardownTestCase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	r := &RewardsParam{
		Id:     mock.Hash().String(),
		Amount: types.Balance{Int: big.NewInt(100)},
		Self:   a1,
		To:     a2,
	}

	if h, err := api.GetUnsignedRewardData(r); err != nil {
		t.Fatal(err)
	} else {
		sign := account1.Sign(h)
		if blk, err := api.GetSendRewardBlock(r, &sign); err != nil {
			t.Fatal(err)
		} else {
			txHash := blk.GetHash()
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}

			if rxBlk, err := api.GetReceiveRewardBlock(&txHash); err != nil {
				t.Fatal(err)
			} else {
				if err := verifier.BlockProcess(rxBlk); err != nil {
					t.Fatal(err)
				}

				if b := api.IsAirdropRewards(blk.Data); !b {
					t.Fatal("IsAirdropRewards failed...")
				}
			}
		}
	}

	r2 := &RewardsParam{
		Id:     mock.Hash().String(),
		Amount: types.Balance{Int: big.NewInt(1000)},
		Self:   a1,
		To:     a2,
	}

	if h, err := api.GetUnsignedConfidantData(r2); err != nil {
		t.Fatal(err)
	} else {
		sign := account1.Sign(h)
		if blk, err := api.GetSendConfidantBlock(r2, &sign); err != nil {
			t.Fatal(err)
		} else {
			txHash := blk.GetHash()
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}

			if rxBlk, err := api.GetReceiveRewardBlock(&txHash); err != nil {
				t.Fatal(err)
			} else {
				if err := verifier.BlockProcess(rxBlk); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	if amount, err := api.GetTotalRewards(r.Id); err != nil {
		t.Fatal(err)
	} else if amount.Cmp(r.Amount.Int) != 0 {
		t.Fatalf("invalid amount, exp: %d, act: %d", r.Amount, amount)
	}

	if detail, err := api.GetRewardsDetail(r.Id); err != nil {
		t.Fatal(err)
	} else if len(detail) != 1 {
		t.Fatal("invalid rewards detail")
	}

	if rewards, err := api.GetConfidantRewards(r2.To); err != nil {
		t.Fatal(err)
	} else if len(rewards) == 0 {
		t.Fatal("invalid confidant rewards...")
	} else {
		t.Log(rewards)
	}

	if detail, err := api.GetConfidantRewordsDetail(r2.To); err != nil {
		t.Fatal(err)
	} else if len(detail) != 1 {
		t.Fatal("invalid confidant detail")
	}

}
