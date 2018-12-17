package protos

import (
	"fmt"
	"math"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

var HeaderBlockHash = "D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D41"
var OpenBlockHash = "D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47"

func TestFrontierReq(t *testing.T) {
	address := types.Address{}
	Req := NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	frbytes, err := FrontierReqToProto(Req)
	if err != nil {
		t.Fatal("FrontierReqToProto error")
	}
	fmt.Println(frbytes)
	frstruct, err := FrontierReqFromProto(frbytes)
	if err != nil {
		t.Fatal("FrontierReqFromProto error")
	}
	if frstruct.StartAddress != address {
		t.Fatal("Address error")
	}
	if frstruct.Count != math.MaxUint32 {
		t.Fatal("Count error")
	}
	if frstruct.Age != math.MaxUint32 {
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
	fr := NewFrontierRsp(Frontier)
	frbytes, err := FrontierResponseToProto(fr)
	if err != nil {
		t.Fatal("FrontierResponseToProto error")
	}
	frontier, err := FrontierResponseFromProto(frbytes)
	if err != nil {
		t.Fatal("FrontierResponseFromProto error")
	}
	if frontier.Frontier.HeaderBlock != Frontier.HeaderBlock {
		t.Fatal("parse Headerblock error")
	}
	if frontier.Frontier.OpenBlock != Frontier.OpenBlock {
		t.Fatal("parse Openblock error")
	}
}
