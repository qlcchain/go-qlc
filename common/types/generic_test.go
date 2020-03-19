/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"reflect"
	"testing"
)

func TestGenericType_Serialize(t *testing.T) {
	genericType := &GenericType{
		"QLC is the best",
	}

	if data, err := genericType.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		g2 := &GenericType{}
		if err := g2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(genericType, g2) {
			t.Fatalf("invalid generic type, exp: %v, act: %v", genericType, g2)
		}
	}
}

func TestGenericKey_Serialize(t *testing.T) {
	genericKey := &GenericKey{
		"QLC is the best",
	}

	if data, err := genericKey.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		g2 := &GenericKey{}
		if err := g2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(genericKey, g2) {
			t.Fatalf("invalid generic key, exp: %v, act: %v", genericKey, g2)
		}
	}
}
