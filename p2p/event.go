package p2p

import (
	"errors"
	"sync"
)

type EventType int16

type EventFunc func(v interface{})
type EventSubscriber chan interface{}

//  Event Type
const (
	EventPublish        EventType = 0
	EventConfirmReq     EventType = 1
	EventConfirmAck     EventType = 2
	EventSyncBlock      EventType = 3
	EventConfirmedBlock EventType = 4
)

type Event struct {
	m           sync.RWMutex
	subscribers map[EventType]map[EventSubscriber]EventFunc
}

func NewEvent() *Event {
	return &Event{
		subscribers: make(map[EventType]map[EventSubscriber]EventFunc),
	}
}

//  adds a new subscriber to Event.
func (e *Event) Subscribe(et EventType, ef EventFunc) EventSubscriber {
	e.m.Lock()
	defer e.m.Unlock()

	sub := make(chan interface{})
	_, ok := e.subscribers[et]
	if !ok {
		e.subscribers[et] = make(map[EventSubscriber]EventFunc)
	}
	e.subscribers[et][sub] = ef

	return sub
}

// UnSubscribe removes the specified subscriber
func (e *Event) UnSubscribe(et EventType, subscriber EventSubscriber) (err error) {
	e.m.Lock()
	defer e.m.Unlock()

	subEvent, ok := e.subscribers[et]
	if !ok {
		err = errors.New("No event type.")
		return err
	}

	delete(subEvent, subscriber)
	close(subscriber)

	return nil
}

//Notify subscribers that Subscribe specified event
func (e *Event) Notify(et EventType, value interface{}) (err error) {
	e.m.RLock()
	defer e.m.RUnlock()

	subs, ok := e.subscribers[et]
	if !ok {
		err = errors.New("No event type.")
		return err
	}

	for _, event := range subs {
		e.NotifySubscriber(event, value)
	}
	return nil
}

func (e *Event) NotifySubscriber(ef EventFunc, value interface{}) {
	if ef == nil {
		return
	}
	//invode subscriber event func
	ef(value)

}

//Notify all event subscribers
func (e *Event) NotifyAll() (errs []error) {
	e.m.RLock()
	defer e.m.RUnlock()

	for eventType, _ := range e.subscribers {
		if err := e.Notify(eventType, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
