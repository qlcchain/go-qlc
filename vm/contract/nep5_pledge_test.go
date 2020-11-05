package contract

import (
	"math/big"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestNep5Pledge_And_Withdraw(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.NEP5PledgeAddress)

	addr1 := account1.Address()
	addr2 := account2.Address()

	am, err := l.GetAccountMeta(addr1)
	if err != nil {
		t.Fatal(err)
	}
	tm := am.Token(cfg.ChainToken())

	param := &cabi.PledgeParam{
		Beneficial:    addr2,
		PledgeAddress: addr1,
		PType:         uint8(cabi.Vote),
		NEP5TxId:      mock.Hash().String(),
	}
	data, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        addr1,
		Balance:        tm.Balance.Sub(types.Balance{Int: big.NewInt(1e8)}),
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.NEP5PledgeAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      0,
		Timestamp:      common.TimeNow().Unix(),
	}

	pledge := Nep5Pledge{}
	if err = pledge.DoSend(ctx, send); err != nil {
		t.Fatal(err)
	}

	//if pending, info, err := pledge.DoPending(send); err != nil {
	//	t.Fatal(err)
	//} else {
	//	t.Log(pending, info)
	//}

	if receiver, err := pledge.GetTargetReceiver(ctx, send); err != nil {
		t.Fatal(err)
	} else if receiver != addr2 {
		t.Fatalf("invalid target receiver, exp: %s, act: %s", addr2.String(), receiver.String())
	}

	//verifier := process.NewLedgerVerifier(l)
	//if err = verifier.BlockProcess(send); err != nil {
	//	t.Fatal(err)
	//}
	if err := updateBlock(l, send); err != nil {
		t.Fatal(err)
	}

	rev := &types.StateBlock{
		Timestamp: time.Now().Unix(),
	}

	if r, err := pledge.DoReceive(ctx, rev, send); err != nil {
		t.Fatal(err)
	} else {
		if len(r) > 0 {
			if err := updateBlock(l, r[0].Block); err != nil {
				t.Fatal(err)
			}
			//verifier.BlockProcess(r[0].Block)
		}
	}

	if _, err := pledge.DoReceive(ctx, rev, send); err == nil {
		t.Fatal()
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	_ = pledge.GetRefundData()

	// patch withdraw time
	pledgeKey := cabi.GetPledgeKey(addr1, param.Beneficial, param.NEP5TxId)
	var pledgeData []byte
	if pledgeData, err = ctx.GetStorage(contractaddress.NEP5PledgeAddress[:], pledgeKey); err != nil {
		t.Fatal(err)
	} else {
		if len(pledgeData) > 0 {
			info, err := cabi.ParsePledgeInfo(pledgeData)
			if err != nil {
				t.Fatal(err)
			} else {
				info.WithdrawTime = time.Now().AddDate(0, 0, -1).Unix()
				//save data
				if pledgeData, err = info.ToABI(); err != nil {
					t.Fatal(err)
				} else {
					if err = ctx.SetStorage(contractaddress.NEP5PledgeAddress[:], pledgeKey, pledgeData); err != nil {
						t.Fatal(err)
					}
				}
			}
		} else {
		}
	}

	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}

	withdrawNep5Pledge := WithdrawNep5Pledge{}

	wp := &cabi.WithdrawPledgeParam{
		Beneficial: param.Beneficial,
		Amount:     big.NewInt(1e8),
		PType:      param.PType,
		NEP5TxId:   param.NEP5TxId,
	}

	if am, err := l.GetAccountMeta(wp.Beneficial); err != nil {
		t.Fatal(err)
	} else {
		tm := am.Token(cfg.ChainToken())
		if tm == nil {
			t.Fatal()
		}
		if abi, err := wp.ToABI(); err != nil {
			t.Fatal(err)
		} else {
			send2 := &types.StateBlock{
				Type:           types.ContractSend,
				Token:          tm.Type,
				Address:        wp.Beneficial,
				Balance:        am.CoinBalance,
				Vote:           types.ToBalance(am.CoinVote),
				Network:        types.ToBalance(am.CoinNetwork),
				Oracle:         types.ToBalance(am.CoinOracle),
				Storage:        types.ToBalance(am.CoinStorage),
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.NEP5PledgeAddress),
				Representative: tm.Representative,
				Data:           abi,
				PoVHeight:      0,
				Timestamp:      common.TimeNow().Unix(),
			}
			send2.Vote = types.ToBalance(send.GetVote().Sub(types.Balance{Int: wp.Amount}))

			if err := withdrawNep5Pledge.DoSend(ctx, send2); err != nil {
				t.Fatal(err)
			}

			if r, err := withdrawNep5Pledge.GetTargetReceiver(ctx, send2); err != nil {
				t.Fatal(err)
			} else if r != addr1 {
				t.Fatalf("invalid receive addr, exp: %s, act: %s", r, addr1)
			}

			if _, _, err := withdrawNep5Pledge.DoPending(send2); err == nil {
				t.Fatal()
			}

			rev := &types.StateBlock{
				Timestamp: time.Now().Unix(),
			}

			if r, err := withdrawNep5Pledge.DoReceive(ctx, rev, send2); err != nil {
				t.Fatal(err)
			} else {
				t.Log(r)
			}

			_ = withdrawNep5Pledge.GetRefundData()
		}
	}
}

