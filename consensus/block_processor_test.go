package consensus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

var (
	l *ledger.Ledger
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "consensus")
	fmt.Println(dir)
	l = ledger.NewLedger(dir)

	return func(t *testing.T) {
		//err := l.db.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//l.CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func parseBlocks(t *testing.T, filename string) (blocks []types.StateBlock) {
	type fileStruct struct {
		Blocks []json.RawMessage `json:"blocks"`
	}
	data, err := ioutil.ReadFile(filename)
	var file fileStruct
	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	for _, data := range file.Blocks {
		var values map[string]interface{}
		json.Unmarshal(data, &values)
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		var blk types.StateBlock
		if err := json.Unmarshal(data, &blk); err != nil {
			t.Fatal(err)
		} else {
			blocks = append(blocks, blk)
		}
	}
	return
}

func TestDpos_BlockBasicInfoCheck(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	dpos := NewDpos(nil, l)
	blks := parseBlocks(t, "testdata/blocks.json")

	err := dpos.bp.addBasicInfo(blks[0], false)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blks[1:] {
		r := dpos.bp.process_receive_one(&b)
		if r == other {
			t.Fatal(r)
		}
		t.Log(r)
	}
	checkInfo(t, l)
}

func checkInfo(t *testing.T, l *ledger.Ledger) {
	blocks, _ := l.GetBlocks()
	fmt.Println("blocks: ")
	for _, b := range blocks {
		fmt.Println(*b)
	}

	fmt.Println("frontiers:")
	fs, _ := l.GetFrontiers()
	for _, f := range fs {
		fmt.Println(f)
	}
	fmt.Println("account: ")
	a, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	ac, _ := l.GetAccountMeta(a)
	fmt.Println(" account1,", ac.Address)
	for _, t := range ac.Tokens {
		fmt.Println("  token, ", t)
	}
	a, _ = types.HexToAddress("qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i")
	ac, _ = l.GetAccountMeta(a)
	fmt.Println(" account2,", ac.Address)
	for _, t := range ac.Tokens {
		fmt.Println("  token, ", t)
	}
}
