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
