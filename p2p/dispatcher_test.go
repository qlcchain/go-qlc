package p2p

import (
	"sync"
	"testing"
)

func TestDispatcher(t *testing.T) {
	dp := NewDispatcher()
	sb := NewSubscriber("", make(chan *Message, 128), false, "test")
	types := sb.MessageType()
	dp.Register(sb)
	mt, _ := dp.subscribersMap.Load(types)
	if mt == nil {
		t.Fatal("register fail")
	}
	dp.Deregister(sb)
	mt2, _ := dp.subscribersMap.Load(types)
	s, _ := mt2.(*sync.Map).Load(sb)
	if s != nil {
		t.Fatal("deregister fail")
	}

}
