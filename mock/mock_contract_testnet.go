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
	  "link": "ae03b950e7284eaedab98571509b58b5d797180dd7a90da4f95fee10a09e23b1",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "povHeight": 0,
	  "timestamp": 1588045333,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
	  "work": "0000000000000000",
	  "signature": "9e8fd29c2271d201f902a69a451932e0afcdc04e74137b66cb555b88738748a0f2cd6550b7a8df0d42960024c02a8de095e744c971b54dbbae2f599c62b90c0c"
	}
`

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
	  "previous": "d9eefd96ecfdd7e8cd985c4f32081a1eb5bda674c5280780485388f1cf01dfb8",
	  "link": "b7902600dfc79387b2601edc347b854d55d6b31142e324a4e54ff00a4c519c91",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "uuDnSNBQPHKpvU3xtss8bEgGpQWs8MaaHi55C25hd02lxWU7FnazE4iOVVR1g8Acb1AK1Uft9O9dLJuwbmY9VIHZqtYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBlZTBlYzI0ZTdlZjRlMTJmZTFiYmI2MTUwMDVkODcyY2Y5NWYwZDJiYjYyMTAzZDVhNDhhYzdlMDdkZjQ4ODI5",
	  "povHeight": 0,
	  "timestamp": 1588045334,
	  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
	  "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
	  "work": "0000000000000000",
	  "signature": "9c3caebfdbd39ec6396537596fee881803f608a758de9e30f9352ff58488ce62f905d975987dd174659c011a9ed7da54192a3264cbf90e29eed5e4e8ace69506"
	}
`

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
	  "link": "36c65df28de6de1184c06d10d5f9ce7572add49127726e9b0a3cef0b683b8bc2",
	  "message": "0000000000000000000000000000000000000000000000000000000000000000",
	  "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAO5rKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABep6hu0FA8cqm9TfG2yzxsSAalBazwxpoeLnkLbmF3TaXFZTsWdrMTiI5VVHWDwBxvUArVR+30710sm7BuZj1Ugdmq1gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBlZTBlYzI0ZTdlZjRlMTJmZTFiYmI2MTUwMDVkODcyY2Y5NWYwZDJiYjYyMTAzZDVhNDhhYzdlMDdkZjQ4ODI5",
	  "povHeight": 0,
	  "timestamp": 1588045335,
	  "extra": "a1252f091bc33fef51bac7a32459f0fc1c1405bdbc3516f3ee1f2d20d9ee456d",
	  "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
	  "work": "0000000000000000",
	  "signature": "3443bc0c926cdd4c69581621e782e22123dd49072dbe4568c704d94b6fedf3dca863624dec3cff6edebce2278aa542a889e2793b32a5df92909bf83f2cff2e00"
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
