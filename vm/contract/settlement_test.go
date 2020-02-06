// +build !testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package contract

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/mock"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
)

const (
	accountBlocks = `[
	{
		"type": "Open",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "100000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "c0d330096ec4ab6ccf5481e06cc54e74b14f534e99e38df486f47d1123cbd1ae",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "747648bafd344347582876662641c4b8ffbf20a85ba01dc559ff930435bc5bad",
		"povHeight": 0,
		"timestamp": 1580997079,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "000000000122e972",
		"signature": "5460905ad2096d1822dc086e8fe375409f9fc87f0e8288ca215a399eb2fee6c6c5fc94b53a18f62fe6d124f869cbac1737b762c9a8f7654d1b7ecacc480f010a"
	},
	{
		"type": "Send",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "40000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "cad0cad8a98813787dc11ba2413afca574f2d62e222fd2644cc33c7d70124d90",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "f82eae0fa0f56a53e9d217140eaa33219c7cb910439501f333383f4d6147618c",
		"povHeight": 0,
		"timestamp": 1580997083,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000002aa56ad",
		"signature": "dd0af652dc5acca94547b5e130a38a2728531235d224e6031f13d9958221fcd4852bee6684fa1b29c5170f71f7f301b64eda8d10208ecc12b2862b34ea049a0e"
	},
	{
		"type": "Open",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te",
		"balance": "60000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "b05f7c462867df6f24b810c0b28b50d709667feb7d870a2b1db23bb3fa491249",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "eb9c1dcccaba3937d8745c364dadb1ca056cfa9540184277ad6fe8af66f81358",
		"povHeight": 0,
		"timestamp": 1580997093,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000002389ad",
		"signature": "5d35efd693b85ccf4a01e4f132aa0a248b328b10024af2a22e65474038a4aea3decb6c414c5563b326f509f4cf5eac852c81317a96a8c349b965849c31e5580d"
	}
]`
)

var (
	// qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq
	priv1, _ = hex.DecodeString("7098c089e66bd66476e3b88df8699bcd4dacdd5e1e5b41b3c598a8a36d851184d992a03b7326b7041f689ae727292d761b329a960f3e4335e0a7dcf2c43c4bcf")
	// qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te
	priv2, _ = hex.DecodeString("31ee4e16826569dc631b969e71bd4c46d5c0df0daeca6933f46586f36f49537cd929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393")
	ac1      = types.NewAccount(priv1)
	ac2      = types.NewAccount(priv2)

	createContractParam = cabi.CreateContractParam{
		PartyA:      ac1.Address(),
		PartyAName:  "c1",
		PartyB:      ac2.Address(),
		PartyBName:  "c2",
		Previous:    mock.Hash(),
		ServiceId:   mock.Hash().String(),
		Mcc:         1,
		Mnc:         2,
		TotalAmount: 100,
		UnitPrice:   2,
		Currency:    "USD",
		SignDate:    time.Now().Unix(),
		SignatureA:  types.ZeroSignature,
	}
)

