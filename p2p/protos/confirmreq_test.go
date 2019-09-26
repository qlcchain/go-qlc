package protos

import (
	"encoding/json"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	testBlockConfirmreq1 = `{
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
	testBlockConfirmreq2 = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "b4badad5bf7aa378c35e92b00003c004ba588c6d0a5907db4b866332697876b4",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "600000",
      "link":"7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`
)

func TestConfirmReqPacket(t *testing.T) {
	blk1 := new(types.StateBlock)
	if err := json.Unmarshal([]byte(testBlockConfirmreq1), &blk1); err != nil {
		t.Fatal(err)
	}

	blk2 := new(types.StateBlock)
	if err := json.Unmarshal([]byte(testBlockConfirmreq2), &blk2); err != nil {
		t.Fatal(err)
	}

	rsp := ConfirmReqBlock{}
	rsp.Blk = append(rsp.Blk, blk1, blk2)

	bytes, err := ConfirmReqBlockToProto(&rsp)
	if err != nil {
		t.Fatal(err)
	}

	block, err := ConfirmReqBlockFromProto(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if blk1.GetPrevious() != block.Blk[0].GetPrevious() {
		t.Fatal("PreviousHash error")
	}

	if blk1.GetHash() != block.Blk[0].GetHash() {
		t.Fatal("hash error")
	}

	if blk2.GetPrevious() != block.Blk[1].GetPrevious() {
		t.Fatal("PreviousHash error")
	}

	if blk2.GetHash() != block.Blk[1].GetHash() {
		t.Fatal("hash error")
	}
}
