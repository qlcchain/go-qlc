package types

import "github.com/cheekybits/genny/generic"

type GenericT generic.Type
type GenericK generic.Type

type GenericType struct {
}

func (t *GenericType) Serialize() ([]byte, error) {
	return nil, nil
}

func (t *GenericType) Deserialize(text []byte) error {
	return nil
}

type GenericKey struct {
}

func (k *GenericKey) Serialize() ([]byte, error) {
	return nil, nil
}

func (k *GenericKey) Deserialize(text []byte) error {
	return nil
}
