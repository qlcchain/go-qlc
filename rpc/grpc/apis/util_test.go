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
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

func setupDefaultUtilApi(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *UtilApi) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewUtilApi(l)

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

func TestNewUtilApi(t *testing.T) {
	testcase, _, api := setupDefaultUtilApi(t)
	defer testcase(t)

	raw := mock.Hash().String()
	pw := "qlcchain"

	if encrypt, err := api.Encrypt(context.Background(), &pb.EncryptRequest{
		Raw:        raw,
		Passphrase: pw,
	}); err != nil {
		t.Fatal(err)
	} else {
		if decrypt, err := api.Decrypt(context.Background(), &pb.DecryptRequest{
			Cryptograph: encrypt.GetValue(),
			Passphrase:  pw,
		}); err != nil {
			t.Fatal(err)
		} else if decrypt.GetValue() != raw {
			t.Fatalf("invalid raw, exp: %s, act: %s", raw, pw)
		}
	}
	balance := types.Balance{Int: big.NewInt(100)}
	n := "QGAS"
	if raw, err := api.BalanceToRaw(context.Background(), &pb.RawBalance{
		Balance:   balance.Int64(),
		Unit:      "QGAS",
		TokenName: n,
	}); err != nil {
		t.Fatal(err)
	} else {
		t.Log(raw.GetValue())
	}

	if raw, err := api.RawToBalance(context.Background(), &pb.RawBalance{
		Balance:   balance.Int64(),
		Unit:      "QGAS",
		TokenName: n,
	}); err != nil {
		t.Fatal(err)
	} else {
		t.Log(raw.GetValue())
	}

	//if raw, err := api.BalanceToRaw(types.Balance{Int: big.NewInt(100)}, "QLC", nil); err != nil {
	//	t.Fatal(err)
	//} else {
	//	t.Log(raw)
	//}
	//
	//if raw, err := api.RawToBalance(types.Balance{Int: big.NewInt(100)}, "QLC", nil); err != nil {
	//	t.Fatal(err)
	//} else {
	//	t.Log(raw)
	//}
}
