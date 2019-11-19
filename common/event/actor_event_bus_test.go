/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

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
			fmt.Printf("PID:%s receive message after sleep:%s\n", c.Self().Id, msg.Message)
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
	sub := NewActorSubscriber(Spawn(func(ctx actor.Context) {
		if msg, ok := ctx.Message().(*testMessage); ok {
			// this is guaranteed to only execute with a max concurrency level of `maxConcurrency`
			fmt.Printf("%v got message %s\n", ctx.Self(), msg.Message)
		}
	}), DefaultActorBus)

	if err := sub.Subscribe(testTopic); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
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
	}), eb)

	if err := sub.SubscribeSync(testTopic); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		eb.Publish(testTopic, &testMessage{Message: strconv.Itoa(i)})
	}

	_ = eb.Close()
}