func setupSettlementTestCase(t *testing.T) (func(t *testing.T), *ledger.Ledger) {
	dir := filepath.Join(cfg.QlcTestDataDir(), "settlement", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := cfg.NewCfgManager(dir)
	c, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	var mintageBlock, genesisBlock types.StateBlock
	for _, v := range c.Genesis.GenesisBlocks {
		_ = json.Unmarshal([]byte(v.Genesis), &genesisBlock)
		_ = json.Unmarshal([]byte(v.Mintage), &mintageBlock)
		genesisInfo := &common.GenesisInfo{
			ChainToken:          v.ChainToken,
			GasToken:            v.GasToken,
			GenesisMintageBlock: mintageBlock,
			GenesisBlock:        genesisBlock,
		}
		common.GenesisInfos = append(common.GenesisInfos, genesisInfo)
	}
	l := ledger.NewLedger(cm.ConfigFile)
	//ctx := vmstore.NewVMContext(l)
	//verifier := process.NewLedgerVerifier(l)
	//
	//for _, v := range common.GenesisInfos {
	//	mb := v.GenesisMintageBlock
	//	gb := v.GenesisBlock
	//	err := ctx.SetStorage(types.MintageAddress[:], v.GenesisBlock.Token[:], v.GenesisBlock.Data)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	if b, err := l.HasStateBlock(mb.GetHash()); !b && err == nil {
	//		if err := l.AddStateBlock(&mb); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	if b, err := l.HasStateBlock(gb.GetHash()); !b && err == nil {
	//		if err := verifier.BlockProcess(&gb); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//}
	//_ = ctx.SaveStorage()

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(accountBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	for i := range blocks {
		block := blocks[i]
		//if err := verifier.BlockProcess(block); err != nil {
		//	t.Fatal(err)
		//}
		if err := updateBlock(l, block); err != nil {
			t.Fatal(err)
		}
	}

	return func(t *testing.T) {
		//err := l.Store.Erase()
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

func updateBlock(l *ledger.Ledger, block *types.StateBlock) error {
	return l.BatchUpdate(func(txn db.StoreTxn) error {
		err := l.AddStateBlock(block, txn)
		if err != nil {
			return err
		}
		am, err := l.GetAccountMetaConfirmed(block.GetAddress(), txn)
		if err != nil && err != ledger.ErrAccountNotFound {
			return fmt.Errorf("get account meta error: %s", err)
		}
		tm, err := l.GetTokenMetaConfirmed(block.GetAddress(), block.GetToken(), txn)
		if err != nil && err != ledger.ErrAccountNotFound && err != ledger.ErrTokenNotFound {
			return fmt.Errorf("get token meta error: %s", err)
		}
		err = updateFrontier(l, block, tm, txn)
		if err != nil {
			return err
		}
		err = updateAccountMeta(l, block, am, txn)
		if err != nil {
			return err
		}
		return nil
	})
}

func TestSettlement_Create_And_Sign_Contract(t *testing.T) {
	teardownTestCase, l := setupSettlementTestCase(t)
	defer teardownTestCase(t)

	a1 := ac1.Address()
	if am, err := l.GetAccountMeta(a1); err != nil {
		t.Fatal(err)
	} else {
		t.Log(util.ToIndentString(am))
	}

	if am, err := l.GetAccountMeta(ac2.Address()); err != nil {
		t.Fatal(err)
	} else {
		t.Log(util.ToIndentString(am))
	}

	ctx := vmstore.NewVMContext(l)

	tm, err := ctx.GetTokenMeta(a1, common.GasToken())
	if err != nil {
		t.Fatal(err)
	}

	param := createContractParam
	param.Previous = tm.Header
	if err = param.Sign(ac1); err != nil {
		t.Fatal(err)
	}
	balance, err := param.Balance()
	if err != nil {
		t.Fatal(err)
	}
	if tm.Balance.Compare(balance) == types.BalanceCompSmaller {
		t.Fatalf("not enough balance, [%s] of [%s]", balance.String(), tm.Balance.String())
	}

	if singedData, err := param.ToABI(); err == nil {
		sb := &types.StateBlock{
			Type:           types.ContractSend,
			Token:          tm.Type,
			Address:        param.PartyA,
			Balance:        tm.Balance.Sub(balance),
			Vote:           types.ZeroBalance,
			Network:        types.ZeroBalance,
			Oracle:         types.ZeroBalance,
			Storage:        types.ZeroBalance,
			Previous:       param.Previous,
			Link:           types.Hash(types.SettlementAddress),
			Representative: tm.Representative,
			Data:           singedData,
			Timestamp:      common.TimeNow().Unix(),
		}

		h := ctx.Cache.Trie().Hash()
		if h != nil {
			povHeader, err := l.GetLatestPovHeader()
			if err != nil {
				t.Fatalf("get pov header error: %s", err)
			}
			sb.PoVHeight = povHeader.GetHeight()
			sb.Extra = *h
		}

		if err := updateBlock(l, sb); err != nil {
			t.Fatal(err)
		}

		createContract := CreateContract{}
		if pendingKey, info, err := createContract.ProcessSend(ctx, sb); err != nil {
			t.Fatal(err)
		} else {
			t.Log(pendingKey, info)
			if err := ctx.SaveStorage(); err != nil {
				t.Fatal(err)
			}
		}

		rev := &types.StateBlock{
			Timestamp: common.TimeNow().Unix(),
		}
		if rb, err := createContract.DoReceive(ctx, rev, sb); err != nil {
			t.Fatal(err)
		} else {
			if len(rb) > 0 {
				t.Log(rb[0].Block)
			} else {
				t.Fatal("fail to generate create contract reward block")
			}
		}

		if contractParams, err := cabi.GetContractsIDByAddressAsPartyA(ctx, &a1); err != nil {
			t.Fatal(err)
		} else {
			a2 := ac2.Address()
			for _, cp := range contractParams {
				t.Log(cp.String())
				if cp.PartyB != a2 {
					t.Fatalf("invalid contract, partyB exp: %s,act: %s", a2.String(), cp.PartyB.String())
				}
				if address, err := cp.Address(); err != nil {
					t.Fatal(err)
				} else {
					sc := cabi.SignContractParam{
						ContractAddress: address,
						ConfirmDate:     time.Now().Unix(),
						SignatureB:      types.Signature{},
					}
					if h, err := types.HashBytes(sc.ContractAddress[:], util.BE_Int2Bytes(sc.ConfirmDate)); err != nil {
						t.Fatal(err)
					} else {
						sc.SignatureB = ac2.Sign(h)
						tm2, err := ctx.GetTokenMeta(a2, common.GasToken())
						if err != nil {
							t.Fatal(err)
						}
						if tm2 == nil {
							t.Fatalf("failed to find token from %s", a2.String())
						}

						signContract := &SignContract{}

						if singedData, err := sc.ToABI(); err == nil {
							sb2 := &types.StateBlock{
								Type:           types.ContractSend,
								Token:          tm.Type,
								Address:        a2,
								Balance:        tm.Balance,
								Vote:           types.ZeroBalance,
								Network:        types.ZeroBalance,
								Oracle:         types.ZeroBalance,
								Storage:        types.ZeroBalance,
								Previous:       tm.Header,
								Link:           types.Hash(types.SettlementAddress),
								Representative: tm.Representative,
								Data:           singedData,
								Timestamp:      common.TimeNow().Unix(),
							}
							if pk, info, err := signContract.ProcessSend(ctx, sb2); err != nil {
								t.Fatal(err)
							} else {
								t.Log(pk, " >>> ", info)
								if err := ctx.SaveStorage(); err != nil {
									t.Fatal(err)
								}

								if available := cabi.IsContractAvailable(ctx, &address); !available {
									t.Fatalf("failed to verify contract %s", address.String())
								} else {
									rev2 := &types.StateBlock{
										Timestamp: common.TimeNow().Unix(),
									}
									if rb, err := signContract.DoReceive(ctx, rev2, sb2); err != nil {
										t.Fatal(err)
									} else {
										if len(rb) > 0 {
											t.Log(rb[0].Block)
										} else {
											t.Fatal("fail to generate sign contract reward block")
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
}

func TestCreateContract_DoGapPov(t *testing.T) {
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "empty",
			args: args{
				ctx:   nil,
				block: nil,
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CreateContract{}
			got, err := c.DoGapPov(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoGapPov() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DoGapPov() got = %v, want %v", got, tt.want)
			}
		})
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
				withSignature: false,
				withPending:   true,
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

func TestProcessCDR_DoGapPov(t *testing.T) {
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx:   nil,
				block: nil,
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProcessCDR{}
			got, err := p.DoGapPov(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoGapPov() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DoGapPov() got = %v, want %v", got, tt.want)
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
				withSignature: false,
				withPending:   true,
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

func TestSignContract_DoGapPov(t *testing.T) {
	type args struct {
		ctx   *vmstore.VMContext
		block *types.StateBlock
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx:   nil,
				block: nil,
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignContract{}
			got, err := s.DoGapPov(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoGapPov() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DoGapPov() got = %v, want %v", got, tt.want)
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
				withSignature: false,
				withPending:   true,
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
