/******************************************************************************/
/* event_manager.go                                                           */
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
	"fmt"
	"sync"
)

// EventData represents data carried with an event. It provides a flexible
// key-value structure for passing arbitrary data between event publishers
// and subscribers.
type EventData map[string]interface{}

// CallbackFunc is the signature for event callback functions
type CallbackFunc func(EventData)

// Subscription represents a single subscription to an event channel.
// It holds the subscription ID, channel name, and the callback function
// to be executed when events are emitted on the channel.
type Subscription struct {
	id       string
	channel  string
	callback CallbackFunc
}

// ID returns the unique identifier for this subscription
func (s *Subscription) ID() string { return s.id }

// Channel returns the name of the channel this subscription is attached to
func (s *Subscription) Channel() string { return s.channel }

// Channel represents an event channel that manages multiple subscriptions.
// Events emitted to a channel are delivered to all active subscribers.
// Channel operations are thread-safe through mutex locking.
type Channel struct {
	name          string
	subscriptions map[string]*Subscription
	mutex         sync.RWMutex
}

// newChannel creates a new event channel with the given name
func newChannel(name string) *Channel {
	return &Channel{
		name:          name,
		subscriptions: make(map[string]*Subscription),
	}
}

// Name returns the name of this channel
func (c *Channel) Name() string { return c.name }

// SubscriptionCount returns the number of active subscriptions on this channel
func (c *Channel) SubscriptionCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.subscriptions)
}

// Emit sends an event to all subscribers on this channel asynchronously.
// Each callback is executed in its own goroutine to prevent blocking.
func (c *Channel) Emit(data EventData) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, sub := range c.subscriptions {
		// Execute each callback in a separate goroutine (non-blocking)
		go sub.callback(data)
	}
}

// EmitSync sends an event to all subscribers on this channel synchronously.
// Callbacks are executed sequentially in the calling goroutine.
func (c *Channel) EmitSync(data EventData) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, sub := range c.subscriptions {
		sub.callback(data)
	}
}

// addSubscription adds a new subscription to this channel
func (c *Channel) addSubscription(sub *Subscription) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.subscriptions[sub.id] = sub
}

// removeSubscription removes a subscription from this channel by ID
func (c *Channel) removeSubscription(subID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.subscriptions, subID)
}

// Manager manages all event channels and subscriptions in the system.
// It provides a centralized way to create channels, subscribe to events,
// and emit events across different parts of the application.
type Manager struct {
	channels       map[string]*Channel
	mutex          sync.RWMutex
	subscriptionID int
}

// NewManager creates a new event manager instance
func NewManager() *Manager {
	return &Manager{
		channels: make(map[string]*Channel),
	}
}

// CreateChannel creates a new event channel with the given name.
// If a channel with this name already exists, it returns the existing channel.
func (em *Manager) CreateChannel(name string) *Channel {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if ch, exists := em.channels[name]; exists {
		return ch
	}

	ch := newChannel(name)
	em.channels[name] = ch
	return ch
}

// Subscribe subscribes a callback function to an event channel.
// If the channel does not exist, it will be created automatically.
// Returns a Subscription object that can be used to unsubscribe later.
//
// Example usage:
//
//	sub := host.EventManager().Subscribe("collision", func(data EventData) {
//	    // Handle collision event
//	})
func (em *Manager) Subscribe(channelName string, callback CallbackFunc) *Subscription {
	em.mutex.Lock()
	em.subscriptionID++
	subID := fmt.Sprintf("sub_%d", em.subscriptionID)
	em.mutex.Unlock()

	// Create channel if it doesn't exist
	ch := em.CreateChannel(channelName)

	sub := &Subscription{
		id:       subID,
		channel:  channelName,
		callback: callback,
	}

	ch.addSubscription(sub)
	return sub
}

// Unsubscribe cancels a subscription, removing it from its channel
func (em *Manager) Unsubscribe(sub *Subscription) {
	em.mutex.RLock()
	ch, exists := em.channels[sub.channel]
	em.mutex.RUnlock()

	if exists {
		ch.removeSubscription(sub.id)
	}
}

// Emit sends an event to all subscribers of the specified channel asynchronously.
// If the channel does not exist, this is a no-op.
func (em *Manager) Emit(channelName string, data EventData) {
	em.mutex.RLock()
	ch, exists := em.channels[channelName]
	em.mutex.RUnlock()

	if exists {
		ch.Emit(data)
	}
}

// EmitSync sends an event to all subscribers of the specified channel synchronously.
// If the channel does not exist, this is a no-op.
func (em *Manager) EmitSync(channelName string, data EventData) {
	em.mutex.RLock()
	ch, exists := em.channels[channelName]
	em.mutex.RUnlock()

	if exists {
		ch.EmitSync(data)
	}
}

// GetChannel retrieves an existing channel by name.
// Returns the channel and true if it exists, nil and false otherwise.
func (em *Manager) GetChannel(name string) (*Channel, bool) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	ch, exists := em.channels[name]
	return ch, exists
}

// RemoveChannel deletes a channel and all its subscriptions
func (em *Manager) RemoveChannel(name string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	delete(em.channels, name)
}

// ChannelCount returns the number of active channels
func (em *Manager) ChannelCount() int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	return len(em.channels)
}

// Clear removes all channels and subscriptions from the manager
func (em *Manager) Clear() {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.channels = make(map[string]*Channel)
	em.subscriptionID = 0
}
