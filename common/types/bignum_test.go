package types

import (
	"math/rand"
	"testing"
)

func TestBigNum_Calculation(t *testing.T) {
	bn1 := NewBigNumFromInt(rand.Int63())
	bn11 := NewBigNumFromBigInt(bn1.ToBigInt())
	if bn1.Cmp(bn11) != 0 {
		t.Fatalf("exp: %v, act: %v", bn1, bn11)
	}

	bn2 := NewBigNumFromUint(uint64(rand.Int63()))
	bn21 := NewBigNumFromBigInt(bn2.ToBigInt())
	if bn2.CmpBigInt(bn21.ToBigInt()) != 0 {
		t.Fatalf("exp: %v, act: %v", bn2, bn21)
	}

	bn3 := new(BigNum).Add(bn1, bn2)
	bi3 := new(BigNum).AddBigInt(bn1.ToBigInt(), bn2.ToBigInt())
	if bi3.Cmp(bn3) != 0 {
		t.Fatalf("exp: %v, act: %v", bi3, bn3)
	}

	bn4 := new(BigNum).Sub(bn1, bn2)
	bi4 := new(BigNum).SubBigInt(bn1.ToBigInt(), bn2.ToBigInt())
	if bi4.Cmp(bn4) != 0 {
		t.Fatalf("exp: %v, act: %v", bi4, bn4)
	}

	bn5 := new(BigNum).Div(bn1, bn2)
	bi5 := new(BigNum).DivBigInt(bn1.ToBigInt(), bn2.ToBigInt())
	if bi5.Cmp(bn5) != 0 {
		t.Fatalf("exp: %v, act: %v", bi5, bn5)
	}

	bn6 := new(BigNum).Mul(bn1, bn2)
	bi6 := new(BigNum).MulBigInt(bn1.ToBigInt(), bn2.ToBigInt())
	if bi6.Cmp(bn6) != 0 {
		t.Fatalf("exp: %v, act: %v", bi6, bn6)
	}
}

func TestBigNum_Marshal(t *testing.T) {
	bn1 := NewBigNumFromInt(rand.Int63())

	dataBin := make([]byte, bn1.Len())
	err := bn1.MarshalBinaryTo(dataBin)
	if err != nil {
		t.Fatal(err)
	}

	bn2 := new(BigNum)
	err = bn2.UnmarshalBinary(dataBin)
	if err != nil {
		t.Fatal(err)
	}

	if bn1.Cmp(bn2) != 0 {
		t.Fatalf("exp: %v, act: %v", bn1, bn2)
	}

	bn3 := bn1.Copy()
	if bn1.Cmp(bn3) != 0 {
		t.Fatalf("exp: %v, act: %v", bn1, bn3)
	}
}
