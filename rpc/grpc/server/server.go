package grpcServer

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	rpc     *grpc.Server
	ledger  ledger.Store
	eb      event.EventBus
	config  *config.Config
	cfgFile string
	lock    sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *zap.SugaredLogger
}

func Start(cfgFile string) (*GRPCServer, error) {
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	ctx, cancel := context.WithCancel(context.Background())
	l := ledger.NewLedger(cfgFile)
	eb := cc.EventBus()

	network, address, err := scheme(cfg.RPC.GRPCConfig.GRPCListenAddress)
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen(network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %s", err)
	}
	gRpcServer := grpc.NewServer()
	pb.RegisterTestAPIServer(gRpcServer, &TestApi{})
	pb.RegisterLedgerAPIServer(gRpcServer, NewLedgerApi(ctx, l, eb, cc))
	reflection.Register(gRpcServer)

	go gRpcServer.Serve(lis)
	if err := newGateway(address, cfg.RPC.GRPCConfig.ListenAddress); err != nil {
		return nil, fmt.Errorf("start gateway: %s", err)
	}
	return &GRPCServer{
		rpc:     gRpcServer,
		ledger:  l,
		eb:      eb,
		config:  cfg,
		cfgFile: cfgFile,
		ctx:     ctx,
		cancel:  cancel,
		logger:  log.NewLogger("rpc"),
	}, nil
}

func newGateway(grpcAddress, gwAddress string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterChainAPIHandlerFromEndpoint(ctx, gwmux, grpcAddress, opts)
	if err != nil {
		return fmt.Errorf("gateway register: %s", err)
	}
	_, address, err := scheme(gwAddress)
	if err != nil {
		return err
	}
	http.ListenAndServe(address, gwmux)
	return nil
}

func (r *GRPCServer) Stop() {
	r.rpc.Stop()
}

func scheme(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}
