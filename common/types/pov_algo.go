package types

import (
	"strings"
)

type PovAlgoType uint32

const (
	ALGO_VERSION_MASK = PovAlgoType(255 << 8)

	ALGO_SHA256D = PovAlgoType(0 << 8)
	ALGO_SCRYPT  = PovAlgoType(1 << 8)
	ALGO_NIST5   = PovAlgoType(2 << 8)
	ALGO_LYRA2Z  = PovAlgoType(3 << 8)
	ALGO_X11     = PovAlgoType(4 << 8)
	ALGO_X16R    = PovAlgoType(5 << 8)
	ALGO_HYBRID  = PovAlgoType(254 << 8)
	ALGO_UNKNOWN = PovAlgoType(255 << 8)
)

func (a PovAlgoType) String() string {
	switch a {
	case ALGO_SHA256D:
		return "SHA256D"
	case ALGO_SCRYPT:
		return "SCRYPT"
	case ALGO_NIST5:
		return "NIST5"
	case ALGO_LYRA2Z:
		return "LYRA2Z"
	case ALGO_X11:
		return "X11"
	case ALGO_X16R:
		return "X16R"
	case ALGO_HYBRID:
		return "HYBRID"
	}

	return "UNKNOWN"
}

func NewPoVHashAlgoFromStr(name string) PovAlgoType {
	upName := strings.ToUpper(name)
	switch upName {
	case "SHA256D":
		return ALGO_SHA256D
	case "SCRYPT":
		return ALGO_SCRYPT
	case "NIST5":
		return ALGO_NIST5
	case "LYRA2Z":
		return ALGO_LYRA2Z
	case "X11":
		return ALGO_X11
	case "X16R":
		return ALGO_X16R
	case "HYBRID":
		return ALGO_HYBRID
	}

	return ALGO_UNKNOWN
}
