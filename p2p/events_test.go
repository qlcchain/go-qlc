package p2p

import (
	"sync"
	"testing"
	"time"
)

var count1, count2 int
var m1, m2 sync.Mutex

func BlockReceiveEvent(v interface{}) {
	m1.Lock()
	defer m1.Unlock()
	count1++
}

func BlockPushEvent(v interface{}) {
	m2.Lock()
	defer m2.Unlock()
	count2++
}

func TestEvents(t *testing.T) {
	count1 = 0
	count2 = 0
	eventQ := NeweventQueue()
	eventQ.Consensus.Subscribe(EventSyncBlock, BlockReceiveEvent)
	sub1 := eventQ.Consensus.Subscribe(EventPublish, BlockPushEvent)
	eventQ.GetEvent("consensus").Notify(EventSyncBlock, "test count1")
	eventQ.GetEvent("consensus").Notify(EventPublish, "test count2")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if count1 != 1 {
		t.Fatal("BlockReceiveType error")
	}
	if count2 != 1 {
		t.Fatal("BlockPushType error")
	}

	eventQ.Consensus.UnSubscribe(EventPublish, sub1)
	eventQ.GetEvent("consensus").Notify(EventSyncBlock, "test count1")
	eventQ.GetEvent("consensus").Notify(EventPublish, "test count2")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if count1 != 2 {
		t.Fatal("UnSubscribe error")
	}
	if count2 != 1 {
		t.Fatal("UnSubscribe error")
	}
}
