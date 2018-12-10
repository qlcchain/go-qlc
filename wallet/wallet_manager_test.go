/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
	"reflect"
	"testing"
)

func TestMarshalWalletIds(t *testing.T) {
	var ids []uuid.UUID
	ids = append(ids, uuid.New())
	ids = append(ids, uuid.New())
	ids = append(ids, uuid.New())
	ids = append(ids, uuid.New())

	for _, id := range ids {
		fmt.Println(id.String())
	}

	bytes, _ := jsoniter.Marshal(ids)
	fmt.Println(bytes)

	var ids2 []uuid.UUID

	err := jsoniter.Unmarshal(bytes, &ids2)

	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids2 {
		fmt.Println(id.String())
	}

	if !reflect.DeepEqual(ids, ids2) {
		t.Fatal("ids!=id2")
	}
}
