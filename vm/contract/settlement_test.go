// +build testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	sync2 "sync"
	"testing"
	"time"

	"github.com/bluele/gcache"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/sync"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi/settlement"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	createContractParam = cabi.CreateContractParam{
		PartyA: cabi.Contractor{
			Address: mock.Address(),
			Name:    "PCCWG",
		},
		PartyB: cabi.Contractor{
			Address: mock.Address(),
			Name:    "HKTCSL",
		},
		Previous: mock.Hash(),
		Services: []cabi.ContractService{
			{
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
			},
		},
		SignDate:  time.Now().AddDate(0, 0, -5).Unix(),
		StartDate: time.Now().AddDate(0, 0, -2).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 2).Unix(),
	}
)

func buildContract(l *ledger.Ledger) (contractAddress, a1, a2 types.Address, err error) {
	a1 = account1.Address()
	a2 = account2.Address()

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)

	if _, err = l.GetAccountMeta(a1); err != nil {
		return
	}

	if _, err = l.GetAccountMeta(a2); err != nil {
		return
	}

	tm, err := ctx.GetTokenMeta(a1, cfg.GasToken())
	if err != nil {
		return
	}

	param := createContractParam
	param.PartyA.Address = a1
	param.PartyB.Address = a2
	param.Previous = tm.Header

	contractAddress, _ = param.Address()

	var balance types.Balance
	if balance, err = param.Balance(); err != nil {
		return
	}
	if tm.Balance.Compare(balance) == types.BalanceCompSmaller {
		err = fmt.Errorf("not enough balance, [%s] of [%s]", balance.String(), tm.Balance.String())
		return
	}

	var abi []byte
	if abi, err = param.ToABI(); err == nil {
		sb := &types.StateBlock{
			Type:    types.ContractSend,
			Token:   tm.Type,
			Address: param.PartyA.Address,
			Balance: tm.Balance.Sub(balance),
			//Vote:           types.ZeroBalance,
			//Network:        types.ZeroBalance,
			//Oracle:         types.ZeroBalance,
			//Storage:        types.ZeroBalance,
			Previous:       param.Previous,
			Link:           types.Hash(contractaddress.SettlementAddress),
			Representative: tm.Representative,
			Data:           abi,
			Timestamp:      common.TimeNow().Unix(),
		}

		sb.Signature = account1.Sign(sb.GetHash())

		h := vmstore.TrieHash(ctx)
		if h != nil {
			povHeader, err2 := l.GetLatestPovHeader()
			if err2 != nil {
				err = err2
				return
			}
			sb.PoVHeight = povHeader.GetHeight()
			sb.Extra = h
		}
		//verifier := process.NewLedgerVerifier(l)
		//if err = verifier.BlockProcess(sb); err != nil {
		//	return
		//}
		if err = updateBlock(l, sb); err != nil {
			return
		}

		createContract := CreateContract{}
		if _, _, err = createContract.ProcessSend(ctx, sb); err != nil {
			return
		} else {
			if err = l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
				return
			}
		}

		return
	}
	return
}

