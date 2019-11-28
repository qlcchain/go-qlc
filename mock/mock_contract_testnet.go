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
		"address": "qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s",
		"balance": 10000000000000,
		"vote": 0,
		"network": 0,
		"storage": 0,
		"oracle": 0,
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "36dfd648f428fb1232a929c640778052e3fe376c6ec482c6557c3babcceb5e94",
		"message": "0000000000000000000000000000000000000000000000000000000000000000",
		"povHeight": 0,
		"timestamp": 1574866017,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
		"work": "0000000000000000",
		"signature": "db4a18b65feff071ef4c1fd6f02a8965a023af1218cf3bfad8eff5f1ea90d62f755abbcf1ae9afe8b0b26940019a1a4952a583a8aee8b3d0877e8328ebeeb103"
		}`

	send = `{
		"type": "ContractSend",
		"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
		"address": "qlc_3n4i9jscmhcfy8ueph5eb15cc3fey55bn9jgh67pwrdqbpkwcsbu4iot7f7s",
		"balance": 9999000000000,
		"vote": 0,
		"network": 0,
		"storage": 0,
		"oracle": 0,
		"previous": "f9ffde70bebec1788a3a0c2088ef560278589a4cdfe73d0a2899cdcdeafcdfba",
		"link": "b7902600dfc79387b2601edc347b854d55d6b31142e324a4e54ff00a4c519c91",
		"message": "0000000000000000000000000000000000000000000000000000000000000000",
		"data": "uuDnSCtW8xxOy1WvkW0yf39W6zYtdKv0nM3d1mOOMQHhQfWD0FA8cqm9TfG2yzxsSAalBazwxpoeLnkLbmF3TaXFZTsAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAyYzIxNThhOThiNDFmNDBjYzk0YjRiNGU2NDlmYzZlYjUzOWRkZDA4OWYyN2Y0ODI1MTVkOWRmNmFkOTJjNzg4",
		"povHeight": 0,
		"timestamp": 1574866017,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
		"work": "0000000000000000",
		"signature": "433d66ab290755476aa2031110d4e71eed2b8b010b91143438ef9a805e7e79b94bd86221ee48ccc5f1e81a06a6179d2bf86739f4ce9b3656981417e27d9ac70a"
		}`

	rece = `{
		"type": "ContractReward",
		"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
		"address": "qlc_1ctpyeg6xktooyaptemzhxdgpfjfgkozb98fuqd895jj19in5xe5rxnxrc34",
		"balance": 0,
		"vote": 1000000000,
		"network": 0,
		"storage": 0,
		"oracle": 0,
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "a8168c30a68bbfe7b8a97f1bbf6fc5b7b964d4e1f3d0bbeb632eb993d8893abd",
		"message": "0000000000000000000000000000000000000000000000000000000000000000",
		"data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAO5rKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABd3o65K1bzHE7LVa+RbTJ/f1brNi10q/Sczd3WY44xAeFB9YPQUDxyqb1N8bbLPGxIBqUFrPDGmh4ueQtuYXdNpcVlOwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAyYzIxNThhOThiNDFmNDBjYzk0YjRiNGU2NDlmYzZlYjUzOWRkZDA4OWYyN2Y0ODI1MTVkOWRmNmFkOTJjNzg4",
		"povHeight": 0,
		"timestamp": 1574866018,
		"extra": "03f303691552a5d690e921530c29f375c70fc97b1a45c632c5e7acc9b933b873",
		"representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
		"work": "0000000000000000",
		"signature": "ce158652357c61807014e8987d93ce0e2085331f8319178c66dd0be180e37ec7ff3ab9d71ae39c2f98a322fd7040211cefb9f4b914407bb533bdc4464433d902"
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
