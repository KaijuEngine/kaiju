/******************************************************************************/
/* checkbox.go                                                                */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package ui

import (
	"kaiju/rendering"
)

const (
	offIdleTexture  = "textures/checkbox-off-idle.png"
	offDownTexture  = "textures/checkbox-off-down.png"
	offHoverTexture = "textures/checkbox-off-hover.png"
	onIdleTexture   = "textures/checkbox-on-idle.png"
	onDownTexture   = "textures/checkbox-on-down.png"
	onHoverTexture  = "textures/checkbox-on-hover.png"
)

const (
	texOffIdle = iota
	texOffDown
	texOffHover
	texOnIdle
	texOnDown
	texOnHover
)

const (
	defaultCheckboxSize = 25
)

type checkboxData struct {
	panelData
	label     *Label
	textures  [6]*rendering.Texture
	isChecked bool
}

func (c *checkboxData) innerPanelData() *panelData { return &c.panelData }

type Checkbox Panel

func (u *UI) AsCheckbox() *Checkbox { return (*Checkbox)(u) }
func (cb *Checkbox) Base() *UI      { return (*UI)(cb) }

func (cb *Checkbox) CheckboxData() *checkboxData {
	return cb.Base().elmData.(*checkboxData)
}

func (p *Panel) ConvertToCheckbox() *Checkbox {
	ld := &checkboxData{
		panelData: *p.PanelData(),
	}
	tc := p.host.TextureCache()
	ld.textures[texOffIdle], _ = tc.Texture(
		offIdleTexture, rendering.TextureFilterLinear)
	ld.textures[texOffDown], _ = tc.Texture(
		offDownTexture, rendering.TextureFilterLinear)
	ld.textures[texOffHover], _ = tc.Texture(
		offHoverTexture, rendering.TextureFilterLinear)
	ld.textures[texOnIdle], _ = tc.Texture(
		onIdleTexture, rendering.TextureFilterLinear)
	ld.textures[texOnDown], _ = tc.Texture(
		onDownTexture, rendering.TextureFilterLinear)
	ld.textures[texOnHover], _ = tc.Texture(
		onHoverTexture, rendering.TextureFilterLinear)
	cb := (*Checkbox)(p)
	cb.elmData = ld
	cb.elmType = ElementTypeCheckbox
	p.Base().AddEvent(EventTypeEnter, cb.onHover)
	p.Base().AddEvent(EventTypeExit, cb.onBlur)
	p.Base().AddEvent(EventTypeDown, cb.onDown)
	p.Base().AddEvent(EventTypeUp, cb.onUp)
	p.Base().AddEvent(EventTypeClick, cb.onClick)
	cb.layout.Scale(defaultCheckboxSize, defaultCheckboxSize)
	p.ensureBGExists(ld.textures[texOffIdle])
	return cb
}

func (cb *Checkbox) onHover() {
	var target *rendering.Texture = nil
	data := cb.CheckboxData()
	if cb.isDown {
		if data.isChecked {
			target = data.textures[texOnDown]
		} else {
			target = data.textures[texOffDown]
		}
	} else {
		if data.isChecked {
			target = data.textures[texOnHover]
		} else {
			target = data.textures[texOffHover]
		}
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onBlur() {
	data := cb.CheckboxData()
	var target *rendering.Texture = nil
	if data.isChecked {
		target = data.textures[texOnIdle]
	} else {
		target = data.textures[texOffIdle]
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onDown() {
	data := cb.CheckboxData()
	var target *rendering.Texture = nil
	if data.isChecked {
		target = data.textures[texOnDown]
	} else {
		target = data.textures[texOffDown]
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onUp() {
	data := cb.CheckboxData()
	var target *rendering.Texture = nil
	if data.isChecked {
		target = data.textures[texOnHover]
	} else {
		target = data.textures[texOffHover]
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onClick() {
	data := cb.CheckboxData()
	cb.SetChecked(!data.isChecked)
}

func (cb *Checkbox) SetChecked(isChecked bool) {
	data := cb.CheckboxData()
	if data.isChecked == isChecked {
		return
	}
	data.isChecked = isChecked
	var target *rendering.Texture = nil
	if data.isChecked {
		if cb.hovering {
			target = data.textures[texOnHover]
		} else {
			target = data.textures[texOnIdle]
		}
	} else {
		if cb.hovering {
			target = data.textures[texOffHover]
		} else {
			target = data.textures[texOffIdle]
		}
	}
	(*Panel)(cb).SetBackground(target)
	(*UI)(cb).requestEvent(EventTypeChange)
}

func (cb Checkbox) IsChecked() bool {
	return cb.CheckboxData().isChecked
}
