package protos

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	HeaderBlockHash = "5f8c4940cc180b2f12ebb13d8cd7deeeee459215cb8613c655f2469f29e7ef45"
	OpenBlockHash   = "5f8c4940cc180b2f12ebb13d8cd7deeeee459215cb8613c655f2469f29e7ef45"
	testHeaderBlock = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "b4badad5bf7aa378c35e92b00003c004ba588c6d0a5907db4b866332697876b4",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "500000",
      "link":"7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`
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

	blk := new(types.StateBlock)
	if err := json.Unmarshal([]byte(testBulkPull), &blk); err != nil {
		t.Fatal(err)
	}

	frb := &types.FrontierBlock{
		Fr:        Frontier,
		HeaderBlk: blk,
	}
	frs := &FrontierResponse{}

	frs.Fs = append(frs.Fs, frb)
	frBytes, err := FrontierResponseToProto(frs)
	if err != nil {
		t.Fatal("FrontierResponseToProto error")
	}
	frontierRsp, err := FrontierResponseFromProto(frBytes)
	if err != nil {
		t.Fatal("FrontierResponseFromProto error")
	}
	if frontierRsp.Fs[0].Fr.HeaderBlock != Frontier.HeaderBlock {
		t.Fatal("parse Headerblock error")
	}
	if frontierRsp.Fs[0].Fr.OpenBlock != Frontier.OpenBlock {
		t.Fatal("parse Openblock error")
	}
	if frontierRsp.Fs[0].HeaderBlk.GetHash() != Frontier.HeaderBlock {
		t.Fatal("hash error")
	}
}
