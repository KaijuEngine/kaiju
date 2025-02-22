/******************************************************************************/
/* ui.go                                                                      */
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
	"kaiju/hid"
	"kaiju/matrix"
	"kaiju/pooling"
	"kaiju/rendering"
	"kaiju/systems/events"
	"kaiju/windowing"
)

type DirtyType = int

const (
	DirtyTypeNone = iota
	DirtyTypeLayout
	DirtyTypeResize
	DirtyTypeGenerated
	DirtyTypeColorChange
	DirtyTypeScissor
	DirtyTypeParent
	DirtyTypeParentLayout
	DirtyTypeParentResize
	DirtyTypeParentGenerated
	DirtyTypeParentReGenerated
	DirtyTypeParentColorChange
	DirtyTypeParentScissor
)

type ElementType = uint8

const (
	ElementTypeLabel = ElementType(iota)
	ElementTypePanel
	ElementTypeButton
	ElementTypeCheckbox
	ElementTypeImage
	ElementTypeInput
	ElementTypeProgressBar
	ElementTypeSelect
	ElementTypeSlider
)

type UIElementData interface {
	innerPanelData() *panelData
}

type UI struct {
	man              *Manager
	entity           engine.Entity
	elmData          UIElementData
	events           [EventTypeEnd]events.Event
	group            *Group
	postLayoutUpdate func()
	render           func()
	layout           Layout
	dragStartPos     matrix.Vec3
	downPos          matrix.Vec2
	elmType          ElementType
	dirtyType        DirtyType
	shaderData       *ShaderData
	textureSize      matrix.Vec2
	lastClick        float64
	poolId           pooling.PoolGroupId
	id               pooling.PoolIndex
	hovering         bool
	cantMiss         bool
	isDown           bool
	isRightDown      bool
	drag             bool
	lastActive       bool
	dontClean        bool
}

func (ui *UI) isActive() bool { return ui.entity.IsActive() }

func (ui *UI) init(textureSize matrix.Vec2, anchor Anchor) {
	if ui.postLayoutUpdate == nil {
		ui.postLayoutUpdate = func() {}
	}
	if ui.render == nil {
		ui.render = func() { ui.events[EventTypeRender].Execute() }
	}
	ui.entity.Init(ui.Host().WorkGroup())
	ui.man.Host.AddEntity(&ui.entity)
	ui.shaderData = &ShaderData{
		ShaderDataBase: rendering.NewShaderDataBase(),
	}
	ui.shaderData.Scissor = matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax}
	ui.entity.AddNamedData(EntityDataName, ui)
	ui.textureSize = textureSize
	ui.layout.initialize(ui, anchor)
	rzId := ui.man.Host.Window.OnResize.Add(func() {
		ui.SetDirty(DirtyTypeResize)
	})
	ui.entity.OnDestroy.Add(func() {
		ui.man.Host.Window.OnResize.Remove(rzId)
		ui.events[EventTypeDestroy].Execute()
		ui.elmData = nil
		ui.man.Remove(ui)
	})
}

func (ui *UI) Entity() *engine.Entity          { return &ui.entity }
func (ui *UI) Layout() *Layout                 { return &ui.layout }
func (ui *UI) hasScissor() bool                { return ui.shaderData.Scissor.X() > -matrix.FloatMax }
func (ui *UI) selfScissor() matrix.Vec4        { return ui.shaderData.Scissor }
func (ui *UI) Host() *engine.Host              { return ui.man.Host }
func (ui *UI) dirty() DirtyType                { return ui.dirtyType }
func (ui *UI) ShaderData() *ShaderData         { return ui.shaderData }
func (ui *UI) IsType(elmType ElementType) bool { return ui.elmType == elmType }
func (ui *UI) Type() ElementType               { return ui.elmType }

func (ui *UI) ExecuteEvent(evtType EventType) bool {
	ui.events[evtType].Execute()
	return !ui.events[evtType].IsEmpty()
}

func (ui *UI) AddEvent(evtType EventType, evt func()) events.Id {
	return ui.events[evtType].Add(evt)
}

func (ui *UI) RemoveEvent(evtType EventType, evtId events.Id) {
	ui.events[evtType].Remove(evtId)
}

func (ui *UI) Event(evtType EventType) *events.Event {
	return &ui.events[evtType]
}

func (ui *UI) cleanDirty() { ui.dirtyType = DirtyTypeNone }

