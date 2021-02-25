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

func TestAirdropRewards(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.RewardsAddress)

	addr1 := account1.Address()
	b := mock.Address()
	am, err := l.GetAccountMeta(addr1)
	if err != nil {
		t.Fatal(err)
	}
	tm := am.Token(cfg.GasToken())

	param := &cabi.RewardsParam{
		Id:         mock.Hash(),
		Beneficial: b,
		TxHeader:   tm.Header,
		RxHeader:   types.Hash{},
		Amount:     big.NewInt(1e8),
		Sign:       types.Signature{},
	}

	if data, err := param.ToUnsignedABI(cabi.MethodNameUnsignedAirdropRewards); err != nil {
		t.Fatal(err)
	} else {
		h := types.HashData(data)
		param.Sign = account1.Sign(h)
		if abi, err := param.ToSignedABI(cabi.MethodNameAirdropRewards); err != nil {
			t.Fatal(err)
		} else {
			sb := &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: addr1,
				Balance: tm.Balance.Sub(types.Balance{Int: param.Amount}),
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.RewardsAddress),
				Representative: tm.Representative,
				Data:           abi,
				PoVHeight:      0,
				Timestamp:      common.TimeNow().Unix(),
			}

			rewards := AirdropRewards{}
			if err := rewards.DoSend(ctx, sb); err != nil {
				t.Fatal(err)
			}

			if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
				t.Fatal(err)
			}

			if pending, info, err := rewards.DoPending(sb); err != nil {
				t.Fatal(err)
			} else {
				t.Log(pending, info)
			}

			if r, err := rewards.GetTargetReceiver(ctx, sb); err != nil {
				t.Fatal(err)
			} else if r != b {
				t.Fatalf("invalid receive address, exp: %s, act: %s", b.String(), r.String())
			}

			_ = rewards.GetRefundData()

			rev := &types.StateBlock{
				Timestamp: time.Now().Unix(),
			}
			if r, err := rewards.DoReceive(ctx, rev, sb); err != nil {
				t.Fatal(err)
			} else {
				if len(r) == 0 {
					t.Fatal()
				}
				t.Log(r[0])
			}
		}
	}
}

func TestConfidantRewards(t *testing.T) {
	testCase, l := setupLedgerForTestCase(t)
	defer testCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.RewardsAddress)

	addr1 := account1.Address()
	b := account2.Address()
	am, err := l.GetAccountMeta(addr1)
	if err != nil {
		t.Fatal(err)
	}
	tm := am.Token(cfg.GasToken())

	param := &cabi.RewardsParam{
		Id:         mock.Hash(),
		Beneficial: b,
		TxHeader:   tm.Header,
		RxHeader:   types.Hash{},
		Amount:     big.NewInt(1e8),
		Sign:       types.Signature{},
	}

	if data, err := param.ToUnsignedABI(cabi.MethodNameUnsignedConfidantRewards); err != nil {
		t.Fatal(err)
	} else {
		h := types.HashData(data)
		param.Sign = account1.Sign(h)
		if abi, err := param.ToSignedABI(cabi.MethodNameConfidantRewards); err != nil {
			t.Fatal(err)
		} else {
			sb := &types.StateBlock{
				Type:    types.ContractSend,
				Token:   tm.Type,
				Address: addr1,
				Balance: tm.Balance.Sub(types.Balance{Int: param.Amount}),
				//Vote:           types.ZeroBalance,
				//Network:        types.ZeroBalance,
				//Oracle:         types.ZeroBalance,
				//Storage:        types.ZeroBalance,
				Previous:       tm.Header,
				Link:           types.Hash(contractaddress.RewardsAddress),
				Representative: tm.Representative,
				Data:           abi,
				PoVHeight:      0,
				Timestamp:      common.TimeNow().Unix(),
			}

			rewards := ConfidantRewards{}
			if err := rewards.DoSend(ctx, sb); err != nil {
				t.Fatal(err)
			}

			if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
				t.Fatal(err)
			}

			if pending, info, err := rewards.DoPending(sb); err != nil {
				t.Fatal(err)
			} else {
				t.Log(pending, info)
			}

			if r, err := rewards.GetTargetReceiver(ctx, sb); err != nil {
				t.Fatal(err)
			} else if r != b {
				t.Fatalf("invalid receive address, exp: %s, act: %s", b.String(), r.String())
			}

			_ = rewards.GetRefundData()

			rev := &types.StateBlock{
				Timestamp: time.Now().Unix(),
			}

			if r, err := rewards.DoReceive(ctx, rev, sb); err != nil {
				t.Fatal(err)
			} else {
				if len(r) == 0 {
					t.Fatal()
				}
				t.Log(r[0])
			}

			rev2 := &types.StateBlock{
				Timestamp: time.Now().Unix(),
			}
			if r, err := rewards.DoReceive(ctx, rev2, sb); err != nil {
				t.Fatal(err)
			} else {
				if len(r) == 0 {
					t.Fatal()
				}
				t.Log(r[0])
			}
		}
	}
}
