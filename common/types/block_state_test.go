package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

var testBlk = `{
      "type": "state",
      "token":"991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "1456778",
      "previous": "247230c7377a661e57d51b17b527198ed52392fb8b99367a234d28ccc378eb05",
      "link": "7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "sender": "IjE1ODExMTEwMDAwMCI=",
      "receiver": "IjE1ODExMTEwMDAwMCI=",
      "message": "1235650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
	  "data": "DCI4Tg==",
	  "quota": 12345612,
	  "timestamp": 783474523,
      "extra": "1235650e78d297037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "representative": "qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "work": "3c82cc724905ee00",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600"
	}
	`

func TestMarshalStateBlock(t *testing.T) {
	blk := StateBlock{}
	//fmt.Println(blk)
	bytes, err := json.Marshal(blk)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
}

func TestUnmarshalStateBlock(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	addr, err := HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if err != nil {
		t.Fatal(err)
	}
	if addr != b.Address {
		t.Fatal("addr != address")
	}

	if !b.Balance.Equal(StringToBalance("1456778")) {
		t.Fatal("balance error")
	}
}

func TestStateBlock(t *testing.T) {
	b, err := NewBlock(State)
	if err != nil {
		t.Fatal(err)
	}
	if sb, ok := b.(*StateBlock); ok {
		sb.Balance = Balance{big.NewInt(123454)}
		sb.Address, _ = HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
		sb.Token, _ = NewHash("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")
		bytes, err := json.Marshal(&sb)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bytes))
	} else {
		t.Fatal("new state block error")
	}
}

func TestStateBlock_Serialize(t *testing.T) {
	//data := []byte{12, 34, 56, 78}
	//fmt.Println(base64.StdEncoding.EncodeToString(data))

	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.Balance)

	buff, err := b.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(buff))
	var b2 StateBlock
	if err = b2.Deserialize(buff); err != nil {
		t.Fatal(err)
	}
	t.Log(b2)

	bytes, _ := json.Marshal(&b2)
	t.Log(string(bytes))

	if !b2.Balance.Equal(b.Balance) {
		t.Fatal("balance error")
	}
	if hex.EncodeToString(b2.Data) != hex.EncodeToString(b.Data) {
		t.Fatal("data error")
	}
}

func TestStateBlock_Clone(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.Balance)
	b1 := b.Clone()

	if reflect.DeepEqual(b, b1) {
		t.Fatal("invalid clone")
	}
	if b.String() != b1.String() {
		t.Fatal("invalid clone ", b.String(), b1.String())
	}
}
