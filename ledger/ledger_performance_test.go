package ledger

import (
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestLedger_AddOrUpdatePerformance(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	for i := 0; i < 20; i++ {
		pt := types.NewPerformanceTime()
		pt.Hash = mock.Hash()
		err := l.AddOrUpdatePerformance(pt)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := l.PerformanceTimes(func(p *types.PerformanceTime) {
		t.Logf("%s", p.String())
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddOrUpdatePerformance2(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pt := types.NewPerformanceTime()
	h := mock.Hash()
	pt.Hash = h

	err := l.AddOrUpdatePerformance(pt)
	if err != nil {
		t.Fatal(err)
	}

	t3 := time.Now().AddDate(0, 0, 1).Unix()
	pt.T3 = t3

	err = l.AddOrUpdatePerformance(pt)
	if err != nil {
		t.Fatal(err)
	}

	exist, err := l.IsPerformanceTimeExist(h)
	if err != nil {
		t.Fatal(err)
	}
	if !exist {
		t.Fatal("error exist")
	}

	pt2, err := l.GetPerformanceTime(h)
	if err != nil {
		t.Fatal(err)
	}

	if pt2.T3 != t3 {
		t.Fatal("err t3z")
	}

	b, err := l.IsPerformanceTimeExist(types.ZeroHash)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatal("error exist2")
	}

}
