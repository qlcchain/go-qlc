package p2p

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	p2pmetrics "github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pps "github.com/libp2p/go-libp2p-pubsub"
	localdiscovery "github.com/libp2p/go-libp2p/p2p/discovery"
	chaincontext "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
	ping "github.com/qlcchain/go-qlc/p2p/pinger"
	"github.com/qlcchain/go-qlc/p2p/pubsub"
	"go.uber.org/zap"

	"github.com/google/uuid"
	//"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/mock"
)

func TestQlcNode(t *testing.T) {
	removeDir := filepath.Join(config.QlcTestDataDir(), "node")
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "node", uuid.New().String(), config.QlcConfigFile)
	cc := chaincontext.NewChainContext(dir)
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

func TestNewNode_NewNode(t *testing.T) {
	//removeDir := filepath.Join(config.QlcTestDataDir(), "node")
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "node", uuid.New().String(), config.QlcConfigFile)
	cc := chaincontext.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19740"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.LogLevel = "error"
	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"127.0.0.1:18889/msg2"}

	cc2 := config.NewCfgManagerWithName("", "")
	cfg2, _ := cc2.Config()
	type args struct {
		config *config.Config
	}
	tests := []struct {
		name    string
		args    args
		want    *QlcNode
		wantErr bool
	}{
		// TODO: Add test cases.
		{"OK", args{cfg}, nil, false},
		{"badconfig", args{cfg2}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewNode(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func getMockQlcNode() (*QlcNode, error) {
	dir := filepath.Join(config.QlcTestDataDir(), "node", uuid.New().String(), config.QlcConfigFile)
	cc := chaincontext.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19740"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.LogLevel = "error"
	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"127.0.0.1:18889/msg2"}
	node, err := NewNode(cfg)
	return node, err
}

func getMockQlcServer() (*QlcService, error) {
	dir := filepath.Join(config.QlcTestDataDir(), "node", uuid.New().String(), config.QlcConfigFile)
	cc := chaincontext.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19750"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.LogLevel = "error"
	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"127.0.0.1:18890/msg2"}
	// node, err := NewNode(cfg)
	// if err != nil {
	// 	return nil, err
	// }
	// http.HandleFunc("/msg2/bootNode", func(w http.ResponseWriter, r *http.Request) {
	// 	bootNode := cfg.P2P.Listen + "/p2p/" + cfg.P2P.ID.PeerID
	// 	_, _ = fmt.Fprintf(w, bootNode)
	// })
	go func() {
		if err := http.ListenAndServe("127.0.0.1:18889", nil); err != nil {
			return
		}
	}()
	//start bootNode
	ns, err := NewQlcService(dir)
	err = ns.Start()
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func TestQlcNode_setRepresentativeNode(t *testing.T) {
	na, _ := getMockQlcNode()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		isRepresentative bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"OK", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, args{true}},
		{"OK", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, args{false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.setRepresentativeNode(tt.args.isRepresentative)
		})
	}
}
func TestQlcNode_updateWhiteList(t *testing.T) {
	na, _ := getMockQlcNode()
	na.cfg.WhiteList.Enable = true
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		id  string
		url string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"badid", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, args{"", ""}},
		{"OK", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, args{"QmapYS7nmZu7f7nUoFraF9J79q2uYZsESBkyraaAwaQSvQ", "/ip4/127.0.0.1/tcp/19740"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.updateWhiteList(tt.args.id, tt.args.url)
		})
	}
}