func TestCreate_And_Terminate_Contract(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	if _, _, _, err := buildContract(l); err != nil {
		t.Fatal(err)
	}
	a1 := account1.Address()
	a2 := account2.Address()

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	//verifier := process.NewLedgerVerifier(l)

	if contractParams, err := cabi.GetContractsIDByAddressAsPartyA(l, &a1); err != nil {
		t.Fatal(err)
	} else {
		if len(contractParams) == 0 {
			t.Fatal("can not find any contact params")
		}
		for _, cp := range contractParams {
			if cp.PartyB.Address != a2 {
				t.Fatalf("invalid contract, partyB exp: %s,act: %s", a2.String(), cp.PartyB.Address.String())
			}
			if address, err := cp.Address(); err != nil {
				t.Fatal(err)
			} else {
				terminateContract := &TerminateContract{}
				tm, err := ctx.GetTokenMeta(a2, cfg.GasToken())
				if err != nil {
					t.Fatal(err)
				}
				param := cabi.TerminateParam{ContractAddress: address, Request: true}
				if abi, err := param.ToABI(); err == nil {
					sb := &types.StateBlock{
						Type:    types.ContractSend,
						Token:   tm.Type,
						Address: a2,
						Balance: tm.Balance,
						//Vote:           types.ZeroBalance,
						//Network:        types.ZeroBalance,
						//Oracle:         types.ZeroBalance,
						//Storage:        types.ZeroBalance,
						Previous:       tm.Header,
						Link:           types.Hash(contractaddress.SettlementAddress),
						Representative: tm.Representative,
						Data:           abi,
						Timestamp:      common.TimeNow().Unix(),
					}

					sb.Signature = account2.Sign(sb.GetHash())

					h := vmstore.TrieHash(ctx)
					if h != nil {
						povHeader, err := l.GetLatestPovHeader()
						if err != nil {
							t.Fatalf("get pov header error: %s", err)
						}
						sb.PoVHeight = povHeader.GetHeight()
						sb.Extra = h
					}

					//if err := verifier.BlockProcess(sb); err != nil {
					//	t.Fatal(err)
					//}
					if err := updateBlock(l, sb); err != nil {
						t.Fatal(err)
					}

					if _, _, err := terminateContract.ProcessSend(ctx, sb); err != nil {
						t.Fatal(err)
					} else {
						if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
							t.Fatal(err)
						} else {
							if r, err := terminateContract.GetTargetReceiver(ctx, sb); err != nil {
								t.Fatal(err)
							} else {
								if r != a1 {
									t.Fatalf("act: %s, exp: %s", r.String(), a1.String())
								}
							}

							if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
								t.Fatal(err)
							} else {
								if c.Status != cabi.ContractStatusRejected {
									t.Fatalf("invalid contract status, exp: %s, act: %s", cabi.ContractStatusRejected.String(), c.Status.String())
								} else {
									rb := &types.StateBlock{
										Timestamp: time.Now().Unix(),
									}

									if _, err := terminateContract.DoReceive(ctx, rb, sb); err != nil {
										t.Fatal(err)
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestEdit_Pre_Next_Stops(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	if _, _, _, err := buildContract(l); err != nil {
		t.Fatal(err)
	}
	a1 := account1.Address()
	a2 := account2.Address()
	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	//verifier := process.NewLedgerVerifier(l)
	if contractParams, err := cabi.GetContractsIDByAddressAsPartyA(l, &a1); err != nil {
		t.Fatal(err)
	} else {
		if len(contractParams) == 0 {
			t.Fatal("can not find any contact params")
		}
		for _, cp := range contractParams {
			//t.Log(cp.String())
			if cp.PartyB.Address != a2 {
				t.Fatalf("invalid contract, partyB exp: %s,act: %s", a2.String(), cp.PartyB.Address.String())
			}
			address, err := cp.Address()
			if err != nil {
				t.Fatal(err)
			}

			// add next stop
			tm, err := ctx.GetTokenMeta(a1, cfg.GasToken())
			if err != nil {
				t.Fatal(err)
			}
			stopParam := &cabi.StopParam{StopName: "HTKCSL", ContractAddress: address}
			abi, err := stopParam.ToABI(cabi.MethodNameAddNextStop)
			if err != nil {
				t.Fatal(err)
			}
			sb := &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: a1,
				Balance: tm.Balance,
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.SettlementAddress),
				Representative: tm.Representative,
				Data:           abi,
				Timestamp:      common.TimeNow().Unix(),
			}

			sb.Signature = account1.Sign(sb.GetHash())

			h := vmstore.TrieHash(ctx)
			if h != nil {
				sb.PoVHeight = 0
				sb.Extra = h
			}

			if err := updateBlock(l, sb); err != nil {
				t.Fatal(err)
			} else {
				addNextStop := &AddNextStop{}
				if _, _, err := addNextStop.ProcessSend(ctx, sb); err != nil {
					t.Fatal(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}

					if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
						t.Fatal(err)
					} else {
						if len(c.NextStops) != 1 {
							t.Fatalf("invalid next stop size: %d", len(c.NextStops))
						}

						if c.NextStops[0] != "HTKCSL" {
							t.Fatalf("invalid next stop, exp: HTKCSL, act: %s", c.NextStops[0])
						}
					}
				}
			}

			// update next stop
			tm, err = ctx.GetTokenMeta(a1, cfg.GasToken())
			if err != nil {
				t.Fatal(err)
			}
			sp := &cabi.UpdateStopParam{
				ContractAddress: address,
				StopName:        "HTKCSL",
				New:             "HTK-CSL",
			}
			abi, err = sp.ToABI(cabi.MethodNameUpdateNextStop)
			if err != nil {
				t.Fatal(err)
			}
			sb = &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: a1,
				Balance: tm.Balance,
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.SettlementAddress),
				Representative: tm.Representative,
				Data:           abi,
				Timestamp:      common.TimeNow().Unix(),
			}

			sb.Signature = account1.Sign(sb.GetHash())

			h = vmstore.TrieHash(ctx)
			if h != nil {
				sb.PoVHeight = 0
				sb.Extra = h
			}

			if err := updateBlock(l, sb); err != nil {
				t.Fatal(err)
			} else {
				updateNextStop := UpdateNextStop{}
				if _, _, err := updateNextStop.ProcessSend(ctx, sb); err != nil {
					t.Fatal(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}

					rb := &types.StateBlock{Timestamp: time.Now().Unix()}
					if _, err := updateNextStop.DoReceive(ctx, rb, sb); err != nil {
						t.Fatal(err)
					}
					if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
						t.Fatal(err)
					} else {
						if len(c.NextStops) != 1 {
							t.Fatalf("invalid next stop size: %d", len(c.NextStops))
						}

						if c.NextStops[0] != "HTK-CSL" {
							t.Fatalf("invalid next stop, exp: HTK-CSL, act: %s", c.NextStops[0])
						}
					}
				}
			}

			// remove next stop
			tm, err = ctx.GetTokenMeta(a1, cfg.GasToken())
			if err != nil {
				t.Fatal(err)
			}
			s2 := &cabi.StopParam{
				ContractAddress: address,
				StopName:        "HTK-CSL",
			}
			abi, err = s2.ToABI(cabi.MethodNameRemoveNextStop)
			if err != nil {
				t.Fatal(err)
			}
			sb = &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: a1,
				Balance: tm.Balance,
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.SettlementAddress),
				Representative: tm.Representative,
				Data:           abi,
				Timestamp:      common.TimeNow().Unix(),
			}

			sb.Signature = account1.Sign(sb.GetHash())

			h = vmstore.TrieHash(ctx)
			if h != nil {
				sb.PoVHeight = 0
				sb.Extra = h
			}

			if err := updateBlock(l, sb); err != nil {
				t.Fatal(err)
			} else {
				removeNextStop := RemoveNextStop{}
				if _, _, err := removeNextStop.ProcessSend(ctx, sb); err != nil {
					t.Fatal(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}
					rb := &types.StateBlock{Timestamp: time.Now().Unix()}
					if _, err := removeNextStop.DoReceive(ctx, rb, sb); err != nil {
						t.Fatal(err)
					}
					if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
						t.Fatal(err)
					} else {
						if len(c.NextStops) != 0 {
							t.Fatalf("invalid next stop size: %d", len(c.NextStops))
						}
					}
				}
			}

			// add pre stop
			tm, err = ctx.GetTokenMeta(a2, cfg.GasToken())
			if err != nil {
				t.Fatal(err)
			}
			stopParam = &cabi.StopParam{StopName: "PCCWG", ContractAddress: address}
			abi, err = stopParam.ToABI(cabi.MethodNameAddPreStop)
			if err != nil {
				t.Fatal(err)
			}
			sb = &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: a2,
				Balance: tm.Balance,
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.SettlementAddress),
				Representative: tm.Representative,
				Data:           abi,
				Timestamp:      common.TimeNow().Unix(),
			}

			sb.Signature = account2.Sign(sb.GetHash())

			h = vmstore.TrieHash(ctx)
			if h != nil {
				sb.PoVHeight = 0
				sb.Extra = h
			}

			if err := updateBlock(l, sb); err != nil {
				t.Fatal(err)
			} else {
				addPreStop := &AddPreStop{}
				if _, _, err := addPreStop.ProcessSend(ctx, sb); err != nil {
					t.Fatal(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}
					rb := &types.StateBlock{Timestamp: time.Now().Unix()}
					if _, err := addPreStop.DoReceive(ctx, rb, sb); err != nil {
						t.Fatal(err)
					}
					if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
						t.Fatal(err)
					} else {
						if len(c.PreStops) != 1 {
							t.Fatalf("invalid next stop size: %d", len(c.PreStops))
						}

						if c.PreStops[0] != "PCCWG" {
							t.Fatalf("invalid next stop, exp: PCCWG, act: %s", c.PreStops[0])
						}
					}
				}
			}

			// update pre stop
			tm, err = ctx.GetTokenMeta(a2, cfg.GasToken())
			if err != nil {
				t.Fatal(err)
			}
			ss2 := &cabi.UpdateStopParam{
				ContractAddress: address,
				StopName:        "PCCWG",
				New:             "PCCW-G",
			}
			abi, err = ss2.ToABI(cabi.MethodNameUpdatePreStop)
			if err != nil {
				t.Fatal(err)
			}
			sb = &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: a2,
				Balance: tm.Balance,
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.SettlementAddress),
				Representative: tm.Representative,
				Data:           abi,
				Timestamp:      common.TimeNow().Unix(),
			}

			sb.Signature = account2.Sign(sb.GetHash())

			h = vmstore.TrieHash(ctx)
			if h != nil {
				sb.PoVHeight = 0
				sb.Extra = h
			}

			if err := updateBlock(l, sb); err != nil {
				t.Fatal(err)
			} else {
				updatePreStop := &UpdatePreStop{}
				if _, _, err := updatePreStop.ProcessSend(ctx, sb); err != nil {
					t.Fatal(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}
					rb := &types.StateBlock{Timestamp: time.Now().Unix()}
					if _, err := updatePreStop.DoReceive(ctx, rb, sb); err != nil {
						t.Fatal(err)
					}
					if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
						t.Fatal(err)
					} else {
						if len(c.PreStops) != 1 {
							t.Fatalf("invalid next stop size: %d", len(c.PreStops))
						}

						if c.PreStops[0] != "PCCW-G" {
							t.Fatalf("invalid next stop, exp: PCCW-G, act: %s", c.PreStops[0])
						}
					}
				}
			}

			// remove pre stop
			tm, err = ctx.GetTokenMeta(a2, cfg.GasToken())
			if err != nil {
				t.Fatal(err)
			}
			ss3 := &cabi.StopParam{
				ContractAddress: address,
				StopName:        "PCCW-G",
			}
			abi, err = ss3.ToABI(cabi.MethodNameRemovePreStop)
			if err != nil {
				t.Fatal(err)
			}
			sb = &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: a2,
				Balance: tm.Balance,
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.SettlementAddress),
				Representative: tm.Representative,
				Data:           abi,
				Timestamp:      common.TimeNow().Unix(),
			}

			sb.Signature = account2.Sign(sb.GetHash())

			h = vmstore.TrieHash(ctx)
			if h != nil {
				sb.PoVHeight = 0
				sb.Extra = h
			}

			if err := updateBlock(l, sb); err != nil {
				t.Fatal(err)
			} else {
				removePreStop := &RemovePreStop{}
				if _, _, err := removePreStop.ProcessSend(ctx, sb); err != nil {
					t.Fatal(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}

					rb := &types.StateBlock{Timestamp: time.Now().Unix()}
					if _, err := removePreStop.DoReceive(ctx, rb, sb); err != nil {
						t.Fatal(err)
					}
					if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
						t.Fatal(err)
					} else {
						if len(c.PreStops) != 0 {
							t.Fatalf("invalid next stop size: %d", len(c.PreStops))
						}
					}
				}
			}
		}
	}
}

