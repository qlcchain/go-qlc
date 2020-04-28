package types

import (
	"encoding/json"

	"github.com/cheekybits/genny/generic"
)

type GenericT generic.Type
type GenericK generic.Type

type GenericType struct {
	Value string
}

func (t *GenericType) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

func (t *GenericType) Deserialize(text []byte) error {
	return json.Unmarshal(text, t)
}

type GenericKey struct {
	Key string
}

func (k *GenericKey) Serialize() ([]byte, error) {
	return json.Marshal(k)
}

func (k *GenericKey) Deserialize(text []byte) error {
	return json.Unmarshal(text, k)
}

type GenericTypeC struct {
	Value string
}

func (t *GenericTypeC) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

func (t *GenericTypeC) Deserialize(text []byte) error {
	return json.Unmarshal(text, t)
}

type GenericKeyC struct {
	Key string
}

func (k *GenericKeyC) Serialize() ([]byte, error) {
	return json.Marshal(k)
}

func (k *GenericKeyC) Deserialize(text []byte) error {
	return json.Unmarshal(text, k)
}