func TestQlcNode_buildHost(t *testing.T) {
	ns, _ := getMockQlcServer()
	na := ns.node
	na.cfg.WhiteList.Enable = true
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"okWhitelistEnable", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, false},
		{"okWhitelistDisable", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, false},
	}
	step := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			switch step {
			case 1:
				na.cfg.WhiteList.Enable = false
			}
			step = step + 1
			if err := node.buildHost(); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.buildHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_pubsubscribe(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		ctx     context.Context
		topic   *pubsub.Topic
		handler pubSubHandler
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    pubsub.Subscription
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			got, err := node.pubsubscribe(tt.args.ctx, tt.args.topic, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.pubsubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QlcNode.pubsubscribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcNode_startLocalDiscovery(t *testing.T) {
	na, _ := getMockQlcNode()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"ok", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if err := node.startLocalDiscovery(); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.startLocalDiscovery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_StartServices(t *testing.T) {
	ns, _ := getMockQlcServer()
	na := ns.node
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"ok", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if err := node.StartServices(); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.StartServices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_startPingService(t *testing.T) {
	//ns,_ := getMockQlcServer()
	//na := ns.node
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		//{"ok", fields{na.ID,na.privateKey,na.cfg,na.ctx,na.cancel,na.localDiscovery,na.host,na.peerStore,na.boostrapAddrs,na.streamManager,na.dis,na.kadDht,na.netService,na.logger,na.MessageTopic,na.pubSub,na.MessageSub,na.isMiner,na.isRepresentative,na.reporter,na.ping,na.connectionGater}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.startPingService()
		})
	}
}

func TestQlcNode_publishPeersInfoToChainContext(t *testing.T) {
	//ns,_ := getMockQlcServer()
	//na := ns.node
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		//{"ok", fields{na.ID,na.privateKey,na.cfg,na.ctx,na.cancel,na.localDiscovery,na.host,na.peerStore,na.boostrapAddrs,na.streamManager,na.dis,na.kadDht,na.netService,na.logger,na.MessageTopic,na.pubSub,na.MessageSub,na.isMiner,na.isRepresentative,na.reporter,na.ping,na.connectionGater}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.publishPeersInfoToChainContext()
		})
	}
}

func TestQlcNode_connectBootstrap(t *testing.T) {
	na, _ := getMockQlcNode()
	pInfoS, _ := convertPeers(na.boostrapAddrs)
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		pInfoS []peer.AddrInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"ok", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, args{pInfoS}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.connectBootstrap(tt.args.pInfoS)
		})
	}
}

func TestQlcNode_startPeerDiscovery(t *testing.T) {
	//ns,_ := getMockQlcServer()
	//na := ns.node
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		//{"ok", fields{na.ID,na.privateKey,na.cfg,na.ctx,na.cancel,na.localDiscovery,na.host,na.peerStore,na.boostrapAddrs,na.streamManager,na.dis,na.kadDht,na.netService,na.logger,na.MessageTopic,na.pubSub,na.MessageSub,na.isMiner,na.isRepresentative,na.reporter,na.ping,na.connectionGater}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.startPeerDiscovery()
		})
	}
}

func TestQlcNode_findPeers(t *testing.T) {
	//na,_ := getMockQlcNode()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		//{"ok", fields{na.ID,na.privateKey,na.cfg,na.ctx,na.cancel,na.localDiscovery,na.host,na.peerStore,na.boostrapAddrs,na.streamManager,na.dis,na.kadDht,na.netService,na.logger,na.MessageTopic,na.pubSub,na.MessageSub,na.isMiner,na.isRepresentative,na.reporter,na.ping,na.connectionGater},false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if err := node.findPeers(); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.findPeers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_updatePeersInfo(t *testing.T) {
	//na,_ := getMockQlcNode()
	//peers, _ := na.dhtFoundPeers()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		peers []peer.AddrInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		//{"ok", fields{na.ID,na.privateKey,na.cfg,na.ctx,na.cancel,na.localDiscovery,na.host,na.peerStore,na.boostrapAddrs,na.streamManager,na.dis,na.kadDht,na.netService,na.logger,na.MessageTopic,na.pubSub,na.MessageSub,na.isMiner,na.isRepresentative,na.reporter,na.ping,na.connectionGater},args{peers}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.updatePeersInfo(tt.args.peers)
		})
	}
}

func TestQlcNode_handleStream(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		s network.Stream
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.handleStream(tt.args.s)
		})
	}
}

