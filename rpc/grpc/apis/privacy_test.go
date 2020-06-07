package apis

import (
	"context"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/google/uuid"
	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	"os"
	"path/filepath"
	"testing"
)

type mockDataTestPrivacyApi struct {
	cfg *config.Config
	eb  event.EventBus
	l   *ledger.Ledger
	cc  *qctx.ChainContext
	sub *event.ActorSubscriber
	api *PrivacyAPI
}

func setupTestCasePrivacyApi(t *testing.T) (func(t *testing.T), *mockDataTestPrivacyApi) {
	t.Parallel()

	md := new(mockDataTestPrivacyApi)

	dir := filepath.Join(config.QlcTestDataDir(), "privacy", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	cfg, _ := cm.Config()
	cfg.Privacy.Enable = true
	cfg.Privacy.PtmNode = filepath.Join(dir, "__UnitTestCase__.ipc")
	_ = cm.CommitAndSave()

	md.l = ledger.NewLedger(cm.ConfigFile)

	md.cc = qctx.NewChainContext(cm.ConfigFile)
	md.cc.Init(nil)

	md.cfg, _ = md.cc.Config()

	md.eb = md.cc.EventBus()

	md.api = NewPrivacyAPI(md.cfg, md.l, md.eb, md.cc)

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

func TestPrivacyApi_Payload(t *testing.T) {
	tearDown, md := setupTestCasePrivacyApi(t)
	defer tearDown(t)

	param := &api.PrivacyDistributeParam{
		PrivateFrom: util.RandomFixedString(64),
		PrivateFor:  []string{util.RandomFixedString(64)},
		RawPayload:  util.RandomFixedBytes(128),
	}
	encKey, err := md.api.DistributeRawPayload(context.Background(), &pb.PrivacyDistributeParam{
		RawPayload:     param.RawPayload,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(encKey.GetValue()) == 0 {
		t.Fatal("encKey is nil")
	}

	pl, err := md.api.GetRawPayload(context.Background(), encKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(pl.GetValue()) == 0 {
		t.Fatal("pl is nil")
	}
}

func TestPrivacyAPI_GetDemoKV(t *testing.T) {
	tearDown, md := setupTestCasePrivacyApi(t)
	defer tearDown(t)

	_, err := md.api.GetDemoKV(context.Background(), &pb.Bytes{})
	if err == nil {
		t.Fatal(err)
	}

	_, err = md.api.GetBlockPrivatePayload(context.Background(), toHash(mock.Hash()))
	if err == nil {
		t.Fatal(err)
	}
}
