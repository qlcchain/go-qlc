package apis

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

type TestApi struct {
}

func (TestApi) Version(context.Context, *empty.Empty) (*proto.VersionResponse4, error) {
	//return &proto.VersionResponse{
	//	BuildTime: version.BuildTime,
	//	Version:   version.Version,
	//	Hash:      version.GitRev,
	//	Mode:      version.Mode,
	//}, nil
	return &proto.VersionResponse4{
		BuildTime: "2020-03-04",
		Version:   "version1",
		Hash:      "686952c9aee99cb5cef1ac156ff72064dfaf3f659d496f2a2c88b614a2101e2b",
		Mode:      "testnet",
	}, nil
}

func (TestApi) Version2(context.Context, *proto.VersionRequest2) (*proto.VersionResponse4, error) {
	return &proto.VersionResponse4{
		BuildTime: "2020-03-04",
		Version:   "version2",
		Hash:      "686952c9aee99cb5cef1ac156ff72064dfaf3f659d496f2a2c88b614a2101e2b",
		Mode:      "testnet",
	}, nil
}

func (TestApi) Version3(ctx context.Context, vr *proto.VersionRequest3) (*proto.VersionResponse4, error) {
	return &proto.VersionResponse4{
		BuildTime: "2020-03-04",
		Version:   "version3-" + vr.Version,
		Hash:      "686952c9aee99cb5cef1ac156ff72064dfaf3f659d496f2a2c88b614a2101e2b",
		Mode:      "testnet",
	}, nil
}

func (TestApi) BlockStream(req *proto.BlockRequest, srv proto.TestAPI_BlockStreamServer) error {
	for {
		time.Sleep(1 * time.Second)
		srv.Send(&proto.BlockResponse{
			BuildTime: time.Now().String(),
		})
	}
}
