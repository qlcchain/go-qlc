/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
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

func buildContactParam(addr1, addr2 types.Address, name1, name2 string) *CreateContractParam {
	return &CreateContractParam{
		PartyA: cabi.Contractor{
			Address: addr1,
			Name:    name1,
		},
		PartyB: cabi.Contractor{
			Address: addr2,
			Name:    name2,
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
		StartDate: time.Now().AddDate(0, 0, -1).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 1).Unix(),
	}
}

//TODO: remove all sleep
func TestSettlementAPI_Integration(t *testing.T) {
	testcase, verifier, api := setupSettlementAPI(t)
	defer testcase(t)

	pccwAddr := account1.Address()
	cslAddr := account2.Address()

	param := buildContactParam(pccwAddr, cslAddr, "PCCWG", "HTK-CSL")

	if address, err := api.ToAddress(&cabi.CreateContractParam{
		PartyA:    param.PartyA,
		PartyB:    param.PartyB,
		Previous:  mock.Hash(),
		Services:  param.Services,
		SignDate:  time.Now().Unix(),
		StartDate: param.StartDate,
		EndDate:   param.EndDate,
	}); err != nil {
		t.Fatal(err)
	} else if address.IsZero() {
		t.Fatal("ToAddress failed")
	}

	if blk, err := api.GetCreateContractBlock(param); err != nil {
		t.Fatal(err)
	} else {
		//t.Log(blk)
		txHash := blk.GetHash()
		blk.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(blk); err != nil {
			t.Fatal(err)
		}
		if rx, err := api.GetSettlementRewardsBlock(&txHash); err != nil {
			t.Fatal(err)
		} else {
			if err := verifier.BlockProcess(rx); err != nil {
				t.Fatal(err)
			}
		}

		for {
			if contracts, _ := api.GetAllContracts(10, offset(0)); len(contracts) > 0 {
				break
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}

		if contracts, err := api.GetContractsAsPartyA(&pccwAddr, 10, offset(0)); err != nil {
			t.Fatal(err)
		} else if len(contracts) != 1 {
			t.Fatalf("invalid GetContractsAsPartyA len, exp: 1, act: %d", len(contracts))
		}

		if contracts, err := api.GetContractsByAddress(&pccwAddr, 10, offset(0)); err != nil {
			t.Fatal(err)
		} else if len(contracts) != 1 {
			t.Fatalf("invalid GetContractsByAddress len, exp: 1, act: %d", len(contracts))
		}

		if contracts, err := api.GetContractsAsPartyB(&cslAddr, 1, offset(0)); err != nil {
			t.Fatal(err)
		} else if len(contracts) != 1 {
			t.Fatalf("invalid contracts len, exp: 1, act: %d", len(contracts))
		} else {
			contractAddr1 := contracts[0].Address

			if blk, err := api.GetSignContractBlock(&SignContractParam{
				ContractAddress: contractAddr1,
				Address:         cslAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(blk.GetHash())
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			// add next stop
			if blk, err := api.GetAddNextStopBlock(&StopParam{
				StopParam: cabi.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL",
				},
				Address: pccwAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			if blk, err := api.GetAddNextStopBlock(&StopParam{
				StopParam: cabi.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL-1",
				},
				Address: pccwAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			// update next stop name
			if blk, err := api.GetUpdateNextStopBlock(&UpdateStopParam{
				UpdateStopParam: cabi.UpdateStopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL-1",
					New:             "HKTCSL-2",
				},
				Address: pccwAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}
			if blk, err := api.GetRemoveNextStopBlock(&StopParam{
				StopParam: cabi.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL-2",
				},
				Address: pccwAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			// add pre stop
			if blk, err := api.GetAddPreStopBlock(&StopParam{
				StopParam: cabi.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG",
				},
				Address: cslAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			if blk, err := api.GetAddPreStopBlock(&StopParam{
				StopParam: cabi.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG-1",
				},
				Address: cslAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			// update pre stop
			if blk, err := api.GetUpdatePreStopBlock(&UpdateStopParam{
				UpdateStopParam: cabi.UpdateStopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG-1",
					New:             "PCCWG-2",
				},
				Address: cslAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			// remove pre stop
			if blk, err := api.GetRemovePreStopBlock(&StopParam{
				StopParam: cabi.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG-2",
				},
				Address: cslAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			// waiting for data save to db
			for {
				if nextStopNames, err := api.GetNextStopNames(&pccwAddr); err != nil {
					t.Fatal(err)
				} else if len(nextStopNames) > 0 {
					t.Logf("next names: %s %s", pccwAddr.String(), nextStopNames)
					if names, err := api.GetPreStopNames(&cslAddr); err != nil {
						t.Fatal(err)
					} else if len(names) > 0 {
						t.Logf("pre names: %s %s", cslAddr.String(), names)
						break
					} else {
						time.Sleep(500 * time.Millisecond)
					}
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}

			if contracts, err := api.GetContractsByStatus(&pccwAddr, "Activated", 10, offset(0)); err != nil {
				t.Fatal(err)
			} else if len(contracts) != 1 {
				t.Fatalf("invalid GetContractsByStatus len, exp: 1, act: %d", len(contracts))
			}

			if a, err := api.GetContractAddressByPartyANextStop(&pccwAddr, "HKTCSL"); err != nil {
				t.Fatal(err)
			} else if *a != contractAddr1 {
				t.Fatalf("invalid contract address, exp: %s, act: %s", contractAddr1.String(), a.String())
			}

			if a, err := api.GetContractAddressByPartyBPreStop(&cslAddr, "PCCWG"); err != nil {
				t.Fatal(err)
			} else if *a != contractAddr1 {
				t.Fatalf("invalid contract address, exp: %s, act: %s", contractAddr1.String(), a.String())
			}

			// pccw upload CDR
			cdr1 := &cabi.CDRParam{
				ContractAddress: types.ZeroAddress,
				Index:           1111,
				SmsDt:           time.Now().Unix(),
				Sender:          "WeChat",
				Destination:     "85257***343",
				SendingStatus:   cabi.SendingStatusSent,
				DlrStatus:       cabi.DLRStatusDelivered,
				PreStop:         "",
				NextStop:        "HKTCSL",
			}

			if blk, err := api.GetProcessCDRBlock(&pccwAddr, cdr1); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			// CSL upload CDR
			cdr2 := &cabi.CDRParam{
				ContractAddress: types.ZeroAddress,
				Index:           1111,
				SmsDt:           time.Now().Unix(),
				Sender:          "WeChat",
				Destination:     "85257***343",
				SendingStatus:   cabi.SendingStatusSent,
				DlrStatus:       cabi.DLRStatusDelivered,
				PreStop:         "PCCWG",
				NextStop:        "",
			}
			if blk, err := api.GetProcessCDRBlock(&cslAddr, cdr2); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			h, err := cdr1.ToHash()
			if err != nil {
				t.Fatal(err)
			}

			if status, err := api.GetCDRStatus(&contractAddr1, h); err != nil {
				t.Fatal(err)
			} else if status.Status != cabi.SettlementStatusSuccess {
				t.Fatalf("invalid cdr state, exp: %s, act: %s", cabi.SettlementStatusSuccess.String(), status.Status.String())
			}

			if data, err := api.GetCDRStatusByCdrData(&contractAddr1, 1111, "WeChat", "85257***343"); err != nil {
				t.Fatal(err)
			} else {
				t.Log(data)
			}

			for {
				if records, _ := api.GetAllCDRStatus(&contractAddr1, 10, offset(0)); len(records) == 1 {
					break
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}

			if records, err := api.GetCDRStatusByDate(&contractAddr1, 0, 0, 10, offset(0)); err != nil {
				t.Fatal(err)
			} else if len(records) != 1 {
				t.Fatalf("invalid GetCDRStatusByDate len, exp: 1, act: %d", len(records))
			}

			if allContracts, err := api.GetAllContracts(10, offset(0)); err != nil {
				t.Fatal(err)
			} else if len(allContracts) != 1 {
				t.Fatalf("invalid GetAllContracts len, exp: 1, act: %d", len(allContracts))
			}

			if report, err := api.GetSummaryReport(&contractAddr1, 0, 0); err != nil {
				t.Fatal(err)
			} else {
				t.Log(report)
			}

			if invoices, err := api.GenerateInvoices(&pccwAddr, 0, 0); err != nil {
				t.Fatal(err)
			} else {
				t.Log(invoices)
			}
			if invoices, err := api.GenerateInvoicesByContract(&contractAddr1, 0, 0); err != nil {
				t.Fatal(err)
			} else {
				t.Log(invoices)
			}
		}
	}
}

func TestSortCDRs(t *testing.T) {
	cdr1 := buildCDRStatus()
	cdr2 := buildCDRStatus()
	r := []*cabi.CDRStatus{cdr1, cdr2}
	sort.Slice(r, func(i, j int) bool {
		return sortCDRFun(r[i], r[j])
	})
	t.Log(r)
}

func buildCDRStatus() *cabi.CDRStatus {
	cdrParam := cabi.CDRParam{
		Index:         1,
		SmsDt:         time.Now().Unix(),
		Sender:        "PCCWG",
		Destination:   "85257***343",
		SendingStatus: cabi.SendingStatusSent,
		DlrStatus:     cabi.DLRStatusDelivered,
	}
	cdr1 := cdrParam
	i, _ := random.Intn(10000)
	cdr1.Index = uint64(i)
	cdr1.SmsDt = time.Now().Add(time.Minute * time.Duration(cdr1.Index)).Unix()

	status := &cabi.CDRStatus{
		Params: map[string][]cabi.CDRParam{
			mock.Address().String(): {cdr1},
		},
		Status: cabi.SettlementStatusSuccess,
	}

	return status
}

func TestSettlementAPI_GetTerminateContractBlock(t *testing.T) {
	testcase, verifier, api := setupSettlementAPI(t)
	defer testcase(t)

	pccwAddr := account1.Address()
	cslAddr := account2.Address()

	param := &CreateContractParam{
		PartyA: cabi.Contractor{
			Address: pccwAddr,
			Name:    "PCCWG",
		},
		PartyB: cabi.Contractor{
			Address: cslAddr,
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
		StartDate: time.Now().AddDate(0, 0, -1).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 1).Unix(),
	}

	if blk, err := api.GetCreateContractBlock(param); err != nil {
		t.Fatal(err)
	} else {
		//t.Log(blk)
		txHash := blk.GetHash()
		blk.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(blk); err != nil {
			t.Fatal(err)
		}

		for {
			if contracts, _ := api.GetAllContracts(10, offset(0)); len(contracts) > 0 {
				break
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}

		if contracts, err := api.GetContractsAsPartyB(&cslAddr, 1, offset(0)); err != nil {
			t.Fatal(err)
		} else if len(contracts) != 1 {
			t.Fatalf("invalid contracts len, exp: 1, act: %d", len(contracts))
		} else {
			c := contracts[0]
			contractAddr := c.Address
			if blk, err := api.GetTerminateContractBlock(&TerminateParam{
				TerminateParam: cabi.TerminateParam{
					ContractAddress: contractAddr,
					Request:         true,
				},
				Address: cslAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(blk.GetHash())
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestSettlementAPI_GetExpiredContracts(t *testing.T) {
	testcase, verifier, api := setupSettlementAPI(t)
	defer testcase(t)

	pccwAddr := account1.Address()
	cslAddr := account2.Address()

	param := &CreateContractParam{
		PartyA: cabi.Contractor{
			Address: pccwAddr,
			Name:    "PCCWG",
		},
		PartyB: cabi.Contractor{
			Address: cslAddr,
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
		StartDate: time.Now().AddDate(0, 0, -30).Unix(),
		EndDate:   time.Now().AddDate(0, 0, -1).Unix(),
	}

	if blk, err := api.GetCreateContractBlock(param); err != nil {
		t.Fatal(err)
	} else {
		txHash := blk.GetHash()
		blk.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(blk); err != nil {
			t.Fatal(err)
		}

		for {
			if contracts, _ := api.GetAllContracts(10, offset(0)); len(contracts) > 0 {
				break
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}

		if contracts, err := api.GetContractsAsPartyB(&cslAddr, 1, offset(0)); err != nil {
			t.Fatal(err)
		} else if len(contracts) != 1 {
			t.Fatalf("invalid contracts len, exp: 1, act: %d", len(contracts))
		} else {
			c := contracts[0]
			contractAddr := c.Address
			if blk, err := api.GetSignContractBlock(&SignContractParam{
				ContractAddress: contractAddr,
				Address:         cslAddr,
			}); err != nil {
				t.Fatal(err)
			} else {
				blk.Signature = account2.Sign(blk.GetHash())
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}
			}

			for {
				if contracts, _ := api.GetAllContracts(10, offset(0)); len(contracts) > 0 && contracts[0].Status == cabi.ContractStatusActivated {
					break
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}

			if contracts, err := api.GetExpiredContracts(&pccwAddr, 10, offset(0)); err != nil {
				t.Fatal(err)
			} else if len(contracts) != 1 {
				t.Fatalf("invalid GetExpiredContracts len, exp: 1, act: %d", len(contracts))
			}
		}
	}
}

type paramContainer struct {
	account1, account2 *types.Account
	name1, name2       string
	pre, next          string
}

func TestSettlementAPI_GenerateMultiPartyInvoice(t *testing.T) {
	testcase, verifier, api := setupSettlementAPI(t)
	defer testcase(t)

	params := []*paramContainer{
		{
			account1: account1,
			account2: account2,
			name1:    "MONTNETS",
			name2:    "PCCWG",
			pre:      "MONTNETS",
			next:     "A2P_PCCWG",
		}, {
			account1: account2,
			account2: account3,
			name1:    "PCCWG",
			name2:    "HKT-CSL",
			pre:      "A2P_PCCWG",
			next:     "CSL Hong Kong @ 3397",
		},
	}

	// prepare two settlement contract
	for idx, p := range params {
		addr1 := p.account1.Address()
		addr2 := p.account2.Address()
		param1 := buildContactParam(addr1, addr2, p.name1, p.name2)

		if blk, err := api.GetCreateContractBlock(param1); err != nil {
			t.Fatal(err)
		} else {
			txHash := blk.GetHash()
			blk.Signature = p.account1.Sign(txHash)
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}

			for {
				if contracts, _ := api.GetAllContracts(10, offset(0)); len(contracts) > idx {
					break
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}

			if contracts, err := api.GetContractsAsPartyB(&addr2, 1, offset(0)); err != nil {
				t.Fatal(err)
			} else if len(contracts) != 1 {
				t.Fatalf("invalid contracts len, exp: 1, act: %d", len(contracts))
			} else {
				contractAddr := contracts[0].Address

				if blk, err := api.GetSignContractBlock(&SignContractParam{
					ContractAddress: contractAddr,
					Address:         addr2,
				}); err != nil {
					t.Fatal(err)
				} else {
					blk.Signature = p.account2.Sign(blk.GetHash())
					if err := verifier.BlockProcess(blk); err != nil {
						t.Fatal(err)
					}
				}

				// add next stop
				if blk, err := api.GetAddNextStopBlock(&StopParam{
					StopParam: cabi.StopParam{
						ContractAddress: contractAddr,
						StopName:        p.next,
					},
					Address: addr1,
				}); err != nil {
					t.Fatal(err)
				} else {
					blk.Signature = p.account1.Sign(txHash)
					if err := verifier.BlockProcess(blk); err != nil {
						t.Fatal(err)
					}
				}

				// add pre stop
				if blk, err := api.GetAddPreStopBlock(&StopParam{
					StopParam: cabi.StopParam{
						ContractAddress: contractAddr,
						StopName:        p.pre,
					},
					Address: addr2,
				}); err != nil {
					t.Fatal(err)
				} else {
					blk.Signature = p.account2.Sign(txHash)
					if err := verifier.BlockProcess(blk); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}

	// upload CDR
	montAddr := account1.Address()
	pccwAddr := account2.Address()
	cslAddr := account3.Address()

	var contractAddr1, contractAddr2 types.Address
	for {
		if c1, err := api.GetContractsAsPartyB(&pccwAddr, 1, offset(0)); err == nil &&
			c1[0].Status == cabi.ContractStatusActivated {
			contractAddr1 = c1[0].Address
			if c2, err := api.GetContractsAsPartyA(&pccwAddr, 1, offset(0)); err == nil &&
				c2[0].Status == cabi.ContractStatusActivated {
				contractAddr2 = c2[0].Address
				t.Log(c1[0].String())
				t.Log(c2[0].String())
				break
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
	cdrCount := 10
	for i := 0; i < cdrCount; i++ {
		template := cabi.CDRParam{
			ContractAddress: types.ZeroAddress,
			Index:           1111,
			SmsDt:           time.Now().Unix(),
			Destination:     "85257***343",
			SendingStatus:   cabi.SendingStatusSent,
			DlrStatus:       cabi.DLRStatusDelivered,
			PreStop:         "",
			NextStop:        "",
		}

		if i%2 == 1 {
			template.Sender = "WeChat"
		} else {
			template.Sender = "Slack"
		}
		template.Index++
		template.Destination = template.Destination[:len(template.Destination)] + strconv.Itoa(i)

		cdr1 := template
		cdr1.NextStop = "A2P_PCCWG"

		if blk, err := api.GetProcessCDRBlock(&montAddr, &cdr1); err != nil {
			t.Fatal(err)
		} else {
			blk.Signature = account1.Sign(blk.GetHash())
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}
		}

		cdr2 := template
		cdr2.PreStop = "MONTNETS"

		if blk, err := api.GetProcessCDRBlock(&pccwAddr, &cdr2); err != nil {
			t.Fatal(err)
		} else {
			blk.Signature = account2.Sign(blk.GetHash())
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}
		}

		cdr3 := template
		cdr3.NextStop = "CSL Hong Kong @ 3397"

		if blk, err := api.GetProcessCDRBlock(&pccwAddr, &cdr3); err != nil {
			t.Fatal(err)
		} else {
			blk.Signature = account2.Sign(blk.GetHash())
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}
		}
		cdr4 := template
		cdr4.PreStop = "A2P_PCCWG"

		if blk, err := api.GetProcessCDRBlock(&cslAddr, &cdr4); err != nil {
			t.Fatal(err)
		} else {
			blk.Signature = account3.Sign(blk.GetHash())
			if err := verifier.BlockProcess(blk); err != nil {
				t.Fatal(err)
			}
		}
	}

	for {
		if status, err := api.GetAllCDRStatus(&contractAddr1, 100, offset(0)); err == nil && len(status) >= cdrCount {
			if status, err := api.GetAllCDRStatus(&contractAddr2, 100, offset(0)); err == nil && len(status) >= cdrCount {
				break
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if invoice, err := api.GenerateMultiPartyInvoice(&contractAddr1, &contractAddr2, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		t.Log(invoice)
	}

	if report, err := api.GenerateMultiPartySummaryReport(&contractAddr1, &contractAddr2, 0, 0); err != nil {
		t.Fatal(err)
	} else {
		t.Log(report)
	}
}
