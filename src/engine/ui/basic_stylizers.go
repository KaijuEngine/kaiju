/******************************************************************************/
/* basic_stylizers.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "weak"

type BasicStylizer struct {
	Parent weak.Pointer[UI]
}

type StretchWidthStylizer struct{ BasicStylizer }
type StretchHeightStylizer struct{ BasicStylizer }
type StretchCenterStylizer struct{ BasicStylizer }
type LeftStylizer struct{ BasicStylizer }
type RightStylizer struct{ BasicStylizer }

func (s StretchWidthStylizer) ProcessStyle(layout *Layout) []error {
	parent := s.Parent.Value()
	if parent == nil || !parent.IsValid() {
		return []error{}
	}
	sw := parent.layout.PixelSize().X()
	pPad := parent.layout.Padding()
	pBorder := parent.layout.Border()
	sw -= pPad.X() + pPad.Z() + pBorder.X() + pBorder.Z()
	p := layout.Padding()
	w := sw - p.X() - p.Z()
	layout.ScaleWidth(w)
	return []error{}
}

func (s StretchHeightStylizer) ProcessStyle(layout *Layout) []error {
	parent := s.Parent.Value()
	if parent == nil || !parent.IsValid() {
		return []error{}
	}
	sh := parent.layout.PixelSize().Y()
	pPad := parent.layout.Padding()
	pBorder := parent.layout.Border()
	sh -= pPad.Y() + pPad.W() + pBorder.Y() + pBorder.W()
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

func (l *Layout) stylizerControlsHeight() bool {
	switch l.Stylizer.(type) {
	case StretchHeightStylizer, StretchCenterStylizer:
		return true
	default:
		return false
	}
}

func (s RightStylizer) ProcessStyle(layout *Layout) []error {
	width := float32(layout.ui.Host().Window.Width())
	parent := s.Parent.Value()
	if parent != nil {
		width = parent.Layout().PixelSize().X()
	}
	selfWidth := layout.PixelSize().X()
	layout.SetInnerOffsetLeft(width - selfWidth)
	return nil
}

func (s LeftStylizer) ProcessStyle(layout *Layout) []error {
	height := float32(layout.ui.Host().Window.Height())
	parent := s.Parent.Value()
	if parent != nil {
		height = parent.Layout().PixelSize().Y()
	}
	selfHeight := layout.PixelSize().Y()
	layout.SetInnerOffsetTop(-height*0.5 + selfHeight*0.5)
	return nil
}
