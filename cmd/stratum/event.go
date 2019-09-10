package main

import "sync"

type EventType int

const (
	EventUpdateApiWork EventType = iota
	EventBCastJobWork
	EventMinerSubmit
	EventJobSubmit
	EventMinerSendRsp
)

const (
	EventMaxChanSize = 1000
)

type EventBus struct {
	sync.Mutex
	allEventSubs map[EventType][]chan Event
}

type Event struct {
	Topic EventType
	Data  interface{}
}

func NewEventBus() *EventBus {
	eb := new(EventBus)
	eb.allEventSubs = make(map[EventType][]chan Event, EventMaxChanSize)
	return eb
}

func (eb *EventBus) Publish(topic EventType, data interface{}) {
	eb.Lock()
	defer eb.Unlock()

	subs := eb.allEventSubs[topic]
	for _, subChan := range subs {
		subChan <- Event{Topic: topic, Data: data}
	}
}

func (eb *EventBus) Subscribe(topic EventType, subChan chan Event) {
	eb.Lock()
	defer eb.Unlock()

	subs := eb.allEventSubs[topic]
	subs = append(subs, subChan)
	eb.allEventSubs[topic] = subs
}

var defaultEventBus *EventBus
var defaultEBOnce sync.Once

func GetDefaultEventBus() *EventBus {
	defaultEBOnce.Do(func() {
		defaultEventBus = NewEventBus()
	})
	return defaultEventBus
}
