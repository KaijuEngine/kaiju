/******************************************************************************/
/* event_manager_test.go                                                      */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package events

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	em := NewManager()
	if em == nil {
		t.Fatal("NewManager should return a non-nil manager")
	}
	if em.ChannelCount() != 0 {
		t.Errorf("New manager should have 0 channels, got %d", em.ChannelCount())
	}
}

func TestCreateChannel(t *testing.T) {
	em := NewManager()
	ch := em.CreateChannel("test-channel")

	if ch == nil {
		t.Fatal("CreateChannel should return a non-nil channel")
	}
	if ch.Name() != "test-channel" {
		t.Errorf("Channel name should be 'test-channel', got '%s'", ch.Name())
	}
	if em.ChannelCount() != 1 {
		t.Errorf("Manager should have 1 channel, got %d", em.ChannelCount())
	}

	// Creating the same channel again should return the existing one
	ch2 := em.CreateChannel("test-channel")
	if ch != ch2 {
		t.Error("CreateChannel should return the same channel instance for the same name")
	}
	if em.ChannelCount() != 1 {
		t.Errorf("Manager should still have 1 channel, got %d", em.ChannelCount())
	}
}

func TestSubscribe(t *testing.T) {
	em := NewManager()
	called := false

	sub := em.Subscribe("test-channel", func(data EventData) {
		called = true
	})

	if sub == nil {
		t.Fatal("Subscribe should return a non-nil subscription")
	}
	if sub.Channel() != "test-channel" {
		t.Errorf("Subscription channel should be 'test-channel', got '%s'", sub.Channel())
	}
	if sub.ID() == "" {
		t.Error("Subscription ID should not be empty")
	}

	ch, exists := em.GetChannel("test-channel")
	if !exists {
		t.Fatal("Channel should exist after subscription")
	}
	if ch.SubscriptionCount() != 1 {
		t.Errorf("Channel should have 1 subscription, got %d", ch.SubscriptionCount())
	}

	// Verify callback wasn't called yet
	if called {
		t.Error("Callback should not be called until event is emitted")
	}
}

func TestEmitSync(t *testing.T) {
	em := NewManager()
	callCount := 0

	em.Subscribe("test-channel", func(data EventData) {
		callCount++
		if val, ok := data["test"]; !ok || val != "value" {
			t.Error("Event data should contain 'test' key with value 'value'")
		}
	})

	em.EmitSync("test-channel", EventData{"test": "value"})

	if callCount != 1 {
		t.Errorf("Callback should be called exactly once, got %d", callCount)
	}
}

func TestEmitAsync(t *testing.T) {
	em := NewManager()
	var callCount atomic.Int32
	var wg sync.WaitGroup

	wg.Add(1)
	em.Subscribe("test-channel", func(data EventData) {
		callCount.Add(1)
		wg.Done()
	})

	em.Emit("test-channel", EventData{"test": "async"})

	// Wait for async callback to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if callCount.Load() != 1 {
			t.Errorf("Callback should be called exactly once, got %d", callCount.Load())
		}
	case <-time.After(time.Second):
		t.Fatal("Async callback did not complete within timeout")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	em := NewManager()
	var callCount atomic.Int32
	var wg sync.WaitGroup

	// Subscribe 5 callbacks to the same channel
	for i := 0; i < 5; i++ {
		wg.Add(1)
		em.Subscribe("test-channel", func(data EventData) {
			callCount.Add(1)
			wg.Done()
		})
	}

	em.Emit("test-channel", EventData{"test": "multiple"})

	// Wait for all async callbacks to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if callCount.Load() != 5 {
			t.Errorf("All 5 callbacks should be called, got %d", callCount.Load())
		}
	case <-time.After(time.Second):
		t.Fatal("Async callbacks did not complete within timeout")
	}
}

func TestUnsubscribe(t *testing.T) {
	em := NewManager()
	callCount := 0

	sub := em.Subscribe("test-channel", func(data EventData) {
		callCount++
	})

	// First emit - callback should be called
	em.EmitSync("test-channel", EventData{"test": "first"})
	if callCount != 1 {
		t.Errorf("Callback should be called once, got %d", callCount)
	}

	// Unsubscribe
	em.Unsubscribe(sub)

	ch, _ := em.GetChannel("test-channel")
	if ch.SubscriptionCount() != 0 {
		t.Errorf("Channel should have 0 subscriptions after unsubscribe, got %d", ch.SubscriptionCount())
	}

	// Second emit - callback should not be called
	em.EmitSync("test-channel", EventData{"test": "second"})
	if callCount != 1 {
		t.Errorf("Callback should still be called only once after unsubscribe, got %d", callCount)
	}
}

func TestGetChannel(t *testing.T) {
	em := NewManager()

	// Non-existent channel
	_, exists := em.GetChannel("non-existent")
	if exists {
		t.Error("GetChannel should return false for non-existent channel")
	}

	// Create a channel
	em.CreateChannel("existing")
	ch, exists := em.GetChannel("existing")
	if !exists {
		t.Error("GetChannel should return true for existing channel")
	}
	if ch.Name() != "existing" {
		t.Errorf("Channel name should be 'existing', got '%s'", ch.Name())
	}
}

