/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import "github.com/cornelk/hashmap"

var cache = hashmap.New(50)

func GetEventBus(id string) EventBus {
	if id == "" {
		return NewActorEventBus()
	}

	if v, ok := cache.GetStringKey(id); ok {
		return v.(EventBus)
	}

	eb := New()
	cache.Set(id, eb)
	return eb
}
