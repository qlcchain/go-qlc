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
	  "link": "e6f766540b2c4f380255844a00602bbcff5aec24b4e1a400256bac902dba9228",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "povHeight": 0,
	  "timestamp": 1588048265,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
	  "work": "00000000001e3828",
	  "signature": "80b6f876db180764a407ad71c8f8e5208f1e667b812199d7299a4d1600dcf57cbabac56755867fd8a5341b03a78fff34d148c26574daff39c643441cacd5dc0f"
	}
     `

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
	  "previous": "184a111e7b975ff5c7e7a0b95bb03979950d2221e9f604a5d0fa9c65720d4088",
	  "link": "b7902600dfc79387b2601edc347b854d55d6b31142e324a4e54ff00a4c519c91",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "uuDnSNBQPHKpvU3xtss8bEgGpQWs8MaaHi55C25hd02lxWU7FnazE4iOVVR1g8Acb1AK1Uft9O9dLJuwbmY9VIHZqtYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEA2MGVlOWMxOGE0ODM5MWFmNTBkZWIxNTcxMzljNzA4Y2E4NTYxYzJkNTcyODE3NDA4MTk5N2MwMTY2NDVhYzIz",
	  "povHeight": 0,
	  "timestamp": 1588048267,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
	  "work": "00000000006e442a",
	  "signature": "f98ea1f306e9aa83ee1e4242840f6bda4e7e7c60e35d346fbfebb251cddb34fa71313299722c800171ccdbcbcf9e198723fd0879e510689b00dc3223c429c007"
	}
	`

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
	  "link": "8700c42155b09bb0d615dcb9c97cbc53238c47ed39c7086505e3b07e37e990b6",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAO5rKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABetOCL0FA8cqm9TfG2yzxsSAalBazwxpoeLnkLbmF3TaXFZTsWdrMTiI5VVHWDwBxvUArVR+30710sm7BuZj1Ugdmq1gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEA2MGVlOWMxOGE0ODM5MWFmNTBkZWIxNTcxMzljNzA4Y2E4NTYxYzJkNTcyODE3NDA4MTk5N2MwMTY2NDVhYzIz",
	  "povHeight": 0,
	  "timestamp": 1588048270,
	  "extra": "babeedfb416c9c1920735a4a2dd215c5917301f1c8e58019d0a30bfba7c32612",
	  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
	  "work": "0000000000519537",
	  "signature": "7d611158466852f4fe25450e98d183f237487e7d68f29394c88d2d7b019e80ba7feaabfe0bb436cef212849041027dd33282f55494789a8ecb5fb63138fc9701"
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
