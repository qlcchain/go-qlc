package grpcServer

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/grpc/apis"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

type GRPCServer struct {
	rpc     *grpc.Server
	ledger  ledger.Store
	eb      event.EventBus
	cfg     *config.Config
	cfgFile string
	cc      *chainctx.ChainContext
	ctx     context.Context
	hServer *http.Server
	logger  *zap.SugaredLogger
}

func Start(cfgFile string, ctx context.Context) (*GRPCServer, error) {
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	l := ledger.NewLedger(cfgFile)
	eb := cc.EventBus()

	network, address, err := scheme(cfg.RPC.GRPCConfig.ListenAddress)
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen(network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %s (%s,%s)", err, network, address)
	}
	qrpc := &GRPCServer{
		ledger:  l,
		eb:      eb,
		cfg:     cfg,
		cfgFile: cfgFile,
		ctx:     ctx,
		cc:      cc,
		logger:  log.NewLogger("grpc"),
	}

	grpcServer := grpc.NewServer(grpc.StreamInterceptor(qrpc.StreamServerInterceptor),
		grpc.UnaryInterceptor(qrpc.UnaryServerInterceptor))

	qrpc.rpc = grpcServer
	qrpc.registerApi()
	reflection.Register(grpcServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			qrpc.logger.Errorf("grpc start error: %s", err)
		}
	}()
	if cfg.RPC.GRPCConfig.HTTPEnable {
		go func() {
			if err := qrpc.newGateway(address); err != nil {
				qrpc.logger.Errorf("grpc start error: %s", err)
			}
		}()
	}
	qrpc.logger.Info("grpc serve successfully")
	return qrpc, nil
}

func (r *GRPCServer) newGateway(grpcAddress string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := registerGWApi(ctx, gwmux, grpcAddress, opts); err != nil {
		r.logger.Errorf("gateway register: %s", err)
		return err
	}
	_, address, err := scheme(r.cfg.RPC.GRPCConfig.HTTPListenAddress)
	if err != nil {
		return err
	}
	srv := &http.Server{Addr: address, Handler: gwmux}
	//return http.ListenAndServe(address, gwmux)
	r.hServer = srv
	return srv.ListenAndServe()
}

func (r *GRPCServer) Stop() {
	if r.hServer != nil {
		if err := r.hServer.Shutdown(context.Background()); err != nil {
			r.logger.Error(err)
		}
	}
	r.rpc.Stop()
	r.logger.Info("grpc stopped")
}

func scheme(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	return u.Scheme, u.Host, nil
}

func (r *GRPCServer) registerApi() {
	pb.RegisterAccountAPIServer(r.rpc, apis.NewAccountApi())
	pb.RegisterBlackHoleAPIServer(r.rpc, apis.NewBlackHoleAPI(r.ledger, r.cc))
	pb.RegisterChainAPIServer(r.rpc, apis.NewChainAPI(r.ledger))
	pb.RegisterContractAPIServer(r.rpc, apis.NewContractAPI(r.cc, r.ledger))
	pb.RegisterLedgerAPIServer(r.rpc, apis.NewLedgerApi(r.ctx, r.ledger, r.eb, r.cc))
	pb.RegisterMetricsAPIServer(r.rpc, apis.NewMetricsAPI())
	pb.RegisterMinerAPIServer(r.rpc, apis.NewMinerAPI(r.cfg, r.ledger))
	pb.RegisterMintageAPIServer(r.rpc, apis.NewMintageAPI(r.cfgFile, r.ledger))
	pb.RegisterNEP5PledgeAPIServer(r.rpc, apis.NewNEP5PledgeAPI(r.cfgFile, r.ledger))
	pb.RegisterNetAPIServer(r.rpc, apis.NewNetApi(r.ledger, r.eb, r.cc))
	pb.RegisterPermissionAPIServer(r.rpc, apis.NewPermissionAPI(r.cfgFile, r.ledger))
	pb.RegisterPovAPIServer(r.rpc, apis.NewPovAPI(r.ctx, r.cfg, r.ledger, r.eb, r.cc))
	pb.RegisterPrivacyAPIServer(r.rpc, apis.NewPrivacyAPI(r.cfg, r.ledger, r.eb, r.cc))
	pb.RegisterPtmKeyAPIServer(r.rpc, apis.NewPtmKeyAPI(r.cfgFile, r.ledger))
	pb.RegisterPublicKeyDistributionAPIServer(r.rpc, apis.NewPublicKeyDistributionAPI(r.cfgFile, r.ledger))
	pb.RegisterRepAPIServer(r.rpc, apis.NewRepAPI(r.cfg, r.ledger))
	pb.RegisterRewardsAPIServer(r.rpc, apis.NewRewardsAPI(r.ledger, r.cc))
	pb.RegisterSettlementAPIServer(r.rpc, apis.NewSettlementAPI(r.ledger, r.cc))
	pb.RegisterUtilAPIServer(r.rpc, apis.NewUtilApi(r.ledger))
}

func registerGWApi(ctx context.Context, gwmux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	if err := pb.RegisterAccountAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterBlackHoleAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterChainAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterContractAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterLedgerAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterMetricsAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterMinerAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterMintageAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterNEP5PledgeAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterNetAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterPermissionAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterPovAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterPrivacyAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterPtmKeyAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterPublicKeyDistributionAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterRepAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterRewardsAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterSettlementAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterUtilAPIHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
		return err
	}
	return nil
}

func (r *GRPCServer) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	r.logger.Debugf("before unary handling. info: %+v \n", info)
	resp, err := handler(ctx, req)
	r.logger.Debugf("after unary handling. resp: %+v \n", resp)
	return resp, err
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (r *GRPCServer) StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	r.logger.Debugf("before stream handling. info: %+v \n", info)
	err := handler(srv, ss)
	r.logger.Debugf("after stream handling. err: %v \n", err)
	return err
}
