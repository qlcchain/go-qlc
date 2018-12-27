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
func (e *Event) Subscribe(eventtype EventType, eventfunc EventFunc) EventSubscriber {
	e.m.Lock()
	defer e.m.Unlock()

	sub := make(chan interface{})
	_, ok := e.subscribers[eventtype]
	if !ok {
		e.subscribers[eventtype] = make(map[EventSubscriber]EventFunc)
	}
	e.subscribers[eventtype][sub] = eventfunc

	return sub
}

// UnSubscribe removes the specified subscriber
func (e *Event) UnSubscribe(eventtype EventType, subscriber EventSubscriber) (err error) {
	e.m.Lock()
	defer e.m.Unlock()

	subEvent, ok := e.subscribers[eventtype]
	if !ok {
		err = errors.New("No event type.")
		return err
	}

	delete(subEvent, subscriber)
	close(subscriber)

	return nil
}

//Notify subscribers that Subscribe specified event
func (e *Event) Notify(eventtype EventType, value interface{}) (err error) {
	e.m.RLock()
	defer e.m.RUnlock()

	subs, ok := e.subscribers[eventtype]
	if !ok {
		err = errors.New("No event type.")
		return err
	}

	for _, event := range subs {
		e.NotifySubscriber(event, value)
	}
	return nil
}

func (e *Event) NotifySubscriber(eventfunc EventFunc, value interface{}) {
	if eventfunc == nil {
		return
	}
	//invode subscriber event func
	eventfunc(value)

}

//Notify all event subscribers
func (e *Event) NotifyAll() (errs []error) {
	e.m.RLock()
	defer e.m.RUnlock()

	for eventtype, _ := range e.subscribers {
		if err := e.Notify(eventtype, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
