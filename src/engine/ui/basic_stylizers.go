/******************************************************************************/
/* basic_stylizers.go                                                         */
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

package ui

type BasicStylizer struct {
	Parent *UI
}

type StretchWidthStylizer struct{ BasicStylizer }
type StretchHeightStylizer struct{ BasicStylizer }
type StretchCenterStylizer struct{ BasicStylizer }
type LeftStylizer struct{ BasicStylizer }
type RightStylizer struct{ BasicStylizer }

func (s StretchWidthStylizer) ProcessStyle(layout *Layout) []error {
	sw := s.Parent.layout.PixelSize().X()
	pPad := s.Parent.layout.Padding()
	sw -= pPad.X() + pPad.Z()
	p := layout.Padding()
	w := sw - p.X() - p.Z()
	layout.ScaleWidth(w)
	return []error{}
}

func (s StretchHeightStylizer) ProcessStyle(layout *Layout) []error {
	sh := s.Parent.layout.PixelSize().Y()
	pPad := s.Parent.layout.Padding()
	sh -= pPad.Y() + pPad.W()
	p := layout.Padding()
	h := sh - p.Y() - p.W()
	layout.ScaleHeight(h)
	return []error{}
}

func (s StretchCenterStylizer) ProcessStyle(layout *Layout) []error {
	errs := StretchWidthStylizer(s).ProcessStyle(layout)
	errs = append(errs, StretchHeightStylizer(s).ProcessStyle(layout)...)
	return errs
}

func (s RightStylizer) ProcessStyle(layout *Layout) []error {
	width := float32(layout.ui.Host().Window.Width())
	if s.Parent != nil {
		width = s.Parent.Layout().PixelSize().X()
	}
	selfWidth := layout.PixelSize().X()
	layout.SetInnerOffsetLeft(width - selfWidth)
	return nil
}

func (s LeftStylizer) ProcessStyle(layout *Layout) []error {
	height := float32(layout.ui.Host().Window.Height())
	if s.Parent != nil {
		height = s.Parent.Layout().PixelSize().Y()
	}
	selfHeight := layout.PixelSize().Y()
	layout.SetInnerOffsetTop(-height*0.5 + selfHeight*0.5)
	return nil
}
