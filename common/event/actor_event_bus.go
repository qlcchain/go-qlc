/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/log"
	"sort"
	"time"

	"github.com/AsynkronIT/protoactor-go/router"

	ct "github.com/qlcchain/go-qlc/common/topic"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/cornelk/hashmap"
)

const (
	maxConcurrency = 10
	maxTimeout     = 30 * time.Second
)

type ActorEventBus struct {
	subscribers *hashmap.HashMap
	queueSize   int
	ctx         *actor.RootContext
}

type subscriberOption struct {
	subscriber *actor.PID
	isSync     bool
	timeout    time.Duration
}

func (eb *ActorEventBus) SubscribeSync(topic ct.TopicType, fn interface{}) error {
	return eb.doSubscribe(topic, fn, true, maxTimeout)
}

func NewActorEventBus() EventBus {
	// root context
	//rootContext := actor.NewRootContext(nil).WithGuardian(actor.NewRestartingStrategy()).
	//	WithSpawnMiddleware(func(next actor.SpawnFunc) actor.SpawnFunc {
	//		return func(id string, props *actor.Props, parentContext actor.SpawnerContext) (pid *actor.PID, e error) {
	//			prop := router.NewRoundRobinPool(maxConcurrency).WithSupervisor(actor.NewRestartingStrategy()).
	//				WithGuardian(actor.NewOneForOneStrategy(5, 1000, func(reason interface{}) actor.Directive {
	//					return actor.StopDirective
	//				}))
	//
	//			return next(id, prop, parentContext)
	//		}
	//	}, opentracing.TracingMiddleware())
	rootContext := actor.NewRootContext(nil).WithGuardian(actor.NewRestartingStrategy())
	return &ActorEventBus{
		subscribers: hashmap.New(defaultHandlerSize),
		queueSize:   defaultQueueSize,
		ctx:         rootContext,
	}
}

func (eb *ActorEventBus) SubscribeSyncTimeout(topic ct.TopicType, fn interface{}, timeout time.Duration) error {
	return eb.doSubscribe(topic, fn, true, timeout)
}

func (eb *ActorEventBus) Subscribe(topic ct.TopicType, fn interface{}) error {
	return eb.doSubscribe(topic, fn, false, 0)
}

func (eb *ActorEventBus) doSubscribe(topic ct.TopicType, subscriber interface{}, isSync bool, timeout time.Duration) error {
	t := string(topic)
	s, ok := subscriber.(*actor.PID)
	if !ok {
		return errors.New("subscriber is not *actor.PID")
	}

	if value, ok := eb.subscribers.GetStringKey(t); ok {
		list := append(value.([]*subscriberOption), &subscriberOption{
			subscriber: s,
			isSync:     isSync,
			timeout:    timeout,
		})
		sort.Slice(list, func(i, j int) bool {
			return bool2Int(list[i].isSync) > bool2Int(list[j].isSync)
		})
		eb.subscribers.Set(t, list)
	} else {
		eb.subscribers.Set(t, []*subscriberOption{{
			subscriber: s,
			isSync:     isSync,
			timeout:    timeout,
		}})
	}

	return nil
}

func bool2Int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (eb *ActorEventBus) Unsubscribe(topic ct.TopicType, subscriber interface{}) error {
	t := string(topic)
	s, ok := subscriber.(*actor.PID)
	if !ok {
		return errors.New("subscriber is not *actor.PID")
	}

	if value, ok := eb.subscribers.GetStringKey(t); ok {
		subscribers := value.([]*subscriberOption)
		for i := range subscribers {
			if subscribers[i].subscriber == s {
				eb.subscribers.Set(t, append(subscribers[0:i], subscribers[i+1:]...))
			}
		}

		return nil
	}

	return fmt.Errorf("topic %s doesn't exist", t)
}

func (eb *ActorEventBus) Publish(topic ct.TopicType, args ...interface{}) {
	t := string(topic)
	msg := args[0]
	for kv := range eb.subscribers.Iter() {
		topicPattern := kv.Key.(string)
		options := kv.Value.([]*subscriberOption)
		if len(options) > 0 && MatchSimple(topicPattern, t) {
			for _, subscriber := range options {
				if subscriber.isSync {
					if err := eb.ctx.RequestFuture(subscriber.subscriber, msg, subscriber.timeout).Wait(); err != nil {
						log.Root.Warn(err)
					}
				} else {
					eb.ctx.Send(subscriber.subscriber, msg)
				}
			}
		}
	}
}

func (eb *ActorEventBus) HasCallback(topic ct.TopicType) bool {
	if v, ok := eb.subscribers.GetStringKey(string(topic)); ok {
		subscribers := v.([]*subscriberOption)
		return len(subscribers) > 0
	}
	return false
}

func (eb *ActorEventBus) CloseTopic(topic ct.TopicType) error {
	t := string(topic)
	if value, ok := eb.subscribers.GetStringKey(t); ok && value != nil {
		var errs []error
		subscribers := value.([]*subscriberOption)
		for i := range subscribers {
			if err := eb.ctx.PoisonFuture(subscribers[i].subscriber).Wait(); err != nil {
				errs = append(errs, err)
			}
		}
		eb.subscribers.Del(t)
		if len(errs) > 0 {
			return fmt.Errorf("close topic[%s]: %s", t, joinErrs(errs...))
		}
	}
	return nil
}

func (eb *ActorEventBus) Close() error {
	var errs []error
	for k := range eb.subscribers.Iter() {
		if err := eb.CloseTopic(ct.TopicType(k.Key.(string))); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return joinErrs(errs...)
	}
	return nil
}

