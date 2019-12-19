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
		  "address": "qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s",
		  "balance": 10000000000000,
		  "vote": 0,
		  "network": 0,
		  "storage": 0,
		  "oracle": 0,
		  "previous": "0000000000000000000000000000000000000000000000000000000000000000",
		  "link": "bd13932faa9afd7536d304d1f2d85ab72457c4eb794b2db0b0bd4649f2745ac9",
		  "message": "0000000000000000000000000000000000000000000000000000000000000000",
		  "povHeight": 0,
		  "timestamp": 1574910154,
		  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
		  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
		  "work": "0000000000519537",
		  "signature": "82fcccd423d36e5972eac210cff435d0eb4a0629b2ff2fdc3bd9621cb3cfe50be1293eb987f91fd140d85236598f6a33322eb0dba3453f2844eb3cb2e9a4140d"
		}`

	send = `
		{
		  "type": "ContractSend",
		  "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
		  "address": "qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s",
		  "balance": 9999000000000,
		  "vote": 0,
		  "network": 0,
		  "storage": 0,
		  "oracle": 0,
		  "previous": "a114552e717e7dde45cdf9df2a93d23b24465bdc54cb220ed6eef49fbe427fa4",
		  "link": "b7902600dfc79387b2601edc347b854d55d6b31142e324a4e54ff00a4c519c91",
		  "message": "0000000000000000000000000000000000000000000000000000000000000000",
		  "data": "uuDnSCtW8xxOy1WvkW0yf39W6zYtdKv0nM3d1mOOMQHhQfWD0FA8cqm9TfG2yzxsSAalBazwxpoeLnkLbmF3TaXFZTsAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEA4NmU3MmRmZDNmNTI0NDZhNDRiMmFjNmQ3ZWUxMzk3MmY3YzQwY2U2NzRlZGI0MTA4MDE4MGNjMjg0MjBkOWQ4",
		  "povHeight": 0,
		  "timestamp": 1574910157,
		  "extra": "0000000000000000000000000000000000000000000000000000000000000000",
		  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
		  "work": "00000000004698e1",
		  "signature": "7fff727d5a46757da1fd570eddc34304875bd213e034ae65dbbec97c6c48d7ff38bc2603c3def3f3ec97ebe12204c5c84733078be887b2f51af07a825d4ab700"
		}`

	rece = `
		{
		  "type": "ContractReward",
		  "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
		  "address": "qlc_1ctpyeg6xktooyaptemzhxdgpfjfgkozb98fuqd895jj19in5xe5rxnxrc34",
		  "balance": 0,
		  "vote": 1000000000,
		  "network": 0,
		  "storage": 0,
		  "oracle": 0,
		  "previous": "0000000000000000000000000000000000000000000000000000000000000000",
		  "link": "6fead67c0bb8e85eee119614076f9e52aee0339546da392436a5309b1747f292",
		  "message": "0000000000000000000000000000000000000000000000000000000000000000",
		  "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAO5rKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABd7GfNK1bzHE7LVa+RbTJ/f1brNi10q/Sczd3WY44xAeFB9YPQUDxyqb1N8bbLPGxIBqUFrPDGmh4ueQtuYXdNpcVlOwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEA4NmU3MmRmZDNmNTI0NDZhNDRiMmFjNmQ3ZWUxMzk3MmY3YzQwY2U2NzRlZGI0MTA4MDE4MGNjMjg0MjBkOWQ4",
		  "povHeight": 0,
		  "timestamp": 1574910159,
		  "extra": "546f100c16381e12599878767b185bb68691fe746939419b1d4d3a8c10d30685",
		  "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
		  "work": "00000000004059ad",
		  "signature": "8ff5b21c3492e570e918efd119ccb446fe3878fed40455ecb2b8773b1484b27c8f331f62cd57ba86825aae2366530b00e07abe0618e9410a601c19f0d3e82303"
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
