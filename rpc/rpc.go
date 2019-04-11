package rpc

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/wallet"
	"go.uber.org/zap"
)

type RPC struct {
	//p2pServer *p2p.Server

	rpcAPIs          []API
	inProcessHandler *Server

	ipcListener net.Listener
	ipcHandler  *Server

	httpWhitelist []string
	httpListener  net.Listener
	httpHandler   *Server

	wsListener net.Listener
	wsHandler  *Server

	wsCli              *WebSocketCli
	config             *config.Config
	DashboardTargetURL string
	NetID              uint `json:"NetID"`

	lock sync.RWMutex

	ledger   *ledger.Ledger
	wallet   *wallet.WalletStore
	relation *relation.Relation
	eb       event.EventBus
	logger   *zap.SugaredLogger
}

func NewRPC(cfg *config.Config, eb event.EventBus) (*RPC, error) {
	rl, err := relation.NewRelation(cfg, eb)
	if err != nil {
		return nil, err
	}
	r := RPC{
		ledger:   ledger.NewLedger(cfg.LedgerDir(), eb),
		wallet:   wallet.NewWalletStore(cfg),
		relation: rl,
		eb:       eb,
		config:   cfg,
		logger:   log.NewLogger("rpc"),
	}
	return &r, nil
}

// startIPC initializes and starts the IPC RPC endpoint.
func (r *RPC) startIPC(apis []API) error {
	if r.config.RPC.IPCEndpoint == "" {
		return nil // IPC disabled.
	}
	listener, handler, err := StartIPCEndpoint(r.config.RPC.IPCEndpoint, apis)
	if err != nil {
		return err
	}
	r.ipcListener = listener
	r.ipcHandler = handler
	r.logger.Info("IPC endpoint opened, ", "url:", r.config.RPC.IPCEndpoint)
	return nil
}

// stopIPC terminates the IPC RPC endpoint.
func (r *RPC) stopIPC() {
	if r.ipcListener != nil {
		r.ipcListener.Close()
		r.ipcListener = nil

		r.logger.Debug("IPC endpoint closed, ", "endpoint:", r.config.RPC.IPCEndpoint)
	}
	if r.ipcHandler != nil {
		r.ipcHandler.Stop()
		r.ipcHandler = nil
	}
}

// startHTTP initializes and starts the HTTP RPC endpoint.
func (r *RPC) startHTTP(endpoint string, apis []API, modules []string, cors []string, vhosts []string, timeouts HTTPTimeouts, exposeAll bool) error {
	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts, timeouts, exposeAll)
	if err != nil {
		return err
	}
	r.logger.Info("HTTP endpoint opened,", " url:", listener.Addr(), ", cors:", strings.Join(cors, ","), ", vhosts:", strings.Join(vhosts, ","))
	// All listeners booted successfully
	//r.httpEndpoint = endpoint
	r.httpListener = listener
	r.httpHandler = handler

	return nil
}

// stopHTTP terminates the HTTP RPC endpoint.
func (r *RPC) stopHTTP() {
	if r.httpListener != nil {
		r.httpListener.Close()
		r.httpListener = nil

		r.logger.Debug("HTTP endpoint closed, ", "endpoint:", r.config.RPC.HTTPEndpoint)
	}
	if r.httpHandler != nil {
		r.httpHandler.Stop()
		r.httpHandler = nil
	}
}

// startWS initializes and starts the websocket RPC endpoint.
func (r *RPC) startWS(endpoint string, apis []API, modules []string, wsOrigins []string, exposeAll bool) error {
	// Short circuit if the WS endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := StartWSEndpoint(endpoint, apis, modules, wsOrigins, exposeAll)
	if err != nil {
		return err
	}
	r.logger.Info("WebSocket endpoint opened, ", "url:", listener.Addr())
	// All listeners booted successfully
	//r.wsEndpoint = endpoint
	r.wsListener = listener
	r.wsHandler = handler

	return nil
}

