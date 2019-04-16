/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"fmt"
	"reflect"
	"sync"
)

// DefaultEventBus - box for handlers and callbacks.
type DefaultEventBus struct {
	handlers map[string][]*eventHandler
	lock     sync.Mutex
	wg       sync.WaitGroup
}

type eventHandler struct {
	callBack      reflect.Value
	flagOnce      bool
	async         bool
	transactional bool
	sync.Mutex
}

// New returns new DefaultEventBus with empty handlers.
func New() EventBus {
	b := &DefaultEventBus{
		make(map[string][]*eventHandler),
		sync.Mutex{},
		sync.WaitGroup{},
	}
	return EventBus(b)
}

// doSubscribe handles the subscription logic and is utilized by the public Subscribe functions
func (eb *DefaultEventBus) doSubscribe(topic string, fn interface{}, handler *eventHandler) error {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	if !(reflect.TypeOf(fn).Kind() == reflect.Func) {
		return fmt.Errorf("%s is not of type reflect.Func", reflect.TypeOf(fn).Kind())
	}
	eb.handlers[topic] = append(eb.handlers[topic], handler)
	return nil
}

// Subscribe subscribes to a topic.
// Returns error if `fn` is not a function.
func (eb *DefaultEventBus) Subscribe(topic string, fn interface{}) error {
	return eb.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), false, false, false, sync.Mutex{},
	})
}

// SubscribeAsync subscribes to a topic with an asynchronous callback
// Transactional determines whether subsequent callbacks for a topic are
// run serially (true) or concurrently (false)
// Returns error if `fn` is not a function.
func (eb *DefaultEventBus) SubscribeAsync(topic string, fn interface{}, transactional bool) error {
	return eb.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), false, true, transactional, sync.Mutex{},
	})
}

// SubscribeOnce subscribes to a topic once. Handler will be removed after executing.
// Returns error if `fn` is not a function.
func (eb *DefaultEventBus) SubscribeOnce(topic string, fn interface{}) error {
	return eb.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), true, false, false, sync.Mutex{},
	})
}

// SubscribeOnceAsync subscribes to a topic once with an asynchronous callback
// Handler will be removed after executing.
// Returns error if `fn` is not a function.
func (eb *DefaultEventBus) SubscribeOnceAsync(topic string, fn interface{}) error {
	return eb.doSubscribe(topic, fn, &eventHandler{
		reflect.ValueOf(fn), true, true, false, sync.Mutex{},
	})
}

// HasCallback returns true if exists any callback subscribed to the topic.
func (eb *DefaultEventBus) HasCallback(topic string) bool {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	_, ok := eb.handlers[topic]
	if ok {
		return len(eb.handlers[topic]) > 0
	}
	return false
}

// Unsubscribe removes callback defined for a topic.
// Returns error if there are no callbacks subscribed to the topic.
func (eb *DefaultEventBus) Unsubscribe(topic string, handler interface{}) error {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	if _, ok := eb.handlers[topic]; ok && len(eb.handlers[topic]) > 0 {
		eb.removeHandler(topic, eb.findHandlerIdx(topic, reflect.ValueOf(handler)))
		return nil
	}
	return fmt.Errorf("topic %s doesn't exist", topic)
}

// Publish executes callback defined for a topic. Any additional argument will be transferred to the callback.
func (eb *DefaultEventBus) Publish(topic string, args ...interface{}) {
	eb.lock.Lock() // will unlock if handler is not found or always after setUpPublish
	defer eb.lock.Unlock()
	for topicPattern, handlers := range eb.handlers {
		if 0 < len(handlers) && MatchSimple(topicPattern, topic) {
			// Handlers slice may be changed by removeHandler and Unsubscribe during iteration,
			// so make a copy and iterate the copied slice.
			copyHandlers := make([]*eventHandler, 0, len(handlers))
			copyHandlers = append(copyHandlers, handlers...)
			for i, handler := range copyHandlers {
				if handler.flagOnce {
					eb.removeHandler(topic, i)
				}
				if !handler.async {
					eb.doPublish(handler, topic, args...)
				} else {
					eb.wg.Add(1)
					if handler.transactional {
						handler.Lock()
					}
					go eb.doPublishAsync(handler, topic, args...)
				}
			}
		}
	}
}

func (eb *DefaultEventBus) doPublish(handler *eventHandler, topic string, args ...interface{}) {
	passedArguments := eb.setUpPublish(topic, args...)
	handler.callBack.Call(passedArguments)
}

func (eb *DefaultEventBus) doPublishAsync(handler *eventHandler, topic string, args ...interface{}) {
	defer eb.wg.Done()
	if handler.transactional {
		defer handler.Unlock()
	}
	eb.doPublish(handler, topic, args...)
}

func (eb *DefaultEventBus) removeHandler(topic string, idx int) {
	if _, ok := eb.handlers[topic]; !ok {
		return
	}
	l := len(eb.handlers[topic])

	if !(0 <= idx && idx < l) {
		return
	}

	copy(eb.handlers[topic][idx:], eb.handlers[topic][idx+1:])
	eb.handlers[topic][l-1] = nil // or the zero value of T
	eb.handlers[topic] = eb.handlers[topic][:l-1]
}

func (eb *DefaultEventBus) findHandlerIdx(topic string, callback reflect.Value) int {
	if _, ok := eb.handlers[topic]; ok {
		for idx, handler := range eb.handlers[topic] {
			if handler.callBack == callback {
				return idx
			}
		}
	}
	return -1
}

func (eb *DefaultEventBus) setUpPublish(topic string, args ...interface{}) []reflect.Value {
	passedArguments := make([]reflect.Value, 0)
	for _, arg := range args {
		passedArguments = append(passedArguments, reflect.ValueOf(arg))
	}
	return passedArguments
}

// WaitAsync waits for all async callbacks to complete
func (eb *DefaultEventBus) WaitAsync() {
	eb.wg.Wait()
}
