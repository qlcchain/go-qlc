// +build  !testnet

package mock

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	open = `
	{
	  "type": "Open",
	  "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
	  "address": "qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf",
	  "balance": "1000000000000",
	  "vote": "0",
	  "network": "0",
	  "storage": "0",
	  "oracle": "0",
	  "previous": "0000000000000000000000000000000000000000000000000000000000000000",
	  "link": "a8b3a0de0db3abc458109639fdb4368ee483615a15c5449ba8665debd730e482",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "povHeight": 0,
	  "timestamp": 1587989546,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
	  "work": "00000000001e3828",
	  "signature": "fd71e436229b810d9dcd5ae2530c4cb051a5279fb064d6758fd1754a9e91a3ce94ebce0bac56383c9d08a41e4e82f05dbb3a88b4df164787206cf9ac6fe9850d"
	}`

	send = `
	{
	  "type": "ContractSend",
	  "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
	  "address": "qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf",
	  "balance": "999000000000",
	  "vote": "0",
	  "network": "0",
	  "storage": "0",
	  "oracle": "0",
	  "previous": "21e15ed8700c0a2d8ebe8c3ebee981a6558dcc5a3781b002d6f4accc16a195ba",
	  "link": "b7902600dfc79387b2601edc347b854d55d6b31142e324a4e54ff00a4c519c91",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "uuDnSNBQPHKpvU3xtss8bEgGpQWs8MaaHi55C25hd02lxWU7FnazE4iOVVR1g8Acb1AK1Uft9O9dLJuwbmY9VIHZqtYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAzYTI5NjNkMzY2YzE5NDI0M2ZhZGUzOGZlNmZmN2Q4NDgwNTk1MmQzNmE1NDk2OTQyMTBkMTI1YTc4MDU0ODU5",
	  "povHeight": 0,
	  "timestamp": 1587989547,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
	  "work": "00000000001e8f73",
	  "signature": "eee6769314bfd0c1a433617f0283ed63bd78140e1e00e9a29f5cea5721fe5cd3c3e02cf702a6ed5e4ab4bb44cb4306eda802daeb8ab10094366efd9bd23cc805"
	}`

	rece = `
	{
	  "type": "ContractReward",
	  "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
	  "address": "qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s",
	  "balance": "0",
	  "vote": "1000000000",
	  "network": "0",
	  "storage": "0",
	  "oracle": "0",
	  "previous": "0000000000000000000000000000000000000000000000000000000000000000",
	  "link": "4ca8501be354fd978c0d9a8ce10955a3173097518ba27683f278b5d56a18d91b",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAO5rKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABes/sr0FA8cqm9TfG2yzxsSAalBazwxpoeLnkLbmF3TaXFZTsWdrMTiI5VVHWDwBxvUArVR+30710sm7BuZj1Ugdmq1gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAzYTI5NjNkMzY2YzE5NDI0M2ZhZGUzOGZlNmZmN2Q4NDgwNTk1MmQzNmE1NDk2OTQyMTBkMTI1YTc4MDU0ODU5",
	  "povHeight": 0,
	  "timestamp": 1587989548,
	  "extra": "c8203d459f5492d56cb4b9a17b110afbc9231448e690407b4cdb2463407d9693",
	  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
	  "work": "0000000000519537",
	  "signature": "be58e163be897f9bf206e28206c4c3e329f5c0864dbb58ca02c267a09a63cdd3e2c9193687f226b1292fb58088278c7c5ff5ce8cc5987aebc1ea887c252c6c08"
	}
		`
)

func ContractBlocks() []*types.StateBlock {
	var openBlock types.StateBlock
	var sendBlock types.StateBlock
	var receBlock types.StateBlock
	_ = json.Unmarshal([]byte(open), &openBlock)
	_ = json.Unmarshal([]byte(send), &sendBlock)
	_ = json.Unmarshal([]byte(rece), &receBlock)
	bs := make([]*types.StateBlock, 0)
	bs = append(bs, &openBlock)
	bs = append(bs, &sendBlock)
	bs = append(bs, &receBlock)
	return bs
}
