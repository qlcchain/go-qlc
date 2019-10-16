/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"fmt"
	"strings"
	"sync"
)

type eventHandlers struct {
	handlers []*eventHandler
	locker   sync.RWMutex
}

func newEventHandler() *eventHandlers {
	return &eventHandlers{
		handlers: make([]*eventHandler, 0),
	}
}

func (eh *eventHandlers) Add(handler *eventHandler) {
	eh.locker.Lock()
	defer eh.locker.Unlock()

	eh.handlers = append(eh.handlers, handler)
}

func (eh *eventHandlers) RemoveCallback(handlerID string) error {
	eh.locker.Lock()
	defer eh.locker.Unlock()
	i := len(eh.handlers)
	for idx, h := range eh.handlers {
		if h.id == handlerID {
			i = idx
			break
		}
	}
	if i != len(eh.handlers) {
		eh.handlers[i].pool.StopWait()
		eh.handlers = append(eh.handlers[:i], eh.handlers[i+1:]...)
		return nil
	}
	return fmt.Errorf("remove callback %s failed", handlerID)
}

func (eh *eventHandlers) Remove(handler *eventHandler) bool {
	eh.locker.Lock()
	defer eh.locker.Unlock()
	i := len(eh.handlers)
	for idx, h := range eh.handlers {
		if h.id == handler.id {
			i = idx
			break
		}
	}
	if i != len(eh.handlers) {
		eh.handlers[i].pool.StopWait()
		eh.handlers = append(eh.handlers[:i], eh.handlers[i+1:]...)
		return true
	} else {
		return false
	}
}

func (eh *eventHandlers) Clear() {
	eh.locker.RLock()
	defer eh.locker.RUnlock()
	for _, h := range eh.handlers {
		h.pool.StopWait()
	}
	eh.handlers = eh.handlers[:0]
}

func (eh *eventHandlers) Size() int {
	eh.locker.RLock()
	defer eh.locker.RUnlock()

	return len(eh.handlers)
}

func (eh *eventHandlers) All() []*eventHandler {
	eh.locker.RLock()
	defer eh.locker.RUnlock()
	result := make([]*eventHandler, len(eh.handlers))
	copy(result, eh.handlers)
	return result
}

func (eh *eventHandlers) String() string {
	sb := strings.Builder{}
	for idx, h := range eh.handlers {
		sb.WriteString(fmt.Sprintf("%d: %v,%v ", idx, h.pool, h.callBack))
	}
	return sb.String()
}
