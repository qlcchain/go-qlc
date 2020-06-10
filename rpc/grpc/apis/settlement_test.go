package apis

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi/settlement"
)

var (
	assetParam = cabi.AssetParam{
		Owner: cabi.Contractor{
			Address: mock.Address(),
			Name:    "HKT-CSL",
		},
		Previous: mock.Hash(),
		Assets: []*cabi.Asset{
			{
				Mcc:         42,
				Mnc:         5,
				TotalAmount: 1000,
				SLAs: []*cabi.SLA{
					{
						SLAType:  cabi.SLATypeLatency,
						Priority: 0,
						Value:    30,
						Compensations: []*cabi.Compensation{
							{
								Low:  50,
								High: 60,
								Rate: 10,
							},
							{
								Low:  60,
								High: 80,
								Rate: 20.5,
							},
						},
					},
				},
			},
		},
		SignDate:  time.Now().Unix(),
		StartDate: time.Now().AddDate(0, 0, 1).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 1).Unix(),
		Status:    cabi.AssetStatusActivated,
	}
)

func setupSettlementAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *SettlementAPI) {
	//t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)
	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewSettlementAPI(l, cc)

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

func buildContactParam(addr1, addr2 types.Address, name1, name2 string) *api.CreateContractParam {
	return &api.CreateContractParam{
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

func toCreateContractParamOfAbi(param *cabi.CreateContractParam) *pbtypes.CreateContractParam {
	services := make([]*pbtypes.ContractService, 0)
	for _, c := range param.Services {
		ct := &pbtypes.ContractService{
			ServiceId:   c.ServiceId,
			Mcc:         c.Mcc,
			Mnc:         c.Mnc,
			TotalAmount: c.TotalAmount,
			UnitPrice:   c.UnitPrice,
			Currency:    c.Currency,
		}
		services = append(services, ct)
	}

	return &pbtypes.CreateContractParam{
		PartyA: &pbtypes.Contractor{
			Address: toAddressValue(param.PartyA.Address),
			Name:    param.PartyA.Name,
		},
		PartyB: &pbtypes.Contractor{
			Address: toAddressValue(param.PartyB.Address),
			Name:    param.PartyB.Name,
		},
		Previous:  toHashValue(param.Previous),
		Services:  services,
		SignDate:  param.SignDate,
		StartDate: param.StartDate,
		EndDate:   param.EndDate,
	}
}

func toCreateContractParam(param *api.CreateContractParam) *pb.CreateContractParam {
	services := make([]*pbtypes.ContractService, 0)
	for _, c := range param.Services {
		ct := &pbtypes.ContractService{
			ServiceId:   c.ServiceId,
			Mcc:         c.Mcc,
			Mnc:         c.Mnc,
			TotalAmount: c.TotalAmount,
			UnitPrice:   c.UnitPrice,
			Currency:    c.Currency,
		}
		services = append(services, ct)
	}

	return &pb.CreateContractParam{
		PartyA: &pbtypes.Contractor{
			Address: toAddressValue(param.PartyA.Address),
			Name:    param.PartyA.Name,
		},
		PartyB: &pbtypes.Contractor{
			Address: toAddressValue(param.PartyB.Address),
			Name:    param.PartyB.Name,
		},
		Services:  services,
		StartDate: param.StartDate,
		EndDate:   param.EndDate,
	}
}

func toCDRParams(params []*cabi.CDRParam) []*pbtypes.CDRParam {
	rs := make([]*pbtypes.CDRParam, 0)
	for _, p := range params {
		r := &pbtypes.CDRParam{
			Index:         p.Index,
			SmsDt:         p.SmsDt,
			Account:       p.Account,
			Sender:        p.Sender,
			Customer:      p.Customer,
			Destination:   p.Destination,
			SendingStatus: int32(p.SendingStatus),
			DlrStatus:     int32(p.DlrStatus),
			PreStop:       p.PreStop,
			NextStop:      p.NextStop,
		}
		rs = append(rs, r)
	}
	return rs
}

func TestSettlementAPI_Integration(t *testing.T) {
	testcase, verifier, settlementApi := setupSettlementAPI(t)
	defer testcase(t)

	pccwAddr := account1.Address()
	cslAddr := account2.Address()

	param := buildContactParam(pccwAddr, cslAddr, "PCCWG", "HTK-CSL")
	if address, err := settlementApi.ToAddress(context.Background(), toCreateContractParamOfAbi(&cabi.CreateContractParam{
		PartyA:    param.PartyA,
		PartyB:    param.PartyB,
		Previous:  mock.Hash(),
		Services:  param.Services,
		SignDate:  time.Now().Unix(),
		StartDate: param.StartDate,
		EndDate:   param.EndDate,
	})); err != nil {
		t.Fatal(err)
	} else if a, _ := toOriginAddress(address); a.IsZero() {
		t.Fatal("ToAddress failed")
	}
	account := "SAP_DIRECTS"
	customer := "SAP Mobile Services"

	if blk, err := settlementApi.GetCreateContractBlock(context.Background(), toCreateContractParam(param)); err != nil {
		t.Fatal(err)
	} else {
		//t.Log(blk)
		bs, err := toOriginStateBlock(blk)
		if err != nil {
			t.Fatal(err)
		}
		txHash := bs.GetHash()
		bs.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(bs); err != nil {
			t.Fatal(err)
		}
		if rx, err := settlementApi.GetSettlementRewardsBlock(context.Background(), &pbtypes.Hash{
			Hash: toHashValue(txHash),
		}); err != nil {
			t.Fatal(err)
		} else {
			br, err := toOriginStateBlock(rx)
			if err != nil {
				t.Fatal(err)
			}
			if err := verifier.BlockProcess(br); err != nil {
				t.Fatal(err)
			}
		}

		if contracts, err := settlementApi.GetContractsAsPartyA(context.Background(), &pb.ContractsByAddressRequest{
			Addr:   toAddressValue(pccwAddr),
			Count:  10,
			Offset: 0,
		}); err != nil {
			t.Fatal(err)
		} else if len(contracts.GetContracts()) != 1 {
			t.Fatalf("invalid GetContractsAsPartyA len, exp: 1, act: %d", len(contracts.GetContracts()))
		}

		if contracts, err := settlementApi.GetContractsByAddress(context.Background(), &pb.ContractsByAddressRequest{
			Addr:   toAddressValue(pccwAddr),
			Count:  10,
			Offset: 0,
		}); err != nil {
			t.Fatal(err)
		} else if len(contracts.GetContracts()) != 1 {
			t.Fatalf("invalid GetContractsByAddress len, exp: 1, act: %d", len(contracts.GetContracts()))
		}

		if contracts, err := settlementApi.GetContractsAsPartyB(context.Background(), &pb.ContractsByAddressRequest{
			Addr:   toAddressValue(cslAddr),
			Count:  10,
			Offset: 0,
		}); err != nil {
			t.Fatal(err)
		} else if len(contracts.GetContracts()) != 1 {
			t.Fatalf("invalid contracts len, exp: 1, act: %d", len(contracts.GetContracts()))
		} else {
			contractAddr1 := contracts.GetContracts()[0].Address

			if blk, err := settlementApi.GetSignContractBlock(context.Background(), &pb.SignContractParam{
				ContractAddress: contractAddr1,
				Address:         toAddressValue(cslAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bk, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bk.Signature = account2.Sign(bk.GetHash())
				if err := verifier.BlockProcess(bk); err != nil {
					t.Fatal(err)
				}
			}

			// add next stop
			if blk, err := settlementApi.GetAddNextStopBlock(context.Background(), &pb.StopParam{
				StopParam: &pbtypes.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL",
				},
				Address: toAddressValue(pccwAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			if blk, err := settlementApi.GetAddNextStopBlock(context.Background(), &pb.StopParam{
				StopParam: &pbtypes.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL-1",
				},
				Address: toAddressValue(pccwAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			// update next stop name
			if blk, err := settlementApi.GetUpdateNextStopBlock(context.Background(), &pb.UpdateStopParam{
				UpdateStopParam: &pbtypes.UpdateStopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL-1",
					NewName:         "HKTCSL-2",
				},
				Address: toAddressValue(pccwAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}
			if blk, err := settlementApi.GetRemoveNextStopBlock(context.Background(), &pb.StopParam{
				StopParam: &pbtypes.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "HKTCSL-2",
				},
				Address: toAddressValue(pccwAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			// add pre stop
			if blk, err := settlementApi.GetAddPreStopBlock(context.Background(), &pb.StopParam{
				StopParam: &pbtypes.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG",
				},
				Address: toAddressValue(cslAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			if blk, err := settlementApi.GetAddPreStopBlock(context.Background(), &pb.StopParam{
				StopParam: &pbtypes.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG-1",
				},
				Address: toAddressValue(cslAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			// update pre stop
			if blk, err := settlementApi.GetUpdatePreStopBlock(context.Background(), &pb.UpdateStopParam{
				UpdateStopParam: &pbtypes.UpdateStopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG-1",
					NewName:         "PCCWG-2",
				},
				Address: toAddressValue(cslAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			// remove pre stop
			if blk, err := settlementApi.GetRemovePreStopBlock(context.Background(), &pb.StopParam{
				StopParam: &pbtypes.StopParam{
					ContractAddress: contractAddr1,
					StopName:        "PCCWG-2",
				},
				Address: toAddressValue(cslAddr),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			if contracts, err := settlementApi.GetContractsByStatus(context.Background(), &pb.ContractsByStatusRequest{
				Addr:   toAddressValue(pccwAddr),
				Status: "Activated",
				Count:  10,
				Offset: 0,
			}); err != nil {
				t.Fatal(err)
			} else if len(contracts.GetContracts()) != 1 {
				t.Fatalf("invalid GetContractsByStatus len, exp: 1, act: %d", len(contracts.GetContracts()))
			}

			if a, err := settlementApi.GetContractAddressByPartyANextStop(context.Background(), &pb.ContractAddressByPartyRequest{
				Addr:     toAddressValue(pccwAddr),
				StopName: "HKTCSL",
			}); err != nil {
				t.Fatal(err)
			} else if a.GetAddress() != contractAddr1 {
				t.Fatalf("invalid contract address, exp: %s, act: %s", contractAddr1, a.String())
			}

			if a, err := settlementApi.GetContractAddressByPartyBPreStop(context.Background(), &pb.ContractAddressByPartyRequest{
				Addr:     toAddressValue(cslAddr),
				StopName: "PCCWG",
			}); err != nil {
				t.Fatal(err)
			} else if a.GetAddress() != contractAddr1 {
				t.Fatalf("invalid contract address, exp: %s, act: %s", contractAddr1, a.String())
			}

			// pccw upload CDR
			cdr1 := &cabi.CDRParam{
				Index:         1111,
				SmsDt:         time.Now().Unix(),
				Account:       account,
				Customer:      customer,
				Sender:        "WeChat",
				Destination:   "85257***343",
				SendingStatus: cabi.SendingStatusSent,
				DlrStatus:     cabi.DLRStatusDelivered,
				PreStop:       "",
				NextStop:      "HKTCSL",
			}

			if blk, err := settlementApi.GetProcessCDRBlock(context.Background(), &pb.ProcessCDRBlockRequest{
				Addr:   toAddressValue(pccwAddr),
				Params: toCDRParams([]*cabi.CDRParam{cdr1}),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			// CSL upload CDR
			cdr2 := &cabi.CDRParam{
				Index:         1111,
				SmsDt:         time.Now().Unix(),
				Sender:        "WeChat",
				Destination:   "85257***343",
				SendingStatus: cabi.SendingStatusSent,
				DlrStatus:     cabi.DLRStatusDelivered,
				PreStop:       "PCCWG",
				NextStop:      "",
			}
			if blk, err := settlementApi.GetProcessCDRBlock(context.Background(), &pb.ProcessCDRBlockRequest{
				Addr:   toAddressValue(cslAddr),
				Params: toCDRParams([]*cabi.CDRParam{cdr2}),
			}); err != nil {
				t.Fatal(err)
			} else {
				bx, err := toOriginStateBlock(blk)
				if err != nil {
					t.Fatal(err)
				}
				bx.Signature = account2.Sign(txHash)
				if err := verifier.BlockProcess(bx); err != nil {
					t.Fatal(err)
				}
			}

			if names, err := settlementApi.GetPreStopNames(context.Background(), &pbtypes.Address{
				Address: cslAddr.String(),
			}); err != nil {
				t.Fatal(err)
			} else if len(names.GetValue()) == 0 {
				t.Fatal("can not find any pre names")
			}

			if names, err := settlementApi.GetNextStopNames(context.Background(), &pbtypes.Address{
				Address: pccwAddr.String(),
			}); err != nil {
				t.Fatal(err)
			} else if len(names.GetValue()) == 0 {
				t.Fatal("can not find any next names")
			}

			h, err := cdr1.ToHash()
			if err != nil {
				t.Fatal(err)
			}

			if status, err := settlementApi.GetCDRStatus(context.Background(), &pb.CDRStatusRequest{
				Addr: contractAddr1,
				Hash: toHashValue(h),
			}); err != nil {
				t.Fatal(err)
			} else if status.GetStatus() != int32(cabi.SettlementStatusSuccess) {
				t.Fatalf("invalid cdr state, exp: %s, act: %d", cabi.SettlementStatusSuccess.String(), int32(status.GetStatus()))
			}

			if data, err := settlementApi.GetCDRStatusByCdrData(context.Background(), &pb.CDRStatusByCdrDataRequest{
				Addr:        contractAddr1,
				Index:       1111,
				Sender:      "WeChat",
				Destination: "85257***343",
			}); err != nil {
				t.Fatal(err)
			} else {
				t.Log(data)
			}

			if records, err := settlementApi.GetCDRStatusByDate(context.Background(), &pb.CDRStatusByDateRequest{
				Addr:   contractAddr1,
				Start:  0,
				End:    0,
				Count:  10,
				Offset: 0,
			}); err != nil {
				t.Fatal(err)
			} else if len(records.GetStatuses()) != 1 {
				t.Fatalf("invalid GetCDRStatusByDate len, exp: 1, act: %d", len(records.GetStatuses()))
			}

			if allContracts, err := settlementApi.GetAllContracts(context.Background(), &pb.Offset{
				Count:  10,
				Offset: 0,
			}); err != nil {
				t.Fatal(err)
			} else if len(allContracts.GetContracts()) != 1 {
				t.Fatalf("invalid GetAllContracts len, exp: 1, act: %d", len(allContracts.GetContracts()))
			}

			if report, err := settlementApi.GetSummaryReport(context.Background(), &pb.ReportRequest{
				Addr:  contractAddr1,
				Start: 0,
				End:   0,
			}); err != nil {
				t.Fatal(err)
			} else {
				t.Log(report)
			}

			//if report, err := settlementApi.GetSummaryReportByAccount(context.Background(), &pb.ReportByAccountRequest{
			//	Addr:    contractAddr1,
			//	Account: account,
			//	Start:   0,
			//	End:     0,
			//}); err != nil {
			//	t.Fatal(err)
			//} else {
			//	if v, ok := report.GetRecords()[account]; !ok {
			//		t.Fatal("can not generate summary invoice")
			//	} else {
			//		t.Log(report, v)
			//	}
			//}

			if invoice, err := settlementApi.GenerateInvoicesByAccount(context.Background(), &pb.InvoicesByAccountRequest{
				Addr:    contractAddr1,
				Account: account,
				Start:   0,
				End:     0,
			}); err != nil {
				t.Fatal(err)
			} else {
				if len(invoice.GetRecords()) == 0 {
					t.Fatal("invalid invoice")
				} else {
					t.Log(util.ToString(invoice.GetRecords()))
				}
			}

			if invoices, err := settlementApi.GenerateInvoices(context.Background(), &pb.InvoicesRequest{
				Addr:  toAddressValue(pccwAddr),
				Start: 0,
				End:   0,
			}); err != nil {
				t.Fatal(err)
			} else {
				t.Log(invoices.GetRecords())
			}
			if invoices, err := settlementApi.GenerateInvoicesByContract(context.Background(), &pb.InvoicesRequest{
				Addr:  contractAddr1,
				Start: 0,
				End:   0,
			}); err != nil {
				t.Fatal(err)
			} else {
				t.Log(invoices.GetRecords())
			}

			//if report, err := settlementApi.GetSummaryReportByCustomer(context.Background(), &pb.ReportByCustomerRequest{
			//	Addr:     contractAddr1,
			//	Customer: customer,
			//	Start:    0,
			//	End:      0,
			//}); err != nil {
			//	t.Fatal(err)
			//} else {
			//	if v, ok := report.GetRecords()[customer]; !ok {
			//		t.Fatal("can not generate summary report")
			//	} else {
			//		t.Log(report, v)
			//	}
			//}

			if invoice, err := settlementApi.GenerateInvoicesByCustomer(context.Background(), &pb.InvoicesByCustomerRequest{
				Addr:     contractAddr1,
				Customer: customer,
				Start:    0,
				End:      0,
			}); err != nil {
				t.Fatal(err)
			} else {
				if len(invoice.GetRecords()) == 0 {
					t.Fatal("invalid invoice")
				} else {
					t.Log(util.ToString(invoice.GetRecords()))
				}
			}
		}
	}
}

func TestSettlementAPI_GetRegisterAssetBlock(t *testing.T) {
	testcase, verifier, settlementApi := setupSettlementAPI(t)
	defer testcase(t)

	r := &api.RegisterAssetParam{
		Owner:     assetParam.Owner,
		Assets:    assetParam.Assets,
		StartDate: assetParam.StartDate,
		EndDate:   assetParam.EndDate,
		Status:    cabi.AssetStatusActivated.String(),
	}
	address := account1.Address()
	r.Owner.Address = address

	t.Log(util.ToIndentString(r))

	if blk, err := settlementApi.GetRegisterAssetBlock(context.Background(), toRegisterAssetParam(r)); err != nil {
		t.Fatal(err)
	} else {
		bs, err := toOriginStateBlock(blk)
		if err != nil {
			t.Fatal(err)
		}
		txHash := bs.GetHash()
		bs.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(bs); err != nil {
			t.Fatal(err)
		}

		if blk, err := settlementApi.GetSettlementRewardsBlock(context.Background(), &pbtypes.Hash{
			Hash: toHashValue(txHash),
		}); err != nil {
			t.Fatal(err)
		} else {
			br, err := toOriginStateBlock(blk)
			if err != nil {
				t.Fatal(err)
			}
			txHash := br.GetHash()
			br.Signature = account1.Sign(txHash)
			if err := verifier.BlockProcess(br); err != nil {
				t.Fatal(err)
			}
		}
	}

	if assets, err := settlementApi.GetAllAssets(context.Background(), &pb.Offset{
		Count:  10,
		Offset: 0,
	}); err != nil {
		t.Fatal(err)
	} else if len(assets.GetParams()) != 1 {
		t.Fatalf("invalid assets len, exp: 1, act: %d", len(assets.GetParams()))
	}

	if assets, err := settlementApi.GetAssetsByOwner(context.Background(), &pb.AssetsByOwnerRequest{
		Owner:  toAddressValue(address),
		Count:  10,
		Offset: 0,
	}); err != nil {
		t.Fatal(err)
	} else if len(assets.GetParams()) != 1 {
		t.Fatalf("invalid assets len, exp: 1, act: %d", len(assets.GetParams()))
	} else {
		t.Log(util.ToIndentString(assets.GetParams()[0]))
		if asset, err := settlementApi.GetAsset(context.Background(), &pbtypes.Address{
			Address: assets.GetParams()[0].Address,
		}); err != nil {
			t.Fatal(err)
		} else if asset == nil {
			t.Fatal("")
		} else {
			t.Log(asset)
		}
	}
}
