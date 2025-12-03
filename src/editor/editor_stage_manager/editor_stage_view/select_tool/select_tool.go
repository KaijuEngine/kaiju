/******************************************************************************/
/* select_tool.go                                                             */
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
