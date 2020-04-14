/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"

	"github.com/google/uuid"

	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	t.Parallel()
	dir := filepath.Join(cfg.QlcTestDataDir(), "settlement", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := cfg.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	l := ledger.NewLedger(cm.ConfigFile)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	verifier := process.NewLedgerVerifier(l)

	for i := range blocks {
		block := blocks[i]

		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
	}

	return func(t *testing.T) {
		//err := l.DBStore.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}

func TestGetContractsByAddress(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	var contracts []*ContractParam

	for i := 0; i < 4; i++ {
		param := buildContractParam()
		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	//if err := ctx.Iterator(contractaddress.SettlementAddress[:], func(key []byte, value []byte) error {
	//	t.Log(hex.EncodeToString(key), " >>> ", hex.EncodeToString(value))
	//	return nil
	//}); err != nil {
	//	t.Fatal(err)
	//}

	if contracts == nil || len(contracts) != 4 {
		t.Fatalf("invalid mock contract data, exp: 4, act: %d", len(contracts))
	}

	a := contracts[0].PartyA.Address

	type args struct {
		ctx  ledger.Store
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "1st",
			args: args{
				ctx:  l,
				addr: &a,
			},
			want:    []*ContractParam{contracts[0]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsByAddress(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContractsByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetContractsByAddress() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("GetContractsByAddress() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func mockContractData(size int) []*ContractParam {
	var contracts []*ContractParam
	accounts := []types.Address{mock.Address(), mock.Address()}

	for i := 0; i < size; i++ {
		cp := createContractParam

		var a1 types.Address
		var a2 types.Address
		if i%2 == 0 {
			a1 = accounts[0]
			a2 = accounts[1]
		} else {
			a1 = accounts[1]
			a2 = accounts[0]
		}
		cp.PartyA.Address = a1
		cp.PartyB.Address = a2

		for _, s := range cp.Services {
			s.Mcc = s.Mcc + uint64(i)
			s.Mnc = s.Mnc + uint64(i)
		}

		cd := time.Now().Unix()
		param := &ContractParam{
			CreateContractParam: cp,
			ConfirmDate:         cd,
		}
		contracts = append(contracts, param)
	}
	return contracts
}

func TestGetAllSettlementContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	for i := 0; i < 4; i++ {
		param := buildContractParam()
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  ledger.Store
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx:  l,
				size: 4,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllSettlementContract(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllSettlementContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.args.size {
				t.Fatalf("invalid size: exp: %d, act: %d", tt.args.size, len(got))
			}
			for _, c := range got {
				t.Log(c.String())
			}
		})
	}
}

func TestGetContracts(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		} else {
			//t.Log(hex.EncodeToString(abi))
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	for _, d := range data {
		if addr, err := d.Address(); err != nil {
			t.Fatal(err)
		} else {
			if contract, err := GetSettlementContract(ctx, &addr); err != nil {
				t.Fatal(err)
			} else {
				if !reflect.DeepEqual(contract, d) {
					t.Fatalf("invalid %v, %v", contract, d)
				}
			}
		}
	}
}

func TestGetContractsIDByAddressAsPartyA(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	a1 := data[0].PartyA.Address

	type args struct {
		ctx  ledger.Store
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  l,
				addr: &a1,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsIDByAddressAsPartyA(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContractsIDByAddressAsPartyA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetContractsIDByAddressAsPartyA() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("GetContractsIDByAddressAsPartyA() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetContractsIDByAddressAsPartyB(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	a2 := data[0].PartyB.Address

	type args struct {
		ctx  ledger.Store
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  l,
				addr: &a2,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsIDByAddressAsPartyB(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetContractsIDByAddressAsPartyB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("TestGetContractsIDByAddressAsPartyB() len(go) != len(tt.want), %d,%d", len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				a1, _ := got[i].Address()
				a2, _ := tt.want[i].Address()
				if a1 != a2 {
					t.Fatalf("TestGetContractsIDByAddressAsPartyB() i[%d] %v,%v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetContractsAddressByPartyANextStop(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	data := mockContractData(2)
	data[0].NextStops = append(data[0].NextStops, "PCCWG")

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	a1 := data[0].PartyA.Address

	type args struct {
		ctx  ledger.Store
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  l,
				addr: &a1,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsAddressByPartyANextStop(tt.args.ctx, tt.args.addr, "PCCWG")
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetContractsAddressByPartyANextStop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			a2, _ := tt.want[0].Address()
			if *got != a2 {
				t.Fatalf("TestGetContractsAddressByPartyANextStop()  %s,%s", got, a2)
			}
		})
	}
}

func TestGetContractsAddressByPartyBPreStop(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	data := mockContractData(2)
	data[0].PreStops = append(data[0].PreStops, "CSL")
	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	for _, d := range data {
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	a2 := data[0].PartyB.Address

	type args struct {
		ctx  ledger.Store
		addr *types.Address
	}
	tests := []struct {
		name    string
		args    args
		want    []*ContractParam
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  l,
				addr: &a2,
			},
			want:    []*ContractParam{data[0]},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContractsAddressByPartyBPreStop(tt.args.ctx, tt.args.addr, "CSL")
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetContractsIDByAddressAsPartyB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			a2, _ := tt.want[0].Address()
			if *got != a2 {
				t.Fatalf("TestGetContractsIDByAddressAsPartyB() %s,%s", got, a2)
			}
		})
	}
}

func TestGetSettlementContract(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	var contracts []*ContractParam

	for i := 0; i < 2; i++ {
		param := buildContractParam()

		if i%2 == 1 {
			param.Status = ContractStatusActiveStage1
		}

		contracts = append(contracts, param)
		a, _ := param.Address()
		abi, _ := param.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	if contracts == nil || len(contracts) != 2 {
		t.Fatal("invalid mock contract data")
	}
	a1 := contracts[0].PartyA.Address
	cdr := cdrParam
	cdr.NextStop = "HKTCSL"

	if c, err := FindSettlementContract(l, &a1, &cdr); err != nil {
		t.Fatal(err)
	} else {
		t.Log(c)
	}
	a2 := mock.Address()
	if _, err := FindSettlementContract(l, &a2, &cdr); err == nil {
		t.Fatal("should find nothing...")
	}

	a3 := contracts[0].PartyB.Address
	cdr3 := cdrParam
	cdr3.PreStop = "PCCWG"
	if _, err := FindSettlementContract(l, &a3, &cdr3); err != nil {
		t.Fatal(err)
	}
}

func buildCDRStatus() *CDRStatus {
	cdr1 := cdrParam
	i, _ := random.Intn(10000)
	cdr1.Index = uint64(i)

	status := &CDRStatus{
		Params: map[string][]CDRParam{
			mock.Address().String(): {cdr1},
		},
		Status: SettlementStatusSuccess,
	}

	return status
}

func TestGetAllCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	contractAddr := mock.Address()

	var data []*CDRStatus
	for i := 0; i < 4; i++ {
		s := buildCDRStatus()
		if h, err := s.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}
			if abi, err := s.ToABI(); err != nil {
				t.Fatal(err)
			} else {
				if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
					t.Fatal(err)
				} else {
					data = append(data, s)
				}
			}
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  ledger.Store
		addr *types.Address
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    []*CDRStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  l,
				addr: &contractAddr,
				size: len(data),
			},
			want:    data,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllCDRStatus(tt.args.ctx, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllCDRStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(data) {
				t.Errorf("GetAllCDRStatus() got = %d, want %d", len(got), len(data))
			}
		})
	}
}

func TestGetCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	contractAddr := mock.Address()
	contractAddr2 := mock.Address()

	s := buildCDRStatus()
	h, err := s.ToHash()
	if err != nil {
		t.Fatal(err)
	} else {
		if abi, err := s.ToABI(); err != nil {
			t.Fatal(err)
		} else {
			if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  ledger.Store
		addr *types.Address
		hash types.Hash
	}
	tests := []struct {
		name    string
		args    args
		want    *CDRStatus
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx:  l,
				addr: &contractAddr,
				hash: h,
			},
			want:    s,
			wantErr: false,
		}, {
			name: "fail",
			args: args{
				ctx:  l,
				addr: &contractAddr,
				hash: mock.Hash(),
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "f2",
			args: args{
				ctx:  l,
				addr: &contractAddr2,
				hash: h,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCDRStatus(tt.args.ctx, tt.args.addr, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCDRStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCDRStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSummaryReport(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	// mock settlement contract
	ac1 := mock.Account()
	ac2 := mock.Account()
	a1 := ac1.Address()
	a2 := ac2.Address()

	param := buildContractParam()
	param.PartyA.Address = a1
	param.PartyB.Address = a2
	param.NextStops = []string{"CSL Hong Kong @ 3397"}
	param.PreStops = []string{"A2P_PCCWG"}

	contractAddr, _ := param.Address()
	abi, _ := param.ToABI()
	if err := SaveContractParam(ctx, &contractAddr, abi[:]); err != nil {
		t.Fatal(err)
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		params := make(map[string][]CDRParam, 0)
		param1 := cdrParam
		if i%2 == 0 {
			param1.Sender = "Slack"
		}
		params[a1.String()] = []CDRParam{param1}
		params[a2.String()] = []CDRParam{param1}

		s := &CDRStatus{
			Params: params,
			Status: SettlementStatusSuccess,
		}
		//t.Log(s.String())
		if h, err := s.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}
			if abi, err := s.ToABI(); err != nil {
				t.Fatal(err)
			} else {
				if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	if report, err := GetSummaryReport(l, &contractAddr, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		t.Log(report)
	}

	if invoices, err := GenerateInvoices(l, &a1, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoices) == 0 {
			t.Fatal("invalid invoice")
		}
		for _, i := range invoices {
			t.Log(util.ToIndentString(i))
		}
	}

	if invoices, err := GenerateInvoicesByContract(l, &contractAddr, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoices) == 0 {
			t.Fatal("invalid invoice")
		}
		t.Log(util.ToIndentString(invoices))
	}
}

func TestGetStopNames(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	// mock settlement contract
	ac1 := mock.Account()
	ac2 := mock.Account()
	a1 := ac1.Address()
	a2 := ac2.Address()

	var contracts []*ContractParam

	param := buildContractParam()
	param.PartyA.Address = a1
	param.PartyB.Address = a2
	param.NextStops = []string{"CSL Hong Kong @ 3397"}
	param.PreStops = []string{"A2P_PCCWG"}
	contracts = append(contracts, param)

	param2 := buildContractParam()
	param2.PartyA.Address = a2
	param2.PartyB.Address = a1
	param2.PreStops = []string{"CSL Hong Kong @ 33971"}
	param2.NextStops = []string{"A2P_PCCWG2"}

	contracts = append(contracts, param2)
	for _, c := range contracts {
		contractAddr, _ := c.Address()
		abi, _ := c.ToABI()
		if err := SaveContractParam(ctx, &contractAddr, abi[:]); err != nil {
			t.Fatal(err)
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	if names, err := GetPreStopNames(ctx, &a1); err != nil {
		t.Fatal(err)
	} else {
		if len(names) != 1 {
			t.Fatalf("invalid len %d", len(names))
		}

		if names[0] != "CSL Hong Kong @ 33971" {
			t.Fatal(names)
		}
	}

	if names, err := GetNextStopNames(ctx, &a1); err != nil {
		t.Fatal(err)
	} else {
		if len(names) != 1 {
			t.Fatalf("invalid len %d", len(names))
		}

		if names[0] != "CSL Hong Kong @ 3397" {
			t.Fatal(names)
		}
	}
}

func TestSaveCDRStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	a1 := mock.Address()
	cdr := buildCDRStatus()

	h, err := cdr.ToHash()
	if err != nil {
		t.Fatal(err)
	}
	if err = SaveCDRStatus(ctx, &a1, &h, cdr); err != nil {
		t.Fatal(err)
	}

	if s, err := GetCDRStatus(l, &a1, h); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(cdr, s) {
			t.Fatalf("invalid cdr, act: %v, exp: %v", s, cdr)
		} else {
			if addresses, err := GetCDRMapping(ctx, &h); err != nil {
				t.Fatal(err)
			} else {
				if len(addresses) != 1 {
					t.Fatalf("invalid address len: %d", len(addresses))
				}

				if !reflect.DeepEqual(addresses[0], &a1) {
					t.Fatalf("invalid address, act: %s, exp: %s", addresses[0].String(), a1.String())
				}
			}
		}
	}
}

func TestGetCDRMapping(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	a1 := mock.Address()
	a2 := mock.Address()
	h := mock.Hash()

	exp := []*types.Address{&a1, &a2}

	if err := saveCDRMapping(ctx, &a1, &h); err != nil {
		t.Fatal(err)
	}
	if err := saveCDRMapping(ctx, &a2, &h); err != nil {
		t.Fatal(err)
	}

	if addressList, err := GetCDRMapping(ctx, &h); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(addressList, exp) {
			t.Fatalf("invalid address, act: %v, exp:%v", addressList, exp)
		}
	}
}

func TestGenerateMultiPartyInvoice(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

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
	for _, c := range contracts {
		contractAddr, _ := c.Address()
		abi, _ := c.ToABI()
		if err := SaveContractParam(ctx, &contractAddr, abi[:]); err != nil {
			t.Fatal(err)
		}
	}
	ca1, _ := contracts[0].Address()
	ca2, _ := contracts[1].Address()

	// upload CDR
	for i := 0; i < 10; i++ {
		template := cdrParam
		if i%2 == 1 {
			template.Sender = "WeChat"
		} else {
			template.Sender = "Slack"
		}
		template.Index++
		template.Destination = template.Destination[:len(template.Destination)] + strconv.Itoa(i)

		p1 := template
		p1.NextStop = "A2P_PCCWG"

		cdr1 := &CDRStatus{
			Params: map[string][]CDRParam{
				a1.String(): {p1},
			},
			Status: SettlementStatusSuccess,
		}

		if h, err := cdr1.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}

			t.Log("p1", h.String())
			//if err := SaveCDRStatus(ctx, &ca1, &h, cdr1); err != nil {
			//	t.Fatal(err)
			//}
		}

		p2 := template
		p2.PreStop = "MONTNETS"

		cdr2 := &CDRStatus{
			Params: map[string][]CDRParam{
				a1.String(): {p1},
				a2.String(): {p2},
			},
			Status: SettlementStatusSuccess,
		}

		if h, err := cdr2.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}
			t.Log("p2", h.String())
			if err := SaveCDRStatus(ctx, &ca1, &h, cdr2); err != nil {
				t.Fatal(err)
			}
		}

		// upload CDR to PCCWG-CSL
		p3 := template
		p3.NextStop = "CSL Hong Kong @ 3397"

		cdr3 := &CDRStatus{
			Params: map[string][]CDRParam{
				a2.String(): {p3},
			},
			Status: SettlementStatusSuccess,
		}

		if h, err := cdr3.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}

			t.Log("p3", h.String())
			//if err := SaveCDRStatus(ctx, &ca2, &h, cdr3); err != nil {
			//	t.Fatal(err)
			//}
		}

		p4 := template
		p4.PreStop = "A2P_PCCWG"

		cdr4 := &CDRStatus{
			Params: map[string][]CDRParam{
				a2.String(): {p3},
				a3.String(): {p4},
			},
			Status: SettlementStatusSuccess,
		}

		if h, err := cdr4.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}

			t.Log("p4", h.String())
			if err := SaveCDRStatus(ctx, &ca2, &h, cdr4); err != nil {
				t.Fatal(err)
			}
		}
	}

	// save to db
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	// generate summary report
	if report, err := GetMultiPartySummaryReport(l, &ca1, &ca2, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		t.Log(report.String())
	}

	// generate invoice
	if invoice, err := GenerateMultiPartyInvoice(l, &ca1, &ca2, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoice) == 0 {
			t.Fatal("invalid invoice")
		}
		t.Log(util.ToIndentString(invoice))
	}
	if cdrs, err := GetMultiPartyCDRStatus(l, &ca1, &ca2); err != nil {
		t.Fatal(err)
	} else {
		if len(cdrs) != 2 {
			t.Fatal("invalid multi-party CDR")
		}
	}
}

