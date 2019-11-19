/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"io"
	"time"

	"github.com/qlcchain/go-qlc/common/topic"
)

// subscriber defines subscription-related bus behavior
type subscriber interface {
	Subscribe(topic topic.TopicType, fn interface{}) error
	SubscribeSync(topic topic.TopicType, fn interface{}) error
	SubscribeSyncTimeout(topic topic.TopicType, fn interface{}, timeout time.Duration) error
	Unsubscribe(topic topic.TopicType, fn interface{}) error
}

// publisher defines publishing-related bus behavior
type publisher interface {
	Publish(topic topic.TopicType, args ...interface{})
}

// controller defines bus control behavior (checking handler's presence, synchronization)
type controller interface {
	HasCallback(topic topic.TopicType) bool
	CloseTopic(topic topic.TopicType) error
}

type EventBus interface {
	subscriber
	publisher
	controller
	io.Closer
}

type Event struct {
	Topic   topic.TopicType
	Message interface{}
}