var (
	errNilSubscriber = errors.New("default subscriber is Nil")
)

type ActorSubscriber struct {
	Bus         EventBus
	subscribers []*subscriberContainer
	subscriber  *actor.PID
	timeout     time.Duration
}

type subscriberContainer struct {
	subscriber *actor.PID
	topic      ct.TopicType
}

func Spawn(fn func(c actor.Context)) *actor.PID {
	return actor.EmptyRootContext.Spawn(actor.PropsFromFunc(func(c actor.Context) {
		fn(c)
	}))
}

// SpawnWithPool spawn a process with pool support
func SpawnWithPool(fn func(c actor.Context)) *actor.PID {
	return actor.EmptyRootContext.Spawn(router.NewRoundRobinPool(maxConcurrency).WithFunc(func(c actor.Context) {
		fn(c)
	}))
}

func NewActorSubscriber(subscriber *actor.PID, bus ...EventBus) *ActorSubscriber {
	var b EventBus
	if len(bus) == 0 {
		b = DefaultActorBus
	} else {
		b = bus[0]
	}
	return &ActorSubscriber{
		Bus:         b,
		subscriber:  subscriber,
		subscribers: make([]*subscriberContainer, 0),
		timeout:     maxTimeout,
	}
}

// WithSubscribe set default handler
func (s *ActorSubscriber) WithSubscribe(subscriber *actor.PID) *ActorSubscriber {
	s.subscriber = subscriber
	return s
}

func (s *ActorSubscriber) WithTimeout(timeout time.Duration) *ActorSubscriber {
	s.timeout = timeout
	return s
}

// SubscribeOne subscribe topic->handler
func (s *ActorSubscriber) SubscribeOne(topic ct.TopicType, subscriber *actor.PID) error {
	if err := s.Bus.Subscribe(topic, subscriber); err != nil {
		return err
	}

	s.subscribers = append(s.subscribers, &subscriberContainer{
		subscriber: subscriber,
		topic:      topic,
	})

	return nil
}

func (s *ActorSubscriber) SubscribeSyncOne(topic ct.TopicType, subscriber *actor.PID) error {
	if subscriber == nil {
		return errNilSubscriber
	}
	if err := s.Bus.SubscribeSyncTimeout(topic, subscriber, s.timeout); err != nil {
		return err
	} else {
		s.subscribers = append(s.subscribers, &subscriberContainer{
			subscriber: subscriber,
			topic:      topic,
		})
	}

	return nil
}

// SubscribeSync multiple topics by default handler
func (s *ActorSubscriber) SubscribeSync(topic ...ct.TopicType) error {
	if s.subscriber == nil {
		return errNilSubscriber
	}
	var errs []error
	for _, t := range topic {
		if err := s.Bus.SubscribeSyncTimeout(t, s.subscriber, s.timeout); err != nil {
			errs = append(errs, err)
		} else {
			s.subscribers = append(s.subscribers, &subscriberContainer{
				subscriber: s.subscriber,
				topic:      t,
			})
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", joinErrs(errs...))
	}
	return nil
}

// Subscribe multiple topics by default handler
func (s *ActorSubscriber) Subscribe(topic ...ct.TopicType) error {
	var errs []error
	if s.subscriber == nil {
		return errors.New("default subscriber is Nil")
	}
	for _, t := range topic {
		if err := s.Bus.Subscribe(t, s.subscriber); err != nil {
			errs = append(errs, err)
		} else {
			s.subscribers = append(s.subscribers, &subscriberContainer{
				subscriber: s.subscriber,
				topic:      t,
			})
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", joinErrs(errs...))
	}
	return nil
}

func (s *ActorSubscriber) Unsubscribe(topic ct.TopicType) error {
	var errs []error
	temp := s.subscribers[:0]
	for _, sc := range s.subscribers {
		if sc.topic == topic {
			if err := s.Bus.Unsubscribe(sc.topic, sc.subscriber); err != nil {
				temp = append(temp, sc)
				errs = append(errs, err)
			}
		} else {
			temp = append(temp, sc)
		}
	}
	s.subscribers = temp
	if len(errs) > 0 {
		return fmt.Errorf("%s", joinErrs(errs...))
	}
	return nil
}

func (s *ActorSubscriber) UnsubscribeAll() error {
	var errs []error
	temp := s.subscribers[:0]
	for _, sc := range s.subscribers {
		if err := s.Bus.Unsubscribe(sc.topic, sc.subscriber); err != nil {
			temp = append(temp, sc)
			errs = append(errs, err)
		}
	}
	s.subscribers = temp
	if len(errs) > 0 {
		return fmt.Errorf("%s", joinErrs(errs...))
	}
	return nil
}

func joinErrs(errs ...error) error {
	var joinErrsR func(string, int, ...error) error
	joinErrsR = func(soFar string, count int, errs ...error) error {
		if len(errs) == 0 {
			if count == 0 {
				return nil
			}
			return fmt.Errorf(soFar)
		}
		current := errs[0]
		next := errs[1:]
		if current == nil {
			return joinErrsR(soFar, count, next...)
		}
		count++
		if count == 1 {
			return joinErrsR(fmt.Sprintf("%s", current), count, next...)
		} else if count == 2 {
			return joinErrsR(fmt.Sprintf("1: %s\n2: %s", soFar, current), count, next...)
		}
		return joinErrsR(fmt.Sprintf("%s\n%d: %s", soFar, count, current), count, next...)
	}
	return joinErrsR("", 0, errs...)
}
