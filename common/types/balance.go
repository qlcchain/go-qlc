package types

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/tinylib/msgp/msgp"
)

func init() {
	msgp.RegisterExtension(BalanceExtensionType, func() msgp.Extension { return new(Balance) })
}

const (
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
)

var (
	// ZeroBalance zero
	ZeroBalance = Balance{big.NewInt(0)}
)

// Balance of account
type Balance struct {
	*big.Int
}

//StringToBalance create balance from string
func StringToBalance(b string) Balance {
	t := new(big.Int)
	t.SetString(b, 10)
	return Balance{t}
}

// BytesToBalance create balance from byte slice
func BytesToBalance(b []byte) Balance {
	t := new(big.Int)
	t.SetBytes(b)
	return Balance{t}
}

// Bytes returns the binary representation of this Balance with the given
// endianness.
func (b Balance) Bytes() []byte {
	if b.Int == nil {
		return ZeroBalance.Bytes()
	}
	return b.Int.Bytes()
}

// Equal reports whether this balance and the given balance are equal.
func (b Balance) Equal(b2 Balance) bool {
	return b.Int != nil && b2.Int != nil && b.Int.Cmp(b2.Int) == 0
}

// Add balances add
func (b Balance) Add(n Balance) Balance {
	if u, i := util.SafeAdd(b.Uint64(), n.Uint64()); !i {
		r := new(big.Int).SetUint64(u)
		return Balance{r}
	}
	return ZeroBalance
}

// Sub balances sub
func (b Balance) Sub(n Balance) Balance {
	if u, b := util.SafeSub(b.Uint64(), n.Uint64()); !b {
		return Balance{new(big.Int).SetUint64(u)}
	}

	return ZeroBalance
}

// Div balances div
func (b Balance) Div(n int64) (Balance, error) {
	if n == 0 {
		return ZeroBalance, errors.New("n should not be zero")
	}
	div := b.Int64() / n
	b1 := new(big.Int).SetInt64(div)
	return Balance{b1}, nil
}

// Mul balances mul
func (b Balance) Mul(n int64) Balance {
	if u, b := util.SafeMul(b.Uint64(), uint64(n)); !b {
		return Balance{new(big.Int).SetUint64(u)}
	}
	return ZeroBalance
}

//Compare two balances
func (b Balance) Compare(n Balance) BalanceComp {
	res := b.Int.Cmp(n.Int)
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

//ExtensionType implements Extension.ExtensionType interface
func (b *Balance) ExtensionType() int8 {
	return BalanceExtensionType
}

//ExtensionType implements Extension.Len interface
func (b *Balance) Len() int { return len(b.Bytes()) }

//ExtensionType implements Extension.UnmarshalBinary interface
func (b *Balance) MarshalBinaryTo(text []byte) error {
	copy(text, b.Bytes())
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (b *Balance) UnmarshalBinary(text []byte) error {
	i := new(big.Int)
	i.SetBytes(text)
	*b = Balance{i}

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (b Balance) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (b *Balance) UnmarshalText(text []byte) error {
	s := util.TrimQuotes(string(text))
	_, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	balance := StringToBalance(s)
	*b = balance
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (b *Balance) MarshalJSON() ([]byte, error) {
	s := ""

	if b.Int == nil {
		s = fmt.Sprintf("\"%s\"", ZeroBalance)
	} else {
		s = fmt.Sprintf("\"%s\"", b.String())
	}
	return []byte(s), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (b *Balance) UnmarshalJSON(text []byte) error {
	return b.UnmarshalText(text)
}

// IsZero check balance is zero
func (b *Balance) IsZero() bool {
	return b.Equal(ZeroBalance)
}
