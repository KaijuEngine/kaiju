package select_tool

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/matrix"
)

type ResultHandler interface {
	TryBoxSelect(screenBox matrix.Vec4)
}

type SelectTool struct {
	panel   *ui.Panel
	uiMan   ui.Manager
	start   matrix.Vec2
	handler ResultHandler
}

func (s *SelectTool) Init(host *engine.Host, handler ResultHandler) {
	s.handler = handler
	s.uiMan.Init(host)
	s.panel = s.uiMan.Add().ToPanel()
	s.panel.Init(nil, ui.ElementTypePanel)
	s.panel.SetColor(matrix.NewColor(1, 1, 1, 0.5))
	s.panel.Base().Hide()
}

func (s *SelectTool) Update() {
	c := &s.uiMan.Host.Window.Cursor
	if c.Pressed() {
		s.start = c.ScreenPosition()
		u := s.panel.Base()
		u.Show()
		l := u.Layout()
		l.SetOffset(s.start.X(), s.start.Y())
		l.Scale(0.0001, 0.0001)
		u.Clean()
	} else if c.Released() {
		if s.start.Distance(c.ScreenPosition()) > 5 {
			s.handler.TryBoxSelect(s.box())
		}
		s.panel.Base().Hide()
	} else if c.Held() && s.panel.Base().Entity().IsActive() {
		box := s.box()
		x, y := box.Left(), box.Top()
		w, h := box.Right()-x, box.Bottom()-y
		u := s.panel.Base()
		l := u.Layout()
		l.SetOffset(x, y)
		l.Scale(w, h)
		u.Clean()
	}
}

func (s *SelectTool) box() matrix.Vec4 {
	if !s.panel.Base().Entity().IsActive() {
		return matrix.NewVec4(-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax)
	}
	p := s.uiMan.Host.Window.Cursor.ScreenPosition()
	return matrix.Vec4{
		min(s.start.X(), p.X()),
		min(s.start.Y(), p.Y()),
		max(s.start.X(), p.X()),
		max(s.start.Y(), p.Y()),
	}
}
