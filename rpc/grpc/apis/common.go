package apis

import pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"

func toStringPoint(s string) *string {
	if s != "" {
		return &s
	}
	return nil
}

func toUInt32PointByProto(c *pb.UInt32) *uint32 {
	if c == nil {
		return nil
	}
	r := c.GetValue()
	return &r
}
