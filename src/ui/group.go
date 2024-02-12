/*****************************************************************************/
/* group.go                                                                  */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package ui

import (
	"kaiju/bitmap"
	"kaiju/engine"
	"sort"
)

type groupRequest struct {
	target    UI
	eventType EventType
}

type Group struct {
	requests []groupRequest
	focus    UI
	updateId int
}

func NewGroup() *Group {
	return &Group{
		requests: make([]groupRequest, 0),
		focus:    nil,
	}
}

func (group *Group) requestEvent(ui UI, eType EventType) {
	if eType < EventTypeInvalid || eType >= EventTypeEnd {
		panic("Invalid UI event type")
	}
	group.requests = append(group.requests, groupRequest{
		target:    ui,
		eventType: eType,
	})
}

func (group *Group) setFocus(ui UI) {
	if group.focus != nil {
		group.focus.ExecuteEvent(EventTypeMiss)
	}
	group.focus = ui
	group.focus.ExecuteEvent(EventTypeClick)
}

func (group *Group) attach(host *engine.Host) {
	group.updateId = host.LateUpdater.AddUpdate(func(dt float64) {
		group.lateUpdate()
	})
}

func (group *Group) detach(host *engine.Host) {
	host.LateUpdater.RemoveUpdate(group.updateId)
	group.updateId = -1
}

func sortRequests(a *groupRequest, b *groupRequest) int {
	return (int)((b.target.Entity().Transform.WorldPosition().Z() -
		a.target.Entity().Transform.WorldPosition().Z()) * 1000.0)
}

func (group *Group) lateUpdate() {
	if len(group.requests) > 0 {
		sort.Slice(group.requests, func(i, j int) bool {
			return sortRequests(&group.requests[i], &group.requests[j]) < 0
		})
		available := bitmap.NewTrue(EventTypeEnd)
		for i := 0; i < len(group.requests); i++ {
			req := &group.requests[i]
			if available.Check(req.eventType) {
				shouldContinue := true
				switch req.eventType {
				case EventTypeEnter:
					fallthrough
				case EventTypeExit:
					fallthrough
				case EventTypeMiss:
					req.target.ExecuteEvent(req.eventType)
				case EventTypeClick:
					fallthrough
				case EventTypeDown:
					fallthrough
				case EventTypeUp:
					fallthrough
				case EventTypeScroll:
					if req.target.ExecuteEvent(req.eventType) {
						shouldContinue = false
					}
				default:
					panic("Invalid UI event type")
				}
				available.Assign(req.eventType, shouldContinue)
			}
		}
		group.requests = group.requests[:0]
	}
}
