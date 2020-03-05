package protos

import (
	"testing"

	"github.com/qlcchain/go-qlc/mock"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestPovBulkPullReq(t *testing.T) {
	var start1, start2 types.Hash
	err := start1.Of("D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of StartHash1 error")
	}
	err = start2.Of("12F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of StartHash2 error")
	}
	var Locators []*types.Hash
	Locators = append(Locators, &start1)
	Locators = append(Locators, &start2)
	povReq := &PovBulkPullReq{
		StartHash:   start1,
		StartHeight: 1000,
		Count:       2,
		PullType:    PovPullTypeBackward,
		Reason:      PovReasonFetch,
		Locators:    Locators,
	}
	data, err := PovBulkPullReqToProto(povReq)
	if err != nil {
		t.Fatal(err)
	}
	r, err := PovBulkPullReqFromProto(data)
	if err != nil {
		t.Fatal(err)
	}
	if r.StartHash != start1 || r.StartHeight != 1000 || r.Count != 2 || r.PullType != PovPullTypeBackward || r.Reason != PovReasonFetch || len(r.Locators) != 2 {
		t.Fatal("there is some error in PovBulkPullReqToProto or PovBulkPullReqFromProto")
	}
}

func TestPovBulkPullRsp(t *testing.T) {
	blk1, _ := mock.GeneratePovBlock(nil, 0)
	blk2, _ := mock.GeneratePovBlock(nil, 0)
	rsp := &PovBulkPullRsp{
		Count:  2,
		Reason: PovReasonFetch,
		Blocks: types.PovBlocks{blk1, blk2},
	}
	data, err := PovBulkPullRspToProto(rsp)
	if err != nil {
		t.Fatal(err)
	}
	r, err := PovBulkPullRspFromProto(data)
	if err != nil {
		t.Fatal(err)
	}
	if r.Reason != rsp.Reason || r.Count != rsp.Count || len(r.Blocks) != len(rsp.Blocks) {
		t.Fatal("there is some error in PovBulkPullRspToProto or PovBulkPullRspFromProto")
	}
}
