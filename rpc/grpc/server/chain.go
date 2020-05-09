package grpcServer

import (
	"context"

	"github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

type chainApi struct {
}

func (chainApi) Version(context.Context, *proto.VersionRequest) (*proto.VersionResponse, error) {
	//return &proto.VersionResponse{
	//	BuildTime: version.BuildTime,
	//	Version:   version.Version,
	//	Hash:      version.GitRev,
	//	Mode:      version.Mode,
	//}, nil
	return &proto.VersionResponse{
		BuildTime: "2020-03-04",
		Version:   "dfea3421",
		Hash:      "686952c9aee99cb5cef1ac156ff72064dfaf3f659d496f2a2c88b614a2101e2b",
		Mode:      "testnet",
	}, nil
}
