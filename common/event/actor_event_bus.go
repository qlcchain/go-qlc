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
	"sort"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor/middleware"
	"github.com/AsynkronIT/protoactor-go/router"

	"github.com/qlcchain/go-qlc/common/hashmap"
	ct "github.com/qlcchain/go-qlc/common/topic"

	"github.com/AsynkronIT/protoactor-go/actor"
)

const (
	defaultQueueSize   = 100
	defaultHandlerSize = 1024
	maxConcurrency     = 10
)

type ActorEventBus struct {
	subscribers *hashmap.HashMap
	queueSize   int
	Context     *actor.RootContext
}

type subscriberOption struct {
	Subscriber *actor.PID
	index      int
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
		Context:     rootContext,
	}
}

func (eb *ActorEventBus) Subscribe(topic ct.TopicType, fn interface{}) error {
	return eb.doSubscribe(topic, fn)
}

func (eb *ActorEventBus) doSubscribe(topic ct.TopicType, subscriber interface{}) error {
	t := string(topic)
	s, ok := subscriber.(*actor.PID)
	if !ok {
		return errors.New("subscriber is not *actor.PID")
	}

	if value, ok := eb.subscribers.GetStringKey(t); ok {
		opts := value.([]*subscriberOption)
		list := append(opts, &subscriberOption{
			Subscriber: s,
			index:      len(opts),
		})
		sort.Slice(list, func(i, j int) bool {
			return list[i].index < list[j].index
		})
		eb.subscribers.Set(t, list)
	} else {
		eb.subscribers.Set(t, []*subscriberOption{{
			Subscriber: s,
			index:      0,
		}})
	}

	return nil
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
			if subscribers[i].Subscriber == s {
				eb.subscribers.Set(t, append(subscribers[0:i], subscribers[i+1:]...))
			}
		}

		return nil
	}

	return fmt.Errorf("topic %s doesn't exist", t)
}

func (eb *ActorEventBus) Publish(topic ct.TopicType, msg interface{}) {
	eb.Subscribers(topic, func(subscriber *actor.PID) {
		eb.Context.Send(subscriber, msg)
	})
}

func (eb *ActorEventBus) PublishFrom(topic ct.TopicType, msg interface{}, publisher interface{}) error {
	p, ok := publisher.(*actor.PID)
	if !ok {
		return errors.New("invalid publisher")
	}
	eb.Subscribers(topic, func(subscriber *actor.PID) {
		eb.Context.RequestWithCustomSender(subscriber, msg, p)
	})

	return nil
}

func (eb *ActorEventBus) Subscribers(topic ct.TopicType, callback func(subscriber *actor.PID)) {
	t := string(topic)
	quitCh := make(chan struct{})
	defer close(quitCh)
	for kv := range eb.subscribers.Iter(quitCh) {
		topicPattern := kv.Key.(string)
		options := kv.Value.([]*subscriberOption)
		if len(options) > 0 && MatchSimple(topicPattern, t) {
			for _, subscriber := range options {
				callback(subscriber.Subscriber)
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
			if err := eb.Context.PoisonFuture(subscribers[i].Subscriber).Wait(); err != nil {
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
	quitCh := make(chan struct{})
	defer close(quitCh)
	for k := range eb.subscribers.Iter(quitCh) {
		if err := eb.CloseTopic(ct.TopicType(k.Key.(string))); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return joinErrs(errs...)
	}
	return nil
}

type ActorSubscriber struct {
	Bus         EventBus
	subscribers []*subscriberContainer
	subscriber  *actor.PID
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
	}
}

// WithSubscribe set default handler
func (s *ActorSubscriber) WithSubscribe(subscriber *actor.PID) *ActorSubscriber {
	s.subscriber = subscriber
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

// Unsubscribe by topic
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

// UnsubscribeAll topic
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

var (
	DefaultPublisher = actor.EmptyRootContext.Spawn(actor.PropsFromFunc(func(c actor.Context) {

	}).WithReceiverMiddleware(middleware.Logger))
	defaultTimeout = 3 * time.Second
)

type ActorPublisher struct {
	Bus       EventBus
	publisher *actor.PID
	timeout   time.Duration
}

// NewActorPublisher with publisher and eventbus
func NewActorPublisher(publisher *actor.PID, bus ...EventBus) *ActorPublisher {
	var b EventBus
	if len(bus) == 0 {
		b = DefaultActorBus
	} else {
		b = bus[0]
	}

	if publisher == nil {
		publisher = DefaultPublisher
	}

	return &ActorPublisher{
		Bus:       b,
		publisher: publisher,
		timeout:   defaultTimeout,
	}
}

// WithTimeout change default timeout
func (p *ActorPublisher) WithTimeout(timeout time.Duration) *ActorPublisher {
	p.timeout = timeout
	return p
}

// Publish msg to topic
func (p *ActorPublisher) Publish(topic ct.TopicType, msg interface{}) error {
	return p.Bus.PublishFrom(topic, msg, p.publisher)
}

func (p *ActorPublisher) PublishFuture(topic ct.TopicType, msg interface{}, callback func(msg interface{}, err error)) {
	eb := p.Bus.(*ActorEventBus)
	eb.Subscribers(topic, func(subscriber *actor.PID) {
		result, err := eb.Context.RequestFuture(subscriber, msg, p.timeout).Result()
		callback(result, err)
	})
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
