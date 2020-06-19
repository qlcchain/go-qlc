package apis

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
)

func setupBlackHoleAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, ledger.Store, *BlackHoleAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewBlackHoleAPI(l, cc)

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

func TestBlackHoleAPI_GetSendBlock(t *testing.T) {
	testcase, verifier, l, api := setupBlackHoleAPI(t)
	defer testcase(t)

	a1 := account1.Address()
	tm, err := l.GetTokenMeta(a1, config.GasToken())
	if err != nil {
		t.Fatal(err)
	}

	param := &cabi.DestroyParam{
		Owner:    a1,
		Previous: tm.Header,
		Token:    config.GasToken(),
		Amount:   big.NewInt(100),
	}
	if param.Sign, err = param.Signature(account1); err != nil {
		t.Fatal(err)
	}

	if blk, err := api.GetSendBlock(context.Background(), &pb.DestroyParam{
		Owner:    toAddressValue(param.Owner),
		Previous: toHashValue(param.Previous),
		Token:    toHashValue(param.Token),
		Amount:   param.Amount.Int64(),
		Sign:     toSignatureValue(param.Sign),
	}); err != nil {
		t.Fatal(err)
	} else {
		sb, err := toOriginStateBlock(blk)
		if err != nil {
			t.Fatal(err)
		}
		txHash := sb.GetHash()
		sb.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(sb); err != nil {
			t.Fatal(err)
		}

		if blk, err := api.GetRewardsBlock(context.Background(), &pbtypes.Hash{
			Hash: toHashValue(txHash),
		}); err != nil {
			t.Fatal(err)
		} else {
			rb, err := toOriginStateBlock(blk)
			if err != nil {
				t.Fatal(err)
			}
			txHash := rb.GetHash()
			rb.Signature = account1.Sign(txHash)
			if err := verifier.BlockProcess(rb); err != nil {
				t.Fatal(err)
			}
		}

		if details, err := api.GetDestroyInfoDetail(context.Background(), &pbtypes.Address{
			Address: toAddressValue(a1),
		}); err != nil {
			t.Fatal(err)
		} else if len(details.GetInfos()) != 1 {
			t.Fatalf("invalid details len, exp: 1, act: %d", len(details.GetInfos()))
		} else {
			t.Log(util.ToString(details))
		}

		info, err := api.GetTotalDestroyInfo(context.Background(), &pbtypes.Address{
			Address: toAddressValue(a1),
		})
		if err != nil {
			t.Fatal(err)
		}
		ra := toOriginBalanceByValue(info.GetBalance())

		if ra.Compare(types.ZeroBalance) == types.BalanceCompEqual {
			t.Fatal("invalid balance")
		} else {
			t.Log(info)
		}
	}
}
