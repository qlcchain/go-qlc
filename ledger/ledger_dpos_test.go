package ledger

import "testing"

func TestLedger_GetLastGapPovHeight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	height := l.GetLastGapPovHeight()
	if height != 0 {
		t.Fatal()
	}

	err := l.SetLastGapPovHeight(100)
	if err != nil {
		t.Fatal(err)
	}

	height = l.GetLastGapPovHeight()
	if height != 100 {
		t.Fatal(height)
	}
}
