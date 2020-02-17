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
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
)

const testTopic = "test"

type testMessage struct {
	Message string
}

func testSubReceive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *testMessage:
		fmt.Printf("PID:%s receive message:%s\n", c.Self().Id, msg.Message)
	}
}

func TestNewActorSubscriber(t *testing.T) {
	subPID1, _ := actor.EmptyRootContext.SpawnNamed(actor.PropsFromFunc(testSubReceive), "sub1")
	subPID2, _ := actor.EmptyRootContext.SpawnNamed(actor.PropsFromFunc(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *testMessage:
			t.Logf("PID:%s receive message after sleep:%s\n", c.Self().Id, msg.Message)
		}
	}), "sub2")
	sub1 := NewActorSubscriber(subPID1)
	sub2 := NewActorSubscriber(subPID2)
	_ = sub1.Subscribe(testTopic)
	_ = sub2.Subscribe(testTopic)

	DefaultActorBus.Publish(testTopic,
		&testMessage{
			Message: "hello",
		},
	)

	DefaultActorBus.Publish(testTopic,
		&testMessage{
			Message: "world",
		},
	)
	_ = DefaultActorBus.Close()
}

func TestLimitEventBus(t *testing.T) {
	sub := NewActorSubscriber(SpawnWithPool(func(ctx actor.Context) {
		if msg, ok := ctx.Message().(*testMessage); ok {
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		}
	}), DefaultActorBus)

	if err := sub.Subscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		DefaultActorBus.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	_ = DefaultActorBus.Close()
}

func TestActorEventBus_LongTimeTask(t *testing.T) {
	eb := NewActorEventBus()
	wg := sync.WaitGroup{}
	sub := NewActorSubscriber(Spawn(func(ctx actor.Context) {
		if msg, ok := ctx.Message().(*testMessage); ok {
			defer func() {
				wg.Done()
			}()
			t1 := time.Now()
			time.Sleep(2 * time.Second)
			fmt.Printf("%v got message %s cost %s\n", ctx.Self(), msg.Message, time.Since(t1))
		}
	}), eb)

	if err := sub.Subscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	wg.Wait()

	_ = eb.Close()
}

func TestActorEventBus_LongTimeTask2(t *testing.T) {
	eb := NewActorEventBus()
	wg := sync.WaitGroup{}
	sub := NewActorSubscriber(SpawnWithPool(func(ctx actor.Context) {
		if msg, ok := ctx.Message().(*testMessage); ok {
			defer func() {
				wg.Done()
			}()
			t1 := time.Now()
			time.Sleep(2 * time.Second)
			fmt.Printf("%v got message %s cost %s\n", ctx.Self(), msg.Message, time.Since(t1))
		}
	}), eb)

	if err := sub.Subscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	wg.Wait()

	_ = eb.Close()
}

func TestNewActorEventBus(t *testing.T) {
	eb := NewActorEventBus()
	sub := NewActorSubscriber(Spawn(func(ctx actor.Context) {
		switch msg := ctx.Message().(type) {
		case *testMessage:
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		case map[string]string:
			fmt.Printf("%v: %s\n", ctx.Self(), msg)
		case *types.Tuple:
			fmt.Printf("%v: %d,%s\n", ctx.Self(), msg.First.(int), msg.Second.(string))
		}
	}), eb)

	if err := sub.Subscribe(testTopic, "hahaha"); err != nil {
		t.Fatal(err)
	}

	eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(100)})
	eb.Publish(testTopic, map[string]string{"k1": "v1", "k2": "v2"})
	eb.Publish(testTopic, types.NewTuple(10, "hahaha"))

	_ = eb.Close()
}

