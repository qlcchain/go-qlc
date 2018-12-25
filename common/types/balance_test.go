package types

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types/internal/uint128"
	"github.com/qlcchain/go-qlc/common/util"
	"testing"
)

func TestParseBalance(t *testing.T) {
	_, err := ParseBalanceString("6000000000")

	if err != nil {
		t.Errorf("ParseBalance err for b1 %s", err)
	}
}

func TestBalance_MarshalJSON(t *testing.T) {
	b1 := Balance{Hi: 0, Lo: 1000}
	bs, err := json.Marshal(&b1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bs))
	b2 := Balance{}
	err = json.Unmarshal(bs, &b2)
	if err != nil {
		t.Fatal(err)
	}
	if b1.Compare(b2) != BalanceCompEqual {
		t.Fatal("b1!=b2")
	}
	t.Log(b1.BigInt())
}

func TestParseBalanceInts(t *testing.T) {
	b1 := Balance(uint128.FromInts(0, 1041321))
	bytes, err := json.Marshal(b1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes), b1.BigInt())
}

func TestParseBalanceString(t *testing.T) {
	s := "600000000000000000"
	text := []byte(s)
	l := len(text)
	if l%2 != 0 {
		text = util.LeftPadBytes(text, 1)
	}

	b, err := ParseBalanceString(s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.String())
}
