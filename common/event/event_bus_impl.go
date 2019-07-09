/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/sync/hashmap"
	"reflect"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
)

const (
	defaultQueueSize   = 100
	defaultHandlerSize = 1024
)

// DefaultEventBus
type DefaultEventBus struct {
	handlers  *hashmap.HashMap
	queueSize int
}

func (eb *DefaultEventBus) Close() error {
	for k := range eb.handlers.Iter() {
		eb.CloseTopic(k.Key.(string))
	}
	return nil
}

type handlerOption struct {
	isSync bool
}

type eventHandler struct {
	callBack reflect.Value
	option   *handlerOption
	pool     *workerpool.WorkerPool
}

// New returns new DefaultEventBus with empty handlers.
func New() EventBus {
	b := &DefaultEventBus{
		handlers:  hashmap.New(defaultHandlerSize),
		queueSize: defaultQueueSize,
	}
	return EventBus(b)
}

func NewEventBus(queueSize int) EventBus {
	b := &DefaultEventBus{
		handlers:  hashmap.New(defaultHandlerSize),
		queueSize: queueSize,
	}
	return EventBus(b)
}

var (
	once  sync.Once
	eb    EventBus
	cache *hashmap.HashMap
)

func init() {
	cache = hashmap.New(50)
}

func SimpleEventBus() EventBus {
	once.Do(func() {
		eb = New()
	})

	return eb
}

func GetEventBus(id string) EventBus {
	if len(id) == 0 {
		return SimpleEventBus()
	}

	if v, ok := cache.GetStringKey(id); ok {
		return v.(EventBus)
	} else {
		eb := New()
		cache.Set(id, eb)
		return eb
	}
}

// doSubscribe handles the subscription logic and is utilized by the public Subscribe functions
func (eb *DefaultEventBus) doSubscribe(topic string, fn interface{}, option *handlerOption) error {
	kind := reflect.TypeOf(fn).Kind()
	if kind != reflect.Func {
		return fmt.Errorf("%s is not of type reflect.Func", kind)
	}

	handler := &eventHandler{
		callBack: reflect.ValueOf(fn),
		option:   option,
		pool:     workerpool.New(eb.queueSize),
	}

	if value, ok := eb.handlers.GetStringKey(topic); ok {
		list := value.(*eventHandlers)
		list.Add(handler)
	} else {
		handlers := newEventHandler()
		handlers.Add(handler)
		eb.handlers.Set(topic, handlers)
	}
	return nil
}

// Subscribe subscribes to a topic.
// Returns error if `fn` is not a function.
func (eb *DefaultEventBus) Subscribe(topic string, fn interface{}) error {
	return eb.doSubscribe(topic, fn, &handlerOption{isSync: false})
}

func (eb *DefaultEventBus) SubscribeSync(topic string, fn interface{}) error {
	return eb.doSubscribe(topic, fn, &handlerOption{isSync: true})
}

// HasCallback returns true if exists any callback subscribed to the topic.
func (eb *DefaultEventBus) HasCallback(topic string) bool {
	if v, ok := eb.handlers.GetStringKey(topic); ok {
		handlers := v.(*eventHandlers)
		return handlers.Size() > 0
	}
	return false
}

// Close unsubscribe all handlers from given topic
func (eb *DefaultEventBus) CloseTopic(topic string) {
	if value, ok := eb.handlers.GetStringKey(topic); ok {
		value.(*eventHandlers).Clear()
		eb.handlers.Del(topic)
	}
}

// Unsubscribe removes callback defined for a topic.
// Returns error if there are no callbacks subscribed to the topic.
//TODO: fix unsubscribe failed
func (eb *DefaultEventBus) Unsubscribe(topic string, handler interface{}) error {
	kind := reflect.TypeOf(handler).Kind()
	if kind != reflect.Func {
		return fmt.Errorf("%s is not of type reflect.Func", kind)
	}

	if value, ok := eb.handlers.GetStringKey(topic); ok {
		if flag := value.(*eventHandlers).RemoveCallback(reflect.ValueOf(handler)); flag {
			//return nil
		}
		return nil
	} else {
		return fmt.Errorf("topic %s doesn't exist", topic)
	}
}

// Publish executes callback defined for a topic. Any additional argument will be transferred to the callback.
func (eb *DefaultEventBus) Publish(topic string, args ...interface{}) {
	rArgs := eb.setUpPublish(topic, args...)
	for kv := range eb.handlers.Iter() {
		topicPattern := kv.Key.(string)
		handlers := kv.Value.(*eventHandlers)
		if handlers.Size() > 0 && MatchSimple(topicPattern, topic) {
			all := handlers.All()
			for _, handler := range all {
				h := handler

				//waiting until the queue is ready
				if h.pool.WaitingQueueSize() >= common.EventBusWaitingQueueSize {
					checkInterval := time.NewTicker(10 * time.Millisecond)

				checkWaitingQueueOut:
					for {
						select {
						case <-checkInterval.C:
							if h.pool.WaitingQueueSize() < common.EventBusWaitingQueueSize {
								checkInterval.Stop()
								break checkWaitingQueueOut
							}
						}
					}
				}

				if h.option.isSync {
					h.pool.SubmitWait(func() {
						h.callBack.Call(rArgs)
					})
				} else {
					h.pool.Submit(func() {
						h.callBack.Call(rArgs)
					})
				}
			}
		}
	}
}

func (eb *DefaultEventBus) setUpPublish(topic string, args ...interface{}) []reflect.Value {
	passedArguments := make([]reflect.Value, 0)
	for _, arg := range args {
		passedArguments = append(passedArguments, reflect.ValueOf(arg))
	}
	return passedArguments
}

func (eb *DefaultEventBus) dump() {
	for kv := range eb.handlers.Iter() {
		fmt.Println(kv.Key.(string), kv.Value.(*eventHandlers).String())
	}
}
