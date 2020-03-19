/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/json"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract"
)

func setupNEP5PledgeAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *NEP5PledgeAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewNEP5PledgeAPI(cc.ConfigFile(), l)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(MockBlocks), &blocks); err != nil {
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

func TestNewNEP5PledgeAPI(t *testing.T) {
	testcase, verifier, api := setupNEP5PledgeAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	a2 := account2.Address()
	contract.SetPledgeTime(0, 0, 0, 0, 0, 1)
	param := &PledgeParam{
		Beneficial:    a2,
		PledgeAddress: a1,
		Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
		PType:         "vote",
		NEP5TxId:      mock.Hash().String(),
	}

	if data, err := api.GetPledgeData(param); err != nil {
		t.Fatal(err)
	} else if len(data) == 0 {
		t.Fatal("invalid GetPledgeData")
	}

	if blk, err := api.GetPledgeBlock(param); err != nil {
		t.Fatal(err)
	} else {
		txHash := blk.GetHash()
		blk.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(blk); err != nil {
			t.Fatal(err)
		}

		if _, err := api.GetPledgeRewardBlock(blk); err != nil {
			t.Fatal(err)
		}

		if txBlk, err := api.GetPledgeRewardBlockBySendHash(txHash); err != nil {
			t.Fatal(err)
		} else {
			txHash := txBlk.GetHash()
			txBlk.Signature = account2.Sign(txHash)
			if err := verifier.BlockProcess(txBlk); err != nil {
				t.Fatal(err)
			}

			if _, err := api.ParsePledgeInfo(txBlk.Data); err != nil {
				t.Fatal(err)
			}
		}

		time.Sleep(2 * time.Second)

		w := &WithdrawPledgeParam{
			Beneficial: a2,
			Amount:     param.Amount,
			PType:      param.PType,
			NEP5TxId:   param.NEP5TxId,
		}

		if info, err := api.GetPledgeInfo(w); err != nil {
			t.Fatal(err)
		} else {
			t.Log(info)
		}

		if info, err := api.GetPledgeInfoWithNEP5TxId(w); err != nil {
			t.Fatal(err)
		} else {
			t.Log(info)
		}
		if info, err := api.GetPledgeInfoWithTimeExpired(w); err != nil {
			t.Fatal(err)
		} else {
			t.Log(info)
		}

		if data, err := api.GetWithdrawPledgeData(w); err != nil {
			t.Fatal(err)
		} else if len(data) == 0 {
			t.Fatal("invalid GetWithdrawPledgeData")
		}

		if blk, err := api.GetWithdrawPledgeBlock(w); err != nil {
			t.Fatal(err)
		} else {
			txHash := blk.GetHash()
			blk.Signature = account2.Sign(txHash)
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}

			if _, err := api.GetWithdrawRewardBlock(blk); err != nil {
				t.Fatal(err)
			}

			if rxBlk, err := api.GetWithdrawRewardBlockBySendHash(txHash); err != nil {
				t.Fatal(err)
			} else {
				txHash := rxBlk.GetHash()
				rxBlk.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(rxBlk); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	if info := api.GetPledgeInfosByPledgeAddress(a1); info == nil {
		t.Fatal()
	} else {
		t.Log(info)
	}

	if amount, err := api.GetPledgeBeneficialTotalAmount(a2); err != nil {
		t.Fatal(err)
	} else if amount.Cmp(big.NewInt(0)) == 0 {
		t.Fatal("invalid amount")
	} else {
		t.Log("amount: ", amount)
	}

	if info := api.GetBeneficialPledgeInfosByAddress(a2); info == nil {
		t.Fatal()
	} else {
		t.Log(info)
	}

	if infos, err := api.GetBeneficialPledgeInfos(a2, "vote"); err != nil || infos == nil {
		t.Fatal(err)
	} else if len(infos.PledgeInfo) == 0 {
		t.Fatal("invalid pledge info")
	}

	if amount, err := api.GetPledgeBeneficialAmount(a2, "vote"); err != nil {
		t.Fatal(err)
	} else if amount.Cmp(big.NewInt(0)) == 0 {
		t.Fatal("invalid amount")
	} else {
		t.Log("amount: ", amount)
	}

	if amount, err := api.GetTotalPledgeAmount(); err != nil {
		t.Fatal(err)
	} else if amount.Cmp(big.NewInt(0)) == 0 {
		t.Fatal("invalid amount")
	} else {
		t.Log("amount: ", amount)
	}

	if info, err := api.GetAllPledgeInfo(); err != nil {
		t.Fatal(err)
	} else if len(info) == 0 {
		t.Fatal("GetAllPledgeInfo")
	}
}

func TestNEP5PledgeAPI_GetPledgeBlock(t *testing.T) {
	testcase, _, api := setupNEP5PledgeAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	param := &PledgeParam{
		Beneficial:    a2,
		PledgeAddress: a1,
		Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
		PType:         "vote",
		NEP5TxId:      mock.Hash().String(),
	}

	type fields struct {
		logger   *zap.SugaredLogger
		l        ledger.Store
		pledge   *contract.Nep5Pledge
		withdraw *contract.WithdrawNep5Pledge
		cc       *qlcchainctx.ChainContext
	}
	type args struct {
		param *PledgeParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: param,
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: &PledgeParam{
					Beneficial:    a2,
					PledgeAddress: mock.Address(),
					Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
					PType:         "vote",
					NEP5TxId:      mock.Hash().String(),
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: &PledgeParam{
					Beneficial:    a2,
					PledgeAddress: a1,
					Amount:        types.Balance{Int: big.NewInt(math.MaxInt64)},
					PType:         "vote",
					NEP5TxId:      mock.Hash().String(),
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: nil,
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "f4",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: &PledgeParam{
					Beneficial:    types.ZeroAddress,
					PledgeAddress: a1,
					Amount:        types.Balance{Int: big.NewInt(math.MaxInt64)},
					PType:         "vote",
					NEP5TxId:      mock.Hash().String(),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &NEP5PledgeAPI{
				logger:   tt.fields.logger,
				l:        tt.fields.l,
				pledge:   tt.fields.pledge,
				withdraw: tt.fields.withdraw,
				cc:       tt.fields.cc,
			}
			_, err := p.GetPledgeBlock(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPledgeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetPledgeData() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestNEP5PledgeAPI_GetPledgeData(t *testing.T) {
	testcase, _, api := setupNEP5PledgeAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	param := &PledgeParam{
		Beneficial:    a2,
		PledgeAddress: a1,
		Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
		PType:         "vote",
		NEP5TxId:      mock.Hash().String(),
	}

	type fields struct {
		logger   *zap.SugaredLogger
		l        ledger.Store
		pledge   *contract.Nep5Pledge
		withdraw *contract.WithdrawNep5Pledge
		cc       *qlcchainctx.ChainContext
	}
	type args struct {
		param *PledgeParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: param,
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: nil,
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "f2",
			fields: fields{
				logger:   api.logger,
				l:        api.l,
				pledge:   api.pledge,
				withdraw: api.withdraw,
				cc:       api.cc,
			},
			args: args{
				param: &PledgeParam{
					Beneficial:    a2,
					PledgeAddress: a1,
					Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
					PType:         "invalid",
					NEP5TxId:      mock.Hash().String(),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &NEP5PledgeAPI{
				logger:   tt.fields.logger,
				l:        tt.fields.l,
				pledge:   tt.fields.pledge,
				withdraw: tt.fields.withdraw,
				cc:       tt.fields.cc,
			}
			_, err := p.GetPledgeData(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPledgeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetPledgeData() got = %v, want %v", got, tt.want)
			//}
		})
	}
}
