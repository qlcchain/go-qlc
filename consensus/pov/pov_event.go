package pov

import (
	"sync"

	"github.com/qlcchain/go-qlc/common/types"
)

type EventListener interface {
	OnPovBlockEvent(event byte, block *types.PovBlock)
}

const (
	EventConnectPovBlock    = byte(1)
	EventDisconnectPovBlock = byte(2)
)

type eventManager struct {
	listenerList []EventListener

	maxHandlerID uint32
	mu           sync.Mutex
}

func newEventManager() *eventManager {
	return &eventManager{
		maxHandlerID: 0,
		listenerList: make([]EventListener, 0),
	}
}

func (em *eventManager) RegisterListener(listener EventListener) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.listenerList = append(em.listenerList, listener)
}

func (em *eventManager) UnRegisterListener(listener EventListener) {
	em.mu.Lock()
	defer em.mu.Unlock()

	for index, ln := range em.listenerList {
		if ln == listener {
			em.listenerList = append(em.listenerList[:index], em.listenerList[index+1:]...)
			break
		}
	}
}

func (em *eventManager) TriggerBlockEvent(event byte, block *types.PovBlock) {
	for _, listener := range em.listenerList {
		listener.OnPovBlockEvent(event, block)
	}
}

func (bc *PovBlockChain) RegisterListener(listener EventListener) {
	bc.em.RegisterListener(listener)
}

func (bc *PovBlockChain) UnRegisterListener(listener EventListener) {
	bc.em.UnRegisterListener(listener)
}
