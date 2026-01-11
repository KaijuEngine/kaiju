/******************************************************************************/
/* event_manager_example_test.go                                              */
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

package events_test

import (
	"fmt"
	"kaiju/engine/systems/events"
)

// Example demonstrates basic usage of the event manager system
func Example_basicUsage() {
	// Create a new event manager
	em := events.NewManager()

	// Subscribe to an event channel
	sub := em.Subscribe("player_damaged", func(data events.EventData) {
		playerID := data["playerID"].(string)
		damage := data["damage"].(int)
		fmt.Printf("Player %s took %d damage\n", playerID, damage)
	})

	// Emit an event synchronously
	em.EmitSync("player_damaged", events.EventData{
		"playerID": "player1",
		"damage":   25,
	})

	// Clean up subscription
	em.Unsubscribe(sub)

	// Output:
	// Player player1 took 25 damage
}

// Example showing how to use the event manager with game entities
func Example_entityCommunication() {
	// This example demonstrates how entities can communicate using events
	em := events.NewManager()

	// Entity 1: Player subscribes to collision events
	playerID := "player1"
	playerHealth := 100

	playerSub := em.Subscribe("collision", func(data events.EventData) {
		if targetID, ok := data["targetID"].(string); ok && targetID == playerID {
			if damage, ok := data["damage"].(int); ok {
				playerHealth -= damage
				fmt.Printf("Player health: %d\n", playerHealth)

				// Emit death event if health depletes (synchronous to ensure deterministic output)
				if playerHealth <= 0 {
					em.EmitSync("entity_death", events.EventData{
						"entityID": playerID,
					})
				}
			}
		}
	})

	// UI System subscribes to death events
	uiSub := em.Subscribe("entity_death", func(data events.EventData) {
		entityID := data["entityID"].(string)
		fmt.Printf("Entity died: %s\n", entityID)
	})

	// Simulate collision events
	em.EmitSync("collision", events.EventData{
		"targetID": playerID,
		"damage":   30,
	})

	em.EmitSync("collision", events.EventData{
		"targetID": playerID,
		"damage":   80,
	})

	// Clean up
	em.Unsubscribe(playerSub)
	em.Unsubscribe(uiSub)

	// Output:
	// Player health: 70
	// Player health: -10
	// Entity died: player1
}

// Example showing multiple subscribers to the same event
func Example_multipleSubscribers() {
	em := events.NewManager()

	// Achievement system subscribes to score events
	em.Subscribe("score_changed", func(data events.EventData) {
		score := data["score"].(int)
		if score >= 1000 {
			fmt.Println("Achievement unlocked!")
		}
	})

	// UI system subscribes to score events
	em.Subscribe("score_changed", func(data events.EventData) {
		score := data["score"].(int)
		fmt.Printf("Score: %d\n", score)
	})

	// Analytics system subscribes to score events
	em.Subscribe("score_changed", func(data events.EventData) {
		score := data["score"].(int)
		fmt.Printf("Logged score: %d\n", score)
	})

	// Emit score change - all three subscribers will receive it
	em.EmitSync("score_changed", events.EventData{
		"score": 1500,
	})

	// Output:
	// Achievement unlocked!
	// Score: 1500
	// Logged score: 1500
}

// Example showing channel management
func Example_channelManagement() {
	em := events.NewManager()

	// Create channels explicitly
	gameChannel := em.CreateChannel("game_events")
	uiChannel := em.CreateChannel("ui_events")

	fmt.Printf("Game channel: %s\n", gameChannel.Name())
	fmt.Printf("UI channel: %s\n", uiChannel.Name())
	fmt.Printf("Total channels: %d\n", em.ChannelCount())

	// Subscribe to channels
	em.Subscribe("game_events", func(data events.EventData) {
		fmt.Println("Game event received")
	})

	// Get channel information
	if ch, exists := em.GetChannel("game_events"); exists {
		fmt.Printf("Subscribers on game_events: %d\n", ch.SubscriptionCount())
	}

	// Remove a channel
	em.RemoveChannel("ui_events")
	fmt.Printf("Channels after removal: %d\n", em.ChannelCount())

	// Output:
	// Game channel: game_events
	// UI channel: ui_events
	// Total channels: 2
	// Subscribers on game_events: 1
	// Channels after removal: 1
}

// Example showing how to integrate with Host
func Example_hostIntegration() {
	// This is a conceptual example showing how to use EventManager with Host
	// In actual code, you would access it through host.EventManager()

	em := events.NewManager()

	// Example: Collision detection system emits events
	em.Subscribe("collision_detected", func(data events.EventData) {
		entityA := data["entityA"].(string)
		entityB := data["entityB"].(string)
		fmt.Printf("Collision between %s and %s\n", entityA, entityB)
	})

	// Physics system can emit collision events (synchronous for deterministic example output)
	em.EmitSync("collision_detected", events.EventData{
		"entityA": "player",
		"entityB": "wall",
	})

	// Output:
	// Collision between player and wall
}
