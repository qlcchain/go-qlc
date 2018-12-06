package p2p

import (
	"sync"
	"testing"
	"time"
)

const BlockReceiveType EventType = 0
const BlockPushType EventType = 1

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
	eventQ.Block.Subscribe(BlockReceiveType, BlockReceiveEvent)
	sub1 := eventQ.Block.Subscribe(BlockPushType, BlockPushEvent)
	eventQ.GetEvent("block").Notify(BlockReceiveType, "test count1")
	eventQ.GetEvent("block").Notify(BlockPushType, "test count2")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if count1 != 1 {
		t.Fatal("BlockReceiveType error")
	}
	if count2 != 1 {
		t.Fatal("BlockPushType error")
	}

	eventQ.Block.UnSubscribe(BlockPushType, sub1)
	eventQ.GetEvent("block").Notify(BlockReceiveType, "test count1")
	eventQ.GetEvent("block").Notify(BlockPushType, "test count2")
	time.Sleep(time.Duration(1) * time.Millisecond)
	if count1 != 2 {
		t.Fatal("UnSubscribe error")
	}
	if count2 != 1 {
		t.Fatal("UnSubscribe error")
	}
}
