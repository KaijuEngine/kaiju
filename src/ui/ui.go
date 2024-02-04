package ui

import (
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/windowing"
)

type DirtyType = int

const (
	DirtyTypeNone = iota
	DirtyTypeLayout
	DirtyTypeResize
	DirtyTypeGenerated
	DirtyTypeReGenerated
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

type UI interface {
	Entity() *engine.Entity
	ExecuteEvent(evtType EventType) bool
	AddEvent(evtType EventType, evt func()) engine.EventId
	RemoveEvent(evtType EventType, evtId engine.EventId)
	Event(evtType EventType) *engine.Event
	Update(deltaTime float64)
	SetDirty(dirtyType DirtyType)
	Layout() *Layout
	ShaderData() *ShaderData
	Clean()
	SetGroup(group *Group)
	Host() *engine.Host
	GenerateScissor()
	hasScissor() bool
	selfScissor() matrix.Vec4
	dirty() DirtyType
	setScissor(scissor matrix.Vec4)
	layoutChanged(dirtyType DirtyType)
	cleanDirty()
	postLayoutUpdate()
}

type uiBase struct {
	host                *engine.Host
	entity              *engine.Entity
	events              [EventTypeEnd]engine.Event
	group               *Group
	dragStartPos        matrix.Vec3
	downPos             matrix.Vec2
	layout              Layout
	dirtyType           DirtyType
	shaderData          ShaderData
	textureSize         matrix.Vec2
	updateId            int
	hovering            bool
	cantMiss            bool
	isDown              bool
	drag                bool
	lastActive          bool
	disconnectedScissor bool
}

func (ui *uiBase) isActive() bool { return ui.updateId != 0 }

func (ui *uiBase) init(host *engine.Host, textureSize matrix.Vec2, anchor Anchor, self UI) {
	ui.host = host
	ui.entity = host.NewEntity()
	ui.shaderData.ShaderDataBase = rendering.NewShaderDataBase()
	ui.shaderData.Scissor = matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax}
	ui.entity.AddNamedData(EntityDataName, self)
	ui.textureSize = textureSize
	ui.layout.initialize(ui, anchor)
	if ui.updateId == 0 {
		ui.updateId = host.Updater.AddUpdate(ui.Update)
	}
	ui.entity.OnDestroy.Add(func() {
		host.Updater.RemoveUpdate(ui.updateId)
	})
}

func (ui *uiBase) Entity() *engine.Entity   { return ui.entity }
func (ui *uiBase) Layout() *Layout          { return &ui.layout }
func (ui *uiBase) hasScissor() bool         { return ui.shaderData.Scissor.X() > -matrix.FloatMax }
func (ui *uiBase) selfScissor() matrix.Vec4 { return ui.shaderData.Scissor }
func (ui *uiBase) Host() *engine.Host       { return ui.host }
func (ui *uiBase) dirty() DirtyType         { return ui.dirtyType }
func (ui *uiBase) ShaderData() *ShaderData  { return &ui.shaderData }
func (ui *uiBase) SetGroup(group *Group)    { ui.group = group }
func (ui *uiBase) postLayoutUpdate()        {}

func (ui *uiBase) ExecuteEvent(evtType EventType) bool {
	ui.events[evtType].Execute()
	return !ui.events[evtType].IsEmpty()
}

func (ui *uiBase) AddEvent(evtType EventType, evt func()) engine.EventId {
	return ui.events[evtType].Add(evt)
}

func (ui *uiBase) RemoveEvent(evtType EventType, evtId engine.EventId) {
	ui.events[evtType].Remove(evtId)
}

func (ui *uiBase) Event(evtType EventType) *engine.Event {
	return &ui.events[evtType]
}

func (ui *uiBase) cleanDirty() { ui.dirtyType = DirtyTypeNone }