func (ui *UI) SetDirty(dirtyType DirtyType) {
	if ui.dirtyType == DirtyTypeNone || ui.dirtyType >= DirtyTypeParent || dirtyType == DirtyTypeGenerated {
		ui.dirtyType = dirtyType
		for i := 0; i < len(ui.entity.Children); i++ {
			kid := ui.entity.Children[i]
			all := AllOnEntity(kid)
			for _, cui := range all {
				if cui.dirty() == DirtyTypeNone || cui.dirty() > DirtyTypeParent {
					// TODO:  Let it know it was from the parent and what type
					if ui.dirtyType < DirtyTypeParent {
						cui.SetDirty(DirtyTypeParent + ui.dirtyType)
					} else {
						cui.SetDirty(ui.dirtyType)
					}
				}
			}
		}
	}
}

func (ui *UI) rootUI() *UI {
	root := &ui.entity
	var rootUI *UI = FirstOnEntity(root)
	for root.Parent != nil {
		if pui := FirstOnEntity(root.Parent); pui != nil {
			root = root.Parent
			rootUI = pui
		} else {
			break
		}
	}
	return rootUI
}

func (ui *UI) Clean() {
	if ui.dontClean {
		return
	}
	root := ui.rootUI()
	tree := []*UI{root}
	var createTree func(target *engine.Entity)
	createTree = func(target *engine.Entity) {
		for _, child := range target.Children {
			cui := FirstOnEntity(child)
			if cui != nil {
				tree = append(tree, cui)
				createTree(child)
			}
		}
	}
	createTree(root.Entity())
	stabilized := false
	maxIterations := 100
	for !stabilized && maxIterations > 0 {
		stabilized = true
		for i := range tree {
			tree[i].cleanDirty()
			tree[i].Layout().update()
			tree[i].postLayoutUpdate()
			stabilized = stabilized && tree[i].dirty() == DirtyTypeNone
		}
		//maxIterations--
	}
	for i := range tree {
		tree[i].GenerateScissor()
		tree[i].render()
	}
}

func (ui *UI) GenerateScissor() {
	target := &ui.entity.Transform
	pos := target.WorldPosition()
	size := target.WorldScale()
	bounds := matrix.Vec4{
		pos.X() - size.X()*0.5,
		pos.Y() - size.Y()*0.5,
		pos.X() + size.X()*0.5,
		pos.Y() + size.Y()*0.5,
	}
	if !ui.entity.IsRoot() {
		p := FirstPanelOnEntity(ui.entity.Parent)
		for p.PanelData().overflow == OverflowVisible && !p.entity.IsRoot() {
			p = FirstPanelOnEntity(p.entity.Parent)
		}
		//if !p.entity.IsRoot() {
		ps := p.Base().selfScissor()
		bounds.SetX(max(bounds.X(), ps.X()))
		bounds.SetY(max(bounds.Y(), ps.Y()))
		bounds.SetZ(min(bounds.Z(), ps.Z()))
		bounds.SetW(min(bounds.W(), ps.W()))
		//}
	}
	ui.setScissor(bounds)
}

func (ui *UI) setScissor(scissor matrix.Vec4) {
	if ui.shaderData.Scissor.Equals(scissor) {
		return
	}
	for i := 0; i < len(ui.entity.Children); i++ {
		cUI := FirstOnEntity(ui.entity.Children[i])
		if cUI != nil {
			cUI.setScissor(scissor)
		}
	}
	ui.shaderData.Scissor = scissor
	me := FirstOnEntity(&ui.entity)
	if me.elmType == ElementTypeLabel {
		ld := me.ToLabel().LabelData()
		for i := range ld.runeDrawings {
			ld.runeDrawings[i].ShaderData.(*rendering.TextShaderData).Scissor = scissor
		}
	}
}

func (ui *UI) requestEvent(evtType EventType) {
	if ui.group != nil {
		ui.group.requestEvent(ui, evtType)
	} else {
		ui.ExecuteEvent(evtType)
	}
}

