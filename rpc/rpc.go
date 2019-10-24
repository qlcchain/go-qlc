package rpc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/wallet"
)

type RPC struct {
	rpcAPIs          []rpc.API
	inProcessHandler *rpc.Server

	ipcListener net.Listener
	ipcHandler  *rpc.Server

	httpWhitelist []string
	httpListener  net.Listener
	httpHandler   *rpc.Server

	wsListener net.Listener
	wsHandler  *rpc.Server

	config             *config.Config
	DashboardTargetURL string
	NetID              uint `json:"NetID"`

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	ledger   *ledger.Ledger
	wallet   *wallet.WalletStore
	relation *relation.Relation
	eb       event.EventBus
	cfgFile  string
	logger   *zap.SugaredLogger
}

func NewRPC(cfgFile string) (*RPC, error) {
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()

	rl, err := relation.NewRelation(cfgFile)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	r := RPC{
		ledger:   ledger.NewLedger(cfgFile),
		wallet:   wallet.NewWalletStore(cfgFile),
		relation: rl,
		eb:       cc.EventBus(),
		config:   cfg,
		cfgFile:  cfgFile,
		ctx:      ctx,
		cancel:   cancel,
		logger:   log.NewLogger("rpc"),
	}
	return &r, nil
}

// startIPC initializes and starts the IPC RPC endpoint.
func (r *RPC) startIPC(apis []rpc.API) error {
	if r.config.RPC.IPCEndpoint == "" {
		return nil // IPC disabled.
	}
	listener, handler, err := rpc.StartIPCEndpoint(r.config.RPC.IPCEndpoint, apis)
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
func (r *RPC) startHTTP(endpoint string, apis []rpc.API, modules []string, cors []string, vhosts []string, timeouts rpc.HTTPTimeouts) error {
	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := r.StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts, timeouts)
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
func (r *RPC) startWS(endpoint string, apis []rpc.API, modules []string, wsOrigins []string, exposeAll bool) error {
	// Short circuit if the WS endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := r.StartWSEndpoint(endpoint, apis, modules, wsOrigins, exposeAll)
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
}

func (r *RPC) Attach() (*rpc.Client, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	//if r.p2pServer == nil {
	//	return nil, ErrNodeStopped
	//}

	if r.inProcessHandler == nil {
		return nil, errors.New("server not started")
	}
	return rpc.DialInProc(r.inProcessHandler), nil
}

// startInProc initializes an in-process RPC endpoint.
func (r *RPC) startInProcess(apis []rpc.API) error {
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()
	for _, api := range apis {
		if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
			r.logger.Info(err)
			return err
		}
		//r.logger.Debug("InProc registered ", "service ", api.Service, " namespace ", api.Namespace)
	}
	r.logger.Info("InProc start successfully")
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
	r.cancel()
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
		if err := r.startHTTP(r.config.RPC.HTTPEndpoint, apis, nil, r.config.RPC.HTTPCors, r.config.RPC.HttpVirtualHosts, rpc.HTTPTimeouts{}); err != nil {
			r.logger.Info(err)
			r.stopInProcess()
			r.stopIPC()
			return err
		}
	}

	if r.config.RPC.Enable && r.config.RPC.WSEnabled {
		apis := r.GetWSApis()
		if err := r.startWS(r.config.RPC.WSEndpoint, apis, nil, r.config.RPC.HTTPCors, false); err != nil {
			r.logger.Info(err)
			r.stopInProcess()
			r.stopIPC()
			r.stopHTTP()
			return err
		}
	}

	return nil
}

// StartHTTPEndpoint starts the HTTP RPC endpoint, configured with cors/vhosts/modules
func (r *RPC) StartHTTPEndpoint(endpoint string, apis []rpc.API, modules []string, cors []string, vhosts []string, timeouts rpc.HTTPTimeouts) (net.Listener, *rpc.Server, error) {
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()

	for _, api := range apis {
		if whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}
			r.logger.Debug("HTTP registered ", "namespace ", api.Namespace)
		}
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	network, address, err := scheme(endpoint)
	if err != nil {
		return nil, nil, err
	}
	if listener, err = net.Listen(network, address); err != nil {
		return nil, nil, err
	}

	hServer := new(http.Server)
	go func(hServer *http.Server) {
		hServer = rpc.NewHTTPServer(cors, vhosts, timeouts, handler)
		hServer.Serve(listener)
		select {
		case <-r.ctx.Done():
			hServer.Close()
		}
	}(hServer)

	return listener, handler, err
}

// StartWSEndpoint starts a websocket endpoint
func (r *RPC) StartWSEndpoint(endpoint string, apis []rpc.API, modules []string, wsOrigins []string, exposeAll bool) (net.Listener, *rpc.Server, error) {
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()
	for _, api := range apis {
		if exposeAll || whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return nil, nil, err
			}
			r.logger.Debug("WebSocket registered ", " service ", api.Service, " namespace ", api.Namespace)
		}
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	network, address, err := scheme(endpoint)
	if err != nil {
		return nil, nil, err
	}

	if listener, err = net.Listen(network, address); err != nil {
		return nil, nil, err
	}

	//go rpc.NewWSServer(wsOrigins, handler).Serve(listener)
	hServer := new(http.Server)
	go func(hServer *http.Server) {
		hServer = rpc.NewWSServer(wsOrigins, handler)
		hServer.Serve(listener)
		select {
		case <-r.ctx.Done():
			hServer.Close()
		}
	}(hServer)

	return listener, handler, err
}

func scheme(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}
