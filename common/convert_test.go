package common

import (
	"github.com/qlcchain/go-qlc/common/types"
	"testing"
)

func TestBalanceToRaw(t *testing.T) {
	srcUint := []string{"qlc", "Kqlc", "QLC", "MQLC"}
	expect := []types.Balance{types.NewBalance(1), types.NewBalance(1e5), types.NewBalance(1e8), types.NewBalance(1e11)}
	ori := types.NewBalance(1)

	for i, s := range srcUint {
		exp, err := BalanceToRaw(ori, s)
		if err != nil || exp.Compare(expect[i]) != types.BalanceCompEqual {
			t.Fatal(err)
		}
	}

	_, err := BalanceToRaw(ori, "test")
	if err == nil {
		t.Fatal()
	}
}

func TestRawToBalance(t *testing.T) {
	srcUint := []string{"qlc", "Kqlc", "QLC", "MQLC"}
	expect := []types.Balance{types.NewBalance(1e11), types.NewBalance(1e6), types.NewBalance(1e3), types.NewBalance(1)}
	ori := types.NewBalance(1e11)

	for i, s := range srcUint {
		exp, err := RawToBalance(ori, s)
		if err != nil || exp.Compare(expect[i]) != types.BalanceCompEqual {
			t.Fatal(err)
		}
	}

	_, err := RawToBalance(ori, "test")
	if err == nil {
		t.Fatal()
	}
}
