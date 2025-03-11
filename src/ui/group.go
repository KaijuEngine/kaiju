/******************************************************************************/
/* group.go                                                                   */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package ui

import (
	"kaiju/engine"
	"log/slog"
	"sort"
	"sync"
)

type requestState = int

const (
	requestStateNone = requestState(iota)
	requestStateStarted
	requestStateProcessed
)

type groupRequest struct {
	target    *UI
	eventType EventType
}

type Group struct {
	requests    []groupRequest
	focus       *UI
	updateId    int
	lock        sync.Mutex
	hadRequests requestState
	isThreaded  bool
}

func NewGroup() *Group {
	return &Group{
		requests: make([]groupRequest, 0),
		focus:    nil,
	}
}

func (group *Group) HasRequests() bool {
	return group.hadRequests != requestStateNone || group.focus != nil
}

func (group *Group) requestEvent(ui *UI, eType EventType) {
	if eType < EventTypeInvalid || eType >= EventTypeEnd {
		slog.Error("Invalid UI event type")
		return
	}
	if group.hadRequests != requestStateStarted {
		if eType != EventTypeMiss {
			group.lock.Lock()
			group.hadRequests = requestStateStarted
			group.lock.Unlock()
		}
	}
	if ui.events[eType].IsEmpty() {
		return
	}
	if group.isThreaded {
		group.lock.Lock()
	}
	group.requests = append(group.requests, groupRequest{
		target:    ui,
		eventType: eType,
	})
	if group.isThreaded {
		group.lock.Unlock()
	}
}

func (group *Group) setFocus(ui *UI) {
	if group.focus != nil && group.focus != ui {
		group.focus.ExecuteEvent(EventTypeMiss)
	}
	group.focus = ui
	if group.focus != nil {
		group.focus.ExecuteEvent(EventTypeClick)
	}
}

func (group *Group) Attach(host *engine.Host) {
	group.updateId = host.UILateUpdater.AddUpdate(func(dt float64) {
		group.lateUpdate()
	})
}

func (group *Group) Detach(host *engine.Host) {
	host.UILateUpdater.RemoveUpdate(group.updateId)
	group.updateId = -1
}

func sortRequests(a *groupRequest, b *groupRequest) bool {
	return a.target.Entity().Transform.WorldPosition().Z() >
		b.target.Entity().Transform.WorldPosition().Z()
}

func (group *Group) lateUpdate() {
	if len(group.requests) > 0 {
		sort.Slice(group.requests, func(i, j int) bool {
			return sortRequests(&group.requests[i], &group.requests[j])
		})
		requestSets := [EventTypeEnd][]groupRequest{}
		for i := range group.requests {
			r := &group.requests[i]
			requestSets[r.eventType] = append(requestSets[r.eventType], *r)
		}
		for i := range requestSets {
			g := requestSets[i]
			shouldContinue := true
			for j := range g {
				req := &g[j]
				if shouldContinue {
					switch req.eventType {
					case EventTypeEnter, EventTypeExit, EventTypeMiss,
						EventTypeKeyDown, EventTypeKeyUp, EventTypeChange,
						EventTypeSubmit:
						req.target.ExecuteEvent(req.eventType)
					case EventTypeClick, EventTypeRightClick, EventTypeDoubleClick, EventTypeDown, EventTypeUp,
						EventTypeDropEnter, EventTypeDropExit, EventTypeDragStart,
						EventTypeDragEnd, EventTypeDrop, EventTypeScroll:
						if req.target.ExecuteEvent(req.eventType) {
							shouldContinue = false
						}
					default:
						slog.Error("Invalid UI event type")
						return
					}
				}
			}
		}
		group.requests = group.requests[:0]
	}
	switch group.hadRequests {
	case requestStateStarted:
		group.hadRequests = requestStateProcessed
	case requestStateProcessed:
		group.hadRequests = requestStateNone
	}
}

func (g *Group) SetThreaded() {
	if !g.isThreaded {
		g.isThreaded = true
	}
}
