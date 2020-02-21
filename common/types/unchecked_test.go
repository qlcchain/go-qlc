package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUnchecked_Serialize(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
	unchecked := &Unchecked{Block: &b, Kind: Synchronized}
	s, err := unchecked.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
	u2 := new(Unchecked)
	if err := u2.Deserialize(s); err != nil {
		t.Fatal(err)
	}
	t.Log(u2)
	if u2.Block.GetHash() != unchecked.Block.GetHash() || u2.Kind != unchecked.Kind {
		t.Fatal()
	}
}
