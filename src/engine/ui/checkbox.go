/******************************************************************************/
/* checkbox.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
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
	if cb.IsDisabled() {
		return
	}
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
	if cb.IsDisabled() {
		return
	}
	data := cb.CheckboxData()
	if data.isChecked {
		cb.setAtlas(checkOnIdleUvX, checkOnIdleUvY)
	} else {
		cb.setAtlas(checkOffIdleUvX, checkOffIdleUvY)
	}
}

func (cb *Checkbox) onDown() {
	if cb.IsDisabled() {
		return
	}
	data := cb.CheckboxData()
	if data.isChecked {
		cb.setAtlas(checkOnDownUvX, checkOnDownUvY)
	} else {
		cb.setAtlas(checkOffDownUvX, checkOffDownUvY)
	}
}

func (cb *Checkbox) onUp() {
	if cb.IsDisabled() {
		return
	}
	data := cb.CheckboxData()
	if data.isChecked {
		cb.setAtlas(checkOnHoverUvX, checkOnHoverUvY)
	} else {
		cb.setAtlas(checkOffHoverUvX, checkOffHoverUvY)
	}
}

func (cb *Checkbox) onClick() {
	if cb.IsDisabled() {
		return
	}
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

func (cb *Checkbox) IsDisabled() bool {
	return cb.Base().IsDisabled()
}

func (cb *Checkbox) SetDisabled(disabled bool) {
	cb.Base().SetDisabled(disabled)
	if disabled {
		data := cb.CheckboxData()
		if data.isChecked {
			cb.setAtlas(checkOnIdleUvX, checkOnIdleUvY)
		} else {
			cb.setAtlas(checkOffIdleUvX, checkOffIdleUvY)
		}
	}
}

func (cb *Checkbox) setAtlas(x, y matrix.Float) {
	cb.shaderData.setUVXY(x/cb.textureSize.X(), y, cb.textureSize.Y())
}