// stopWS terminates the websocket RPC endpoint.
func (r *RPC) stopWS() {
	if r.wsListener != nil {
		r.wsListener.Close()
		r.wsListener = nil
		r.logger.Debug("WebSocket endpoint closed, ", "endpoint:", r.config.RPC.WSEndpoint)
	}
	if r.wsHandler != nil {
		r.wsHandler.Stop()
		r.wsHandler = nil
	}
	if r.wsCli != nil {
		r.wsCli.Close()
	}
}

func (r *RPC) Attach() (*Client, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	//if r.p2pServer == nil {
	//	return nil, ErrNodeStopped
	//}

	if r.inProcessHandler == nil {
		return nil, errors.New("server not started")
	}
	return DialInProc(r.inProcessHandler), nil
}

// startInProc initializes an in-process RPC endpoint.
func (r *RPC) startInProcess(apis []API) error {
	// Register all the APIs exposed by the services
	handler := NewServer()
	for _, api := range apis {
		if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
			r.logger.Info(err)
			return err
		}
		//r.logger.Debug("InProc registered ", "service ", api.Service, " namespace ", api.Namespace)
	}
	r.inProcessHandler = handler
	return nil
}

// stopInProc terminates the in-process RPC endpoint.
func (r *RPC) stopInProcess() {
	if r.inProcessHandler != nil {
		r.inProcessHandler.Stop()
		r.inProcessHandler = nil
	}
}

func (r *RPC) StopRPC() {
	r.stopInProcess()
	if r.config.RPC.Enable && r.config.RPC.IPCEnabled {
		r.stopIPC()
	}
	if r.config.RPC.Enable && r.config.RPC.HTTPEnabled {
		r.stopHTTP()
	}
	if r.config.RPC.Enable && r.config.RPC.WSEnabled {
		r.stopWS()
	}

}

func (r *RPC) StartRPC() error {

	// Init rpc log
	//rpcapi.Init(node.config.DataDir, node.config.LogLevel, node.config.TestTokenHexPrivKey, node.config.TestTokenTti)

	// Start the various API endpoints, terminating all in case of errors
	if err := r.startInProcess(r.GetInProcessApis()); err != nil {
		return err
	}

	//Start rpc
	if r.config.RPC.Enable && r.config.RPC.IPCEnabled {
		api := r.GetIpcApis()
		if err := r.startIPC(api); err != nil {
			r.stopInProcess()
			return err
		}
	}

	if r.config.RPC.Enable && r.config.RPC.HTTPEnabled {
		apis := r.GetHttpApis()
		if err := r.startHTTP(r.config.RPC.HTTPEndpoint, apis, nil, r.config.RPC.HTTPCors, r.config.RPC.HttpVirtualHosts, HTTPTimeouts{}, false); err != nil {
			r.logger.Info(err)
			r.stopInProcess()
			r.stopIPC()
			return err
		}
	}

	if r.config.RPC.Enable && r.config.RPC.WSEnabled {
		apis := r.GetWSApis()
		if err := r.startWS(r.config.RPC.WSEndpoint, apis, nil, []string{}, false); err != nil {
			r.logger.Info(err)
			r.stopInProcess()
			r.stopIPC()
			r.stopHTTP()
			return err
		}
	}
	//if len(r.config.DashboardTargetURL) > 0 {
	//	apis := api.GetPublicApis()
	//	if len(r.config.PublicModules) != 0 {
	//		apis = api.GetApis(r.config.PublicModules...)
	//	}
	//
	//	targetUrl := r.config.DashboardTargetURL + "/ws/gvite/" + strconv.FormatUint(uint64(r.config.NetID), 10) + "@" + hex.EncodeToString(node.p2pServer.PrivateKey.PubByte())
	//
	//	u, e := url.Parse(targetUrl)
	//	if e != nil {
	//		return e
	//	}
	//	if u.Scheme != "ws" && u.Scheme != "wss" {
	//		return errors.New("DashboardTargetURL need match WebSocket Protocol.")
	//	}
	//
	//	cli, server, e := StartWSCliEndpoint(u, apis, nil, r.config.WSExposeAll)
	//	if e != nil {
	//		cli.Close()
	//		server.Stop()
	//		return e
	//	} else {
	//		r.wsCli = cli
	//	}
	//}

	return nil
}

func scheme(endpoint string) (string, string, error) {
	fmt.Println(endpoint)
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}
