package types

import "math/big"

type BigNum struct {
	big.Int
}

func NewBigNumFromInt(x int64) *BigNum {
	bn := new(BigNum)
	bn.Int.SetInt64(x)
	return bn
}

func NewBigNumFromUint(x uint64) *BigNum {
	bn := new(BigNum)
	bn.Int.SetUint64(x)
	return bn
}

func NewBigNumFromBigInt(int *big.Int) *BigNum {
	bn := new(BigNum)
	bn.Int.SetBytes(int.Bytes())
	return bn
}

func (bn *BigNum) ToBigInt() *big.Int {
	return new(big.Int).SetBytes(bn.Int.Bytes())
}

func (bn *BigNum) Cmp(y *BigNum) int {
	return bn.Int.Cmp(&y.Int)
}

func (bn *BigNum) CmpBigInt(y *big.Int) int {
	return bn.Int.Cmp(y)
}

func (bn *BigNum) Add(x *BigNum, y *BigNum) *BigNum {
	bn.Int.Add(&x.Int, &y.Int)
	return bn
}

func (bn *BigNum) AddBigInt(x *big.Int, y *big.Int) *BigNum {
	bn.Int.Add(x, y)
	return bn
}

func (bn *BigNum) Sub(x *BigNum, y *BigNum) *BigNum {
	bn.Int.Sub(&x.Int, &y.Int)
	return bn
}

func (bn *BigNum) SubBigInt(x *big.Int, y *big.Int) *BigNum {
	bn.Int.Sub(x, y)
	return bn
}

func (bn *BigNum) Div(x *BigNum, y *BigNum) *BigNum {
	bn.Int.Div(&x.Int, &y.Int)
	return bn
}

func (bn *BigNum) DivBigInt(x *big.Int, y *big.Int) *BigNum {
	bn.Int.Div(x, y)
	return bn
}

func (bn *BigNum) Mul(x *BigNum, y *BigNum) *BigNum {
	bn.Int.Mul(&x.Int, &y.Int)
	return bn
}

func (bn *BigNum) MulBigInt(x *big.Int, y *big.Int) *BigNum {
	bn.Int.Mul(x, y)
	return bn
}

func (bn *BigNum) Copy() *BigNum {
	copyBN := new(BigNum)
	copyBN.Int.SetBytes(bn.Int.Bytes())
	return copyBN
}

//ExtensionType implements Extension.ExtensionType interface
func (bn *BigNum) ExtensionType() int8 {
	return BigNumExtensionType
}

//ExtensionType implements Extension.Len interface
func (bn *BigNum) Len() int { return len(bn.Int.Bytes()) }

//ExtensionType implements Extension.UnmarshalBinary interface
func (bn *BigNum) MarshalBinaryTo(text []byte) error {
	copy(text, bn.Int.Bytes())
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (bn *BigNum) UnmarshalBinary(text []byte) error {
	bn.Int.SetBytes(text)
	return nil
}