func TestCreate_And_Sign_Contract(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	a1 := account1.Address()
	a2 := account2.Address()
	if am, err := l.GetAccountMeta(a1); err != nil {
		t.Fatal(err)
	} else {
		t.Log(util.ToIndentString(am))
	}

	if am, err := l.GetAccountMeta(a2); err != nil {
		t.Fatal(err)
	} else {
		t.Log(util.ToIndentString(am))
	}

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	//verifier := process.NewLedgerVerifier(l)
	tm, err := ctx.GetTokenMeta(a1, cfg.GasToken())
	if err != nil {
		t.Fatal(err)
	}

	param := createContractParam
	param.PartyA.Address = a1
	param.PartyB.Address = a2
	param.Previous = tm.Header

	balance, err := param.Balance()
	if err != nil {
		t.Fatal(err)
	}
	if tm.Balance.Compare(balance) == types.BalanceCompSmaller {
		t.Fatalf("not enough balance, [%s] of [%s]", balance.String(), tm.Balance.String())
	}

	if abi, err := param.ToABI(); err == nil {
		sb := &types.StateBlock{
			Type:    types.ContractSend,
			Token:   tm.Type,
			Address: param.PartyA.Address,
			Balance: tm.Balance.Sub(balance),
			//Vote:           types.ZeroBalance,
			//Network:        types.ZeroBalance,
			//Oracle:         types.ZeroBalance,
			//Storage:        types.ZeroBalance,
			Previous:       param.Previous,
			Link:           types.Hash(contractaddress.SettlementAddress),
			Representative: tm.Representative,
			Data:           abi,
			Timestamp:      common.TimeNow().Unix(),
		}

		sb.Signature = account1.Sign(sb.GetHash())

		h := vmstore.TrieHash(ctx)
		if h != nil {
			povHeader, err := l.GetLatestPovHeader()
			if err != nil {
				t.Fatalf("get pov header error: %s", err)
			}
			sb.PoVHeight = povHeader.GetHeight()
			sb.Extra = h
		}

		if err := updateBlock(l, sb); err != nil {
			t.Fatal(err)
		}

		createContract := CreateContract{}
		if pendingKey, info, err := createContract.ProcessSend(ctx, sb); err != nil {
			t.Fatal(err)
		} else {
			t.Log(pendingKey, info)
			if _, _, err := createContract.ProcessSend(ctx, sb); err != nil {
				t.Fatal(err)
			}
			if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
				t.Fatal(err)
			}

			if r, err := createContract.GetTargetReceiver(ctx, sb); err != nil {
				t.Fatal(err)
			} else {
				if r != a2 {
					t.Fatalf("exp: %s, act: %s", a2.String(), r.String())
				}
			}
		}

		rev := &types.StateBlock{
			Timestamp: common.TimeNow().Unix(),
		}
		if rb, err := createContract.DoReceive(ctx, rev, sb); err != nil {
			t.Fatal(err)
		} else {
			if len(rb) > 0 {
				rb1 := rb[0].Block
				rb1.Signature = account1.Sign(rb1.GetHash())
				t.Log(rb1.String())
			} else {
				t.Fatal("fail to generate create contract reward block")
			}
		}

		if contractParams, err := cabi.GetContractsIDByAddressAsPartyA(l, &a1); err != nil {
			t.Fatal(err)
		} else {
			if len(contractParams) == 0 {
				t.Fatal("can not find any contact params")
			}
			for _, cp := range contractParams {
				if cp.PartyB.Address != a2 {
					t.Fatalf("invalid contract, partyB exp: %s,act: %s", a2.String(), cp.PartyB.Address.String())
				}
				if address, err := cp.Address(); err != nil {
					t.Fatal(err)
				} else {
					sc := cabi.SignContractParam{
						ContractAddress: address,
						ConfirmDate:     time.Now().Unix(),
					}
					tm2, err := ctx.GetTokenMeta(a2, cfg.GasToken())
					if err != nil {
						t.Fatal(err)
					}
					if tm2 == nil {
						t.Fatalf("failed to find token from %s", a2.String())
					}

					signContract := &SignContract{}

					if singedData, err := sc.ToABI(); err == nil {
						sb2 := &types.StateBlock{
							Type:    types.ContractSend,
							Token:   tm2.Type,
							Address: a2,
							Balance: tm2.Balance,
							//Vote:           types.ZeroBalance,
							//Network:        types.ZeroBalance,
							//Oracle:         types.ZeroBalance,
							//Storage:        types.ZeroBalance,
							Previous:       tm2.Header,
							Link:           types.Hash(contractaddress.SettlementAddress),
							Representative: tm.Representative,
							Data:           singedData,
							Timestamp:      common.TimeNow().Unix(),
						}

						sb2.Signature = account2.Sign(sb2.GetHash())
						if pk, info, err := signContract.ProcessSend(ctx, sb2); err != nil {
							t.Fatal(err)
						} else {
							t.Log(pk, " >>> ", info)
							if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
								t.Fatal(err)
							}

							if c, err := cabi.GetSettlementContract(ctx, &address); err != nil {
								t.Fatal(err)
							} else {
								if !c.IsAvailable() {
									t.Fatalf("failed to verify contract %s", address.String())
								}

								rev2 := &types.StateBlock{
									Timestamp: common.TimeNow().Unix(),
								}
								if rb, err := signContract.DoReceive(ctx, rev2, sb2); err != nil {
									t.Fatal(err)
								} else {
									if len(rb) > 0 {
										rb2 := rb[0].Block
										rb2.Signature = account2.Sign(rb2.GetHash())
										t.Log(rb2.String())

										// add prestop
										tm, err = ctx.GetTokenMeta(a1, cfg.GasToken())
										if err != nil {
											t.Fatal(err)
										}
										stopParam := &cabi.StopParam{StopName: "HTKCSL", ContractAddress: address}
										abi, err := stopParam.ToABI(cabi.MethodNameAddNextStop)
										if err != nil {
											t.Fatal(err)
										}
										sb := &types.StateBlock{
											Type:    types.ContractSend,
											Token:   tm.Type,
											Address: a1,
											Balance: tm.Balance,
											//Vote:           types.ZeroBalance,
											//Network:        types.ZeroBalance,
											//Oracle:         types.ZeroBalance,
											//Storage:        types.ZeroBalance,
											Previous:       tm.Header,
											Link:           types.Hash(contractaddress.SettlementAddress),
											Representative: tm.Representative,
											Data:           abi,
											Timestamp:      common.TimeNow().Unix(),
										}

										sb.Signature = account1.Sign(sb.GetHash())

										h := vmstore.TrieHash(ctx)
										if h != nil {
											sb.PoVHeight = 0
											sb.Extra = h
										}

										if err := updateBlock(l, sb); err != nil {
											t.Fatal(err)
										}
										addNextStop := &AddNextStop{}
										if pendingKey, info, err := addNextStop.ProcessSend(ctx, sb); err != nil {
											t.Fatal(err)
										} else {
											t.Log(pendingKey, info)
											if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
												t.Fatal(err)
											}
										}

										rev := &types.StateBlock{
											Timestamp: common.TimeNow().Unix(),
										}
										if rb, err := addNextStop.DoReceive(ctx, rev, sb); err != nil {
											t.Fatal(err)
										} else {
											if len(rb) > 0 {
												rb1 := rb[0].Block
												rb1.Signature = account1.Sign(rb1.GetHash())
												t.Log(rb1.String())
											} else {
												t.Fatal("fail to generate add next stop reward block")
											}
										}

										// add prestop
										tm2, err := ctx.GetTokenMeta(a2, cfg.GasToken())
										if err != nil {
											t.Fatal(err)
										}

										stopParam = &cabi.StopParam{StopName: "PCCWG", ContractAddress: address}
										abi, err = stopParam.ToABI(cabi.MethodNameAddPreStop)
										if err != nil {
											t.Fatal(err)
										}
										sb2 = &types.StateBlock{
											Type:    types.ContractSend,
											Token:   tm2.Type,
											Address: a2,
											Balance: tm.Balance,
											//Vote:           types.ZeroBalance,
											//Network:        types.ZeroBalance,
											//Oracle:         types.ZeroBalance,
											//Storage:        types.ZeroBalance,
											Previous:       tm2.Header,
											Link:           types.Hash(contractaddress.SettlementAddress),
											Representative: tm2.Representative,
											Data:           abi,
											Timestamp:      common.TimeNow().Unix(),
										}

										sb2.Signature = account2.Sign(sb.GetHash())

										h = vmstore.TrieHash(ctx)
										if h != nil {
											sb2.PoVHeight = 0
											sb2.Extra = h
										}

										if err := updateBlock(l, sb2); err != nil {
											t.Fatal(err)
										}
										addPreStop := &AddPreStop{}
										if pendingKey, info, err := addPreStop.ProcessSend(ctx, sb2); err != nil {
											t.Fatal(err)
										} else {
											t.Log(pendingKey, info)
											if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
												t.Fatal(err)
											}
										}

										rev2 := &types.StateBlock{
											Timestamp: common.TimeNow().Unix(),
										}
										if rb, err := addPreStop.DoReceive(ctx, rev2, sb2); err != nil {
											t.Fatal(err)
										} else {
											if len(rb) > 0 {
												rb1 := rb[0].Block
												rb1.Signature = account2.Sign(rb1.GetHash())
												t.Log(rb1.String())
											} else {
												t.Fatal("fail to generate add pre stop reward block")
											}
										}
										// start process CDR as ac1
										cdrContract := &ProcessCDR{}

										tm, err = ctx.GetTokenMeta(a1, cfg.GasToken())
										if err != nil {
											t.Fatal(err)
										}
										cdr1 := &cabi.CDRParamList{
											ContractAddress: address,
											Params: []*cabi.CDRParam{
												{
													Index:         1,
													SmsDt:         time.Now().Unix(),
													Sender:        "WeChat",
													Destination:   "85257***343",
													SendingStatus: 0,
													DlrStatus:     0,
													PreStop:       "",
													NextStop:      "HKTCSL",
												},
											},
										}
										abi, err = cdr1.ToABI()
										if err != nil {
											t.Fatal(err)
										}
										sb = &types.StateBlock{
											Type:    types.ContractSend,
											Token:   tm.Type,
											Address: a1,
											Balance: tm.Balance,
											//Vote:           types.ZeroBalance,
											//Network:        types.ZeroBalance,
											//Oracle:         types.ZeroBalance,
											//Storage:        types.ZeroBalance,
											Previous:       tm.Header,
											Link:           types.Hash(contractaddress.SettlementAddress),
											Representative: tm.Representative,
											Data:           abi,
											Timestamp:      common.TimeNow().Unix(),
										}

										sb.Signature = account1.Sign(sb.GetHash())

										h = vmstore.TrieHash(ctx)
										if h != nil {
											sb.PoVHeight = 0
											sb.Extra = h
										}

										if err := updateBlock(l, sb); err != nil {
											t.Fatal(err)
										}

										if pk, pi, err := cdrContract.ProcessSend(ctx, sb); err != nil {
											t.Fatal(err)
										} else {
											t.Log(pk, pi)
										}
										// start process CDR as ac2
										tm2, err = ctx.GetTokenMeta(a2, cfg.GasToken())
										if err != nil {
											t.Fatal(err)
										}
										cdr2 := &cabi.CDRParamList{
											ContractAddress: address,
											Params: []*cabi.CDRParam{
												{
													Index:         1,
													SmsDt:         time.Now().Unix(),
													Sender:        "WeChat",
													Destination:   "85257***343",
													SendingStatus: 0,
													DlrStatus:     0,
													PreStop:       "PCCWG",
													NextStop:      "",
												},
											},
										}
										abi, err = cdr2.ToABI()
										if err != nil {
											t.Fatal(err)
										}
										sb = &types.StateBlock{
											Type:    types.ContractSend,
											Token:   tm2.Type,
											Address: a2,
											Balance: tm2.Balance,
											//Vote:           types.ZeroBalance,
											//Network:        types.ZeroBalance,
											//Oracle:         types.ZeroBalance,
											//Storage:        types.ZeroBalance,
											Previous:       tm2.Header,
											Link:           types.Hash(contractaddress.SettlementAddress),
											Representative: tm2.Representative,
											Data:           abi,
											Timestamp:      common.TimeNow().Unix(),
										}

										sb.Signature = account2.Sign(sb.GetHash())

										h = vmstore.TrieHash(ctx)
										if h != nil {
											sb.PoVHeight = 0
											sb.Extra = h
										}

										if err := updateBlock(l, sb); err != nil {
											t.Fatal(err)
										}

										rb := &types.StateBlock{
											Timestamp: time.Now().Unix(),
										}

										if _, err := cdrContract.DoReceive(ctx, rb, sb); err != nil {
											t.Fatal(err)
										}

										if pk, pi, err := cdrContract.ProcessSend(ctx, sb); err != nil {
											t.Fatal(err)
										} else {
											t.Log(pk, pi)

											if hash, err := cdr1.Params[0].ToHash(); err != nil {
												t.Fatal(err)
											} else {
												if status, err := cabi.GetCDRStatus(ctx, &address, hash); err != nil {
													t.Fatal(err)
												} else {
													t.Log(status)
												}
											}
										}
									} else {
										t.Fatal("fail to generate sign contract reward block")
									}
								}
							}
						}
					} else {
						t.Fatal(err)
					}
				}
			}
		}
	}
}

