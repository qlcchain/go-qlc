package types

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/tinylib/msgp/msgp"

	"github.com/qlcchain/go-qlc/common/types/internal/uint128"
	"github.com/qlcchain/go-qlc/common/types/internal/util"
	"github.com/shopspring/decimal"
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
	units = map[string]decimal.Decimal{
		"raw":  decimal.New(1, 0),
		"qlc":  decimal.New(1, 2),
		"Kqlc": decimal.New(1, 5),
		"Mqlc": decimal.New(1, 8),
		"Gqlc": decimal.New(1, 11),
	}

	// ZeroBalance zero
	ZeroBalance = Balance(uint128.Uint128{})

	//ErrBadBalanceSize bad size
	ErrBadBalanceSize = errors.New("balances should be 16 bytes in size")
)

// Balance of account
type Balance uint128.Uint128

// ParseBalance parses the given balance string.
func ParseBalance(s string, unit string) (Balance, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return ZeroBalance, err
	}

	// zero is a special case
	if d.Equals(decimal.Zero) {
		return ZeroBalance, nil
	}

	d = d.Mul(units[unit])
	c := d.Coefficient()
	f := bigPow(10, int64(d.Exponent()))
	i := c.Mul(c, f)

	bytes := i.Bytes()
	balanceBytes := make([]byte, BalanceSize)
	copy(balanceBytes[len(balanceBytes)-len(bytes):], bytes)

	var balance Balance
	if err := balance.UnmarshalBinary(balanceBytes); err != nil {
		return ZeroBalance, err
	}

	return balance, nil
}

//ParseBalanceInts create balance from uint64
func ParseBalanceInts(hi uint64, lo uint64) Balance {
	return Balance(uint128.FromInts(hi, lo))
}

//ParseBalanceString create balance from string
func ParseBalanceString(b string) Balance {
	balance, _ := uint128.FromString(b)
	return Balance(balance)
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

// UnitString returns a decimal representation of this uint128 converted to the
// given unit.
func (b Balance) UnitString(unit string, precision int32) string {
	d := decimal.NewFromBigInt(b.BigInt(), 0)
	return d.DivRound(units[unit], BalanceMaxPrecision).Truncate(precision).String()
}

// String implements the fmt.Stringer interface. It returns the balance in Mxrb
// with maximum precision.
func (b Balance) String() string {
	return b.UnitString("Mqlc", BalanceMaxPrecision)
}

func bigPow(base int64, exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(base), big.NewInt(exp), nil)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (b Balance) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (b *Balance) UnmarshalText(text []byte) error {
	balance, err := ParseBalance(string(text), "Mqlc")
	if err != nil {
		return err
	}

	*b = balance
	return nil
}

//ExtensionType implements Extension.ExtensionType interface
func (b *Balance) ExtensionType() int8 {
	return BalanceExtensionType
}

//ExtensionType implements Extension.Len interface
func (b *Balance) Len() int { return BalanceSize }

//ExtensionType implements Extension.UnmarshalBinary interface
func (b *Balance) MarshalBinaryTo(text []byte) error {
	copy(text, b.Bytes(binary.LittleEndian))
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

//MarshalJSON implements json.Marshaler interface
func (b *Balance) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}