func TestGetContractsByStatus(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	for _, d := range data {
		d.Status = ContractStatusActivated
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		} else {
			//t.Log(hex.EncodeToString(abi))
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	addr := data[0].PartyA.Address

	if contracts, err := GetContractsByStatus(l, &addr, ContractStatusActivated); err != nil {
		t.Fatal(err)
	} else {
		if len(contracts) != 2 {
			t.Fatalf("invalid GetContractsByStatus len, exp: 2, got: %d", len(contracts))
		}
	}
}

func TestGetExpiredContracts(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	data := mockContractData(2)

	if len(data) != 2 {
		t.Fatalf("invalid mock data, %v", data)
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	for _, d := range data {
		d.Status = ContractStatusActivated
		d.EndDate = time.Now().AddDate(0, 0, -1).Unix()
		a, _ := d.Address()
		abi, _ := d.ToABI()
		if err := SaveContractParam(ctx, &a, abi[:]); err != nil {
			t.Fatal(err)
		} else {
			//t.Log(hex.EncodeToString(abi))
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	addr := data[0].PartyA.Address

	if contracts, err := GetExpiredContracts(l, &addr); err != nil {
		t.Fatal(err)
	} else {
		if len(contracts) != 2 {
			t.Fatalf("invalid GetContractsByStatus len, exp: 2, got: %d", len(contracts))
		}
	}
}

func TestGetAllAssert(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	addr1 := mock.Address()
	size := 10
	for i := 0; i < size; i++ {
		template := assetParam
		addr := mock.Address()
		if i%2 == 0 {
			addr = addr1
		}
		template.Owner.Address = addr
		template.Previous = mock.Hash()
		a := &template
		if abi, err := a.ToABI(); err != nil {
			t.Fatal(err)
		} else {
			if err = SaveAssetParam(ctx, abi); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	if asserts, err := GetAllAsserts(l); err != nil {
		t.Fatal(err)
	} else if len(asserts) != size {
		t.Fatalf("invalid assert size, exp: %d, act: %d", size, len(asserts))
	}

	if asserts, err := GetAssertsByAddress(l, &addr1); err != nil {
		t.Fatal(err)
	} else if len(asserts) != size/2 {
		t.Fatalf("invalid assert size, exp: %d, act: %d", size/2, len(asserts))
	}
}

func TestGetAssetParam(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	template := assetParam
	template.Owner.Address = mock.Address()
	template.Previous = mock.Hash()
	a := &template

	if abi, err := a.ToABI(); err != nil {
		t.Fatal(err)
	} else {
		if err = SaveAssetParam(ctx, abi); err != nil {
			t.Fatal(err)
		}

		if err = SaveAssetParam(ctx, abi); err != nil {
			t.Fatal(err)
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	if hash, err := a.ToAddress(); err != nil {
		t.Fatal(err)
	} else {
		if param, err := GetAssetParam(ctx, hash); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(param, a) {
			t.Fatalf("invalid param, exp: %v, act: %v", a, param)
		}
	}
}

func TestGetContractParam(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	a1 := mock.Address()
	a2 := mock.Address()

	// Montnets-PCCWG
	param1 := buildContractParam()
	param1.PartyA.Address = a1
	param1.PartyB.Address = a2
	param1.NextStops = []string{"A2P_PCCWG"}
	param1.PreStops = []string{"MONTNETS"}

	contractAddr, _ := param1.Address()
	data, _ := param1.ToABI()
	if err := SaveContractParam(ctx, &contractAddr, data[:]); err != nil {
		t.Fatal(err)
	}

	if _, err := GetContractParam(ctx, &contractAddr); err != nil {
		t.Fatal(err)
	}
}

func TestGetSummaryReportByAccount(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	a1 := mock.Address()
	a2 := mock.Address()

	// Montnets-PCCWG
	param1 := buildContractParam()
	param1.PartyA.Address = a1
	param1.PartyB.Address = a2
	param1.NextStops = []string{"A2P_PCCWG"}
	param1.PreStops = []string{"MONTNETS"}

	contractAddr, _ := param1.Address()
	data, _ := param1.ToABI()
	if err := SaveContractParam(ctx, &contractAddr, data[:]); err != nil {
		t.Fatal(err)
	}

	var records []*CDRStatus
	const account = "SAP_DIRECTS"
	for i := 0; i < 15; i++ {
		cdr1 := cdrParam
		if i%2 == 0 {
			cdr1.Account = account
		}
		i, _ := random.Intn(10000)
		cdr1.Index = uint64(i)

		s := &CDRStatus{
			Params: map[string][]CDRParam{
				a1.String(): {cdr1},
				a2.String(): {cdr1},
			},
			Status: SettlementStatusSuccess,
		}
		if i == 14 {
			s = &CDRStatus{Status: SettlementStatusStage1}
		}
		if h, err := s.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}
			if abi, err := s.ToABI(); err != nil {
				t.Fatal(err)
			} else {
				if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
					t.Fatal(err)
				} else {
					records = append(records, s)
				}
			}
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	if report, err := GetSummaryReportByAccount(l, &contractAddr, account, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if v, ok := report.Records[account]; !ok {
			t.Fatal("can not generate summary invoice")
		} else {
			t.Log(report, v)
		}
	}

	if invoice, err := GenerateInvoicesByAccount(l, &contractAddr, account, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoice) == 0 {
			t.Fatal("invalid invoice")
		} else {
			t.Log(util.ToString(invoice))
		}
	}
}

func TestGetSummaryReportByCustomer(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	a1 := mock.Address()
	a2 := mock.Address()

	// Montnets-PCCWG
	param1 := buildContractParam()
	param1.PartyA.Address = a1
	param1.PartyB.Address = a2
	param1.NextStops = []string{"A2P_PCCWG"}
	param1.PreStops = []string{"MONTNETS"}

	contractAddr, _ := param1.Address()
	data, _ := param1.ToABI()
	if err := SaveContractParam(ctx, &contractAddr, data[:]); err != nil {
		t.Fatal(err)
	}

	var records []*CDRStatus
	const customer = "SAP Mobile Services"
	for i := 0; i < 15; i++ {
		cdr1 := cdrParam
		if i%2 == 0 {
			cdr1.Customer = customer
		}
		i, _ := random.Intn(10000)
		cdr1.Index = uint64(i)

		s := &CDRStatus{
			Params: map[string][]CDRParam{
				a1.String(): {cdr1},
				a2.String(): {cdr1},
			},
			Status: SettlementStatusSuccess,
		}

		if i == 14 {
			s = &CDRStatus{Status: SettlementStatusStage1}
		}
		if h, err := s.ToHash(); err != nil {
			t.Fatal(err)
		} else {
			if h.IsZero() {
				t.Fatal("invalid hash")
			}
			if abi, err := s.ToABI(); err != nil {
				t.Fatal(err)
			} else {
				if err := ctx.SetStorage(contractAddr[:], h[:], abi); err != nil {
					t.Fatal(err)
				} else {
					records = append(records, s)
				}
			}
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	if report, err := GetSummaryReportByCustomer(l, &contractAddr, customer, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if v, ok := report.Records[customer]; !ok {
			t.Fatal("can not generate summary report")
		} else {
			t.Log(report, v)
		}
	}

	if invoice, err := GenerateInvoicesByCustomer(l, &contractAddr, customer, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		if len(invoice) == 0 {
			t.Fatal("invalid invoice")
		} else {
			t.Log(util.ToString(invoice))
		}
	}
}
