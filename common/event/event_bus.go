/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"io"

	"github.com/qlcchain/go-qlc/common"
)

//subscriber defines subscription-related bus behavior
type subscriber interface {
	Subscribe(topic common.TopicType, fn interface{}) error
	SubscribeSync(topic common.TopicType, fn interface{}) error
	Unsubscribe(topic common.TopicType, handler interface{}) error
}

//publisher defines publishing-related bus behavior
type publisher interface {
	Publish(topic common.TopicType, args ...interface{})
}

//controller defines bus control behavior (checking handler's presence, synchronization)
type controller interface {
	HasCallback(topic common.TopicType) bool
	CloseTopic(topic common.TopicType)
}

type EventBus interface {
	subscriber
	publisher
	controller
	io.Closer
}
