package protos

import (
	"testing"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	testBlockConfirmack = `{
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

func TestConfirmAckBlockPacket(t *testing.T) {
	addrString := "qlc_38nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8"
	address, err := types.HexToAddress(addrString)
	if err != nil {
		t.Fatal("HexToAddress error")
	}
	signString := "148AA79F002D747E4E262B0CC2F7B5FAB121C9362C8DB5906DC40B91147A57DAA827DF4321D0D8DED972C2469C72B4191E3AF9A69A67FC893462DCE19E9E7005"
	var sign types.Signature
	err = sign.Of(signString)
	if err != nil {
		t.Fatal("sign error")
	}
	blk, err := types.NewBlock(types.State)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = json.Unmarshal([]byte(testBlockConfirmack), &blk); err != nil {
		t.Fatal(err)
	}
	rsp := ConfirmAckBlock{
		Account:   address,
		Signature: sign,
		Sequence:  0,
		Blk:       blk,
	}
	bytes, err := ConfirmAckBlockToProto(&rsp)
	if err != nil {
		t.Fatal(err)
	}
	ack, err := ConfirmAckBlockFromProto(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if blk.GetType() != ack.Blk.GetType() {
		t.Fatal("type error")
	}
	if blk.GetPrevious() != ack.Blk.GetPrevious() {
		t.Fatal("PreviousHash error")
	}
	if blk.GetHash() != ack.Blk.GetHash() {
		t.Fatal("hash error")
	}
	if sign != ack.Signature {
		t.Fatal("parse signature error")
	}
	if address != ack.Account {
		t.Fatal("parse address error")
	}
}
