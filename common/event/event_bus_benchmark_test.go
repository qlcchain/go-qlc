/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

import (
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
)

func setupTestCase(b *testing.B) (func(b *testing.B), EventBus) {
	bus := New()
	return func(tb *testing.B) {
		bus.WaitAsync()
	}, bus
}

type block struct {
	Hash types.Hash
}

func BenchmarkPublish(b *testing.B) {
	teardownTestCase, bus := setupTestCase(b)
	defer teardownTestCase(b)

	b.ReportAllocs()

	fn := func(blk *block) {
		//b.Log(blk.Index)
	}
	for i := 0; i < 10; i++ {
		_ = bus.Subscribe("block", fn)
	}
	defer teardownTestCase(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish("block", &block{Hash: types.ZeroHash})
	}
}

func BenchmarkPublish2(b *testing.B) {
	teardownTestCase, bus := setupTestCase(b)
	defer teardownTestCase(b)

	b.ReportAllocs()

	fn := func(blk block) {
		//b.Log(blk.Index)
	}
	for i := 0; i < 10; i++ {
		_ = bus.Subscribe("block", fn)
	}
	defer teardownTestCase(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish("block", block{Hash: types.ZeroHash})
	}
}

func BenchmarkPublish3(b *testing.B) {
	teardownTestCase, bus := setupTestCase(b)
	defer teardownTestCase(b)

	b.ReportAllocs()

	fn := func(i int) {
		//b.Log(blk.Index)
	}
	for i := 0; i < 10; i++ {
		_ = bus.Subscribe("index", fn)
	}
	defer teardownTestCase(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish("index", i)
	}
}

func BenchmarkSubscribe_Block(b *testing.B) {
	teardownTestCase, bus := setupTestCase(b)
	defer teardownTestCase(b)

	b.ReportAllocs()

	fn := func(blk block) {
		time.Sleep(10 * time.Millisecond)
	}
	for i := 0; i < 10; i++ {
		_ = bus.Subscribe("block", fn)
	}
	defer teardownTestCase(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish("block", block{Hash: types.ZeroHash})
	}
}

func BenchmarkAsync(b *testing.B) {
	teardownTestCase, bus := setupTestCase(b)
	defer teardownTestCase(b)

	b.ReportAllocs()
	fn := func(blk block) {
		//b.Log(blk.Index)
	}
	for i := 0; i < 1000; i++ {
		_ = bus.SubscribeAsync("block", fn, false)
	}
	defer teardownTestCase(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Publish("block", block{Hash: types.ZeroHash})
	}
}

func BenchmarkAsync_Block(b *testing.B) {
	teardownTestCase, bus := setupTestCase(b)
	defer teardownTestCase(b)

	b.ReportAllocs()
	fn := func(blk block) {
		time.Sleep(10 * time.Millisecond)
		//b.Log(blk.Index)
	}
	for i := 0; i < 1000; i++ {
		_ = bus.SubscribeAsync("block", fn, true)
	}
	defer teardownTestCase(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Publish("block", block{Hash: types.ZeroHash})
	}
}
