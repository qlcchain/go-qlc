/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

var (
	// qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq
	priv1, _ = hex.DecodeString("7098c089e66bd66476e3b88df8699bcd4dacdd5e1e5b41b3c598a8a36d851184d992a03b7326b7041f689ae727292d761b329a960f3e4335e0a7dcf2c43c4bcf")
	// qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te
	priv2, _ = hex.DecodeString("31ee4e16826569dc631b969e71bd4c46d5c0df0daeca6933f46586f36f49537cd929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393")
	account1 = types.NewAccount(priv1)
	account2 = types.NewAccount(priv2)
)

func setupSettlementAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *SettlementAPI) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)
	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewSettlement(l, cc)

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

func Test_calculateRange(t *testing.T) {
	type args struct {
		size   int
		count  int
		offset *int
	}
	tests := []struct {
		name      string
		args      args
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{
			name: "f1",
			args: args{
				size:   2,
				count:  2,
				offset: offset(0),
			},
			wantStart: 0,
			wantEnd:   2,
			wantErr:   false,
		}, {
			name: "f2",
			args: args{
				size:   2,
				count:  10,
				offset: nil,
			},
			wantStart: 0,
			wantEnd:   2,
			wantErr:   false,
		}, {
			name: "overflow",
			args: args{
				size:   2,
				count:  10,
				offset: offset(2),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f3",
			args: args{
				size:   2,
				count:  10,
				offset: offset(1),
			},
			wantStart: 1,
			wantEnd:   2,
			wantErr:   false,
		}, {
			name: "f4",
			args: args{
				size:   2,
				count:  0,
				offset: offset(1),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f5",
			args: args{
				size:   2,
				count:  0,
				offset: offset(-1),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f6",
			args: args{
				size:   2,
				count:  -1,
				offset: offset(0),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		}, {
			name: "f7",
			args: args{
				size:   10,
				count:  3,
				offset: offset(3),
			},
			wantStart: 3,
			wantEnd:   6,
			wantErr:   false,
		}, {
			name: "f8",
			args: args{
				size:   0,
				count:  3,
				offset: offset(3),
			},
			wantStart: 0,
			wantEnd:   0,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, err := calculateRange(tt.args.size, tt.args.count, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStart != tt.wantStart {
				t.Errorf("calculateRange() gotStart = %v, want %v", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("calculateRange() gotEnd = %v, want %v", gotEnd, tt.wantEnd)
			}
		})
	}
}

func offset(o int) *int {
	return &o
}

func TestSettlementAPI_GetCreateContractBlock(t *testing.T) {
	testcase, verifier, api := setupSettlementAPI(t)
	defer testcase(t)

	if blk, err := api.GetCreateContractBlock(&CreateContractParam{
		PartyA: cabi.Contractor{
			Address: account1.Address(),
			Name:    "PCCWG",
		},
		PartyB: cabi.Contractor{
			Address: account2.Address(),
			Name:    "HTK-CSL",
		},
		Services: []cabi.ContractService{{
			ServiceId:   mock.Hash().String(),
			Mcc:         1,
			Mnc:         2,
			TotalAmount: 10,
			UnitPrice:   2,
			Currency:    "USD",
		}, {
			ServiceId:   mock.Hash().String(),
			Mcc:         22,
			Mnc:         1,
			TotalAmount: 30,
			UnitPrice:   4,
			Currency:    "USD",
		}},
		StartDate: time.Now().AddDate(0, 0, 1).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 1).Unix(),
	}); err != nil {
		t.Fatal(err)
	} else {
		t.Log(blk)
		txHash := blk.GetHash()
		blk.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(blk); err != nil {
			t.Fatal(err)
		}
		if rx, err := api.GetSettlementRewardsBlock(&txHash); err != nil {
			t.Fatal(err)
		} else {
			t.Log(rx)
		}
	}
}
