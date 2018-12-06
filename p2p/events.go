package p2p

type eventQueue struct {
	Block *Event
}

func NeweventQueue() *eventQueue {
	Block := NewEvent()
	return &eventQueue{
		Block: Block,
	}
}

func (eq *eventQueue) GetEvent(eventName string) *Event {
	switch eventName {
	case "block":
		return eq.Block
	default:
		logger.Debug("Unknow event")
		return nil
	}
}
