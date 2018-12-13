package types

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common/types/internal/uint128"
	"testing"

	"github.com/json-iterator/go"
)

func TestMarshalStateBlock(t *testing.T) {
	blk := StateBlock{}
	//fmt.Println(blk)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	bytes, err := json.Marshal(blk)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
}

func TestUnmarshalStateBlock(t *testing.T) {
	test_blk := `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "247230c7377a661e57d51b17b527198ed52392fb8b99367a234d28ccc378eb05",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "00000000000000000000000000015f90",
      "link":"7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949",
      "work": "3c82cc724905ee00"
	}
	`
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var b StateBlock
	err := json.Unmarshal([]byte(test_blk), &b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(b)
}

func TestStateBlock(t *testing.T) {
	b, err := NewBlock(State)
	if err != nil {
		t.Fatal(err)
	}
	if sb, ok := b.(*StateBlock); ok {
		sb.Balance = Balance(uint128.Uint128{Lo: 120000000})
		sb.Address, _ = HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
		sb.Token, _ = NewHash("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")
		bytes, err := jsoniter.Marshal(&sb)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bytes))
	} else {
		t.Fatal("new state block error")
	}
}
