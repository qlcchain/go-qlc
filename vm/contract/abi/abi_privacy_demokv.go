package abi

import (
	"strings"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
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
	data, err := ctx.GetStorage(contractaddress.PrivacyDemoKVAddress[:], sKey)
	if err != nil {
		return nil, err
	}

	return data, nil
}
