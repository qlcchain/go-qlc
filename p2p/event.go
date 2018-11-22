package p2p

import (
	"github.com/qlcchain/go-qlc/common/types"
)

type Event struct {
	MsgType MessageType
	Blocks  types.Block
}
type Observer interface {
	OnNotify(*Event)
}

type Subject interface {
	Regist(Observer)
	Deregist(Observer)
	Notify(*Event)
}

type ConcreteSubject struct {
	Observers map[Observer]struct{}
}

func (cs *ConcreteSubject) Regist(ob Observer) {
	cs.Observers[ob] = struct{}{}
}

func (cs *ConcreteSubject) Deregist(ob Observer) {
	delete(cs.Observers, ob)
}

func (cs *ConcreteSubject) Notify(e *Event) {
	for ob, _ := range cs.Observers {
		ob.OnNotify(e)
	}
}
