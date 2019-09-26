package process

import (
	"fmt"
	"testing"

	"github.com/qlcchain/go-qlc/mock"
)

func TestProcess_SyncBlockProcess(t *testing.T) {
	teardownTestCase, _, lv := setupTestCase(t)
	defer teardownTestCase(t)

	var bc, _ = mock.BlockChain()
	if err := lv.BlockSyncProcess(bc[0]); err != nil {
		t.Fatal(err)
	}
	t.Log("bc hash", bc[0].GetHash())
	for i, b := range bc[1:] {
		fmt.Println(i + 1)
		fmt.Println("bc.previous", b.GetPrevious())
		if r, err := lv.BlockSyncCheck(b); r != Progress || err != nil {
			t.Fatal(r, err)
		}
		if err := lv.BlockSyncProcess(b); err != nil {
			t.Fatal(Other, err)
		}
		if err := lv.BlockSyncDone(); err != nil {
			t.Fatal(err)
		}
	}
}
