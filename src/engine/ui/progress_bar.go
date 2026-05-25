/******************************************************************************/
/* progress_bar.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"weak"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type ProgressBar Panel

type progressBarData struct {
	panelData
	fgPanel *Panel
	value   float32
}

func (u *UI) ToProgressBar() *ProgressBar { return (*ProgressBar)(u) }
func (p *ProgressBar) Base() *UI          { return (*UI)(p) }

func (p *ProgressBar) data() *progressBarData {
	return p.elmData.(*progressBarData)
}

func (p *ProgressBar) Init(fgTexture, bgTexture *rendering.Texture) {
	pd := &progressBarData{
		value: 0,
	}
	p.elmData = pd
	p.Base().ToPanel().Init(nil, ElementTypeProgressBar)
	man := p.man.Value()
	man.beginDirtyBatch()
	defer man.endDirtyBatch()
	panel := man.Add().ToPanel()
	fgPanel := man.Add().ToPanel()
	panel.Init(bgTexture, ElementTypePanel)
	fgPanel.Init(fgTexture, ElementTypePanel)
	fgPanel.layout.Stylizer = StretchCenterStylizer{BasicStylizer{weak.Make(p.Base())}}
	panel.AddChild(fgPanel.Base())
	pd.fgPanel = fgPanel
}

func (b *ProgressBar) SetValue(value float32) {
	data := b.data()
	data.value = value
	w := b.entity.Transform.WorldScale().X()
	data.fgPanel.layout.ScaleWidth(w*data.value + 1)
}

func (b *ProgressBar) Value() float32 {
	return b.data().value
}

func (b *ProgressBar) SetFGColor(fgColor matrix.Color) {
	b.data().fgPanel.SetColor(fgColor)
}

func (b *ProgressBar) SetBGColor(bgColor matrix.Color) {
	(*Panel)(b).SetColor(bgColor)
}
