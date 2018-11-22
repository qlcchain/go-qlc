package main

import (
	"fmt"
	"time"

	"github.com/json-iterator/go"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/p2p"
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
)

type Observer1 struct {
}
type Observer2 struct {
}

func (ob1 *Observer1) OnNotify(event *p2p.Event) {
	fmt.Printf("ob1 receive messagetype:%s,block id %d:\n", event.MsgType, event.Blocks.ID())

}
func (ob2 *Observer2) OnNotify(event *p2p.Event) {
	fmt.Printf("ob2 receive messagetype:%s,block id %d:\n", event.MsgType, event.Blocks.ID())
}
func main() {
	cfg := config.DefaultlConfig
	node, err := p2p.NewQlcService(cfg)
	if err != nil {
		fmt.Println(err)
	}
	obs1 := new(Observer1)
	obs2 := new(Observer2)

	node.MessageEvent().Regist(obs1)
	node.MessageEvent().Regist(obs2)
	node.Start()
	blk1, err := types.NewBlock(byte(types.State))
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err = json.Unmarshal([]byte(test_block1), &blk1); err != nil {
		fmt.Println(err)
	}
	blockBytes, err := blk1.MarshalMsg(nil)
	data := make([]byte, 1+len(blockBytes))
	data[0] = byte(types.State)
	copy(data[1:], blockBytes)
	for {
		node.Broadcast(p2p.PublishRequest, data)
		time.Sleep(time.Duration(2) * time.Second)
	}
	select {}
}
