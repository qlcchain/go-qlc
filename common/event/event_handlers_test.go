/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/gammazero/workerpool"
)

func TestEventHandler(t *testing.T) {
	handlers := newEventHandler()

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				//time.Sleep(time.Millisecond)
				fmt.Println("size: ", handlers.Size())
			}
		}
	}(ctx)

	wg := sync.WaitGroup{}

	fn := func() {}
	handler1 := &eventHandler{
		callBack: reflect.ValueOf(fn),
		pool:     workerpool.New(2),
		id:       uuid.New().String(),
	}
	handlers.Add(handler1)

	fnId1 := handler1.id

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			handler := &eventHandler{
				callBack: reflect.ValueOf(func() {
					fmt.Println(i)
				}),
				pool: workerpool.New(2),
				id:   uuid.New().String(),
			}
			handlers.Add(handler)
			fmt.Printf("add %v\n", handler)
		}(i)
	}

	wg.Wait()

	if err := handlers.RemoveCallback(fnId1); err != nil {
		t.Fatal("remove fn failed")
	}

	if err := handlers.RemoveCallback("v2"); err == nil {
		t.Fatal("remove failed")
	}

	cache := handlers.All()
	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			flag := handlers.Remove(cache[i])
			t.Logf("remove %d, %v ==> %t\n", i, cache[i], flag)
		}(i)
	}

	wg.Wait()

	t.Log("finally ", handlers.Size())

	cancel()
}