func (ui *UI) eventUpdates() {
	cursor := &ui.man.Host.Window.Cursor
	mouse := &ui.man.Host.Window.Mouse
	if cursor.Moved() {
		pos := ui.cursorPos(cursor)
		ui.containedCheck(cursor, &ui.entity)
		if ui.isDown && !ui.drag {
			w := ui.Host().Window.Width()
			h := ui.Host().Window.Height()
			wmm, hmm, _ := ui.man.Host.Window.SizeMM()
			threshold := max(windowing.DPI2PX(w, wmm, 1), windowing.DPI2PX(h, hmm, 1))
			if ui.downPos.Distance(pos) > float32(threshold) {
				ui.dragStartPos = ui.entity.Transform.WorldPosition()
				ui.drag = true
				ui.requestEvent(EventTypeDragStart)
			}
		}
	}
	if cursor.Pressed() {
		ui.containedCheck(cursor, &ui.entity)
		if ui.hovering && !ui.isDown {
			ui.isDown = true
			ui.downPos = ui.cursorPos(cursor)
			ui.requestEvent(EventTypeDown)
			ui.cantMiss = true
		}
	}
	if mouse.Pressed(hid.MouseButtonRight) {
		ui.containedCheck(cursor, &ui.entity)
		if ui.hovering && !ui.isRightDown {
			ui.isRightDown = true
		}
	}
	if cursor.Released() {
		if ui.hovering {
			ui.requestEvent(EventTypeUp)
			if cursor.HasDragData() {
				ui.requestEvent(EventTypeDrop)
			}
		}
		if ui.lastActive {
			if ui.isDown {
				ui.isDown = false
				dragged := false
				if ui.drag {
					p := ui.entity.Transform.WorldPosition()
					dragged = ui.dragStartPos.Distance(p) > 5
				}
				ui.drag = false
				ui.requestEvent(EventTypeDragEnd)
				if ui.hovering && !dragged {
					rt := ui.man.Host.Runtime()
					if rt-ui.lastClick < dblCLickTime && !ui.events[EventTypeDoubleClick].IsEmpty() {
						ui.requestEvent(EventTypeDoubleClick)
						ui.lastClick = 0
					} else {
						ui.requestEvent(EventTypeClick)
						ui.lastClick = rt
					}
				}
			} else if !ui.hovering && !ui.cantMiss {
				ui.requestEvent(EventTypeMiss)
			}
			ui.cantMiss = false
		}
	}
	if mouse.Released(hid.MouseButtonRight) {
		if ui.isRightDown && ui.hovering {
			ui.requestEvent(EventTypeRightClick)
		}
		ui.isRightDown = false
	}
	if mouse.Scrolled() && ui.hovering {
		ui.requestEvent(EventTypeScroll)
	}
}

func (ui *UI) Update(deltaTime float64) {
	if ui.dirtyType != DirtyTypeNone {
		ui.Clean()
	}
	ui.lastActive = ui.entity.IsActive()
}

func (ui *UI) cursorPos(cursor *hid.Cursor) matrix.Vec2 {
	pos := cursor.Position()
	pos[matrix.Vx] -= matrix.Float(ui.man.Host.Window.Width()) * 0.5
	pos[matrix.Vy] -= matrix.Float(ui.man.Host.Window.Height()) * 0.5
	return pos
}

func (ui *UI) containedCheck(cursor *hid.Cursor, entity *engine.Entity) {
	cp := ui.cursorPos(cursor)
	contained := entity.Transform.ContainsPoint2D(cp)
	if contained && ui.hasScissor() {
		contained = ui.shaderData.Scissor.ScreenAreaContains(cp.X(), cp.Y())
	}
	if !ui.hovering && contained {
		ui.hovering = true
		ui.requestEvent(EventTypeEnter)
		if cursor.HasDragData() {
			ui.requestEvent(EventTypeDropEnter)
		}
	} else if ui.hovering && !contained {
		ui.hovering = false
		ui.requestEvent(EventTypeExit)
		if cursor.HasDragData() {
			ui.requestEvent(EventTypeDropExit)
		}
	}
}

func (ui *UI) changed() {
	ui.ExecuteEvent(EventTypeChange)
}

func (ui *UI) layoutChanged(dirtyType DirtyType) {
	ui.SetDirty(dirtyType)
}

func (ui *UI) cleanIfNeeded() {
	if ui.anyChildDirty() {
		ui.Clean()
	}
}

func (ui *UI) anyChildDirty() bool {
	if ui.dirtyType != DirtyTypeNone {
		return true
	}
	for i := range ui.entity.Children {
		cui := FirstOnEntity(ui.entity.Children[i])
		if cui != nil && cui.anyChildDirty() {
			return true
		}
	}
	return false
}

func (ui *UI) updateFromManager(deltaTime float64) {
	if !ui.isActive() {
		return
	}
	switch ui.elmType {
	case ElementTypeInput:
		ui.ToInput().update(deltaTime)
	case ElementTypeLabel:
		ui.Update(deltaTime)
	case ElementTypePanel:
		ui.ToPanel().update(deltaTime)
	case ElementTypeButton:
		ui.ToPanel().update(deltaTime)
	case ElementTypeSelect:
		ui.ToSelect().update(deltaTime)
	case ElementTypeSlider:
		ui.ToSlider().update(deltaTime)
	case ElementTypeImage:
		ui.ToImage().update(deltaTime)
	case ElementTypeCheckbox:
		ui.ToPanel().update(deltaTime)
	}
}

func (ui *UI) Hide() {
	ui.entity.Deactivate()
}

func (ui *UI) Show() {
	ui.entity.Activate()
}
