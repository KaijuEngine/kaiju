/******************************************************************************/
/* progress_bar.go                                                            */
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

import (
	"kaiju/matrix"
	"kaiju/rendering"
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
	panel := p.man.Value().Add().ToPanel()
	fgPanel := p.man.Value().Add().ToPanel()
	panel.Init(bgTexture, ElementTypePanel)
	fgPanel.Init(fgTexture, ElementTypePanel)
	fgPanel.layout.Stylizer = StretchCenterStylizer{BasicStylizer{p.Base()}}
	panel.AddChild(fgPanel.Base())
	pd.fgPanel = fgPanel
}

func (b *ProgressBar) SetValue(value float32) {
	data := b.data()
	data.value = value
	w := b.entity.Transform.WorldScale().X()
	data.fgPanel.layout.ScaleWidth(w*data.value + 1)
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
