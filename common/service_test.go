/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type key int

const errorKey key = 0

func TestTaskLifecycle(t *testing.T) {
	tl := &ServiceLifecycle{}
	t.Log(tl.Status, tl.String())
	b := tl.PreInit()
	if !b {
		t.Fatal("PreInit error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PostInit()
	if !b {
		t.Fatal("PostInit error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PreStart()
	if !b {
		t.Fatal("PreStart error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PostStart()
	if !b {
		t.Fatal("PostStart error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PreStop()
	if !b {
		t.Fatal("PreStop error")
	}
	t.Log(tl.Status, tl.String())
	b = tl.PostStop()
	if !b {
		t.Fatal("PostStop error")
	}
	t.Log(tl.Status, tl.String())
}

type waitService struct {
	ServiceLifecycle
}

func (w *waitService) Init() error {
	if !w.PreInit() {
		return errors.New("pre init fail")
	}
	defer w.PostInit()
	return nil
}

func (w *waitService) Start() error {
	if !w.PreStart() {
		return errors.New("pre init fail")
	}
	defer w.PostStart()

	time.Sleep(time.Duration(2) * time.Second)
	return nil
}

func (w *waitService) Stop() error {
	if !w.PreStop() {
		return errors.New("pre init fail")
	}
	defer w.PostStop()
	return nil
}

func (w *waitService) Status() int32 {
	return w.State()
}

func TestServiceStatus(t *testing.T) {
	w := &waitService{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				ctx = context.WithValue(ctx, errorKey, "check services timeout")
				return
			default:
				if w.State() == int32(Started) {
					t.Log("started...")
					cancel()
					return
				}
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()

	err := w.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = w.Start()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-ctx.Done():
		fmt.Println(ctx.Err(), ctx.Value(errorKey))
	}
	fmt.Println("done")
}
