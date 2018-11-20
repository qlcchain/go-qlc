package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMarshalStateBlock(t *testing.T) {
	blk := StateBlock{}
	fmt.Println(blk)
	bytes, err := json.Marshal(blk)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
}
