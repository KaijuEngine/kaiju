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
	"kaiju/rendering"
	"kaiju/systems/events"
	"kaiju/windowing"
)

type DirtyType = int
type ElementType = int

const (
	DirtyTypeNone DirtyType = iota
	DirtyTypeLayout
	DirtyTypeResize
	DirtyTypeGenerated
	DirtyTypeColorChange
	DirtyTypeParent
	DirtyTypeParentLayout
	DirtyTypeParentResize
	DirtyTypeParentGenerated
	DirtyTypeParentReGenerated
	DirtyTypeParentColorChange
)

const (
	ElementTypeInput ElementType = iota
	ElementTypeLabel
	ElementTypePanel
	ElementTypeButton
	ElementTypeSelect
	ElementTypeSlider
	ElementTypeSprite
	ElementTypeCheckbox
)

type UI struct {
	man                      *Manager
	data                     interface{}
	overrideShaderDefinition string
	entity                   engine.Entity
	group                    *Group
	poolId                   int
	id                       uint8
	events                   [EventTypeEnd]events.Event
	resizeId                 events.Id
	postLayoutUpdate         func()
	scissor                  matrix.Vec4
	dragStartPos             matrix.Vec3
	downPos                  matrix.Vec2
	layout                   Layout
	dirtyType                DirtyType
	lastClick                float64
	destroyEvtId             events.Id
	elmType                  ElementType
	hovering                 bool
	cantMiss                 bool
	isDown                   bool
	isRightDown              bool
	drag                     bool
	lastActive               bool
	useBlending              bool
	dontClean                bool
}

func (ui *UI) ButtonData() *ButtonData     { return ui.data.(*ButtonData) }
func (ui *UI) CheckboxData() *CheckboxData { return ui.data.(*CheckboxData) }
func (ui *UI) InputData() *InputData       { return ui.data.(*InputData) }
func (ui *UI) LabelData() *LabelData       { return ui.data.(*LabelData) }
func (ui *UI) SliderData() *SliderData     { return ui.data.(*SliderData) }
func (ui *UI) SpriteData() *SpriteData     { return ui.data.(*SpriteData) }
func (ui *UI) SelectData() *SelectData     { return ui.data.(*SelectData) }
func (ui *UI) PanelData() *PanelData       { return ui.data.(*PanelData) }

func (ui *UI) CleanDirty() { ui.dirtyType = DirtyTypeNone }

func (ui *UI) hasScissor() bool { return ui.scissor.X() > -matrix.FloatMax }

func (ui *UI) Event(evtType EventType) *events.Event { return &ui.events[evtType] }

func (ui *UI) SetGroup(group *Group) {
	ui.group = group
	for i := range ui.entity.Children {
		FirstOnEntity(ui.entity.Children[i]).group = group
	}
}

func (ui *UI) Render() { ui.ExecuteEvent(EventTypeRender) }

func (ui *UI) AddEvent(evtType EventType, call func()) events.Id {
	return ui.events[evtType].Add(call)
}

func (ui *UI) RemoveEvent(evtType EventType, id events.Id) {
	ui.events[evtType].Remove(id)
}

func (ui *UI) Changed() { ui.ExecuteEvent(EventTypeChange) }

func (ui *UI) layoutChanged(dirtyType DirtyType) {
	ui.SetDirty(dirtyType)
}

func (ui *UI) FindByName(name string) *UI {
	target := ui.entity.FindByName(name)
	if target != nil {
		return FirstOnEntity(target)
	}
	return nil
}

func (ui *UI) ChildByIndex(index int) *UI {
	return FirstOnEntity(ui.entity.Children[index])
}

func (ui *UI) DestroyChildren() {
	for i := range ui.entity.Children {
		ui.entity.Children[i].Destroy()
	}
}

func (ui *UI) onActivate() {
	if ui.elmType == ElementTypeLabel {
		ld := ui.LabelData()
		for i := range ld.runeDrawings {
			ld.runeDrawings[i].ShaderData.Activate()
		}
	} else {
		pd := ui.PanelData()
		for i := range pd.drawings {
			pd.drawings[i].ShaderData.Activate()
		}
		ui.SetDirty(DirtyTypeLayout)
	}
}

