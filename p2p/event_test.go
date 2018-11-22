package p2p

import (
	"fmt"
	"testing"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	test_block1 = `{
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

	test_block2 = `{
      "type": "state",
      "address":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "previousHash": "0000000000000000000000000000000000000000000000000000000000000000",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "100000",
      "link":"84533798231c7fb7e78a796f7090c1d26db58eab284115beeb07825b187c1780",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`
	test_block3 = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "84533798231c7fb7e78a796f7090c1d26db58eab284115beeb07825b187c1780",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "200000",
      "link":"7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`
)

type Observer1 struct {
}
type Observer2 struct {
}

func (ob1 *Observer1) OnNotify(event *Event) {
	fmt.Printf("ob1 receive messagetype:%s,block id %d:\n", event.MsgType, event.Blocks.ID())

}
func (ob2 *Observer2) OnNotify(event *Event) {
	fmt.Printf("ob2 receive messagetype:%s,block id %d:\n", event.MsgType, event.Blocks.ID())
}
func TestEvent(t *testing.T) {
	cs := &ConcreteSubject{
		Observers: make(map[Observer]struct{}),
	}

	obs1 := new(Observer1)
	obs2 := new(Observer2)

	cs.Regist(obs1)
	cs.Regist(obs2)
	blk, err := types.NewBlock(byte(types.State))
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = json.Unmarshal([]byte(test_block1), &blk); err != nil {
		t.Fatal(err)
	}
	e := &Event{
		MsgType: "publish",
		Blocks:  blk,
	}
	cs.Notify(e)
}
