/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

//subscriber defines subscription-related bus behavior
type subscriber interface {
	Subscribe(topic string, fn interface{}) error
	SubscribeAsync(topic string, fn interface{}, transactional bool) error
	SubscribeOnce(topic string, fn interface{}) error
	SubscribeOnceAsync(topic string, fn interface{}) error
	Unsubscribe(topic string, handler interface{}) error
}

//publisher defines publishing-related bus behavior
type publisher interface {
	Publish(topic string, args ...interface{})
}

//controller defines bus control behavior (checking handler's presence, synchronization)
type controller interface {
	HasCallback(topic string) bool
	WaitAsync()
}

type EventBus interface {
	subscriber
	publisher
	controller
}
