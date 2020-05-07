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

func TestPublicKeyTypeFromString(t *testing.T) {
	src := []string{"ed25519", "rsa4096", "invalid"}
	expect := []uint16{PublicKeyTypeED25519, PublicKeyTypeRSA4096, PublicKeyTypeInvalid}

	for i, s := range src {
		if PublicKeyTypeFromString(s) != expect[i] {
			t.Fatal(s)
		}
	}
}

func TestPublicKeyTypeToString(t *testing.T) {
	src := []uint16{PublicKeyTypeED25519, PublicKeyTypeRSA4096, PublicKeyTypeInvalid}
	expect := []string{"ed25519", "rsa4096", "invalid"}

	for i, s := range src {
		if PublicKeyTypeToString(s) != expect[i] {
			t.Fatal(s)
		}
	}
}

func TestPtmKeyBtypeFromString(t *testing.T) {
	src := []string{PtmKeyVBtypeStrDefault, PtmKeyVBtypeStrA2p, PtmKeyVBtypeStrDod, PtmKeyVBtypeStrCloud, PtmKeyVBtypeStrInvaild}
	expect := []uint16{PtmKeyVBtypeDefault, PtmKeyVBtypeA2p, PtmKeyVBtypeDod, PtmKeyVBtypeCloud, PtmKeyVBtypeInvaild}

	for i, s := range src {
		if PtmKeyBtypeFromString(s) != expect[i] {
			t.Fatal(s)
		}
	}
}

func TestPtmKeyBtypeToString(t *testing.T) {
	src := []uint16{PtmKeyVBtypeDefault, PtmKeyVBtypeA2p, PtmKeyVBtypeDod, PtmKeyVBtypeCloud, PtmKeyVBtypeInvaild}
	expect := []string{PtmKeyVBtypeStrDefault, PtmKeyVBtypeStrA2p, PtmKeyVBtypeStrDod, PtmKeyVBtypeStrCloud, PtmKeyVBtypeStrInvaild}

	for i, s := range src {
		if PtmKeyBtypeToString(s) != expect[i] {
			t.Fatal(s)
		}
	}
}
