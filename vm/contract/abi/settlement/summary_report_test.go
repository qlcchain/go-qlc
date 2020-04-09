/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestNewSummaryResult(t *testing.T) {
	r := newSummaryResult()

	for i := 0; i < 20; i++ {
		r.UpdateState("WeChat", "partyA", i%3 == 0, i%2 == 0)
		r.UpdateState("WeChat", "partyB", i%2 == 0, i%3 == 0)
		r.UpdateState("Slack", "partyA", i%2 == 0, i%3 == 0)
		r.UpdateState("Slack", "partyB", i%3 == 0, i%2 == 0)
	}

	r.DoCalculate()
	t.Log(r.String())
}

func TestMultiPartySummaryResult_UpdateState(t *testing.T) {
	r := newMultiPartySummaryResult()

	for i := 0; i < 20; i++ {
		r.UpdateState("WeChat", "partyB", i%2 == 0, i%3 == 0)
		r.UpdateState("WeChat", "partyC", i%2 == 0, i%3 == 0)
		r.UpdateState("Slack", "partyA", i%2 == 0, i%3 == 0)
		r.UpdateState("Slack", "partyB", i%3 == 0, i%2 == 0)
		r.UpdateState("Slack", "partyC", i%3 == 0, i%2 == 0)
	}

	r.DoCalculate()
	t.Log(r.String())
}

func Test_verifyMultiPartyAddress(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l)

	a1 := mock.Address()
	a2 := mock.Address()
	a3 := mock.Address()

	//prepare two contracts
	var contracts []*ContractParam

	// Montnets-PCCWG
	param1 := buildContractParam()
	param1.PartyA.Address = a1
	param1.PartyB.Address = a2
	param1.NextStops = []string{"A2P_PCCWG"}
	param1.PreStops = []string{"MONTNETS"}
	contracts = append(contracts, param1)

	// PCCWG-CSL
	param2 := buildContractParam()
	param2.PartyA.Address = a2
	param2.PartyB.Address = a3
	param2.NextStops = []string{"CSL Hong Kong @ 3397"}
	param2.PreStops = []string{"A2P_PCCWG"}
	contracts = append(contracts, param2)

	// mock
	param3 := buildContractParam()
	param3.PartyA.Address = mock.Address()
	param3.PartyB.Address = mock.Address()
	param3.NextStops = []string{"CSL Hong Kong @ 3397"}
	param3.PreStops = []string{"A2P_PCCWG"}
	contracts = append(contracts, param3)

	for _, c := range contracts {
		contractAddr, _ := c.Address()
		abi, _ := c.ToABI()
		if err := SaveContractParam(ctx, &contractAddr, abi[:]); err != nil {
			t.Fatal(err)
		}
	}

	// save to db
	if err := ctx.SaveStorage(); err != nil {
		t.Fatal(err)
	}

	ca1, _ := contracts[0].Address()
	ca2, _ := contracts[1].Address()
	ca3, _ := contracts[2].Address()

	type args struct {
		ctx        *vmstore.VMContext
		firstAddr  *types.Address
		secondAddr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    *ContractParam
		want1   *ContractParam
		want2   bool
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:        ctx,
				firstAddr:  &ca1,
				secondAddr: &ca2,
			},
			want:    nil,
			want1:   nil,
			want2:   true,
			wantErr: false,
		}, {
			name: "f1",
			args: args{
				ctx:        ctx,
				firstAddr:  &ca1,
				secondAddr: &a2,
			},
			want:    nil,
			want1:   nil,
			want2:   false,
			wantErr: true,
		}, {
			name: "f2",
			args: args{
				ctx:        ctx,
				firstAddr:  &ca1,
				secondAddr: &ca3,
			},
			want:    nil,
			want1:   nil,
			want2:   false,
			wantErr: true,
		}, {
			name: "f3",
			args: args{
				ctx:        ctx,
				firstAddr:  &a2,
				secondAddr: &ca2,
			},
			want:    nil,
			want1:   nil,
			want2:   false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, got2, err := verifyMultiPartyAddress(tt.args.ctx, tt.args.firstAddr, tt.args.secondAddr)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyMultiPartyAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("verifyMultiPartyAddress() got = %v, want %v", got, tt.want)
			//}
			//if !reflect.DeepEqual(got1, tt.want1) {
			//	t.Errorf("verifyMultiPartyAddress() got1 = %v, want %v", got1, tt.want1)
			//}
			if got2 != tt.want2 {
				t.Errorf("verifyMultiPartyAddress() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