func (ui *UI) onDeactivate() {
	if ui.elmType == ElementTypeLabel {
		ld := ui.LabelData()
		for i := range ld.runeDrawings {
			ld.runeDrawings[i].ShaderData.Deactivate()
		}
	} else {
		pd := ui.PanelData()
		for i := range pd.drawings {
			pd.drawings[i].ShaderData.Deactivate()
		}
	}
}

func (ui *UI) Init(man *Manager, anchor Anchor, elmType ElementType, construct interface{}) {
	ui.elmType = elmType
	ui.man = man
	ui.scissor = matrix.NewVec4(-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax)
	if ui.postLayoutUpdate == nil {
		ui.postLayoutUpdate = func() {}
	}
	ui.entity.Init(false)
	ui.man.host.AddEntity(&ui.entity)
	ui.entity.Transform.absolutePositions = true
	ui.entity.AddNamedData(EntityDataName, ui)
	ui.layout.initialize(ui, anchor)
	if ui.resizeId == 0 {
		ui.resizeId = ui.man.host.Window.OnResize.Add(ui.windowResize)
	}
	if ui.destroyEvtId == 0 {
		ui.destroyEvtId = ui.entity.OnDestroy.Add(ui.onEntityDestroy)
	}
	ui.entity.OnActivate.Add(ui.onActivate)
	ui.entity.OnDeactivate.Add(ui.onDeactivate)
	initTypeUI(elmType, ui, construct)
}

func (ui *UI) update(deltaTime float64) {
	if ui.entity.IsActive() {
		ui.eventUpdates()
	}
	ui.lastActive = ui.entity.IsActive()
}

func (ui *UI) requestEvent(evtType EventType) {
	if ui.group != nil {
		ui.group.requestEvent(ui, evtType)
	} else {
		ui.ExecuteEvent(evtType)
	}
}

