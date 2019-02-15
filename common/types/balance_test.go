package types

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"
)

func TestBalance_MarshalJSON(t *testing.T) {
	b1 := StringToBalance("123456789")

	bs, err := json.Marshal(&b1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bs))

	b2 := Balance{new(big.Int)}
	//b2 := Balance{}
	err = json.Unmarshal(bs, &b2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b2)
	if b1.Compare(b2) != BalanceCompEqual {
		t.Fatal("b1!=b2")
	}
	t.Log(b1)
}

func TestStringToBalance(t *testing.T) {
	type args struct {
		b string
	}
	tests := []struct {
		name string
		args args
		want Balance
	}{
		{"StringToBalance1", args{"123445"}, Balance{big.NewInt(123445)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToBalance(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesToBalance(t *testing.T) {
	i := 31415926
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(i))
	balance := Balance{big.NewInt(int64(i))}
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want Balance
	}{
		{"BytesToBalance1", args{bs}, balance},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesToBalance(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesToBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBalance_Equal(t *testing.T) {
	type fields struct {
		Int *big.Int
	}
	type args struct {
		b2 Balance
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"zero", fields{big.NewInt(0)}, args{ZeroBalance}, true},
		{"normal", fields{big.NewInt(22)}, args{ZeroBalance}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Balance{
				Int: tt.fields.Int,
			}
			if got := b.Equal(tt.args.b2); got != tt.want {
				t.Errorf("Balance.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBalance_MarshalBinaryTo(t *testing.T) {
	b := Balance{big.NewInt(123)}
	buff := make([]byte, b.Len())
	_ = b.MarshalBinaryTo(buff)
	t.Log(b)

	b2 := ZeroBalance
	err := b2.UnmarshalBinary(buff)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b2)
	if !b.Equal(b2) {
		t.Fatal("b != b2")
	}
}

func TestBalance_IsZero(t *testing.T) {
	type fields struct {
		Int *big.Int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"iszero1", fields{Int: big.NewInt(0)}, true},
		{"iszero2", fields{Int: big.NewInt(222)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Balance{
				Int: tt.fields.Int,
			}
			if got := b.IsZero(); got != tt.want {
				t.Errorf("Balance.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBalance_Sub(t *testing.T) {
	bb1 := Balance{big.NewInt(10000)}
	bb2 := Balance{big.NewInt(12000)}
	t.Log("bb1-bb2: ", bb1.Sub(bb2))
}

func TestBalance_Add(t *testing.T) {
	bb1 := Balance{big.NewInt(10000)}
	bb2 := Balance{big.NewInt(12000)}
	t.Log("bb1+bb2: ", bb1.Add(bb2))
}

func TestBalance_Div(t *testing.T) {
	bb1 := Balance{big.NewInt(10000)}
	bb2, err := bb1.Div(2)
	if err != nil {
		t.Fatal("err should be nil")
	}
	bb3, err := bb1.Div(0)
	if err == nil {
		t.Fatal("err should not be nil")
	}
	t.Log("bb1,bb2: ", bb1, bb2, bb3)
}