func TestRemoveChannel(t *testing.T) {
	em := NewManager()
	em.CreateChannel("test-channel")

	if em.ChannelCount() != 1 {
		t.Errorf("Manager should have 1 channel, got %d", em.ChannelCount())
	}

	em.RemoveChannel("test-channel")

	if em.ChannelCount() != 0 {
		t.Errorf("Manager should have 0 channels after removal, got %d", em.ChannelCount())
	}

	_, exists := em.GetChannel("test-channel")
	if exists {
		t.Error("Channel should not exist after removal")
	}
}

func TestClear(t *testing.T) {
	em := NewManager()

	// Create multiple channels with subscriptions
	em.Subscribe("channel1", func(data EventData) {})
	em.Subscribe("channel2", func(data EventData) {})
	em.Subscribe("channel3", func(data EventData) {})

	if em.ChannelCount() != 3 {
		t.Errorf("Manager should have 3 channels, got %d", em.ChannelCount())
	}

	em.Clear()

	if em.ChannelCount() != 0 {
		t.Errorf("Manager should have 0 channels after Clear, got %d", em.ChannelCount())
	}
}

func TestEmitToNonExistentChannel(t *testing.T) {
	em := NewManager()

	// Should not panic or error
	em.Emit("non-existent", EventData{"test": "value"})
	em.EmitSync("non-existent", EventData{"test": "value"})
}

func TestConcurrentSubscribe(t *testing.T) {
	em := NewManager()
	var wg sync.WaitGroup
	subscriberCount := 100

	// Subscribe from multiple goroutines concurrently
	for i := 0; i < subscriberCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.Subscribe("concurrent-channel", func(data EventData) {})
		}()
	}

	wg.Wait()

	ch, exists := em.GetChannel("concurrent-channel")
	if !exists {
		t.Fatal("Channel should exist after concurrent subscriptions")
	}

	if ch.SubscriptionCount() != subscriberCount {
		t.Errorf("Channel should have %d subscriptions, got %d", subscriberCount, ch.SubscriptionCount())
	}
}

func TestConcurrentEmit(t *testing.T) {
	em := NewManager()
	var callCount atomic.Int32
	var wg sync.WaitGroup

	emitCount := 100
	wg.Add(emitCount)

	em.Subscribe("concurrent-emit", func(data EventData) {
		callCount.Add(1)
		wg.Done()
	})

	// Emit from multiple goroutines concurrently
	for i := 0; i < emitCount; i++ {
		go func(idx int) {
			em.Emit("concurrent-emit", EventData{"index": idx})
		}(i)
	}

	// Wait for all callbacks to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if callCount.Load() != int32(emitCount) {
			t.Errorf("Callback should be called %d times, got %d", emitCount, callCount.Load())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Concurrent emit callbacks did not complete within timeout")
	}
}

func TestEventDataTypes(t *testing.T) {
	em := NewManager()

	em.Subscribe("test-types", func(data EventData) {
		// Test string
		if val, ok := data["string"].(string); !ok || val != "test" {
			t.Error("String value should be retrievable")
		}

		// Test int
		if val, ok := data["int"].(int); !ok || val != 42 {
			t.Error("Int value should be retrievable")
		}

		// Test float
		if val, ok := data["float"].(float64); !ok || val != 3.14 {
			t.Error("Float value should be retrievable")
		}

		// Test bool
		if val, ok := data["bool"].(bool); !ok || val != true {
			t.Error("Bool value should be retrievable")
		}

		// Test nested map
		if nested, ok := data["nested"].(map[string]string); !ok {
			t.Error("Nested map should be retrievable")
		} else if nested["key"] != "value" {
			t.Error("Nested map value should be correct")
		}
	})

	em.EmitSync("test-types", EventData{
		"string": "test",
		"int":    42,
		"float":  3.14,
		"bool":   true,
		"nested": map[string]string{"key": "value"},
	})
}

func TestUniqueSubscriptionIDs(t *testing.T) {
	em := NewManager()
	ids := make(map[string]bool)

	// Create multiple subscriptions
	for i := 0; i < 100; i++ {
		sub := em.Subscribe("test-channel", func(data EventData) {})
		if ids[sub.ID()] {
			t.Errorf("Duplicate subscription ID found: %s", sub.ID())
		}
		ids[sub.ID()] = true
	}
}

func BenchmarkEmitSync(b *testing.B) {
	em := NewManager()
	em.Subscribe("bench-channel", func(data EventData) {
		// Minimal work in callback
	})

	data := EventData{"test": "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.EmitSync("bench-channel", data)
	}
}

func BenchmarkEmitAsync(b *testing.B) {
	em := NewManager()
	em.Subscribe("bench-channel", func(data EventData) {
		// Minimal work in callback
	})

	data := EventData{"test": "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Emit("bench-channel", data)
	}
}

func BenchmarkSubscribe(b *testing.B) {
	em := NewManager()
	callback := func(data EventData) {}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Subscribe("bench-channel", callback)
	}
}
