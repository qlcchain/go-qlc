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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupMintageAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, ledger.Store, *MintageAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewMintageAPI(cc.ConfigFile(), l)

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
	}, verifier, l, api
}

func TestNewMintageApi(t *testing.T) {
	testcase, verifier, l, mintageApi := setupMintageAPI(t)
	defer testcase(t)

	a := account1.Address()

	tm, err := l.GetTokenMeta(a, config.ChainToken())
	if err != nil {
		t.Fatal(err)
	}

	param := &api.MintageParams{
		SelfAddr:    a,
		PrevHash:    tm.Header,
		TokenName:   "Test",
		TokenSymbol: "QT",
		TotalSupply: "100",
		Decimals:    8,
		Beneficial:  a,
		NEP5TxId:    mock.Hash().String(),
	}

	if data, err := mintageApi.GetMintageData(context.Background(), toMintageParams(param)); err != nil {
		t.Fatal(err)
	} else if len(data.GetValue()) == 0 {
		t.Fatal("invalid mintage data")
	}
	contract.SetMinMintageTime(0, 0, 0, 0, 0, 1)

	if blk, err := mintageApi.GetMintageBlock(context.Background(), toMintageParams(param)); err != nil {
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

		if rxBlk, err := mintageApi.GetRewardBlock(context.Background(), blk); err != nil {
			t.Fatal(err)
		} else {
			br, err := toOriginStateBlock(rxBlk)
			if err != nil {
				t.Fatal(err)
			}
			txHash := br.GetHash()
			br.Signature = account1.Sign(txHash)
			if err := verifier.BlockProcess(br); err != nil {
				t.Fatal(err)
			}

			ti, err := mintageApi.ParseTokenInfo(context.Background(), toBytes(br.Data))
			if err != nil {
				t.Fatal(err)
			}

			if data, err := mintageApi.GetWithdrawMintageData(context.Background(), &pbtypes.Hash{
				Hash: ti.TokenId,
			}); err != nil {
				t.Fatal(err)
			} else if len(data.GetValue()) == 0 {
				t.Fatal("invalid GetWithdrawMintageData")
			}

			time.Sleep(2 * time.Second)

			if blk, err := mintageApi.GetWithdrawMintageBlock(context.Background(), &pb.WithdrawParams{
				SelfAddr: a.String(),
				TokenId:  ti.TokenId,
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

				if rxBlk, err := mintageApi.GetWithdrawRewardBlock(context.Background(), blk); err != nil {
					t.Fatal(err)
				} else {
					bp, err := toOriginStateBlock(rxBlk)
					if err != nil {
						t.Fatal(err)
					}
					txHash := bp.GetHash()
					bp.Signature = account1.Sign(txHash)
					if err := verifier.BlockProcess(bp); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
}
