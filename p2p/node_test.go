package p2p

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func TestQlcNode(t *testing.T) {
	removeDir := filepath.Join(config.QlcTestDataDir(), "node")
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "node", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19740"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.LogLevel = "error"

	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"127.0.0.1:18889/msg2"}
	http.HandleFunc("/msg2/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg.P2P.Listen + "/p2p/" + cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe("127.0.0.1:18889", nil); err != nil {
			t.Fatal(err)
		}
	}()
	//start bootNode
	q, err := NewQlcService(dir)
	err = q.Start()
	if err != nil {
		t.Fatal(err)
	}
	//remove test file
	defer func() {
		err = q.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = q.msgService.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.RemoveAll(removeDir)
		if err != nil {
			t.Fatal(err)
		}
	}()

	q.node.setRepresentativeNode(true)
	if !q.node.isRepresentative {
		t.Fatal("set representative error")
	}
	id := q.node.GetID()
	if id != q.node.ID.Pretty() {
		t.Fatal("get node id error")
	}
	if q.node.StreamManager() == nil {
		t.Fatal("get stream manager error")
	}
	blk1 := mock.StateBlockWithoutWork()
	q.node.netService.msgEvent.Publish(topic.EventBroadcast, &EventBroadcastMsg{Type: PublishReq, Message: blk1})
	q.node.netService.msgEvent.Publish(topic.EventSendMsgToSingle, &EventSendMsgToSingleMsg{Type: PublishReq, Message: blk1})
	q.node.netService.msgEvent.Publish(topic.EventRepresentativeNode, true)
	q.node.netService.msgEvent.Publish(topic.EventConsensusSyncFinished, &topic.EventP2PSyncStateMsg{P2pSyncState: topic.SyncFinish})
	q.node.netService.msgEvent.Publish(topic.EventFrontiersReq, &EventFrontiersReqMsg{PeerID: q.node.ID.Pretty()})
}
