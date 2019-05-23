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
	fn2 := func() {}
	handlers.Add(&eventHandler{
		callBack: reflect.ValueOf(fn),
		pool:     workerpool.New(2),
	})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			handler := &eventHandler{
				callBack: reflect.ValueOf(func() {
					fmt.Println(i)
				}),
				pool: workerpool.New(2),
			}
			handlers.Add(handler)
			fmt.Printf("add %v\n", handler)
		}(i)
	}

	wg.Wait()

	v := reflect.ValueOf(fn)
	flag1 := handlers.RemoveCallback(v)
	if !flag1 {
		t.Fatal("remove failed")
	}

	v2 := reflect.ValueOf(fn2)
	flag2 := handlers.RemoveCallback(v2)
	if flag2 {
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
