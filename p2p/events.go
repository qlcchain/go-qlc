package p2p

type EventQueue struct {
	Consensus *Event
}

func NewEventQueue() *EventQueue {
	consensus := NewEvent()
	return &EventQueue{
		Consensus: consensus,
	}
}

func (eq *EventQueue) GetEvent(eventName string) *Event {
	switch eventName {
	case "consensus":
		return eq.Consensus
	default:
		return nil
	}
}
