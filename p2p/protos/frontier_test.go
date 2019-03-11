package protos

import (
	"math"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	HeaderBlockHash = "D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D41"
	OpenBlockHash   = "D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47"
)

func TestFrontierReq(t *testing.T) {
	address := types.Address{}
	Req := NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	frBytes, err := FrontierReqToProto(Req)
	if err != nil {
		t.Fatal("FrontierReqToProto error")
	}
	t.Log(frBytes)
	frStruct, err := FrontierReqFromProto(frBytes)
	if err != nil {
		t.Fatal("FrontierReqFromProto error")
	}
	if frStruct.StartAddress != address {
		t.Fatal("Address error")
	}
	if frStruct.Count != math.MaxUint32 {
		t.Fatal("Count error")
	}
	if frStruct.Age != math.MaxUint32 {
		t.Fatal("Age error")
	}
}

func TestFrontierRsp(t *testing.T) {
	Frontier := new(types.Frontier)
	err := Frontier.HeaderBlock.Of(HeaderBlockHash)
	if err != nil {
		t.Fatal("string to headerhash error")
	}
	err = Frontier.OpenBlock.Of(OpenBlockHash)
	if err != nil {
		t.Fatal("string to openblockhash error")
	}
	fr := NewFrontierRsp(Frontier, 5)
	frBytes, err := FrontierResponseToProto(fr)
	if err != nil {
		t.Fatal("FrontierResponseToProto error")
	}
	frontierRsp, err := FrontierResponseFromProto(frBytes)
	if err != nil {
		t.Fatal("FrontierResponseFromProto error")
	}
	if frontierRsp.Frontier.HeaderBlock != Frontier.HeaderBlock {
		t.Fatal("parse Headerblock error")
	}
	if frontierRsp.Frontier.OpenBlock != Frontier.OpenBlock {
		t.Fatal("parse Openblock error")
	}
	if frontierRsp.TotalFrontierNum != fr.TotalFrontierNum {
		t.Fatal("parse TotalFrontierNum error")
	}
}
