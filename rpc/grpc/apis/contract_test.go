package apis

import (
	"context"
	"encoding/hex"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/google/uuid"
	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	"os"
	"path/filepath"
	"testing"
)

type mockDataContractApi struct {
	l   ledger.Store
	cc  *qctx.ChainContext
	eb  event.EventBus
	sub *event.ActorSubscriber
}

func setupTestCaseContractApi(t *testing.T) (func(t *testing.T), *mockDataContractApi) {
	md := new(mockDataContractApi)

	dir := filepath.Join(config.QlcTestDataDir(), "rewards", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	cfg, _ := cm.Config()
	cfg.Privacy.Enable = true
	cfg.Privacy.PtmNode = filepath.Join(dir, "__UnitTestCase__.ipc")
	_ = cm.CommitAndSave()

	l := ledger.NewLedger(cm.ConfigFile)
	setLedgerStatus(l, t)
	povBlk, povTd := mock.GenerateGenesisPovBlock()
	l.AddPovBlock(povBlk, povTd)
	l.AddPovBestHash(povBlk.GetHeight(), povBlk.GetHash())
	l.SetPovLatestHeight(povBlk.GetHeight())
	md.l = l

	md.cc = qctx.NewChainContext(cm.ConfigFile)
	md.cc.Init(func() error {
		return nil
	})

	md.eb = md.cc.EventBus()

	md.cc.Start()

	md.sub = event.NewActorSubscriber(event.Spawn(func(ctx actor.Context) {
		switch msgReq := ctx.Message().(type) {
		case *topic.EventPrivacySendReqMsg:
			msgReq.RspChan <- &topic.EventPrivacySendRspMsg{
				EnclaveKey: util.RandomFixedBytes(32),
			}
		case *topic.EventPrivacyRecvReqMsg:
			msgReq.RspChan <- &topic.EventPrivacyRecvRspMsg{
				RawPayload: util.RandomFixedBytes(128),
			}
		}
	}), md.eb)

	err := md.sub.Subscribe(topic.EventPrivacySendReq, topic.EventPrivacyRecvReq)
	if err != nil {
		t.Fatal(err)
	}

	return func(t *testing.T) {
		_ = md.sub.UnsubscribeAll()
		_ = md.cc.Stop()
		_ = md.eb.Close()
		err := md.l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestNewContractApi(t *testing.T) {
	tearDown, md := setupTestCaseContractApi(t)
	defer tearDown(t)

	api := NewContractAPI(md.cc, md.l)
	if abi, err := api.GetAbiByContractAddress(context.Background(), &pbtypes.Address{
		Address: contractaddress.BlackHoleAddress.String(),
	}); err != nil {
		t.Fatal(err)
	} else if len(abi.GetValue()) == 0 {
		t.Fatal("invalid abi")
	} else {
		if data, err := api.PackContractData(context.Background(), &pb.PackContractDataRequest{
			AbiStr:     abi.GetValue(),
			MethodName: "Destroy",
			Params: []string{mock.Address().String(),
				mock.Hash().String(), mock.Hash().String(), "111", types.ZeroSignature.String()},
		}); err != nil {
			t.Fatal(err)
		} else if len(data.GetValue()) == 0 {
			t.Fatal("invalid data")
		} else {
			t.Log(hex.EncodeToString(data.GetValue()))
		}
	}

	if addressList, _ := api.ContractAddressList(context.Background(), nil); len(addressList.GetAddresses()) == 0 {
		t.Fatal("can not find any on-chain contract")
	}
}
