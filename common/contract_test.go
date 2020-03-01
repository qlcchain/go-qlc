package common

import "testing"

func TestOracleStringToType(t *testing.T) {
	src := []string{"email", "weChat", "invalid"}
	expect := []uint32{OracleTypeEmail, OracleTypeWeChat, OracleTypeInvalid}

	for i, s := range src {
		if OracleStringToType(s) != expect[i] {
			t.Fatal(s)
		}
	}
}

func TestOracleTypeToString(t *testing.T) {
	src := []uint32{OracleTypeEmail, OracleTypeWeChat, OracleTypeInvalid}
	expect := []string{"email", "weChat", "invalid"}

	for i, s := range src {
		if OracleTypeToString(s) != expect[i] {
			t.Fatal(s)
		}
	}
}
