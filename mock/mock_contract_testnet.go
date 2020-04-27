// +build  testnet

package mock

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	open = `	
	{
	  "type": "Open",
	  "token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
	  "address": "qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf",
	  "balance": "1000000000000",
	  "vote": "0",
	  "network": "0",
	  "storage": "0",
	  "oracle": "0",
	  "previous": "0000000000000000000000000000000000000000000000000000000000000000",
	  "link": "5dd8bca2db0037a7aed77b0bf3b505e72144b995d5cb6cddde74a588c820bf83",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "povHeight": 0,
	  "timestamp": 1587974607,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
	  "work": "0000000000000000",
	  "signature": "2fd8895f29e18714e6e893bdd0abdd9160623d77fd322a6d9bb01435a05555c6d6eb7cf5887f26879b5b83552f83a79e050a44235b575d9661ba15e8cbf8be04"
	}`

	send = `
	{
	  "type": "ContractSend",
	  "token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
	  "address": "qlc_17mppebrj5kocjtr9i1wfxa1ooc9xqtgyqbemgr8wsjxck1xmcppp9abciaf",
	  "balance": "999000000000",
	  "vote": "0",
	  "network": "0",
	  "storage": "0",
	  "oracle": "0",
	  "previous": "33fd25517fb4ef0205223c11638af7bc7efe6abd32d7f34a739a6848cedc1c86",
	  "link": "b7902600dfc79387b2601edc347b854d55d6b31142e324a4e54ff00a4c519c91",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "uuDnSNBQPHKpvU3xtss8bEgGpQWs8MaaHi55C25hd02lxWU7FnazE4iOVVR1g8Acb1AK1Uft9O9dLJuwbmY9VIHZqtYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBmZTZhYWE1NjhmMWE2YjRhYzA2ZTdlMGNiMWFmMDc5N2M3ZDNjNDg4Yjc3YmVhMWEyNmE0YjQ2MGNkMjdkYjlh",
	  "povHeight": 0,
	  "timestamp": 1587974608,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
	  "work": "0000000000000000",
	  "signature": "b08d32f421897520b3b18e8c42a14a7b6d9f4de0c41bab731c18b5580a3e81945f00ef6b5a36318050821756f77356ee367fec63b5c18bf099e300d8cbbaec0b"
	}`

	rece = `
	{
	  "type": "ContractReward",
	  "token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
	  "address": "qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s",
	  "balance": "0",
	  "vote": "1000000000",
	  "network": "0",
	  "storage": "0",
	  "oracle": "0",
	  "previous": "0000000000000000000000000000000000000000000000000000000000000000",
	  "link": "25f09bf0da5c5ff0f36eb4e099e7690cd944798b87ee1ff8eb855e1bb80a62aa",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAO5rKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABeppQo0FA8cqm9TfG2yzxsSAalBazwxpoeLnkLbmF3TaXFZTsWdrMTiI5VVHWDwBxvUArVR+30710sm7BuZj1Ugdmq1gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBmZTZhYWE1NjhmMWE2YjRhYzA2ZTdlMGNiMWFmMDc5N2M3ZDNjNDg4Yjc3YmVhMWEyNmE0YjQ2MGNkMjdkYjlh",
	  "povHeight": 0,
	  "timestamp": 1587974609,
	  "extra": "e811f2a36c62d3b68dbf66d099efd4e3ff2fd9cc96182c92d359698c0c48a60e",
	  "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
	  "work": "0000000000000000",
	  "signature": "28b573aedf57349248a609f3594645fc9bb9f8b87eb410bb01819341ce87cb5f1541b9181ac9503ff34e33b32bcc430374a536af3f7672e1c6bdce8fd16aec0f"
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
