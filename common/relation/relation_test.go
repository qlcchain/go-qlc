package relation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestRegisterRelation(t *testing.T) {
	if err := RegisterInterface(new(StructA)); err != nil {
		t.Fatal(err)
	}

	a := new(StructA)
	a.Address = mock.Address()
	bytes, err := ConvertToBytes(a)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := ConvertToInterface(bytes)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(reflect.TypeOf(r2).String())
	if o, ok := r2.(*StructA); ok {
		if o.Address != a.Address {
			t.Fatal()
		}
	} else {
		t.Fatal()
	}
	if o, ok := r2.(types.Serializer); ok {
		fmt.Println(o.Serialize())
	} else {
		t.Fatal()
	}
	if o, ok := r2.(types.Convert); ok {
		fmt.Println(o.ConvertToSchema())
	} else {
		t.Fatal()
	}
}

type StructA struct {
	Hash    types.Hash    `json:"hash"`
	Address types.Address `json:"address"`
	Typ     int           `json:"typ"`
	Time    int64         `json:"time"`
}

func (a *StructA) Serialize() ([]byte, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (a *StructA) Deserialize(text []byte) error {
	return json.Unmarshal(text, a)
}

func (a *StructA) ConvertToSchema() ([]types.Schema, error) {
	fmt.Println("RelationConvert")
	return nil, nil
}

func TestRegisterRelation_bytes(t *testing.T) {
	bs := mock.Hash().Bytes()
	b, err := ConvertToBytes(bs)
	if err != nil {
		t.Fatal(err)
	}
	re, err := ConvertToInterface(b)
	if err != nil {
		t.Fatal(err)
	}
	if re != nil {
		t.Fatal()
	}
}
