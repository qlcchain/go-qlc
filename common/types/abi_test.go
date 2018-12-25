/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/base64"
	"github.com/json-iterator/go"
	"testing"
)

func TestAbi_MarshalText(t *testing.T) {
	hash, _ := NewHash("FFF8FF5DF1B6ED4FC7F300848931416581AE742999A2399563842F579E018D6B")
	abi := ContractAbi{Abi: []byte("6060604052341561000F57600080FD5B336000806101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF021916908373FFFFFFFFFFFFFFFF FFFF"), AbiLength: 64, AbiHash: hash}

	str, err := jsoniter.MarshalToString(&abi)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(str)

	bytes, err := base64.StdEncoding.DecodeString("NjA2MDYwNDA1MjM0MTU2MTAwMEY1NzYwMDA4MEZENUIzMzYwMDA4MDYxMDEwMDBBODE1NDgxNzNGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGRkZGMDIxOTE2OTA4MzczRkZGRkZGRkZGRkZGRkZGRiBGRkZG")
	if err != nil {
		t.Error(err)
	}

	t.Log(string(bytes))
}
