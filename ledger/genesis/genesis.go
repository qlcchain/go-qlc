package genesis

import (
	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
)

type Genesis struct {
	Block         types.StateBlock
	WorkThreshold uint64
}

var (
	Test_genesis_data = `{
			"type": "state",
			"address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
			"previousHash": "0000000000000000000000000000000000000000000000000000000000000000",
			"representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
			"balance": "600000",
			"link":"2845d6627542d95a0a2a54b0dbb6217e384304baa8ded8664bf258d7b9469fe0",
			"signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
			"token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
			"work": "0b61e26528833017" 
		}
		`

	Test_genesis_data2 = `{
			"type": "state",
			"address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
			"previousHash": "0000000000000000000000000000000000000000000000000000000000000000",
			"representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
			"balance": "10000000",
			"link":"5670d55b612313711ce76e80ac81d80478d5d04d4cebc399aeace07ae05dd299",
			"signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
			"token":"991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949",
			"work": "0b61e26528833017" 
		}
		`
)

func Get(genStr string) (*Genesis, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var gen types.StateBlock
	err := json.Unmarshal([]byte(genStr), &gen)
	if err != nil {
		return nil, err
	}
	return &Genesis{gen, uint64(0x0fffffc000000000)}, nil
}
