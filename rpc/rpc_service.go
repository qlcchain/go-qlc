package rpc

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
)

type RPCService struct {
	rpc *RPC
}

func NewRPCService(cfg *config.Config, dpos *consensus.DposService) *RPCService {
	return &RPCService{NewRPC(cfg, dpos)}
}

func (rs *RPCService) Init() error {
	rs.rpc.logger.Info("rpc service init")
	return nil
}

func (rs *RPCService) Start() error {
	err := rs.rpc.StartRPC()
	if err != nil {
		return err
	}
	return nil
}

func (rs *RPCService) Stop() error {
	rs.rpc.logger.Info("rpc stopping...")
	rs.rpc.StopRPC()
	return nil
}

func (rs *RPCService) Status() int32 {
	panic("implement me")
}

func (rs *RPCService) RPC() *RPC {
	return rs.rpc
}
