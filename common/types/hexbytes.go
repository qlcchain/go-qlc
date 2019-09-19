package types

import (
	"encoding/hex"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/tinylib/msgp/msgp"
)

type HexBytes []byte

func init() {
	msgp.RegisterExtension(HexBytesExtensionType, func() msgp.Extension { return new(HexBytes) })
}

func NewHexBytesFromHex(hexStr string) HexBytes {
	hbs := HexBytes{}
	err := hbs.SetHex(hexStr)
	if err != nil {
		return nil
	}
	return hbs
}

func NewHexBytesFromData(data []byte) HexBytes {
	hbs := make([]byte, len(data))
	copy(hbs[:], data)
	return hbs
}

func (hbs *HexBytes) SetHex(hexString string) error {
	s := util.TrimQuotes(hexString)
	size := hex.DecodedLen(len(s))

	hbs2 := make([]byte, size)
	if _, err := hex.Decode(hbs2[:], []byte(s)); err != nil {
		return err
	}

	*hbs = hbs2
	return nil
}

func (hbs HexBytes) String() string {
	return hex.EncodeToString(hbs[:])
}

//ExtensionType implements Extension.ExtensionType interface
func (hbs HexBytes) ExtensionType() int8 {
	return HexBytesExtensionType
}

//ExtensionType implements Extension.Len interface
func (hbs HexBytes) Len() int {
	return len(hbs[:])
}

//ExtensionType implements Extension.MarshalBinaryTo interface
func (hbs HexBytes) MarshalBinaryTo(text []byte) error {
	copy(text, hbs[:])
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (hbs *HexBytes) UnmarshalBinary(text []byte) error {
	*hbs = make([]byte, len(text))
	copy((*hbs)[:], text)
	return nil
}

//UnmarshalText implements encoding.TextUnmarshaler
func (hbs *HexBytes) UnmarshalText(text []byte) error {
	s := util.TrimQuotes(string(text))
	err := hbs.SetHex(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface.
func (hbs HexBytes) MarshalText() (text []byte, err error) {
	return []byte(hbs.String()), nil
}

func (hbs HexBytes) Bytes() []byte {
	return hbs[:]
}