func TestCreateContract_GetDescribe(t *testing.T) {
	tests := []struct {
		name string
		want Describe
	}{
		{
			name: "default",
			want: Describe{
				specVer:   SpecVer2,
				signature: true,
				pending:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CreateContract{}
			if got := c.GetDescribe(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDescribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContract_GetFee(t *testing.T) {
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		args    args
		want    types.Balance
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx:   nil,
				block: nil,
			},
			want:    types.ZeroBalance,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CreateContract{}
			got, err := c.GetFee(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFee() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateContract_GetRefundData(t *testing.T) {
	tests := []struct {
		name string
		want []byte
	}{
		{
			name: "default",
			want: []byte{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CreateContract{}
			if got := c.GetRefundData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRefundData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessCDR_GetDescribe(t *testing.T) {
	tests := []struct {
		name string
		want Describe
	}{
		{
			name: "default",
			want: Describe{
				specVer:   SpecVer2,
				signature: true,
				pending:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessCDR{}
			if got := p.GetDescribe(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDescribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessCDR_GetFee(t *testing.T) {
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		args    args
		want    types.Balance
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx:   nil,
				block: nil,
			},
			want:    types.ZeroBalance,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessCDR{}
			got, err := p.GetFee(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFee() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessCDR_GetRefundData(t *testing.T) {
	tests := []struct {
		name string
		want []byte
	}{
		{
			name: "default",
			want: []byte{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessCDR{}
			if got := p.GetRefundData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRefundData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignContract_GetDescribe(t *testing.T) {
	tests := []struct {
		name string
		want Describe
	}{
		{
			name: "",
			want: Describe{
				specVer:   SpecVer2,
				signature: true,
				pending:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignContract{}
			if got := s.GetDescribe(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDescribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignContract_GetFee(t *testing.T) {
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		args    args
		want    types.Balance
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				ctx:   nil,
				block: nil,
			},
			want:    types.ZeroBalance,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignContract{}
			got, err := s.GetFee(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFee() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignContract_GetRefundData(t *testing.T) {
	tests := []struct {
		name string
		want []byte
	}{
		{
			name: "default",
			want: []byte{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignContract{}
			if got := s.GetRefundData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRefundData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newLocker(t *testing.T) {
	l := &locker{cache: gcache.New(100).LRU().LoaderFunc(func(key interface{}) (i interface{}, err error) {
		return sync.NewMutex(), nil
	}).Expiration(time.Second * 2).Build()}
	k := []byte("test")

	var m1 *sync.Mutex
	if m, err := l.Get(k); err != nil {
		t.Fatal(err)
	} else {
		m1 = m
		m1.Lock()
		if b := m1.IsLocked(); !b {
			t.Fatal("invalid lock status")
		}
		m1.Unlock()
		if m2, err := l.Get(k); err != nil {
			t.Fatal(err)
		} else {
			if m1 != m2 {
				t.Fatalf("invalid m1: %v, m2: %v", m1, m2)
			} else {
				t.Log(m1, m2)
			}
		}
	}

	time.Sleep(3 * time.Second)

	all := l.cache.GetALL(true)
	if len(all) != 0 {
		for k, v := range all {
			t.Log(k, v)
		}
		t.Fatal("len should be zero")
	}

	if m3, err := l.Get(k); err != nil {
		t.Fatal(err)
	} else {
		if m3 == m1 {
			t.Fatalf("invalid expire locker, %v,%v", m1, m3)
		} else {
			t.Log(m1, m3)
		}
	}

	all = l.cache.GetALL(true)

	for k, v := range all {
		t.Log("all2: ", k, v)
	}
}

func Test_add(t *testing.T) {
	type args struct {
		s    []string
		name string
		want []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				s:    nil,
				name: "11",
				want: []string{"11"},
			},
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				s:    []string{"22", "11"},
				name: "11",
				want: []string{"22", "11"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := add(tt.args.s, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("add() error = %v, wantErr %v", err, tt.wantErr)
				if !reflect.DeepEqual(got, tt.args.want) {
					t.Errorf("should be the same when err, got = %v, want = %v", got, tt.args.want)
				}
			} else {
				if !reflect.DeepEqual(got, tt.args.want) {
					t.Errorf("add() got = %v, want = %v", got, tt.args.want)
				}
			}
		})
	}
}

func Test_remove(t *testing.T) {
	type args struct {
		s    []string
		name string
		want []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				s:    []string{"11", "22"},
				name: "22",
				want: []string{"11"},
			},
			wantErr: false,
		}, {
			name: "f1",
			args: args{
				s:    nil,
				name: "22",
				want: nil,
			},
			wantErr: true,
		}, {
			name: "f2",
			args: args{
				s:    []string{"11", "22"},
				name: "33",
				want: []string{"11", "22"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := remove(tt.args.s, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("remove() error = %v, wantErr %v", err, tt.wantErr)
				if !reflect.DeepEqual(got, tt.args.want) {
					t.Errorf("should be the same when err, got = %v, want = %v", got, tt.args.want)
				}
			} else {
				sort.Strings(tt.args.want)
				if !reflect.DeepEqual(got, tt.args.want) {
					t.Errorf("remove() got = %v, want = %v", got, tt.args.want)
				}
			}
		})
	}
}

func Test_update(t *testing.T) {
	type args struct {
		s    []string
		old  string
		new  string
		want []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				s:    []string{"11", "22"},
				old:  "11",
				new:  "33",
				want: []string{"22", "33"},
			},
			wantErr: false,
		}, {
			name: "f1",
			args: args{
				s:    nil,
				old:  "11",
				new:  "33",
				want: nil,
			},
			wantErr: true,
		}, {
			name: "f2",
			args: args{
				s:    []string{"11", "22"},
				old:  "33",
				new:  "111",
				want: []string{"11", "22"},
			},
			wantErr: true,
		}, {
			name: "f3",
			args: args{
				s:    []string{"11", "22"},
				old:  "22",
				new:  "11",
				want: []string{"11", "22"},
			},
			wantErr: true,
		}, {
			name: "same",
			args: args{
				s:    []string{"11", "22"},
				old:  "22",
				new:  "22",
				want: []string{"11", "22"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := update(tt.args.s, tt.args.old, tt.args.new); (err != nil) != tt.wantErr {
				t.Errorf("update() error = %v, wantErr %v", err, tt.wantErr)
				if !reflect.DeepEqual(got, tt.args.want) {
					t.Errorf("should be the same when err, got = %v, want = %v", got, tt.args.want)
				}
			} else {
				sort.Strings(tt.args.want)
				if !reflect.DeepEqual(got, tt.args.want) {
					t.Errorf("update() got = %v, want = %v", got, tt.args.want)
				}
			}
		})
	}
}

func Test_verifyStopName(t *testing.T) {
	s := []string{"HKTCSL"}
	b, i := verifyStopName(s, "HKT-CSL")
	t.Log(b, i)
}

func Test_internalContract_DoPending(t *testing.T) {
	i := internalContract{}
	i.DoGap(nil, nil)
	_, _, _ = i.DoPending(nil)
	_ = i.DoReceiveOnPov(nil, nil, 0, nil, nil)
	_ = i.DoSendOnPov(nil, nil, 0, nil)
	_, _ = i.GetTargetReceiver(nil, nil)
}

func Test_timeString(t *testing.T) {
	t.Log(timeString(0))
	t.Log(timeString(-1))
	t.Log(timeString(time.Now().Unix()))
}

func TestRegisterAsset_ProcessSend(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
	//verifier := process.NewLedgerVerifier(l)

	a1 := account1.Address()
	tm, err := ctx.GetTokenMeta(a1, cfg.GasToken())
	if err != nil {
		t.Fatal(err)
	}

	assetParam := &cabi.AssetParam{
		Owner: cabi.Contractor{
			Address: a1,
			Name:    "HKT-CSL",
		},
		Previous: tm.Header,
		Assets: []*cabi.Asset{
			{
				Mcc:         42,
				Mnc:         5,
				TotalAmount: 1000,
				SLAs:        nil,
			},
		},
		SignDate:  time.Now().Unix(),
		StartDate: time.Now().AddDate(0, 0, 1).Unix(),
		EndDate:   time.Now().AddDate(1, 0, 1).Unix(),
		Status:    cabi.AssetStatusActivated,
	}

	if abi, err := assetParam.ToABI(); err == nil {
		sb := &types.StateBlock{
			Type:    types.ContractSend,
			Token:   tm.Type,
			Address: a1,
			Balance: tm.Balance,
			//Vote:           types.ZeroBalance,
			//Network:        types.ZeroBalance,
			//Oracle:         types.ZeroBalance,
			//Storage:        types.ZeroBalance,
			Previous:       tm.Header,
			Link:           types.Hash(contractaddress.SettlementAddress),
			Representative: tm.Representative,
			Data:           abi,
			Timestamp:      common.TimeNow().Unix(),
		}

		sb.Signature = account1.Sign(sb.GetHash())

		h := vmstore.TrieHash(ctx)
		if h != nil {
			povHeader, err := l.GetLatestPovHeader()
			if err != nil {
				t.Fatalf("get pov header error: %s", err)
			}
			sb.PoVHeight = povHeader.GetHeight()
			sb.Extra = h
		}

		if err := updateBlock(l, sb); err != nil {
			t.Fatal(err)
		}

		registerAsset := RegisterAsset{}
		if pendingKey, info, err := registerAsset.ProcessSend(ctx, sb); err != nil {
			t.Fatal(err)
		} else {
			t.Log(pendingKey, info)
			if _, _, err := registerAsset.ProcessSend(ctx, sb); err != nil {
				t.Fatal(err)
			}
			if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
				t.Fatal(err)
			}
		}

		rev := &types.StateBlock{
			Timestamp: common.TimeNow().Unix(),
		}
		if rb, err := registerAsset.DoReceive(ctx, rev, sb); err != nil {
			t.Fatal(err)
		} else {
			if len(rb) > 0 {
				rb1 := rb[0].Block
				rb1.Signature = account1.Sign(rb1.GetHash())
				t.Log(rb1.String())
			} else {
				t.Fatal("fail to generate create contract reward block")
			}
		}
	} else {
		t.Fatal(err)
	}
}

func TestProcessCDR_save(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	if ca, a1, a2, err := buildContract(l); err != nil {
		t.Fatal(err)
	} else {
		t.Log(ca.String())
		t.Log(a1.String())
		t.Log(a2.String())
		p := &ProcessCDR{}
		ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
		data, err := cabi.GetContractParam(ctx, &ca)
		if err != nil {
			t.Fatal(err)
		}
		param, err := cabi.ParseContractParam(data)
		if err != nil {
			t.Fatal(err)
		}

		var cdrs []*cabi.CDRParam
		for i := 0; i < 10; i++ {
			cdr := &cabi.CDRParam{
				Index:         uint64(10000 + i),
				SmsDt:         time.Now().Unix(),
				Sender:        "PCCWG",
				Destination:   "85257***34" + strconv.Itoa(i),
				SendingStatus: cabi.SendingStatusSent,
				DlrStatus:     cabi.DLRStatusDelivered}
			cdrs = append(cdrs, cdr)
		}

		wg := sync2.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()

			sb := &types.StateBlock{
				Address:   a1,
				Timestamp: time.Now().Unix(),
			}

			for _, cdr := range cdrs {
				ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
				h, _ := cdr.ToHash()
				key, _ := types.HashBytes(ca[:], h[:])
				t.Log(a1.String(), ": ", key.String())
				if err = p.save(ctx, sb, &ca, param, cdr); err != nil {
					t.Log(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			sb := &types.StateBlock{
				Address:   a2,
				Timestamp: time.Now().Unix(),
			}

			for _, cdr := range cdrs {
				ctx := vmstore.NewVMContext(l, &contractaddress.SettlementAddress)
				h, _ := cdr.ToHash()
				key, _ := types.HashBytes(ca[:], h[:])
				t.Log(a2.String(), ": ", key.String())
				if err = p.save(ctx, sb, &ca, param, cdr); err != nil {
					t.Log(err)
				} else {
					if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
						t.Fatal(err)
					}
				}
			}
		}()

		//ctx1, cancelFunc := context.WithCancel(context.Background())
		//defer cancelFunc()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				all := lockerCache.cache.GetALL(true)
				i := len(all)
				j := 0
				for _, v := range all {
					if !v.(*sync.Mutex).IsLocked() {
						j++
					}
				}
				if i == j {
					break
				}

				s := PrintLocker()
				if s != "" {
					t.Log(s)
				}
			}
		}()

		wg.Wait()
		//cancelFunc()

		time.Sleep(time.Second)

		if records, err := cabi.GetCDRStatusByDate(l, &ca, 0, 0); err != nil {
			t.Fatal(err)
		} else {
			if len(records) != 10 {
				t.Fatalf("invalid records len, %d", len(records))
			} else {
				// for _, record := range records {
				// 	if len(record.Params) != 2 {
				// 		t.Fatalf("invalid record, %s", record.String())
				// 	} else {
				// 		//t.Log(record.String())
				// 	}
				// }
			}
		}
	}
}
