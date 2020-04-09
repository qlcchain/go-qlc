package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/google/uuid"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

type mockDataTestPrivacyApi struct {
	cfg *config.Config
	eb  event.EventBus
	l   *ledger.Ledger
	cc  *qctx.ChainContext
	sub *event.ActorSubscriber
	api *PrivacyApi
}

func setupTestCasePrivacyApi(t *testing.T) (func(t *testing.T), *mockDataTestPrivacyApi) {
	t.Parallel()

	md := new(mockDataTestPrivacyApi)

	dir := filepath.Join(config.QlcTestDataDir(), "privacy", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()

	md.l = ledger.NewLedger(cm.ConfigFile)

	md.cc = qctx.NewChainContext(cm.ConfigFile)
	md.cc.Init(nil)

	md.cfg, _ = md.cc.Config()

	md.eb = md.cc.EventBus()

	md.api = NewPrivacyApi(md.cfg, md.l, md.eb, md.cc)

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

	param := &PrivacyDistributeParam{
		PrivateFrom: util.RandomFixedString(64),
		PrivateFor:  []string{util.RandomFixedString(64)},
		RawPayload:  util.RandomFixedBytes(128),
	}
	encKey, err := md.api.DistributeRawPayload(param)
	if err != nil {
		t.Fatal(err)
	}
	if len(encKey) == 0 {
		t.Fatal("encKey is nil")
	}

	pl, err := md.api.GetRawPayload(encKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(pl) == 0 {
		t.Fatal("pl is nil")
	}
}
