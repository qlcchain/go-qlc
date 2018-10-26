package genesis

import (
	"github.com/qlcchain/go-qlc/common/types"
)

type Genesis struct {
	Block types.StateBlock
}

var (
	test_genesis_data = `{
			"type": "state",
			"address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
			"previousHash": "991cf191094c40f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728948",
			"representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
			"balance": "6000000000000000000",
			"link":"5670d55b612313711ce76e85ac81d80478d5d04d4cebc399aeace07ae05dd299",
			"signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
			"token":"991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949",
			"work": "3c82cc724905ee00"
		}
		`
)

func Get() (*Genesis, error) {
	gen, err := types.ParseStateBlock([]byte(test_genesis_data))
	if err != nil {
		return nil, err
	}
	return &Genesis{*gen}, nil
}
