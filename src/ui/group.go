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
