package relation

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/pkg/errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
)

var strLength = 10
var suffix = []byte("qlcInterface")
var identityLength = strLength + len(suffix)
var relationMap = make(map[string]structInfo)

type structInfo struct {
	value      reflect.Value
	identityID []byte
}

func RegisterInterface(con interface{}) error {
	t := reflect.ValueOf(con)
	typ := t.Type().String()
	if _, ok := con.(types.Serializer); !ok {
		return fmt.Errorf("%s not implement Serializer interface", typ)
	}
	if _, ok := con.(types.Convert); !ok {
		return fmt.Errorf("%s not implement Convert interface", typ)
	}
	if _, ok := relationMap[typ]; ok {
		return errors.Errorf("%s defined repeated", typ)
	}
	sr := structInfo{
		value:      t,
		identityID: identityID(),
	}
	relationMap[typ] = sr
	return nil
}

func ConvertToBytes(con interface{}) ([]byte, error) {
	if obj, ok := con.(types.Serializer); ok {
		val, err := obj.Serialize()
		if err != nil {
			return nil, err
		}
		if _, ok := con.(types.Convert); ok {
			typ := reflect.TypeOf(con).String()
			if sr, ok := relationMap[typ]; ok {
				val = append(val, sr.identityID...)
			} else {
				return nil, fmt.Errorf("%s has not regiseted", typ)
			}
		}
		return val, nil
	} else if r, ok := con.([]byte); ok {
		return r, nil
	} else {
		return nil, errors.New("invalid typ to convert")
	}
}

func ConvertToInterface(val []byte) (types.Convert, error) { //if val is not a Convert type, can not Deserialize because not register
	if len(val) > identityLength {
		identity := val[len(val)-identityLength:]
		if m, err := getStructById(identity); err == nil {
			args := []reflect.Value{reflect.ValueOf(val[:len(val)-identityLength])}
			typ := m.value
			v := typ.MethodByName("Deserialize").Call(args)
			if len(v) <= 0 || !v[0].IsNil() {
				return nil, fmt.Errorf("call method error : %s", v[0])
			} else {
				return typ.Interface().(types.Convert), nil
			}
		}
	}
	return nil, nil
}

func identityID() []byte {
	var id []byte
	for {
		id = make([]byte, 0)
		b := make([]byte, strLength)
		_ = random.Bytes(b)
		id = append(id, b...)
		id = append(id, suffix...)
		if _, err := getStructById(id); err != nil {
			break
		}
	}
	return id
}

func getStructById(id []byte) (structInfo, error) {
	for _, v := range relationMap {
		if bytes.EqualFold(v.identityID, id) {
			return v, nil
		}
	}
	return structInfo{}, errors.New("not found struct map")
}