func TestNep5Pledge_DoSend(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.NEP5PledgeAddress)

	addr1 := account1.Address()
	addr2 := account2.Address()

	am, err := l.GetAccountMeta(addr1)
	if err != nil {
		t.Fatal(err)
	}
	tm := am.Token(cfg.ChainToken())

	param := &cabi.PledgeParam{
		Beneficial:    addr2,
		PledgeAddress: addr1,
		PType:         uint8(cabi.Network),
		NEP5TxId:      mock.Hash().String(),
	}
	data, err := param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	temp := types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        addr1,
		Balance:        tm.Balance.Sub(types.Balance{Int: big.NewInt(2000 * 1e8)}),
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Previous:       tm.Header,
		Link:           types.Hash(contractaddress.NEP5PledgeAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      0,
		Timestamp:      common.TimeNow().Unix(),
	}

	send := &temp

	send2 := temp
	send2.Address = mock.Address()

	send3 := temp
	send3.Address = addr2

	send4 := temp
	send4.Balance = tm.Balance.Sub(types.Balance{Int: big.NewInt(1e8)})

	send5 := temp
	send5.Data = []byte{}

	send6 := temp
	send6.Token = cfg.GasToken()

	type fields struct {
		BaseContract BaseContract
	}
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				BaseContract: BaseContract{},
			},
			args: args{
				ctx:   ctx,
				block: send,
			},
			wantErr: false,
		}, {
			name: "f1",
			fields: fields{
				BaseContract: BaseContract{},
			},
			args: args{
				ctx:   ctx,
				block: &send2,
			},
			wantErr: true,
		}, {
			name: "f3",
			fields: fields{
				BaseContract: BaseContract{},
			},
			args: args{
				ctx:   ctx,
				block: &send3,
			},
			wantErr: true,
		}, {
			name: "f4",
			fields: fields{
				BaseContract: BaseContract{},
			},
			args: args{
				ctx:   ctx,
				block: &send4,
			},
			wantErr: true,
		}, {
			name: "f5",
			fields: fields{
				BaseContract: BaseContract{},
			},
			args: args{
				ctx:   ctx,
				block: &send5,
			},
			wantErr: true,
		}, {
			name: "f6",
			fields: fields{
				BaseContract: BaseContract{},
			},
			args: args{
				ctx:   ctx,
				block: &send6,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ne := &Nep5Pledge{
				BaseContract: tt.fields.BaseContract,
			}
			if err := ne.DoSend(tt.args.ctx, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("DoSend() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