func (ui *UI) containedCheck(cursor *hid.Cursor, entity *engine.Entity) {
	cp := ui.cursorPos(cursor)
	contained := entity.Transform.ContainsPoint2D(cp)
	if contained && ui.hasScissor() {
		contained = ui.scissor.ScreenAreaContains(cp.X(), cp.Y())
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

func (ui *UI) eventUpdates() {
	cursor := &ui.man.host.Window.Cursor
	if cursor.Moved() {
		pos := ui.cursorPos(cursor)
		ui.containedCheck(cursor, &ui.entity)
		if ui.isDown && !ui.drag {
			w := ui.man.host.Window.Width()
			h := ui.man.host.Window.Height()
			wmm, hmm, _ := ui.man.host.Window.SizeMM()
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
			ui.group.setDownElement(ui)
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
				if ui.hovering && !dragged && ui.group.downIsMe(ui) {
					rt := ui.man.host.Runtime()
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
	mouse := &ui.man.host.Window.Mouse
	if mouse.Released(hid.MouseButtonRight) {
		if ui.isRightDown && ui.hovering {
			ui.requestEvent(EventTypeRightClick)
		}
	}
	if ui.man.host.Window.Mouse.Scrolled() && ui.hovering {
		ui.requestEvent(EventTypeScroll)
	}
}

func (ui *UI) onEntityDestroy() {
	ui.man.host.Window.OnResize.Remove(ui.resizeId)
	if ui.group != nil {
		ui.group.removeUI(ui)
	}
	ui.man.Remove(ui)
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

func (ui *UI) RootCleanIfNeeded() {
	root := ui.Root()
	if root.anyChildDirty() {
		root.Clean()
	}
}

func (ui *UI) updateFromManager(deltaTime float64) {
	updateTypeUI(ui.elmType, ui, deltaTime)
}

func (ui *UI) SetDirty(dirtyType DirtyType) {
	if ui.dirtyType == DirtyTypeNone ||
		ui.dirtyType >= DirtyTypeParent ||
		dirtyType == DirtyTypeGenerated {
		ui.dirtyType = dirtyType
		for i := range ui.entity.Children {
			kid := ui.entity.Children[i]
			all := AllOnEntity(kid)
			for j := range len(all) {
				cui := all[j]
				if cui.dirtyType == DirtyTypeNone ||
					cui.dirtyType > DirtyTypeParent {
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

func (ui *UI) Root() *UI {
	if ui.entity.IsRoot() {
		return ui
	}
	root := &ui.entity
	rootUI := FirstOnEntity(root)
	for root.Parent != nil {
		pui := FirstOnEntity(root.Parent)
		if pui != nil {
			root = root.Parent
			rootUI = pui
		} else {
			break
		}
	}
	return rootUI
}

func createTree(tree *[]*UI, target *engine.Entity) {
	for i := range target.Children {
		child := target.Children[i]
		cui := FirstOnEntity(child)
		if cui != nil && !cui.dontClean {
			*tree = append(*tree, cui)
			createTree(tree, child)
		}
	}
}

func (ui *UI) Clean() {
	if ui.dontClean {
		return
	}
	root := ui.Root()
	tree := []*UI{root}
	createTree(&tree, &root.entity)
	stabilized := false
	for stabilized {
		stabilized = true
		for i := range len(tree) {
			tree[i].CleanDirty()
			tree[i].layout.update()
			tree[i].postLayoutUpdate()
			stabilized = stabilized && tree[i].dirtyType == DirtyTypeNone
		}
	}
	for i := range len(tree) {
		tree[i].GenerateScissor()
		tree[i].Render()
	}
}

func (ui *UI) GenerateScissor() {
	target := &ui.entity.Transform
	pos := target.WorldPosition()
	size := target.WorldScale()
	bounds := matrix.NewVec4(
		pos.X()-size.X()*0.5,
		pos.Y()-size.Y()*0.5,
		pos.X()+size.X()*0.5,
		pos.Y()+size.Y()*0.5,
	)
	if !ui.entity.IsRoot() {
		p := FirstOnEntity(ui.entity.Parent)
		for p != nil && !p.entity.IsRoot() &&
			p.elmType == ElementTypePanel && p.PanelData().overflow == OverflowVisible {
			pp := FirstOnEntity(p.entity.Parent)
			if pp != nil && pp.elmType == ElementTypePanel {
				p = pp
			}
		}
		if p != nil {
			ps := p.scissor
			bounds.SetX(max(bounds.X(), ps.X()))
			bounds.SetY(max(bounds.Y(), ps.Y()))
			bounds.SetZ(min(bounds.Z(), ps.Z()))
			bounds.SetW(min(bounds.W(), ps.W()))
		}
	}
	ui.SetScissor(bounds)
}

func (ui *UI) SetScissor(scissor matrix.Vec4) {
	if ui.scissor.Equals(scissor) {
		return
	}
	for i := range ui.entity.Children {
		cui := FirstOnEntity(ui.entity.Children[i])
		if cui != nil {
			cui.SetScissor(scissor)
		}
	}
	ui.scissor = scissor
	if ui.elmType == ElementTypeLabel {
		l := ui.LabelData()
		for i := range l.runeDrawings {
			sd := l.runeDrawings[i].ShaderData.(*rendering.TextShaderData)
			sd.Scissor = ui.scissor
		}
	} else {
		pd := ui.PanelData()
		for i := range pd.drawings {
			pd.drawings[i].ShaderData.(*ShaderData).Scissor = ui.scissor
		}
	}
}

func (ui *UI) Show() { ui.entity.Activate() }
func (ui *UI) Hide() { ui.entity.Deactivate() }

func (ui *UI) EnforceBlending() {
	if ui.useBlending {
		return
	}
	ui.useBlending = true
	if ui.elmType == ElementTypeLabel {
		ui.SetDirty(DirtyTypeGenerated)
	} else {
		// TODO:  Consolidate these lines into a single function call?
		p := ui.PanelData()
		recreateDrawings(ui)
		for i := range p.drawings {
			ui.man.host.Drawings.AddDrawing(&p.drawings[i])
		}
	}
}

func (ui *UI) ExecuteEvent(evtType EventType) bool {
	if ui.events[evtType].IsEmpty() {
		return false
	}
	ui.events[evtType].ExecuteWithSender(ui)
	return true
}

func (ui *UI) cursorPos(cursor *hid.Cursor) matrix.Vec2 {
	pos := cursor.Position()
	pos[matrix.Vx] -= matrix.Float(ui.man.host.Window.Width()) * 0.5
	pos[matrix.Vy] -= matrix.Float(ui.man.host.Window.Height()) * 0.5
	return pos
}

func (ui *UI) windowResize() {
	ui.SetDirty(DirtyTypeResize)
}