func (ui *uiBase) SetDirty(dirtyType DirtyType) {
	if ui.dirtyType == DirtyTypeNone || ui.dirtyType >= DirtyTypeParent || dirtyType == DirtyTypeGenerated || dirtyType == DirtyTypeReGenerated {
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

func (ui *uiBase) rootUI() UI {
	root := ui.entity
	var rootUI UI = FirstOnEntity(root)
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

func (ui *uiBase) Clean() {
	root := ui.rootUI()
	tree := []UI{root}
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
	for !stabilized {
		stabilized = true
		for i := range tree {
			tree[i].cleanDirty()
			tree[i].Layout().update()
			tree[i].postLayoutUpdate()
			stabilized = stabilized && tree[i].dirty() == DirtyTypeNone
		}
	}
	for i := range tree {
		tree[i].GenerateScissor()
		if l, ok := tree[i].(*Label); ok {
			l.render()
		} else if p, ok := tree[i].(*Panel); ok {
			p.shaderData.setSize2d(ui, ui.textureSize.X(), ui.textureSize.Y())
		}
	}
}

func (ui *uiBase) GenerateScissor() {
	ui.disconnectedScissor = false
	target := &ui.entity.Transform
	if !ui.entity.IsRoot() {
		target = &ui.entity.Parent.Transform
	}
	pos := target.WorldPosition()
	size := target.WorldScale()
	bounds := matrix.Vec4{
		pos.X() - size.X()*0.5,
		pos.Y() - size.Y()*0.5,
		pos.X() + size.X()*0.5,
		pos.Y() + size.Y()*0.5,
	}
	ui.setScissor(bounds)
}

func (ui *uiBase) setScissor(scissor matrix.Vec4) {
	if ui.disconnectedScissor {
		return
	}
	if ui.shaderData.Scissor.Equals(scissor) {
		return
	}
	ui.shaderData.Scissor = scissor
	for i := 0; i < len(ui.entity.Children); i++ {
		cUI := FirstOnEntity(ui.entity.Children[i])
		if cUI != nil {
			cUI.setScissor(scissor)
		}
	}
	ui.shaderData.Scissor = scissor
	if ui.dirtyType == DirtyTypeNone {
		ui.SetDirty(DirtyTypeScissor)
	}
}

func (ui *uiBase) requestEvent(evtType EventType) {
	if ui.group != nil {
		ui.group.requestEvent(ui, evtType)
	} else {
		ui.ExecuteEvent(evtType)
	}
}

func (ui *uiBase) Update(deltaTime float64) {
	cursor := &ui.host.Window.Cursor
	if cursor.Moved() {
		pos := ui.cursorPos(cursor)
		ui.containedCheck(cursor, ui.entity)
		if ui.isDown && !ui.drag {
			w := ui.Host().Window.Width()
			h := ui.Host().Window.Height()
			wmm, hmm, _ := ui.host.Window.GetDPI()
			threshold := max(windowing.DPI2PX(w, wmm, 4), windowing.DPI2PX(h, hmm, 4))
			if ui.downPos.Distance(pos) > float32(threshold) {
				ui.dragStartPos = ui.entity.Transform.WorldPosition()
				ui.drag = true
			}
		}
	}
	if cursor.Pressed() {
		ui.containedCheck(cursor, ui.entity)
		if ui.hovering && !ui.isDown {
			ui.isDown = true
			ui.downPos = ui.cursorPos(cursor)
			ui.requestEvent(EventTypeDown)
			ui.cantMiss = true
		}
	}
	if cursor.Released() {
		if ui.hovering {
			ui.requestEvent(EventTypeUp)
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
				if ui.hovering && !dragged {
					ui.requestEvent(EventTypeClick)
				}
			} else if !ui.hovering && !ui.cantMiss {
				ui.requestEvent(EventTypeMiss)
			}
			ui.cantMiss = false
		}
	}
	if ui.host.Window.Mouse.Scrolled() && ui.hovering {
		ui.requestEvent(EventTypeScroll)
	}
	if ui.dirtyType != DirtyTypeNone {
		ui.Clean()
	}
	ui.lastActive = ui.entity.IsActive()
}

func (ui *uiBase) cursorPos(cursor *hid.Cursor) matrix.Vec2 {
	pos := cursor.Position()
	pos[matrix.Vx] -= matrix.Float(ui.host.Window.Width()) * 0.5
	pos[matrix.Vy] -= matrix.Float(ui.host.Window.Height()) * 0.5
	return pos
}

func (ui *uiBase) containedCheck(cursor *hid.Cursor, entity *engine.Entity) {
	cp := ui.cursorPos(cursor)
	contained := entity.Transform.ContainsPoint2D(cp)
	if contained && ui.hasScissor() {
		contained = ui.shaderData.Scissor.ScreenAreaContains(cp.X(), cp.Y())
	}
	if !ui.hovering && contained {
		ui.hovering = true
		ui.requestEvent(EventTypeEnter)
	} else if ui.hovering && !contained {
		ui.hovering = false
		ui.requestEvent(EventTypeExit)
	}
}

func (ui *uiBase) changed() {
	ui.ExecuteEvent(EventTypeChange)
}

func (ui *uiBase) DisconnectParentScissor() {
	if ui.hasScissor() {
		ui.setScissor(matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax})
		ui.GenerateScissor()
	}
	ui.disconnectedScissor = true
}

func (ui *uiBase) layoutChanged(dirtyType DirtyType) {
	ui.SetDirty(dirtyType)
	//if ui.Entity().Parent != nil {
	//	if pui := FirstOnEntity(ui.Entity().Parent); pui != nil {
	//		if pui.dirty() == DirtyTypeNone {
	//			pui.SetDirty(DirtyTypeParentLayout)
	//		} else {
	//			pui.SetDirty(DirtyTypeReGenerated)
	//		}
	//	}
	//}
}
