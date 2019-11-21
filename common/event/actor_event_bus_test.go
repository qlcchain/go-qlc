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
	"go.uber.org/atomic"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"

	"github.com/AsynkronIT/protoactor-go/actor"
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

func TestSort(t *testing.T) {
	type tmp struct {
		i int
		b bool
	}
	var subs []*tmp
	for i := 0; i < 10; i++ {
		subs = append(subs, &tmp{
			i: i,
			b: i%2 == 0,
		})
	}

	sort.Slice(subs, func(i, j int) bool {
		return bool2Int(subs[i].b) > bool2Int(subs[j].b)
	})

	for _, v := range subs {
		t.Log(v)
	}
}

func TestActorEventBus_SubscribeSync(t *testing.T) {
	eb := NewActorEventBus()
	sub := NewActorSubscriber(Spawn(func(ctx actor.Context) {
		if msg, ok := ctx.Message().(*testMessage); ok {
			time.Sleep(100 * time.Millisecond)
			// this is guaranteed to only execute with a max concurrency level of `maxConcurrency`
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		}
	}), eb).WithTimeout(time.Second)

	if err := sub.SubscribeSync(testTopic); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	_ = eb.Close()
}

func TestActorEventBus_SubscribeSyncTimeout(t *testing.T) {
	eb := NewActorEventBus()
	var i atomic.Int32
	sub := NewActorSubscriber(Spawn(func(ctx actor.Context) {
		switch msg := ctx.Message().(type) {
		case *actor.Started:
			ctx.SetReceiveTimeout(100 * time.Millisecond)
		case *actor.ReceiveTimeout:
			fmt.Println("ReceiveTimeout: ")
		case *testMessage:
			time.Sleep(time.Second)
			i.Add(1)
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		}
	}), eb).WithTimeout(100 * time.Millisecond)

	if err := sub.SubscribeSync(testTopic); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	if i.Load() == 10 {
		t.Fatal("timeout failed")
	}

	//_ = eb.Close()
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

	for i := 0; i < 100; i++ {
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

	if err := eb.SubscribeSync("xxxx", pid); err != nil {
		t.Fatal(err)
	}

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

func TestNewActorSubscriber1(t *testing.T) {
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
	sub := NewActorSubscriber(nil, eb).WithTimeout(time.Second * 10).WithSubscribe(pid)
	t1 := topic.TopicType("haha1")
	t2 := topic.TopicType("haha2")

	if err := sub.SubscribeOne(t1, pid); err != nil {
		t.Fatal(err)
	}
	if b := eb.HasCallback(t1); !b {
		t.Fatal(t1, " callback check failed")
	}
	if b := eb.HasCallback(t2); b {
		t.Fatal(t2, " callback check failed")
	}

	if err := sub.SubscribeSyncOne(t2, pid); err != nil {
		t.Fatal(err)
	}

	if err := sub.SubscribeSyncOne(t1, pid); err != nil {
		t.Fatal(err)
	}

	if err := sub.SubscribeSync(t1); err != nil {
		t.Fatal(err)
	}

	if len(sub.subscribers) != 4 {
		t.Fatal("invalid sub len")
	}

	if err := sub.Unsubscribe(t1); err != nil {
		t.Fatal(err)
	}

	if len(sub.subscribers) != 1 {
		t.Fatal("invalid sub len after Unsubscribe")
	}

	if err := sub.UnsubscribeAll(); err != nil {
		t.Fatal(err)
	}

	if len(sub.subscribers) != 0 {
		t.Fatal("invalid sub len after UnsubscribeAll")
	}
}

func Test_joinErrs(t *testing.T) {
	t.Log(joinErrs(fmt.Errorf("failed %d", 100), errors.New("hahaha")))
}
