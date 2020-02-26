/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"testing"
)

func TestGetEventBus(t *testing.T) {
	eb0 := DefaultActorBus
	eb1 := GetEventBus("")
	if eb0 != eb1 {
		t.Fatal("invalid default eb")
	}

	id1 := "111111"
	eb2 := GetEventBus(id1)
	eb3 := GetEventBus(id1)

	if eb2 != eb3 {
		t.Fatal("invalid eb of same id")
	}

	id2 := "222222"
	eb4 := GetEventBus(id2)
	if eb3 == eb4 {
		t.Fatal("invalid eb of diff ids")
	}
}

func TestGetFeedEventBus(t *testing.T) {
	eb0 := DefaultFeedBus
	eb1 := GetFeedEventBus("")
	if eb0 != eb1 {
		t.Fatal("invalid default eb")
	}

	id1 := "111111"
	eb2 := GetFeedEventBus(id1)
	eb3 := GetFeedEventBus(id1)

	if eb2 != eb3 {
		t.Fatal("invalid eb of same id")
	}

	id2 := "222222"
	eb4 := GetFeedEventBus(id2)
	if eb3 == eb4 {
		t.Fatal("invalid eb of diff ids")
	}
}
