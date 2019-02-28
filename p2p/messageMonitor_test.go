package p2p

import (
	"bytes"
	"math"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/test/mock"
)

func TestMarshalMessage(t *testing.T) {
	blk := mock.StateBlock()
	data1, err := marshalMessage(PublishReq, blk)
	if err != nil {
		t.Fatal("Marshal PublishReq err1")
	}
	pushBlock := protos.PublishBlock{
		Blk: blk,
	}
	data2, err := protos.PublishBlockToProto(&pushBlock)
	if err != nil {
		t.Fatal("Marshal PublishReq err2")
	}
	if bytes.Compare(data1, data2) != 0 {
		t.Fatal("Marshal PublishReq err3")
	}
	data3, err := marshalMessage(ConfirmReq, blk)
	if err != nil {
		t.Fatal("Marshal ConfirmReq err1")
	}
	packet := &protos.ConfirmReqBlock{
		Blk: blk,
	}
	data4, err := protos.ConfirmReqBlockToProto(packet)
	if err != nil {
		t.Fatal("Marshal ConfirmReq err2")
	}
	if bytes.Compare(data3, data4) != 0 {
		t.Fatal("Marshal ConfirmReq err3")
	}
	var va protos.ConfirmAckBlock
	a := mock.Account()
	va.Sequence = 0
	va.Blk = blk
	va.Account = a.Address()
	va.Signature = a.Sign(blk.GetHash())
	data5, err := marshalMessage(ConfirmAck, &va)
	if err != nil {
		t.Fatal("Marshal ConfirmAck err1")
	}
	data6, err := protos.ConfirmAckBlockToProto(&va)
	if err != nil {
		t.Fatal("Marshal ConfirmAck err2")
	}
	if bytes.Compare(data5, data6) != 0 {
		t.Fatal("Marshal ConfirmAck err3")
	}
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	data7, err := marshalMessage(FrontierRequest, Req)
	if err != nil {
		t.Fatal("Marshal FrontierRequest err1")
	}
	data8, err := protos.FrontierReqToProto(Req)
	if err != nil {
		t.Fatal("Marshal FrontierRequest err2")
	}
	if bytes.Compare(data7, data8) != 0 {
		t.Fatal("Marshal FrontierRequest err3")
	}
	zeroFrontier := new(types.Frontier)
	data9, err := marshalMessage(FrontierRsp, zeroFrontier)
	if err != nil {
		t.Fatal("Marshal FrontierRsp err1")
	}
	f := protos.NewFrontierRsp(zeroFrontier)
	data10, err := protos.FrontierResponseToProto(f)
	if err != nil {
		t.Fatal("Marshal FrontierRsp err2")
	}
	if bytes.Compare(data9, data10) != 0 {
		t.Fatal("Marshal FrontierRsp err3")
	}
	b := &protos.BulkPullReqPacket{
		StartHash: types.ZeroHash,
		EndHash:   types.ZeroHash,
	}
	data11, err := marshalMessage(BulkPullRequest, b)
	if err != nil {
		t.Fatal("Marshal BulkPullRequest err1")
	}
	data12, err := protos.BulkPullReqPacketToProto(b)
	if err != nil {
		t.Fatal("Marshal BulkPullRequest err2")
	}
	if bytes.Compare(data11, data12) != 0 {
		t.Fatal("Marshal BulkPullRequest err3")
	}
	data13, err := marshalMessage(BulkPullRsp, blk)
	if err != nil {
		t.Fatal("Marshal BulkPullRsp err1")
	}
	r := &protos.BulkPullRspPacket{
		Blk: blk,
	}
	data14, err := protos.BulkPullRspPacketToProto(r)
	if err != nil {
		t.Fatal("Marshal BulkPullRsp err2")
	}
	if bytes.Compare(data13, data14) != 0 {
		t.Fatal("Marshal BulkPullRsp err3")
	}
	data15, err := marshalMessage(BulkPushBlock, blk)
	if err != nil {
		t.Fatal("Marshal BulkPushBlock err1")
	}
	push := &protos.BulkPush{
		Blk: blk,
	}
	data16, err := protos.BulkPushBlockToProto(push)
	if err != nil {
		t.Fatal("Marshal BulkPushBlock err2")
	}
	if bytes.Compare(data15, data16) != 0 {
		t.Fatal("Marshal BulkPushBlock err3")
	}
}
