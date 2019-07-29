/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	bus := New()
	if bus == nil {
		t.Log("New EventBus not created!")
		t.Fail()
	}
}

func TestSubscribe(t *testing.T) {
	bus := NewEventBus(runtime.NumCPU())

	if bus.Subscribe("test", func() {}) != nil {
		t.Fail()
	}

	if bus.Subscribe("test", 2) == nil {
		t.Fail()
	}
}

func TestSubscribeSync(t *testing.T) {
	bus := NewEventBus(runtime.NumCPU())

	counter := int64(0)
	topic := common.TopicType("test")
	if bus.SubscribeSync(topic, func() {
		atomic.AddInt64(&counter, 1)
		t.Log("sub1")
	}) != nil {
		t.Fail()
	}
	if bus.Subscribe(topic, func() {
		t.Log("sub2")
	}) != nil {
		t.Fail()
	}

	bus.Publish(topic)

	if counter != int64(1) {
		t.Fatal("invalid ", counter)
	}

	_ = bus.Close()
}

func TestUnsubscribe(t *testing.T) {
	bus := NewEventBus(runtime.NumCPU())

	handler := func() {}

	_ = bus.Subscribe("test", handler)

	if err := bus.Unsubscribe("test", handler); err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if err := bus.Unsubscribe("unexisted", func() {}); err == nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestUnsubscribe2(t *testing.T) {
	bus := SimpleEventBus()

	handler := func() {}

	_ = bus.Subscribe("test", handler)

	t.Log(bus.(*DefaultEventBus).handlers.Len())
	if value, ok := bus.(*DefaultEventBus).handlers.GetStringKey("test"); ok {
		t.Log(value.(*eventHandlers).Size())
	}

	if err := bus.Unsubscribe("test", handler); err != nil {
		fmt.Println(err)
		t.Fail()
	}
	t.Log(bus.(*DefaultEventBus).handlers.Len())
	if value, ok := bus.(*DefaultEventBus).handlers.GetStringKey("test"); ok {
		t.Log(value.(*eventHandlers).Size())
	}
	if err := bus.Unsubscribe("unexisted", func() {}); err == nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestClose(t *testing.T) {
	bus := NewEventBus(runtime.NumCPU())

	handler := func() {}

	_ = bus.Subscribe("test", handler)

	original, ok := bus.(*DefaultEventBus)
	if !ok {
		fmt.Println("Could not cast message bus to its original type")
		t.Fail()
	}

	if 0 == original.handlers.Len() {
		fmt.Println("Did not subscribed handler to topic")
		t.Fail()
	}

	bus.CloseTopic("test")

	if 0 != original.handlers.Len() {
		fmt.Println("Did not unsubscribed handlers from topic")
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	bus := NewEventBus(runtime.NumCPU())

	var wg sync.WaitGroup
	wg.Add(2)

	first := false
	second := false

	_ = bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		first = v
	})

	_ = bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		second = v
	})

	bus.Publish("topic", true)

	wg.Wait()

	if first == false || second == false {
		t.Fatal(first, second)
	}
}

func TestHandleError(t *testing.T) {
	bus := NewEventBus(runtime.NumCPU())
	_ = bus.Subscribe("topic", func(out chan<- error) {
		out <- errors.New("I do throw error")
	})

	out := make(chan error)
	defer close(out)

	bus.Publish("topic", out)

	if <-out == nil {
		t.Fail()
	}
}

func TestHasCallback(t *testing.T) {
	bus := New()
	err := bus.Subscribe("topic", func() {})
	if err != nil {
		t.Fatal(err)
	}
	if bus.HasCallback("topic_topic") {
		t.Fail()
	}
	if !bus.HasCallback("topic") {
		t.Fail()
	}
}

func TestGetEventBus(t *testing.T) {
	eb0 := SimpleEventBus()
	eb1 := GetEventBus("")
	if eb0 != eb1 {
		t.Fatal("invalid default eb")
	}

	id1 := "111111"
	eb2 := GetEventBus(id1)
	eb3 := GetEventBus(id1)

	if eb2 != eb3 {
		t.Fatal("invalid eb of same id")
	}

	id2 := "222222"
	eb4 := GetEventBus(id2)
	if eb3 == eb4 {
		t.Fatal("invalid eb of diff ids")
	}
}

func TestEventSubscribe(t *testing.T) {
	bus := NewEventBus(runtime.NumCPU())
	topic := common.TopicType("test")

	counter := int64(0)
	_ = bus.Subscribe(topic, func(i int64) {
		fmt.Println("sub1", i, atomic.AddInt64(&counter, 1))
	})

	_ = bus.Subscribe(topic, func(i int64) {
		time.Sleep(time.Second)
		fmt.Println("sub2", i, atomic.AddInt64(&counter, 1))
	})

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			fmt.Println("publish: ", i)
			bus.Publish(topic, int64(i))
		}
	}()
	wg.Wait()
	_ = bus.Close()

	//if atomic.LoadInt64(&counter) != 10 {
	//	t.Fatal("invalid sub", atomic.LoadInt64(&counter))
	//}
	t.Log("result", atomic.LoadInt64(&counter))
}
