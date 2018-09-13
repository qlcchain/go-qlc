package types

import (
	"fmt"
	"testing"
)

var (
	b1Units = map[string]string{
		"raw":  "60000000000000000",
		"qlc":  "600000000000000",
		"Kqlc": "600000000000",
		"Mqlc": "600000000",
		"Gqlc": "600000",
	}
	b2Units = map[string]string{
		"raw":  "1",
		"qlc":  "0.01",
		"Kqlc": "0.00001",
		"Mqlc": "0.00000001",
		"Gqlc": "0.00000000001",
	}
	b1TruncatedUnits = map[string]string{
		"raw":  "60000000000000000",
		"qlc":  "600000000000000",
		"Kqlc": "600000000000",
		"Mqlc": "600000000",
		"Gqlc": "600000",
	}
	compare = func(b Balance, m map[string]string, p int32) error {
		for unit, s := range m {
			res := b.UnitString(unit, p)
			if res != s {
				return fmt.Errorf("(%s) expected: %s, got: %s", unit, s, res)
			}
		}

		return nil
	}
)

func TestParseBalance(t *testing.T) {
	_, err := ParseBalance("600000000.00000000", "Mqlc")
	_, e := ParseBalance("1", "raw")
	if err != nil {
		t.Errorf("ParseBalance err for b1 %s", err)
	}
	if e != nil {
		t.Errorf("ParseBalance err for b2 %s", e)
	}

}

func TestQlcBalance(t *testing.T) {
	b1, err := ParseBalance("600000000.00000000", "Mqlc")
	b2, e := ParseBalance("1", "raw")
	if err != nil {
		t.Errorf("ParseBalance err for b1 %s", err)
	}
	if e != nil {
		t.Errorf("ParseBalance err for b2 %s", e)
	}
	compare(b1, b1Units, BalanceMaxPrecision)
	compare(b1, b1TruncatedUnits, 6)
	compare(b2, b2Units, BalanceMaxPrecision)

	if b1.String() != b1Units["Mqlc"] {
		t.Errorf("unexpected fmt.Stringer result")
	}

	for unit, s := range b1Units {
		b, err := ParseBalance(s, unit)
		if err != nil {
			t.Error(err)
			continue
		}

		if !b.Equal(b1) {
			t.Errorf("(%s) expected: %s, got: %s\n", unit, b1.UnitString(unit, BalanceMaxPrecision), b.UnitString(unit, BalanceMaxPrecision))
		}
	}

	for unit, s := range b1TruncatedUnits {
		b, err := ParseBalance(s, unit)
		if err != nil {
			t.Error(err)
			continue
		}

		res := b.UnitString(unit, 6)
		if res != s {
			t.Errorf("(%s) expected: %s, got: %s\n", unit, s, res)
		}
	}
}
