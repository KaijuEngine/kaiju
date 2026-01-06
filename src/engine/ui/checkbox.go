/******************************************************************************/
/* checkbox.go                                                                */
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
	"kaiju/rendering"
)

const (
	inputAtlas       = "input_atlas.png"
	checkTexSize     = 32
	checkOffIdleUvX  = 128
	checkOffIdleUvY  = 64
	checkOffDownUvX  = 128
	checkOffDownUvY  = 0
	checkOffHoverUvX = 128
	checkOffHoverUvY = 32
	checkOnIdleUvX   = 160
	checkOnIdleUvY   = 32
	checkOnDownUvX   = 128
	checkOnDownUvY   = 96
	checkOnHoverUvX  = 160
	checkOnHoverUvY  = 0
)

const (
	texOffIdle = iota
	texOffDown
	texOffHover
	texOnIdle
	texOnDown
	texOnHover
)

type checkboxData struct {
	panelData
	label     *Label
	isChecked bool
}

func (c *checkboxData) innerPanelData() *panelData { return &c.panelData }

type Checkbox Panel

func (u *UI) ToCheckbox() *Checkbox { return (*Checkbox)(u) }
func (cb *Checkbox) Base() *UI      { return (*UI)(cb) }

func (cb *Checkbox) CheckboxData() *checkboxData {
	return cb.Base().elmData.(*checkboxData)
}

func (cb *Checkbox) Init() {
	ld := &checkboxData{}
	cb.elmData = ld
	base := cb.Base()
	p := base.ToPanel()
	host := p.man.Value().Host
	tc := host.TextureCache()
	tex, _ := tc.Texture(inputAtlas, rendering.TextureFilterLinear)
	p.Init(tex, ElementTypeCheckbox)
	p.shaderData.Size2D.SetZ(checkTexSize)
	p.shaderData.Size2D.SetW(checkTexSize)
	cb.shaderData.setUVSize(checkTexSize/cb.textureSize.X(), checkTexSize/cb.textureSize.Y())
	cb.setAtlas(checkOffIdleUvX, checkOffIdleUvY)
	base.AddEvent(EventTypeEnter, cb.onHover)
	base.AddEvent(EventTypeExit, cb.onBlur)
	base.AddEvent(EventTypeDown, cb.onDown)
	base.AddEvent(EventTypeUp, cb.onUp)
	base.AddEvent(EventTypeClick, cb.onClick)
}

func (cb *Checkbox) onHover() {
	data := cb.CheckboxData()
	if cb.flags.isDown() {
		if data.isChecked {
			cb.setAtlas(checkOnDownUvX, checkOnDownUvY)
		} else {
			cb.setAtlas(checkOffDownUvX, checkOffDownUvY)
		}
	} else {
		if data.isChecked {
			cb.setAtlas(checkOnHoverUvX, checkOnHoverUvY)
		} else {
			cb.setAtlas(checkOffHoverUvX, checkOffHoverUvY)
		}
	}
}

func (cb *Checkbox) onBlur() {
	data := cb.CheckboxData()
	if data.isChecked {
		cb.setAtlas(checkOnIdleUvX, checkOnIdleUvY)
	} else {
		cb.setAtlas(checkOffIdleUvX, checkOffIdleUvY)
	}
}

func (cb *Checkbox) onDown() {
	data := cb.CheckboxData()
	if data.isChecked {
		cb.setAtlas(checkOnDownUvX, checkOnDownUvY)
	} else {
		cb.setAtlas(checkOffDownUvX, checkOffDownUvY)
	}
}

func (cb *Checkbox) onUp() {
	data := cb.CheckboxData()
	if data.isChecked {
		cb.setAtlas(checkOnHoverUvX, checkOnHoverUvY)
	} else {
		cb.setAtlas(checkOffHoverUvX, checkOffHoverUvY)
	}
}

func (cb *Checkbox) onClick() {
	data := cb.CheckboxData()
	cb.SetChecked(!data.isChecked)
}

func (cb *Checkbox) SetCheckedWithoutEvent(isChecked bool) {
	data := cb.CheckboxData()
	if data.isChecked == isChecked {
		return
	}
	data.isChecked = isChecked
	if data.isChecked {
		if cb.flags.hovering() {
			cb.setAtlas(checkOnHoverUvX, checkOnHoverUvY)
		} else {
			cb.setAtlas(checkOnIdleUvX, checkOnIdleUvY)
		}
	} else {
		if cb.flags.hovering() {
			cb.setAtlas(checkOffHoverUvX, checkOffHoverUvY)
		} else {
			cb.setAtlas(checkOffIdleUvX, checkOffIdleUvY)
		}
	}
}

func (cb *Checkbox) SetChecked(isChecked bool) {
	cb.SetCheckedWithoutEvent(isChecked)
	(*UI)(cb).requestEvent(EventTypeChange)
}

func (cb *Checkbox) IsChecked() bool {
	return cb.CheckboxData().isChecked
}

func (cb *Checkbox) setAtlas(x, y float32) {
	cb.shaderData.setUVXY(x/cb.textureSize.X(), y, cb.textureSize.Y())
}
