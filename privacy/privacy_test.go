package privacy

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

type mockDataPrivacy struct {
	cc      *qctx.ChainContext
	cfg     *config.Config
	eb      event.EventBus
	l       ledger.Store
	privacy *Controller
}

func setupPrivacyTestCase(t *testing.T, scheme string) (func(t *testing.T), *mockDataPrivacy) {
	t.Parallel()
	md := &mockDataPrivacy{}

	uid := uuid.New().String()
	rootDir := filepath.Join(config.QlcTestDataDir(), uid)
	_ = os.RemoveAll(rootDir)

	cm := config.NewCfgManager(rootDir)
	_, _ = cm.Load()

	md.cc = qctx.NewChainContext(cm.ConfigFile)
	_ = md.cc.Init(nil)

	md.cfg, _ = md.cc.Config()
	md.cfg.Privacy.Enable = true
	if scheme == "http" {
		md.cfg.Privacy.PtmNode = "http://127.0.0.1:9999/__UnitTestCase__"
	} else {
		md.cfg.Privacy.PtmNode = "unix:" + filepath.Join(rootDir, "__UnitTestCase__.ipc")
	}

	md.l = ledger.NewLedger(cm.ConfigFile)

	md.eb = md.cc.EventBus()

	md.privacy = NewController(md.cc)
	err := md.privacy.Init()
	if err != nil {
		t.Fatal(err)
	}
	md.privacy.ptm.SetFakeMode(true)

	return func(t *testing.T) {
		err := md.l.Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll(rootDir)
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestController_Send(t *testing.T) {
	tearDown, md := setupPrivacyTestCase(t, "unix")
	defer tearDown(t)

	err := md.privacy.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	msgReq := &topic.EventPrivacySendReqMsg{}
	msgReq.RawPayload = util.RandomFixedBytes(64)
	msgReq.PrivateFrom = util.RandomFixedString(32)
	msgReq.PrivateFor = append(msgReq.PrivateFor, util.RandomFixedString(32))
	msgReq.RspChan = make(chan *topic.EventPrivacySendRspMsg, 1)

	md.eb.Publish(topic.EventPrivacySendReq, msgReq)
	time.Sleep(10 * time.Millisecond)

	select {
	case msgRsp := <-msgReq.RspChan:
		t.Log("got send response", msgRsp)
	case <-time.After(time.Second):
		t.Fatal("timeout to get send response")
	}

	err = md.privacy.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestController_Receive(t *testing.T) {
	tearDown, md := setupPrivacyTestCase(t, "http")
	defer tearDown(t)

	err := md.privacy.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	msgReq := &topic.EventPrivacyRecvReqMsg{}
	msgReq.EnclaveKey = util.RandomFixedBytes(64)
	msgReq.RspChan = make(chan *topic.EventPrivacyRecvRspMsg, 1)

	md.eb.Publish(topic.EventPrivacyRecvReq, msgReq)
	time.Sleep(10 * time.Millisecond)

	select {
	case msgRsp := <-msgReq.RspChan:
		t.Log("got recv response", msgRsp)
	case <-time.After(time.Second):
		t.Fatal("timeout to get recv response")
	}

	err = md.privacy.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestController_Debug(t *testing.T) {
	tearDown, md := setupPrivacyTestCase(t, "http")
	defer tearDown(t)

	err := md.privacy.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	feb := md.cc.FeedEventBus()
	msgReq := &topic.EventRPCSyncCallMsg{Name: "Debug.PrivacyInfo"}
	msgReq.In = make(map[string]interface{})
	msgReq.Out = make(map[string]interface{})
	msgRsp := feb.RpcSyncCall(msgReq)
	time.Sleep(10 * time.Millisecond)

	if msgRsp == nil {
		t.Fatal("msgRsp is nil")
	}

	err = md.privacy.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
