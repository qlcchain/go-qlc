package apis

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

func setupRewardsTestCase(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *RewardsAPI) {
	//t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := chainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewRewardsAPI(l, cc)

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

func TestRewardsAPI_GetUnsignedRewardData(t *testing.T) {
	teardownTestCase, _, rewardApi := setupRewardsTestCase(t)
	defer teardownTestCase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	param := &api.RewardsParam{
		Id:     mock.Hash().String(),
		Amount: types.Balance{Int: big.NewInt(100)},
		Self:   a1,
		To:     a2,
	}
	r, err := rewardApi.GetUnsignedRewardData(context.Background(), toRewardsParam(param))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.GetHash())
}

func TestRewardsAPI_GetUnsignedConfidantData(t *testing.T) {
	teardownTestCase, _, rewardApi := setupRewardsTestCase(t)
	defer teardownTestCase(t)

	a1 := account1.Address()
	a2 := account2.Address()
	param := &api.RewardsParam{
		Id:     mock.Hash().String(),
		Amount: types.Balance{Int: big.NewInt(100)},
		Self:   a1,
		To:     a2,
	}

	r, err := rewardApi.GetUnsignedConfidantData(context.Background(), toRewardsParam(param))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r.GetHash())
}

func TestNewRewardsAPI(t *testing.T) {
	teardownTestCase, verifier, rewardApi := setupRewardsTestCase(t)
	defer teardownTestCase(t)

	a1 := account1.Address()
	a2 := account2.Address()

	r1 := &api.RewardsParam{
		Id:     mock.Hash().String(),
		Amount: types.Balance{Int: big.NewInt(100)},
		Self:   a1,
		To:     a2,
	}

	if h, err := rewardApi.GetUnsignedRewardData(context.Background(), toRewardsParam(r1)); err != nil {
		t.Fatal(err)
	} else {
		hr, err := toOriginHashByValue(h.GetHash())
		if err != nil {
			t.Fatal(err)
		}
		sign := account1.Sign(hr)
		if blk, err := rewardApi.GetSendRewardBlock(context.Background(), &pb.RewardsParamWithSign{
			Param: toRewardsParam(r1),
			Sign:  toSignatureValue(sign),
		}); err != nil {
			t.Fatal(err)
		} else {
			bs, err := toOriginStateBlock(blk)
			if err != nil {
				t.Fatal(err)
			}
			txHash := bs.GetHash()
			if err := verifier.BlockProcess(bs); err != nil {
				t.Fatal(err)
			}

			if rxBlk, err := rewardApi.GetReceiveRewardBlock(context.Background(), toHash(txHash)); err != nil {
				t.Fatal(err)
			} else {
				br, err := toOriginStateBlock(rxBlk)
				if err != nil {
					t.Fatal(err)
				}

				if err := verifier.BlockProcess(br); err != nil {
					t.Fatal(err)
				}

				if b, _ := rewardApi.IsAirdropRewards(context.Background(), toBytes(bs.Data)); !b.GetValue() {
					t.Fatal("IsAirdropRewards failed...")
				}
			}
		}
	}

	r2 := &api.RewardsParam{
		Id:     mock.Hash().String(),
		Amount: types.Balance{Int: big.NewInt(1000)},
		Self:   a1,
		To:     a2,
	}

	if h, err := rewardApi.GetUnsignedConfidantData(context.Background(), toRewardsParam(r2)); err != nil {
		t.Fatal(err)
	} else {
		ho, err := toOriginHash(h)
		if err != nil {
			t.Fatal(err)
		}
		sign := account1.Sign(ho)
		if blk, err := rewardApi.GetSendConfidantBlock(context.Background(), &pb.RewardsParamWithSign{
			Param: toRewardsParam(r2),
			Sign:  toSignatureValue(sign),
		}); err != nil {
			t.Fatal(err)
		} else {
			bs2, err := toOriginStateBlock(blk)
			if err != nil {
				t.Fatal(err)
			}

			txHash := bs2.GetHash()
			if err := verifier.BlockProcess(bs2); err != nil {
				t.Fatal(err)
			}

			if rxBlk, err := rewardApi.GetReceiveRewardBlock(context.Background(), toHash(txHash)); err != nil {
				t.Fatal(err)
			} else {
				br2, err := toOriginStateBlock(rxBlk)
				if err != nil {
					t.Fatal(err)
				}
				if err := verifier.BlockProcess(br2); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	if amount, err := rewardApi.GetTotalRewards(context.Background(), toString(r1.Id)); err != nil {
		t.Fatal(err)
	} else if toOriginBalanceByValue(amount.GetValue()).Cmp(r1.Amount.Int) != 0 {
		t.Fatalf("invalid amount, exp: %d, act: %d", r1.Amount, amount.GetValue())
	}

	if detail, err := rewardApi.GetRewardsDetail(context.Background(), toString(r1.Id)); err != nil {
		t.Fatal(err)
	} else if len(detail.GetInfos()) != 1 {
		t.Fatal("invalid rewards detail")
	}

	if rewards, err := rewardApi.GetConfidantRewards(context.Background(), toAddress(r2.To)); err != nil {
		t.Fatal(err)
	} else if len(rewards.GetRewards()) == 0 {
		t.Fatal("invalid confidant rewards...")
	} else {
		t.Log(rewards)
	}

	if detail, err := rewardApi.GetConfidantRewordsDetail(context.Background(), toAddress(r2.To)); err != nil {
		t.Fatal(err)
	} else if len(detail.GetInfos()) != 1 {
		t.Fatal("invalid confidant detail")
	}
}
