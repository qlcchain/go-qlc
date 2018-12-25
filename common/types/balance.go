package types

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/qlcchain/go-qlc/common/types/internal/uint128"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/tinylib/msgp/msgp"
	"math/big"
)

func init() {
	msgp.RegisterExtension(BalanceExtensionType, func() msgp.Extension { return new(Address) })
}

const (
	// BalanceSize represents the size of a balance in bytes.
	BalanceSize = 16
	// BalanceMaxPrecision  balance max precision
	BalanceMaxPrecision = 11
)

// BalanceComp compare
type BalanceComp byte

const (
	//BalanceCompEqual equal compare
	BalanceCompEqual BalanceComp = iota
	//BalanceCompBigger bigger compare
	BalanceCompBigger
	//BalanceCompSmaller smaller compare
	BalanceCompSmaller
	BalanceExtensionType = 100
)

var (
	// ZeroBalance zero
	ZeroBalance = Balance(uint128.Uint128{})

	//ErrBadBalanceSize bad size
	ErrBadBalanceSize = errors.New("balances should be 16 bytes in size")
)

// Balance of account
type Balance uint128.Uint128

type Amount uint128.Uint128

//ParseBalanceInts create balance from uint64
func ParseBalanceInts(hi uint64, lo uint64) Balance {
	return Balance(uint128.FromInts(hi, lo))
}

//ParseBalanceString create balance from string
func ParseBalanceString(b string) (Balance, error) {
	balance, err := uint128.FromString(b)
	if err != nil {
		return ZeroBalance, err
	}
	return Balance(balance), nil
}

// Bytes returns the binary representation of this Balance with the given
// endianness.
func (b Balance) Bytes(order binary.ByteOrder) []byte {
	bytes := uint128.Uint128(b).GetBytes()

	switch order {
	case binary.BigEndian:
		return bytes
	case binary.LittleEndian:
		return util.ReverseBytes(bytes)
	default:
		panic("unsupported byte order")
	}
}

// Equal reports whether this balance and the given balance are equal.
func (b Balance) Equal(b2 Balance) bool {
	return uint128.Uint128(b).Equal(uint128.Uint128(b2))
}

// Add balances add
func (b Balance) Add(n Balance) Balance {
	return Balance(uint128.Uint128(b).Add(uint128.Uint128(n)))
}

// Sub balances sub
func (b Balance) Sub(n Balance) Balance {
	return Balance(uint128.Uint128(b).Sub(uint128.Uint128(n)))
}

//Compare two balances
func (b Balance) Compare(n Balance) BalanceComp {
	res := uint128.Uint128(b).Compare(uint128.Uint128(n))
	switch res {
	case 1:
		return BalanceCompBigger
	case -1:
		return BalanceCompSmaller
	case 0:
		return BalanceCompEqual
	default:
		panic("unexpected comparison result")
	}
}

// BigInt convert balance to int
func (b Balance) BigInt() *big.Int {
	i := big.NewInt(0)
	i.SetBytes(b.Bytes(binary.BigEndian))
	return i
}

// String implements the fmt.Stringer interface. It returns the balance in Mqlc
// with maximum precision.
func (b Balance) String() string {
	return b.BigInt().String()
}

//ExtensionType implements Extension.ExtensionType interface
func (b *Balance) ExtensionType() int8 {
	return BalanceExtensionType
}

//ExtensionType implements Extension.Len interface
func (b *Balance) Len() int { return BalanceSize }

//ExtensionType implements Extension.UnmarshalBinary interface
func (b Balance) MarshalBinaryTo(text []byte) error {
	copy(text, b.Bytes(binary.BigEndian))
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (b *Balance) UnmarshalBinary(text []byte) error {
	if len(text) != BalanceSize {
		return ErrBadBalanceSize
	}

	*b = Balance(uint128.FromBytes(text))
	return nil

}

// MarshalText implements the encoding.TextMarshaler interface.
func (b Balance) MarshalText() ([]byte, error) {
	//return []byte(strings.TrimLeft(hex.EncodeToString(b.Bytes(binary.BigEndian)), "0")), nil
	return []byte(hex.EncodeToString(b.Bytes(binary.BigEndian))), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (b *Balance) UnmarshalText(text []byte) error {
	//l := len(text)
	//if l%2 != 0 {
	//	text = util.LeftPadBytes(text, 1)
	//}
	balance, err := ParseBalanceString(string(text))
	if err != nil {
		return err
	}
	*b = balance
	return nil
}