func TestQlcNode_GetID(t *testing.T) {
	na, _ := getMockQlcNode()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{"ok", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, na.ID.Pretty()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if got := node.GetID(); got != tt.want {
				t.Errorf("QlcNode.GetID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcNode_StreamManager(t *testing.T) {
	na, _ := getMockQlcNode()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name   string
		fields fields
		want   *StreamManager
	}{
		// TODO: Add test cases.
		{"ok", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, na.streamManager},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if got := node.StreamManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QlcNode.StreamManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQlcNode_stopHost(t *testing.T) {
	na, _ := getMockQlcNode()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"nohost", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if err := node.stopHost(); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.stopHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_SetQlcService(t *testing.T) {
	ns, _ := getMockQlcServer()
	na := ns.node
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		ns *QlcService
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"ok", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, args{ns}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.SetQlcService(tt.args.ns)
		})
	}
}

func TestQlcNode_Stop(t *testing.T) {
	na, _ := getMockQlcNode()
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"nohost", fields{na.ID, na.privateKey, na.cfg, na.ctx, na.cancel, na.localDiscovery, na.host, na.peerStore, na.boostrapAddrs, na.streamManager, na.dis, na.kadDht, na.netService, na.logger, na.MessageTopic, na.pubSub, na.MessageSub, na.isMiner, na.isRepresentative, na.reporter, na.ping, na.connectionGater}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if err := node.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_BroadcastMessage(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		messageName MessageType
		value       interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.BroadcastMessage(tt.args.messageName, tt.args.value)
		})
	}
}

func TestQlcNode_SendMessageToPeer(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		messageName MessageType
		value       interface{}
		peerID      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if err := node.SendMessageToPeer(tt.args.messageName, tt.args.value, tt.args.peerID); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.SendMessageToPeer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_handleSubscription(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		ctx     context.Context
		sub     pubsub.Subscription
		handler pubSubHandler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.handleSubscription(tt.args.ctx, tt.args.sub, tt.args.handler)
		})
	}
}

func TestQlcNode_processMessage(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		ctx       context.Context
		pubSubMsg pubsub.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			if err := node.processMessage(tt.args.ctx, tt.args.pubSubMsg); (err != nil) != tt.wantErr {
				t.Errorf("QlcNode.processMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQlcNode_GetBandwidthStats(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		stats *p2pmetrics.Stats
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.GetBandwidthStats(tt.args.stats)
		})
	}
}

func TestQlcNode_getBootNode(t *testing.T) {
	type fields struct {
		ID               peer.ID
		privateKey       crypto.PrivKey
		cfg              *config.Config
		ctx              context.Context
		cancel           context.CancelFunc
		localDiscovery   localdiscovery.Service
		host             host.Host
		peerStore        peerstore.Peerstore
		boostrapAddrs    []string
		streamManager    *StreamManager
		dis              *discovery.RoutingDiscovery
		kadDht           *dht.IpfsDHT
		netService       *QlcService
		logger           *zap.SugaredLogger
		MessageTopic     *pubsub.Topic
		pubSub           *libp2pps.PubSub
		MessageSub       pubsub.Subscription
		isMiner          bool
		isRepresentative bool
		reporter         p2pmetrics.Reporter
		ping             *ping.Pinger
		connectionGater  *ConnectionGater
	}
	type args struct {
		urls []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &QlcNode{
				ID:               tt.fields.ID,
				privateKey:       tt.fields.privateKey,
				cfg:              tt.fields.cfg,
				ctx:              tt.fields.ctx,
				cancel:           tt.fields.cancel,
				localDiscovery:   tt.fields.localDiscovery,
				host:             tt.fields.host,
				peerStore:        tt.fields.peerStore,
				boostrapAddrs:    tt.fields.boostrapAddrs,
				streamManager:    tt.fields.streamManager,
				dis:              tt.fields.dis,
				kadDht:           tt.fields.kadDht,
				netService:       tt.fields.netService,
				logger:           tt.fields.logger,
				MessageTopic:     tt.fields.MessageTopic,
				pubSub:           tt.fields.pubSub,
				MessageSub:       tt.fields.MessageSub,
				isMiner:          tt.fields.isMiner,
				isRepresentative: tt.fields.isRepresentative,
				reporter:         tt.fields.reporter,
				ping:             tt.fields.ping,
				connectionGater:  tt.fields.connectionGater,
			}
			node.getBootNode(tt.args.urls)
		})
	}
}

func Test_accessHttpServer(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := accessHttpServer(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("accessHttpServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("accessHttpServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
