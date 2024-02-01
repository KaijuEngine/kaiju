package ui

import (
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

type ProgressBar Panel

type progressBarData struct {
	fgPanel *Panel
	value   float32
}

func (p *ProgressBar) data() *progressBarData {
	return p.localData.(*progressBarData)
}

func NewProgressBar(host *engine.Host, fgTexture, bgTexture *rendering.Texture, anchor Anchor) *ProgressBar {
	panel := NewPanel(host, bgTexture, anchor)
	fgPanel := NewPanel(host, fgTexture, AnchorStretchCenter)
	panel.AddChild(fgPanel)
	panel.localData = &progressBarData{fgPanel: fgPanel, value: 0.0}
	return (*ProgressBar)(panel)
}

func (b *ProgressBar) SetValue(value float32) {
	data := b.data()
	data.value = value
	w := b.entity.Transform.WorldScale().X()
	data.fgPanel.layout.SetStretch(1, 1, w-(w*data.value)+1, 1)
}

func (b ProgressBar) Value() float32 {
	return b.data().value
}

func (b *ProgressBar) SetFGColor(fgColor matrix.Color) {
	b.data().fgPanel.SetColor(fgColor)
}

func (b *ProgressBar) SetBGColor(bgColor matrix.Color) {
	(*Panel)(b).SetColor(bgColor)
}
