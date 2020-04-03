package abi

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"strings"

	"github.com/qlcchain/go-qlc/vm/abi"
)

const (
	JsonPrivacyDemoKV = `
	[
		{"type":"function","name":"PrivacyDemoKVSet","inputs":[
			{"name":"key","type":"bytes"},
			{"name":"value","type":"bytes"}
		]}
	]`

	MethodNamePrivacyDemoKVSet = "PrivacyDemoKVSet"
)

const (
	PrivacyDemoKVStorageTypeKV byte = iota
)

var (
	PrivacyDemoKVABI, _ = abi.JSONToABIContract(strings.NewReader(JsonPrivacyDemoKV))
)

type PrivacyDemoKVABISetPara struct {
	Key   []byte
	Value []byte
}

func PrivacyKVGetValue(ctx *vmstore.VMContext, key []byte) ([]byte, error) {
	var sKey []byte
	sKey = append(sKey, PrivacyDemoKVStorageTypeKV)
	sKey = append(sKey, key...)
	data, err := ctx.GetStorage(types.PrivacyDemoKVAddress[:], sKey)
	if err != nil {
		return nil, err
	}

	return data, nil
}