func TestActorEventBus_HasCallback(t *testing.T) {
	eb := NewActorEventBus()
	pid := Spawn(func(ctx actor.Context) {
		switch msg := ctx.Message().(type) {
		case *testMessage:
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		case map[string]string:
			fmt.Printf("%v: %s\n", ctx.Self(), msg)
		case *types.Tuple:
			fmt.Printf("%v: %d,%s\n", ctx.Self(), msg.First.(int), msg.Second.(string))
		}
	})

	sub := NewActorSubscriber(pid, eb)

	topic2 := topic.TopicType("hahaha")
	if err := sub.Subscribe(testTopic, topic2); err != nil {
		t.Fatal(err)
	}

	if b := eb.HasCallback(testTopic); !b {
		t.Fatal("invalid ", testTopic)
	}

	if b := eb.HasCallback("dfdaf"); b {
		t.Fatal("invalid topic check ")
	}

	for i := 0; i < 10; i++ {
		eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	eb.Publish(testTopic, map[string]string{"k1": "v1", "k2": "v2"})
	eb.Publish(testTopic, types.NewTuple(10, "hahaha"))

	if err := eb.CloseTopic(testTopic); err != nil {
		t.Fatal(err)
	}

	if b := eb.HasCallback(testTopic); b {
		t.Fatal("invalid close topic ", testTopic)
	}

	if err := sub.SubscribeOne(testTopic, pid); err != nil {
		t.Fatal(err)
	}

	if err := eb.Unsubscribe(testTopic, pid); err != nil {
		t.Fatal(err)
	}

	_ = eb.Close()
}

func Test_joinErrs(t *testing.T) {
	t.Log(joinErrs(fmt.Errorf("failed %d", 100), errors.New("hahaha")))
}

func TestActorEventBus_Subscribe(t *testing.T) {
	eb := NewActorEventBus()
	sub := NewActorSubscriber(Spawn(func(ctx actor.Context) {
		if msg, ok := ctx.Message().(*testMessage); ok {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		}
	}), eb)

	if err := sub.Subscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	sub.WithSubscribe(Spawn(func(ctx actor.Context) {
		if msg, ok := ctx.Message().(*testMessage); ok {
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		}
	}))

	if err := sub.Subscribe("hahah"); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	if err := sub.Unsubscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	if err := sub.UnsubscribeAll(); err != nil {
		t.Fatal(err)
	}

	_ = eb.Close()
}

func TestNewPublisher(t *testing.T) {
	type pingMsg struct {
		Message string
	}

	type pongMsg struct {
		Message string
	}
	eb := NewActorEventBus()

	if err := NewActorSubscriber(Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *actor.Started:
			c.SetReceiveTimeout(100 * time.Millisecond)
		case *pingMsg:
			fmt.Println("publisher req message: " + msg.Message)
			c.Respond(&pongMsg{Message: "hello " + msg.Message})
		}
	}), eb).Subscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	signal := make(chan interface{})
	publisher := Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *pongMsg:
			fmt.Println("subscriber resp message: ", msg.Message)
			signal <- struct{}{}
		}
	})
	actorPublisher := NewActorPublisher(publisher, eb)
	if err := actorPublisher.Publish(testTopic, &pingMsg{Message: "qlcchain"}); err != nil {
		t.Fatal(err)
	}

	<-signal
	_ = eb.Close()
}

func TestNewPublisher2(t *testing.T) {
	type pingMsg struct {
		Message string
	}

	type pongMsg struct {
		Message string
	}
	eb := NewActorEventBus()

	if err := NewActorSubscriber(Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		//case *actor.Started:
		//	c.SetReceiveTimeout(100 * time.Millisecond)
		case *pingMsg:
			fmt.Println("publisher req message: " + msg.Message)
			c.Respond(&pongMsg{Message: "hello " + msg.Message})
		}
	}), eb).Subscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	publisher := NewActorPublisher(nil, eb)
	publisher.PublishFuture(testTopic, &pingMsg{Message: "qlcchain"}, func(msg interface{}, err error) {
		if err == nil {
			switch msg := msg.(type) {
			case *pongMsg:
				fmt.Println("subscriber resp message: ", msg.Message)
			}
		} else {
			t.Error(err)
		}
	})

	_ = eb.Close()
}
