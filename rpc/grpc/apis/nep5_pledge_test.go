package apis

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	"github.com/qlcchain/go-qlc/vm/contract"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"time"
)

func setupNEP5PledgeAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *NEP5PledgeAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "pledgeApi", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	pledgeApi := NewNEP5PledgeAPI(cc.ConfigFile(), l)

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
	}, verifier, pledgeApi
}

func TestNewNEP5PledgeAPI(t *testing.T) {
	testcase, verifier, pledgeApi := setupNEP5PledgeAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	a2 := account2.Address()
	contract.SetPledgeTime(0, 0, 0, 0, 0, 1)
	param := &api.PledgeParam{
		Beneficial:    a2,
		PledgeAddress: a1,
		Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
		PType:         "vote",
		NEP5TxId:      mock.Hash().String(),
	}

	if data, err := pledgeApi.GetPledgeData(context.Background(), toPledgeParam(param)); err != nil {
		t.Fatal(err)
	} else if len(data.GetValue()) == 0 {
		t.Fatal("invalid GetPledgeData")
	}

	if blk, err := pledgeApi.GetPledgeBlock(context.Background(), toPledgeParam(param)); err != nil {
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

		if _, err := pledgeApi.GetPledgeRewardBlock(context.Background(), blk); err != nil {
			t.Fatal(err)
		}

		if txBlk, err := pledgeApi.GetPledgeRewardBlockBySendHash(context.Background(), &pbtypes.Hash{
			Hash: txHash.String(),
		}); err != nil {
			t.Fatal(err)
		} else {
			br, err := toOriginStateBlock(txBlk)
			if err != nil {
				t.Fatal(err)
			}
			txHash := br.GetHash()
			br.Signature = account2.Sign(txHash)
			if err := verifier.BlockProcess(br); err != nil {
				t.Fatal(err)
			}

			if r, err := pledgeApi.ParsePledgeInfo(context.Background(), &pb.Bytes{
				Value: txBlk.Data,
			}); err != nil {
				t.Fatal(err)
			} else {
				t.Log(r)
			}
		}

		time.Sleep(2 * time.Second)

		w := &api.WithdrawPledgeParam{
			Beneficial: a2,
			Amount:     param.Amount,
			PType:      param.PType,
			NEP5TxId:   param.NEP5TxId,
		}

		if info, err := pledgeApi.GetPledgeInfo(context.Background(), toWithdrawPledgeParam(w)); err != nil {
			t.Fatal(err)
		} else {
			t.Log(info)
		}

		if info, err := pledgeApi.GetPledgeInfoWithNEP5TxId(context.Background(), toWithdrawPledgeParam(w)); err != nil {
			t.Fatal(err)
		} else {
			t.Log(info)
		}
		if info, err := pledgeApi.GetPledgeInfoWithTimeExpired(context.Background(), toWithdrawPledgeParam(w)); err != nil {
			t.Fatal(err)
		} else {
			t.Log(info)
		}

		if data, err := pledgeApi.GetWithdrawPledgeData(context.Background(), toWithdrawPledgeParam(w)); err != nil {
			t.Fatal(err)
		} else if len(data.GetValue()) == 0 {
			t.Fatal("invalid GetWithdrawPledgeData")
		}

		if blk, err := pledgeApi.GetWithdrawPledgeBlock(context.Background(), toWithdrawPledgeParam(w)); err != nil {
			t.Fatal(err)
		} else {
			bs, err := toOriginStateBlock(blk)
			if err != nil {
				t.Fatal(err)
			}
			txHash := bs.GetHash()
			bs.Signature = account2.Sign(txHash)
			if err := verifier.BlockProcess(bs); err != nil {
				t.Fatal(err)
			}

			if _, err := pledgeApi.GetWithdrawRewardBlock(context.Background(), blk); err != nil {
				t.Fatal(err)
			}

			if rxBlk, err := pledgeApi.GetWithdrawRewardBlockBySendHash(context.Background(), &pbtypes.Hash{
				Hash: toHashValue(txHash),
			}); err != nil {
				t.Fatal(err)
			} else {
				br, err := toOriginStateBlock(rxBlk)
				if err != nil {
					t.Fatal(err)
				}

				txHash := br.GetHash()
				br.Signature = account1.Sign(txHash)
				//if err := verifier.BlockProcess(rxBlk); err != nil {
				//	t.Fatal(err)
				//}
			}
		}
	}

	if info, err := pledgeApi.GetPledgeInfosByPledgeAddress(context.Background(), &pbtypes.Address{
		Address: a1.String(),
	}); err != nil {
		t.Fatal()
	} else {
		t.Log(info)
	}

	if amount, err := pledgeApi.GetPledgeBeneficialTotalAmount(context.Background(), &pbtypes.Address{
		Address: toAddressValue(a2),
	}); err != nil {
		t.Fatal(err)
	} else if toOriginBalanceByValue(amount.GetValue()).Cmp(big.NewInt(0)) == 0 {
		t.Fatal("invalid amount")
	} else {
		t.Log("amount: ", amount)
	}

	if info, err := pledgeApi.GetBeneficialPledgeInfosByAddress(context.Background(), &pbtypes.Address{
		Address: a2.String(),
	}); err != nil {
		t.Fatal()
	} else {
		t.Log(info.GetTotalAmounts())
	}

	if infos, err := pledgeApi.GetBeneficialPledgeInfos(context.Background(), &pb.BeneficialPledgeRequest{
		Beneficial: toAddressValue(a2),
		PType:      "vote",
	}); err != nil || infos == nil {
		t.Fatal(err)
	} else if len(infos.GetPledgeInfos()) == 0 {
		t.Fatal("invalid pledge info")
	}

	if amount, err := pledgeApi.GetPledgeBeneficialAmount(context.Background(), &pb.BeneficialPledgeRequest{
		Beneficial: toAddressValue(a2),
		PType:      "vote",
	}); err != nil {
		t.Fatal(err)
	} else if toOriginBalanceByValue(amount.GetValue()).Cmp(big.NewInt(0)) == 0 {
		t.Fatal("invalid amount")
	} else {
		t.Log("amount: ", amount)
	}

	if amount, err := pledgeApi.GetTotalPledgeAmount(context.Background(), nil); err != nil {
		t.Fatal(err)
	} else if toOriginBalanceByValue(amount.GetValue()).Cmp(big.NewInt(0)) == 0 {
		t.Fatal("invalid amount")
	} else {
		t.Log("amount: ", amount)
	}

	if info, err := pledgeApi.GetAllPledgeInfo(context.Background(), nil); err != nil {
		t.Fatal(err)
	} else if len(info.GetPledgeInfos()) == 0 {
		t.Fatal("GetAllPledgeInfo")
	}
}

func TestNEP5PledgeAPI_GetPledgeBlock(t *testing.T) {
	testcase, _, pledgeApi := setupNEP5PledgeAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	param := &api.PledgeParam{
		Beneficial:    a2,
		PledgeAddress: a1,
		Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
		PType:         "vote",
		NEP5TxId:      mock.Hash().String(),
	}

	r, err := pledgeApi.GetPledgeBlock(context.Background(), toPledgeParam(param))
	if err != nil {
		t.Fatal(err)
	}
	br, err := toOriginStateBlock(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(br.GetHash())
}

func TestNEP5PledgeAPI_GetPledgeData(t *testing.T) {
	testcase, _, pledgeApi := setupNEP5PledgeAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	param := &api.PledgeParam{
		Beneficial:    a2,
		PledgeAddress: a1,
		Amount:        types.Balance{Int: big.NewInt(100 * 1e8)},
		PType:         "vote",
		NEP5TxId:      mock.Hash().String(),
	}

	_, err := pledgeApi.GetPledgeData(context.Background(), toPledgeParam(param))
	if err != nil {
		t.Fatal(err)
	}
}
