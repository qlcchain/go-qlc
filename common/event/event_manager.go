/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"github.com/cornelk/hashmap"
)

var (
	cache          = hashmap.New(50)
	cacheFeedBusID = hashmap.New(50)

	DefaultActorBus = NewActorEventBus()
	DefaultFeedBus  = NewFeedEventBus()
)

func GetEventBus(id string) EventBus {
	if id == "" {
		return DefaultActorBus
	}

	if v, ok := cache.GetStringKey(id); ok {
		return v.(EventBus)
	}

	eb := NewActorEventBus()
	cache.Set(id, eb)
	return eb
}

func GetFeedEventBus(id string) *FeedEventBus {
	if id == "" {
		return DefaultFeedBus
	}

	if v, ok := cacheFeedBusID.GetStringKey(id); ok {
		return v.(*FeedEventBus)
	}

	eb := NewFeedEventBus()
	cacheFeedBusID.Set(id, eb)
	return eb
}
